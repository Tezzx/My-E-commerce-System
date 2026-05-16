package repository

import (
	"context"
	"errors"
	"order-payment-system/internal/model"

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
	var goods model.Goods
	err = g.db.Where("id = ?", goodsID).Select("price, goodsnum,goodsname").First(&goods).Error
	if err != nil {
		return 0, 0, "", err
	}
	return goods.Price, goods.Goodsnum, goods.Goodsname, nil
}

// 查询所有商品
func (g *GoodsRepo) GetGoodsList() ([]model.Goods, error) {
	var goods []model.Goods
	err := g.db.Find(&goods).Error
	return goods, err
}
