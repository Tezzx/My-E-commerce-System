package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/logger"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var orderSeq uint32

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

func (o *OrderService) CreateOrder(ctx context.Context, userID uint, goodsID uint, buyNum uint) (string, error) {
	// ========== 1. 查询商品 & 扣减库存==========
	price, _, goodsName, err := o.goodsRepo.GetGoodsByID(goodsID)
	if err != nil {
		logger.Ctx(ctx).Warn("创建订单 - 商品不存在", zap.Uint("goods_id", goodsID))
		return "", err
	}

	success, err := o.goodsRepo.DeductStock(goodsID, int64(buyNum))
	if err != nil {
		logger.Log.Error("创建订单 - 扣减库存失败",
			zap.Uint("goods_id", goodsID),
			zap.Uint("buyNum", buyNum),
			zap.Error(err))
		return "", err
	}
	if !success {
		logger.Log.Warn("创建订单 - 库存不足",
			zap.Uint("goods_id", goodsID),
			zap.Uint("buyNum", buyNum))
		return "", errors.New("库存不足")
	}

	// ========== 2.设置库存补偿保护 ==========
	needCompensate := true
	defer func() {
		if needCompensate {
			if compErr := o.goodsRepo.IncrementStock(goodsID, int64(buyNum)); compErr != nil {
				logger.Log.Error("创建订单 - 补偿库存失败！需人工介入",
					zap.Uint("goods_id", goodsID),
					zap.Uint("buyNum", buyNum),
					zap.Uint("user_id", userID),
					zap.Error(compErr))
			} else {
				logger.Log.Info("创建订单 - 库存补偿成功",
					zap.Uint("goods_id", goodsID),
					zap.Uint("buyNum", buyNum),
					zap.Uint("user_id", userID))
			}
		}
	}()

	// ========== 3. 构建订单消息==========
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

	// ========== 4.获取通道 + 启用 Confirm + 阻塞等待 Ack ==========
	ch, err := o.mq.Channel()
	if err != nil {
		logger.Log.Error("创建订单 - 获取MQ通道失败", zap.String("order_no", orderNo), zap.Error(err))
		return "", err
	}
	defer ch.Close()

	// 启用 Publisher Confirm 模式
	if err := ch.Confirm(false); err != nil {
		logger.Ctx(ctx).Error("创建订单 - 启用Confirm模式失败", zap.String("order_no", orderNo), zap.Error(err))
		return "", err
	}

	// 传递 traceID
	traceID := ""
	if val := ctx.Value("trace_id"); val != nil {
		traceID = val.(string)
	}
	headers := amqp.Table{"x-trace-id": traceID}

	//注册确认监听
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	// 发布持久化消息
	if err := ch.Publish(
		"",
		"order_create_queue",
		false,
		false,
		amqp.Publishing{
			Headers:      headers,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	); err != nil {
		logger.Ctx(ctx).Error("创建订单 - 发布MQ消息失败", zap.String("order_no", orderNo), zap.Error(err))
		return "", err
	}

	// 阻塞等待 Broker Ack
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			logger.Ctx(ctx).Error("创建订单 - 消息被Broker拒绝(Nack)", zap.String("order_no", orderNo))
			return "", errors.New("消息被Broker拒绝(Nack)")
		}

	case <-time.After(3 * time.Second):
		logger.Ctx(ctx).Error("创建订单 - 等待Broker确认超时", zap.String("order_no", orderNo))
		return "", errors.New("等待Broker确认超时")
	}

	// ========== 5.确认成功 ==========
	needCompensate = false
	logger.Ctx(ctx).Info("创建订单 - 消息已安全抵达Broker",
		zap.String("order_no", orderNo))

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

func (o *OrderService) CancelOrder(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		logger.Log.Warn("取消订单 - 订单不存在",
			zap.String("order_no", orderNo))
		return err
	}
	if order.Status != 0 {
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

	return nil
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
	now := time.Now()
	timeStr := now.Format("20060102150405")
	ms := now.UnixMilli() % 1000
	seq := atomic.AddUint32(&orderSeq, 1) % 10000

	return fmt.Sprintf("%s%03d%d%d%04d", timeStr, ms, userID, goodsID, seq)
}

func (o *OrderService) IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return strings.Contains(err.Error(), "Duplicate entry")
}
