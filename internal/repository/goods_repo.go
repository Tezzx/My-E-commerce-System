package repository

import (
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

// 扣减商品库存
func (g *GoodsRepo) ReduceStock(goodsID uint, num uint) error {
	result := g.db.Model(&model.Goods{}).
		Where("id = ? AND goodsnum >= ?", goodsID, num).
		UpdateColumn("goodsnum", gorm.Expr("goodsnum - ?", num))

	if result.RowsAffected == 0 {
		return errors.New("库存不足")
	}
	return result.Error
}

// 查询所有商品
func (g *GoodsRepo) GetGoodsList() ([]model.Goods, error) {
	var goods []model.Goods
	err := g.db.Find(&goods).Error
	return goods, err
}

// 增加库存
func (g *GoodsRepo) RestoreStock(tx *gorm.DB, goodsID uint, num uint) error {

	result := tx.Model(&model.Goods{}).
		Where("id = ?", goodsID).
		UpdateColumn("goodsnum", gorm.Expr("goodsnum + ?", num))

	return result.Error
}
