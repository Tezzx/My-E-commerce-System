package service

import (
	"encoding/json"
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo *repository.OrderRepo
	goodsRepo *repository.GoodsRepo
	mq        *amqp.Connection
}

// NewOrderService 构造函数
func NewOrderService(orderRepo *repository.OrderRepo, goodsRepo *repository.GoodsRepo, mq *amqp.Connection) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		goodsRepo: goodsRepo,
		mq:        mq,
	}
}

// 创建订单
func (o *OrderService) CreateOrder(userID uint, goodsID uint, buyNum uint) (string, error) {

	price, _, goodsName, err := o.goodsRepo.GetGoodsByID(goodsID)
	if err != nil {
		return "", err
	}
	//原子化扣库存
	success, err := o.goodsRepo.DeductStock(goodsID, int64(buyNum))
	if err != nil {
		return "", err
	}
	if !success {
		return "", errors.New("库存不足")
	}

	orderNo := o.generateOrderNo(goodsID, userID)
	totalPrice := price * buyNum

	msg := &model.Order{
		OrderNo:    orderNo,
		UserID:     userID,
		GoodsID:    goodsID,
		GoodsName:  goodsName,
		Price:      price,
		BuyNum:     buyNum,
		TotalPrice: totalPrice,
		Status:     0,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	ch, err := o.mq.Channel()
	if err != nil {
		return "", err
	}
	defer ch.Close()

	err = ch.Publish(
		"",                   // exchange
		"order_create_queue", // routing key = queue name（direct 模式）
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent, // 持久化消息，防止 RabbitMQ 重启丢失
		},
	)
	if err != nil {
		return "", err
	}

	return orderNo, nil
}

func (o *OrderService) SaveOrder(order *model.Order) error {
	return o.orderRepo.SaveOrder(order)
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
func (o *OrderService) generateOrderNo(goodsID, userID uint) string {
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

func (o *OrderService) AddQueue(id string) error {
	return o.orderRepo.AddQueue(id)
}
