package repository

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"

	"golang.org/x/sync/singleflight"

	"order-payment-system/internal/model"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GoodsRepo struct {
	db  *gorm.DB
	rdb *redis.Client
	sf  *singleflight.Group
}

func NewGoodsRepo(db *gorm.DB, rdb *redis.Client) *GoodsRepo {
	return &GoodsRepo{
		db:  db,
		rdb: rdb,
		sf:  &singleflight.Group{},
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
	if result.RowsAffected == 0 {
		return errors.New("库存不足或商品不存在")
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

	//直接查 Redis
	val, err := g.rdb.Get(ctx, key).Result()
	if err == nil {
		// 防缓存穿透
		if val == "EMPTY" {
			return 0, 0, "", errors.New("商品不存在")
		}

		// 正常解析 JSON
		var goods model.Goods
		if err := json.Unmarshal([]byte(val), &goods); err != nil {
			g.rdb.Del(ctx, key)
			return 0, 0, "", errors.New("缓存数据异常")
		}
		return goods.Price, goods.Goodsnum, goods.Goodsname, nil
	} else if err != redis.Nil {
		return 0, 0, "", err
	}

	//防击穿 (Singleflight)

	type goodsResult struct {
		Price     uint
		GoodsNum  uint
		GoodsName string
	}

	result, err, _ := g.sf.Do(key, func() (interface{}, error) {

		var goods model.Goods
		dbErr := g.db.Select("id, price, goodsnum, goodsname").
			Where("id = ?", goodsID).
			First(&goods).Error

		// 防缓存穿透
		if dbErr == gorm.ErrRecordNotFound {
			g.rdb.Set(ctx, key, "EMPTY", 5*time.Minute)
			return nil, errors.New("商品不存在")
		} else if dbErr != nil {
			return nil, dbErr
		}

		data, _ := json.Marshal(goods)

		// 防缓存雪崩
		ttl := time.Hour + time.Duration(rand.Intn(10))*time.Minute
		g.rdb.Set(ctx, key, data, ttl)

		return &goodsResult{
			Price:     goods.Price,
			GoodsNum:  goods.Goodsnum,
			GoodsName: goods.Goodsname,
		}, nil
	})

	if err != nil {
		return 0, 0, "", err
	}

	res := result.(*goodsResult)
	return res.Price, res.GoodsNum, res.GoodsName, nil
}

// 查询所有商品
func (g *GoodsRepo) GetGoodsList() ([]model.Goods, error) {
	var goods []model.Goods
	err := g.db.Find(&goods).Error
	return goods, err
}

// GetGoodsListWithPage 分页查询商品列表
func (g *GoodsRepo) GetGoodsListWithPage(page, size int, categoryID *uint, status *int) ([]model.Goods, int64, error) {
	var goods []model.Goods
	var total int64

	query := g.db.Model(&model.Goods{})
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := query.Order("id desc").Offset(offset).Limit(size).Find(&goods).Error; err != nil {
		return nil, 0, err
	}

	return goods, total, nil
}

// UpdateGoods 更新商品信息
func (g *GoodsRepo) UpdateGoods(goodsID uint, updates map[string]interface{}) error {
	return g.db.Model(&model.Goods{}).Where("id = ?", goodsID).Updates(updates).Error
}

// DeleteGoods 软删除商品（将状态置为下架并删除）
func (g *GoodsRepo) DeleteGoods(goodsID uint) error {
	return g.db.Delete(&model.Goods{}, goodsID).Error
}
