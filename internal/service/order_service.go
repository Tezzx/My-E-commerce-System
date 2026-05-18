package service

import (
	"encoding/json"
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/logger"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type OrderService struct {
	orderRepo *repository.OrderRepo
	goodsRepo *repository.GoodsRepo
	mq        *amqp.Connection
}

func NewOrderService(orderRepo *repository.OrderRepo, goodsRepo *repository.GoodsRepo, mq *amqp.Connection) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		goodsRepo: goodsRepo,
		mq:        mq,
	}
}

func (o *OrderService) CreateOrder(userID uint, goodsID uint, buyNum uint) (string, error) {
	price, _, goodsName, err := o.goodsRepo.GetGoodsByID(goodsID)
	if err != nil {
		logger.Log.Warn("创建订单 - 商品不存在", zap.Uint("goods_id", goodsID))
		return "", err
	}

	success, err := o.goodsRepo.DeductStock(goodsID, int64(buyNum))
	if err != nil {
		logger.Log.Error("创建订单 - 扣减库存失败",
			zap.Uint("goods_id", goodsID),
			zap.Uint("buy_num", buyNum),
			zap.Error(err))
		return "", err
	}
	if !success {
		logger.Log.Warn("创建订单 - 库存不足",
			zap.Uint("goods_id", goodsID),
			zap.Uint("buy_num", buyNum))
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
		logger.Log.Error("创建订单 - 消息序列化失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return "", err
	}

	ch, err := o.mq.Channel()
	if err != nil {
		logger.Log.Error("创建订单 - 获取MQ通道失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return "", err
	}
	defer ch.Close()

	err = ch.Publish(
		"",
		"order_create_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		logger.Log.Error("创建订单 - 发布MQ消息失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return "", err
	}

	return orderNo, nil
}

func (o *OrderService) SaveOrder(order *model.Order) error {
	err := o.orderRepo.SaveOrder(order)
	if err != nil {
		logger.Log.Error("保存订单到数据库失败",
			zap.String("order_no", order.OrderNo),
			zap.Error(err))
	}
	return err
}

func (o *OrderService) AddQueue(orderNo string) error {
	err := o.orderRepo.AddQueue(orderNo)
	if err != nil {
		logger.Log.Error("添加订单到超时队列失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
	}
	return err
}

func (o *OrderService) CancelOrder(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		logger.Log.Warn("取消订单 - 订单不存在",
			zap.String("order_no", orderNo))
		return err
	}
	if order.Status != 0 {
		_ = o.orderRepo.DelQueue(orderNo)
		return nil
	}

	err = o.goodsRepo.IncrementStock(order.GoodsID, int64(order.BuyNum))
	if err != nil {
		logger.Log.Error("取消订单 - 回滚库存失败",
			zap.String("order_no", orderNo),
			zap.Uint("goods_id", order.GoodsID),
			zap.Uint("buy_num", order.BuyNum),
			zap.Error(err))
		return err
	}

	err = o.orderRepo.CancelOrderStatus(orderNo)
	if err != nil {
		logger.Log.Error("取消订单 - 更新状态失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return err
	}

	err = o.orderRepo.DelQueue(orderNo)
	if err != nil {
		logger.Log.Warn("取消订单 - 从队列删除失败（可容忍）",
			zap.String("order_no", orderNo),
			zap.Error(err))
	}
	return nil
}

func (o *OrderService) PayOrder(orderNo string) error {
	err := o.orderRepo.ChangeStatusToPayed(orderNo)
	if err != nil {
		logger.Log.Error("支付订单 - 更新状态失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
	}
	return err
}

func (o *OrderService) GetUserOrderList(userID uint) ([]model.Order, error) {
	orders, err := o.orderRepo.GetUserOrderList(userID)
	if err != nil {
		logger.Log.Error("获取用户订单列表失败",
			zap.Uint("user_id", userID),
			zap.Error(err))
	}
	return orders, err
}

func (o *OrderService) generateOrderNo(goodsID, userID uint) string {
	return time.Now().Format("20060102150405") +
		strconv.Itoa(int(userID)) +
		strconv.Itoa(int(goodsID))
}
