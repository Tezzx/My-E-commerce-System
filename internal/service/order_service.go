package service

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo *repository.OrderRepo
	goodsRepo *repository.GoodsRepo // 依赖商品仓库：校验商品+扣库存
}

// NewOrderService 构造函数
func NewOrderService(orderRepo *repository.OrderRepo, goodsRepo *repository.GoodsRepo) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		goodsRepo: goodsRepo,
	}
}

// 创建订单
func (o *OrderService) CreateOrder(userID uint, goodsID uint, buyNum uint) (*model.Order, error) {

	price, stock, goodsName, err := o.goodsRepo.GetGoodsByID(goodsID)
	if err != nil {
		return nil, err
	}

	if stock < buyNum {
		return nil, err
	}

	err = o.goodsRepo.ReduceStock(goodsID, buyNum)
	if err != nil {
		return nil, err
	}

	orderNo := generateOrderNo(goodsID, userID)
	totalPrice := price * buyNum

	order := &model.Order{
		OrderNo:    orderNo,
		UserID:     userID,
		GoodsID:    goodsID,
		GoodsName:  goodsName,
		Price:      price,
		BuyNum:     buyNum,
		TotalPrice: totalPrice,
		Status:     0,
	}

	err = o.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}
	err = o.orderRepo.AddQueue(orderNo)
	return order, nil
}

// 获取当前用户的所有订单
func (o *OrderService) GetUserOrderList(userID uint) ([]model.Order, error) {
	return o.orderRepo.GetUserOrderList(userID)
}

// 支付订单
func (o *OrderService) PayOrder(orderID uint) error {
	return o.orderRepo.UpdateOrderPayStatus(orderID)
}

// 生成唯一订单号
func generateOrderNo(goodsID, userID uint) string {
	return time.Now().Format("20060102150405") +
		strconv.Itoa(int(userID)) +
		strconv.Itoa(int(goodsID))
}

func (o *OrderService) GetUserOrders(userID uint) ([]model.Order, error) {
	return o.orderRepo.GetAllOrdersByUserID(userID)
}

// 取消超时订单并回滚库存
func (o *OrderService) CancelTimeoutOrder(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return err
	}

	db := o.orderRepo.ReturnDB()

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("order_no = ? AND status = ?", orderNo, 0).Update("status", 2).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Goods{}).Where("id = ?", order.GoodsID).UpdateColumn("goodsnum", gorm.Expr("goodsnum + ?", order.BuyNum)).Error; err != nil {
			return err
		}

		return nil
	})
}
