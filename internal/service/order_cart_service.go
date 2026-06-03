package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order-payment-system/internal/model"
	"order-payment-system/pkg/logger"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// CreateCartOrder 从购物车创建订单
func (o *OrderService) CreateCartOrder(ctx context.Context, userID uint, cartItems []model.CartItem, addressID uint, addressSnapshot string) (string, error) {
	if len(cartItems) == 0 {
		return "", errors.New("无选中商品")
	}

	var total uint = 0
	var orderItems []model.OrderItem
	var successDeductions []model.CartItem // 记录成功扣库存的商品，用于回滚

	// 1. 遍历商品扣减Redis库存
	for _, item := range cartItems {
		if !item.Selected {
			continue
		}

		price, _, goodsName, err := o.goodsRepo.GetGoodsByID(item.GoodsID)
		if err != nil {
			o.rollbackRedisStock(successDeductions)
			return "", fmt.Errorf("商品 %d 不存在", item.GoodsID)
		}

		success, err := o.goodsRepo.DeductStock(item.GoodsID, int64(item.Quantity))
		if err != nil || !success {
			o.rollbackRedisStock(successDeductions)
			return "", fmt.Errorf("商品 %s 库存不足", goodsName)
		}

		successDeductions = append(successDeductions, item)
		total += price * item.Quantity
		orderItems = append(orderItems, model.OrderItem{
			GoodsID:   item.GoodsID,
			GoodsName: goodsName,
			Price:     price,
			Quantity:  item.Quantity,
		})
	}

	if len(orderItems) == 0 {
		return "", errors.New("无选中商品")
	}

	// 2. 构造主订单
	now := time.Now()
	orderNo := fmt.Sprintf("%s%03d%d%d", now.Format("20060102150405"), now.UnixMilli()%1000, userID, 0)

	msg := &model.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		GoodsID:         0,
		GoodsName:       "购物车合并订单",
		Price:           0,
		BuyNum:          0,
		TotalPrice:      total,
		Status:          model.OrderStatusPending,
		AddressID:       &addressID,
		AddressSnapshot: addressSnapshot,
		OrderItems:      orderItems,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		o.rollbackRedisStock(successDeductions)
		return "", err
	}

	// 3. 发布 MQ 消息
	resultChan := make(chan error, 1)
	o.publishChan <- &publishTask{
		exchange:   "",
		routingKey: "order_create_queue",
		publishing: amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         msgBytes,
			Headers: amqp.Table{
				"x-trace-id": ctx.Value("trace_id"),
			},
		},
		resultChan: resultChan,
	}

	// 4. 等待确认
	select {
	case err := <-resultChan:
		if err != nil {
			o.rollbackRedisStock(successDeductions)
			logger.Ctx(ctx).Error("购物车订单创建失败 - MQ发布出错", zap.Error(err))
			return "", err
		}
	case <-time.After(3 * time.Second):
		o.rollbackRedisStock(successDeductions)
		logger.Ctx(ctx).Error("购物车订单创建失败 - MQ发布超时")
		return "", errors.New("发布消息超时")
	}

	return orderNo, nil
}

// 辅助方法：回滚批量库存
func (o *OrderService) rollbackRedisStock(deductedItems []model.CartItem) {
	for _, item := range deductedItems {
		_ = o.goodsRepo.IncrementStock(item.GoodsID, int64(item.Quantity))
	}
}
