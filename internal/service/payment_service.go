package service

import (
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
)

type PaymentService struct {
	orderRepo *repository.OrderRepo
	userRepo  *repository.UserRepo
	goodsRepo *repository.GoodsRepo
}

func NewPaymentService(orderRepo *repository.OrderRepo, userRepo *repository.UserRepo, goodsRepo *repository.GoodsRepo) *PaymentService {
	return &PaymentService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		goodsRepo: goodsRepo,
	}
}

func (p *PaymentService) GetOrder(orderNo string) (*model.Order, error) {
	return p.orderRepo.GetOrderByOrderNo(orderNo)
}

func (p *PaymentService) Settling(order *model.Order) error {
	if order.Status != 0 {
		return errors.New("订单已支付或失效")
	}
	err := p.userRepo.Deduct(order.UserID, order.TotalPrice)
	if err != nil {
		return err
	}
	err = p.goodsRepo.ReduceStock(order.GoodsID, order.BuyNum)
	if err != nil {
		return err
	}
	err = p.orderRepo.ChangeStatus(order.OrderNo)
	if err != nil {
		return err
	}
	err = p.orderRepo.ChangeTime(order.OrderNo)
	if err != nil {
		return err
	}
	return nil
}
