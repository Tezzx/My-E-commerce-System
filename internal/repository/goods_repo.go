package repository

import (
	"context"
	"encoding/json"
	"errors"
	"order-payment-system/internal/model"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GoodsRepo struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewGoodsRepo(db *gorm.DB, rdb *redis.Client) *GoodsRepo {
	return &GoodsRepo{
		db:  db,
		rdb: rdb,
	}
}

//-----------------业务--------------------------

// 扣减商品库存(redis)
func (g *GoodsRepo) DeductStock(goodsID uint, quantity int64) (bool, error) {
	ctx := context.Background()

	luaScript := `
local current = redis.call('HGET', KEYS[1], ARGV[1])
if not current then
    return -1  -- 商品不存在
end
local stock = tonumber(current)
local needed = tonumber(ARGV[2])
if stock >= needed then
    redis.call('HINCRBY', KEYS[1], ARGV[1], -needed)
    return stock - needed  -- 返回剩余库存
else
    return -2  -- 库存不足
end
`

	const stockKey = "goods"

	result, err := g.rdb.Eval(ctx, luaScript, []string{stockKey}, goodsID, quantity).Result()
	if err != nil {
		return false, err
	}

	// 解析 Lua 脚本返回值
	switch v := result.(type) {
	case int64:
		if v == -1 {
			return false, errors.New("商品不存在")
		} else if v == -2 {
			return false, nil // 库存不足
		}
		return true, nil
	case string:
		if v == "-1" {
			return false, errors.New("商品不存在")
		} else if v == "-2" {
			return false, nil
		}
		return true, nil
	default:
		return false, errors.New("unexpected return type from redis lua script")
	}
}

// 扣除库存(mysql)
func (g *GoodsRepo) DeductStockSQL(goodsID uint, quantity int64) error {
	result := g.db.Model(&model.Goods{}).
		Where("id = ? AND goodsnum >= ?", goodsID, quantity).
		UpdateColumn("goodsnum", gorm.Expr("goodsnum - ?", quantity))

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 回增库存(redis)
func (g *GoodsRepo) IncrementStock(goodsID uint, quantity int64) error {
	ctx := context.Background()
	luaScript := `
local exists = redis.call('HEXISTS', KEYS[1], ARGV[1])
if exists == 1 then
    return redis.call('HINCRBY', KEYS[1], ARGV[1], ARGV[2])
else
    return redis.error_reply('商品不存在')
end
`
	_, err := g.rdb.Eval(ctx, luaScript, []string{"goods"}, goodsID, quantity).Result()
	return err
}

// -------------------辅助功能-------------------
// CreateGoods 创建商品，商品名重复则覆盖旧数据
func (g *GoodsRepo) CreateGoods(goods *model.Goods) error {
	//根据商品名称查询是否已存在
	var existGoods model.Goods
	err := g.db.Where("goodsname = ?", goods.Goodsname).First(&existGoods).Error

	if err == gorm.ErrRecordNotFound {
		return g.db.Create(goods).Error
	}
	if err != nil {
		return err
	}

	return g.db.Model(&existGoods).Omit("id").Updates(goods).Error
}

// 根据商品ID获取 商品价格 和 库存数量和商品名
func (g *GoodsRepo) GetGoodsByID(goodsID uint) (price, goodsNum uint, goodsName string, err error) {
	ctx := context.Background()
	key := "goods:info:" + strconv.FormatUint(uint64(goodsID), 10)

	// 1. 尝试从 Redis 读取
	val, err := g.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// 2. 缓存未命中，回源到 MySQL
		var goods model.Goods
		dbErr := g.db.Select("id, price, goodsnum, goodsname").
			Where("id = ?", goodsID).
			First(&goods).Error
		if dbErr != nil {
			return 0, 0, "", dbErr
		}

		// 3. 回填缓存
		data, _ := json.Marshal(goods)
		// 设置 TTL 避免永久脏数据
		g.rdb.Set(ctx, key, data, 1*time.Hour)

		return goods.Price, goods.Goodsnum, goods.Goodsname, nil
	} else if err != nil {
		return 0, 0, "", err
	}

	// 4. Redis 命中，解析 JSON
	var goods model.Goods
	if err := json.Unmarshal([]byte(val), &goods); err != nil {
		// JSON 损坏，可选择删除 key 并回源
		g.rdb.Del(ctx, key)
		return g.GetGoodsByID(goodsID) // 递归回源（或改用 DB 查询）
	}

	return goods.Price, goods.Goodsnum, goods.Goodsname, nil
}

// 查询所有商品
func (g *GoodsRepo) GetGoodsList() ([]model.Goods, error) {
	var goods []model.Goods
	err := g.db.Find(&goods).Error
	return goods, err
}
