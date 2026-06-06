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

// GetGoodsListWithPage 分页获取商品列表
func (s *GoodsService) GetGoodsListWithPage(page, size int, categoryID *uint, status *int) ([]model.Goods, int64, error) {
	goods, total, err := s.goodsRepo.GetGoodsListWithPage(page, size, categoryID, status)
	if err != nil {
		logger.Log.Error("分页获取商品列表失败", zap.Error(err))
		return nil, 0, errs.GetGoodsError
	}
	return goods, total, nil
}

// UpdateGoods 更新商品信息
func (s *GoodsService) UpdateGoods(goodsID uint, updates map[string]interface{}) error {
	err := s.goodsRepo.UpdateGoods(goodsID, updates)
	if err != nil {
		logger.Log.Error("更新商品失败", zap.Uint("goods_id", goodsID), zap.Error(err))
		return err
	}
	return nil
}

// DeleteGoods 删除商品
func (s *GoodsService) DeleteGoods(goodsID uint) error {
	err := s.goodsRepo.DeleteGoods(goodsID)
	if err != nil {
		logger.Log.Error("删除商品失败", zap.Uint("goods_id", goodsID), zap.Error(err))
		return err
	}
	return nil
}
