package service

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
)

type GoodsService struct {
	goodsRepo *repository.GoodsRepo
}

func NewGoodsService(goodsRepo *repository.GoodsRepo) *GoodsService {
	return &GoodsService{
		goodsRepo: goodsRepo,
	}
}

// 创建商品
func (s *GoodsService) CreateGoods(goods *model.Goods) error {
	return s.goodsRepo.CreateGoods(goods)
}

func (s *GoodsService) GetGoodsInfoByID(goodsID uint) (price, goodsNum uint, goodsName string, err error) {
	return s.goodsRepo.GetGoodsByID(goodsID)
}

func (g *GoodsService) GetGoodsList() ([]model.Goods, error) {
	return g.goodsRepo.GetGoodsList()
}
