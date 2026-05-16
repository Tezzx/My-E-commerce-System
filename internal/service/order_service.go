package service

import (
	"encoding/json"
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"strconv"
	"time"

	"github.com/streadway/amqp"
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

// --扣redis库存并加入队列--加入超时队列--订单存入mysql
// --加redis库存--改变订单状态
// 创建订单（扣库存然后传消息队列）
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

// 把订单存到mysql
func (o *OrderService) SaveOrder(order *model.Order) error {
	return o.orderRepo.SaveOrder(order)
}

// 把订单加入超时队列
func (o *OrderService) AddQueue(orderNo string) error {
	return o.orderRepo.AddQueue(orderNo)
}

// 取消超时订单并回滚库存
func (o *OrderService) CancelOrder(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return err
	}
	if order.Status != 0 {
		_ = o.orderRepo.DelQueue(orderNo)
		return nil
	}
	err = o.goodsRepo.IncrementStock(order.GoodsID, int64(order.BuyNum))
	if err != nil {
		return err
	}
	err = o.orderRepo.CancelOrderStatus(orderNo)
	if err != nil {
		return err
	}
	err = o.orderRepo.DelQueue(orderNo)
	if err != nil {
		return err
	}
	return nil
}

// 支付订单
func (o *OrderService) PayOrder(orderNo string) error {
	return o.orderRepo.ChangeStatusToPayed(orderNo)
}

// -----------------------辅助功能------------------------
// 获取当前用户的所有订单
func (o *OrderService) GetUserOrderList(userID uint) ([]model.Order, error) {
	return o.orderRepo.GetUserOrderList(userID)
}

// 生成唯一订单号
func (o *OrderService) generateOrderNo(goodsID, userID uint) string {
	return time.Now().Format("20060102150405") +
		strconv.Itoa(int(userID)) +
		strconv.Itoa(int(goodsID))
}
