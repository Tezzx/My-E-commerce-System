package service

import (
	"order-payment-system/internal/errs"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/logger"

	"go.uber.org/zap"
)

type GoodsService struct {
	goodsRepo *repository.GoodsRepo
}

func NewGoodsService(goodsRepo *repository.GoodsRepo) *GoodsService {
	return &GoodsService{
		goodsRepo: goodsRepo,
	}
}

// CreateGoods 创建商品
func (s *GoodsService) CreateGoods(goods *model.Goods) error {
	err := s.goodsRepo.CreateGoods(goods)
	if err != nil {
		logger.Log.Error("创建商品失败",
			zap.Uint("goods_id", goods.ID),
			zap.String("goods_name", goods.Goodsname),
			zap.Error(err))
		return errs.CreateGoodsError
	}

	return nil
}

// GetGoodsInfoByID 根据 ID 获取商品信息
func (s *GoodsService) GetGoodsInfoByID(goodsID uint) (price, goodsNum uint, goodsName string, err error) {
	price, goodsNum, goodsName, err = s.goodsRepo.GetGoodsByID(goodsID)
	if err != nil {
		logger.Log.Warn("商品未找到",
			zap.Uint("goods_id", goodsID))
		return 0, 0, "", errs.GoodsNotFound
	}
	return price, goodsNum, goodsName, nil
}

// GetGoodsList 获取商品列表
func (g *GoodsService) GetGoodsList() ([]model.Goods, error) {
	goodsList, err := g.goodsRepo.GetGoodsList()
	if err != nil {
		logger.Log.Error("获取商品列表失败", zap.Error(err))
		return nil, errs.GetGoodsError
	}
	return goodsList, nil
}
