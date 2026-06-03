package service

import (
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/types"
)

type CartService struct {
	cartRepo  *repository.CartRepo
	goodsRepo *repository.GoodsRepo
}

func NewCartService(cartRepo *repository.CartRepo, goodsRepo *repository.GoodsRepo) *CartService {
	return &CartService{
		cartRepo:  cartRepo,
		goodsRepo: goodsRepo,
	}
}

func (s *CartService) AddToCart(userID uint, goodsID uint, quantity uint) error {
	// 1. 检查商品是否存在
	_, _, _, err := s.goodsRepo.GetGoodsByID(goodsID)
	if err != nil {
		return errors.New("商品不存在")
	}

	// 2. 检查购物车是否已有该商品
	item, err := s.cartRepo.GetItem(userID, goodsID)
	if err != nil {
		return err
	}

	if item == nil {
		// 添加新记录
		newItem := &model.CartItem{
			UserID:   userID,
			GoodsID:  goodsID,
			Quantity: quantity,
			Selected: true,
		}
		return s.cartRepo.AddItem(newItem)
	}

	// 累计数量
	item.Quantity += quantity
	return s.cartRepo.UpdateItem(item)
}

func (s *CartService) UpdateCart(userID uint, goodsID uint, quantity uint) error {
	item, err := s.cartRepo.GetItem(userID, goodsID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("购物车内无此商品")
	}

	item.Quantity = quantity
	return s.cartRepo.UpdateItem(item)
}

func (s *CartService) DeleteCart(userID uint, goodsID uint) error {
	return s.cartRepo.DeleteItem(userID, goodsID)
}

func (s *CartService) ToggleSelect(userID uint, goodsID uint, selected bool) error {
	item, err := s.cartRepo.GetItem(userID, goodsID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("购物车内无此商品")
	}

	item.Selected = selected
	return s.cartRepo.UpdateItem(item)
}

func (s *CartService) GetCartList(userID uint) (*types.CartListResp, error) {
	items, err := s.cartRepo.GetCartByUserID(userID)
	if err != nil {
		return nil, err
	}

	var resp = &types.CartListResp{
		Items:      make([]types.CartItemResp, 0),
		TotalPrice: 0,
	}

	for _, v := range items {
		price, _, goodsName, err := s.goodsRepo.GetGoodsByID(v.GoodsID)
		if err != nil {
			continue // 如果商品已下架/删除，跳过
		}

		resp.Items = append(resp.Items, types.CartItemResp{
			ID:        v.ID,
			GoodsID:   v.GoodsID,
			GoodsName: goodsName,
			Price:     price,
			Quantity:  v.Quantity,
			Selected:  v.Selected,
		})

		if v.Selected {
			resp.TotalPrice += price * v.Quantity
		}
	}

	return resp, nil
}
