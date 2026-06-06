package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order-payment-system/internal/errs"
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
	orderRepo   *repository.OrderRepo
	goodsRepo   *repository.GoodsRepo
	addressRepo *repository.AddressRepo
	mq          *amqp.Connection
	mqChannel   *amqp.Channel
	confirmChan chan amqp.Confirmation
	publishChan chan *publishTask
}

type publishTask struct {
	exchange   string
	routingKey string
	publishing amqp.Publishing
	resultChan chan error
}

func NewOrderService(orderRepo *repository.OrderRepo, goodsRepo *repository.GoodsRepo, addressRepo *repository.AddressRepo, mq *amqp.Connection) *OrderService {
	ch, err := mq.Channel()
	if err != nil {
		panic(fmt.Errorf("failed to create channel: %w", err))
	}
	if err := ch.Confirm(false); err != nil {
		panic(fmt.Errorf("failed to enable confirm mode: %w", err))
	}
	confirmChan := ch.NotifyPublish(make(chan amqp.Confirmation, 100))

	s := &OrderService{
		orderRepo:   orderRepo,
		goodsRepo:   goodsRepo,
		mq:          mq,
		mqChannel:   ch,
		confirmChan: confirmChan,
		publishChan: make(chan *publishTask, 1024),
	}

	go s.asyncPublisher()
	return s
}

func (s *OrderService) asyncPublisher() {
	// deliveryTag 从 1 开始自增，与 RabbitMQ broker 的确认序号一一对应
	var deliveryTag uint64
	pending := make(map[uint64]*publishTask)

	for {
		select {
		case task := <-s.publishChan:
			err := s.mqChannel.Publish(
				task.exchange,
				task.routingKey,
				false,
				false,
				task.publishing,
			)
			if err != nil {
				task.resultChan <- err
				continue
			}

			// 仅在发布成功后递增，保证与 broker 分配的 DeliveryTag 一致
			deliveryTag++
			pending[deliveryTag] = task

		case confirm := <-s.confirmChan:
			if task, ok := pending[confirm.DeliveryTag]; ok {
				delete(pending, confirm.DeliveryTag)
				if confirm.Ack {
					task.resultChan <- nil
				} else {
					task.resultChan <- errors.New("消息被Broker拒绝(Nack)")
				}
			}
		}
	}
}

func (o *OrderService) CreateOrder(ctx context.Context, userID uint, goodsID uint, buyNum uint, addressID *uint, addressSnapshot string) (string, error) {
	// ========== 1. 查询商品 & 扣减库存 ==========
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
		return "", fmt.Errorf("%w：商品ID=%d 请求数量=%d", errs.InsufficientStock, goodsID, buyNum)
	}

	// ========== 2. 设置库存补偿保护 ==========
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

	// ========== 3. 构建订单消息 ==========
	orderNo := o.generateOrderNo(goodsID, userID)
	totalPrice := price * buyNum
	msg := &model.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		GoodsID:         goodsID,
		GoodsName:       goodsName,
		Price:           price,
		BuyNum:          buyNum,
		TotalPrice:      totalPrice,
		Status:          model.OrderStatusPending,
		AddressID:       addressID,
		AddressSnapshot: addressSnapshot,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error("创建订单 - 消息序列化失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return "", err
	}

	// ========== 4. 提交到 asyncPublisher 并等待结果 ==========
	traceID := ""
	if val := ctx.Value("trace_id"); val != nil {
		traceID = val.(string)
	}
	headers := amqp.Table{"x-trace-id": traceID}

	resultChan := make(chan error, 1)
	o.publishChan <- &publishTask{
		exchange:   "",
		routingKey: "order_create_queue",
		publishing: amqp.Publishing{
			Headers:      headers,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
		resultChan: resultChan,
	}

	select {
	case err := <-resultChan:
		if err != nil {
			logger.Ctx(ctx).Error("创建订单 - 消息发布失败", zap.String("order_no", orderNo), zap.Error(err))
			return "", err
		}
	case <-time.After(3 * time.Second):
		logger.Ctx(ctx).Error("创建订单 - 等待MQ确认超时", zap.String("order_no", orderNo))
		return "", errs.OrderMQAckTimeout
	}

	// ========== 5. 确认成功 ==========
	needCompensate = false
	logger.Ctx(ctx).Info("创建订单 - 消息已安全抵达Broker", zap.String("order_no", orderNo))
	return orderNo, nil
}

func (o *OrderService) SaveOrder(order *model.Order) error {
	err := o.orderRepo.SaveOrder(order)
	if err != nil {
		logger.Log.Error("保存订单到数据库失败",
			zap.String("order_no", order.OrderNo),
			zap.Error(err))
		return err
	}

	// 记录订单创建日志
	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   order.OrderNo,
		OldStatus: -1,
		NewStatus: model.OrderStatusPending,
		Operator:  order.UserID,
		Remark:    "订单创建",
	})
	return nil
}

func (o *OrderService) CancelOrder(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		logger.Log.Warn("取消订单 - 订单不存在",
			zap.String("order_no", orderNo))
		return err
	}
	if !model.CanTransition(order.Status, model.OrderStatusCancelled) {
		logger.Log.Warn("取消订单 - 当前状态不允许取消",
			zap.String("order_no", orderNo),
			zap.String("status", model.OrderStatusNames[order.Status]))
		return fmt.Errorf("%w：当前状态【%s】", errs.OrderStatusInvalid, model.OrderStatusNames[order.Status])
	}

	// 先更新订单状态（乐观锁），成功后再回滚库存，避免数据不一致
	err = o.orderRepo.CancelOrderStatus(orderNo)
	if err != nil {
		logger.Log.Error("取消订单 - 更新状态失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return err
	}

	// 状态更新成功后回滚库存
	if len(order.OrderItems) > 0 {
		for _, item := range order.OrderItems {
			if err := o.goodsRepo.IncrementStock(item.GoodsID, int64(item.Quantity)); err != nil {
				logger.Log.Error("取消订单 - 回滚库存失败！需人工介入",
					zap.String("order_no", orderNo),
					zap.Uint("goods_id", item.GoodsID),
					zap.Uint("buy_num", item.Quantity),
					zap.Error(err))
			}
		}
	} else {
		if err := o.goodsRepo.IncrementStock(order.GoodsID, int64(order.BuyNum)); err != nil {
			logger.Log.Error("取消订单 - 回滚库存失败！需人工介入",
				zap.String("order_no", orderNo),
				zap.Uint("goods_id", order.GoodsID),
				zap.Uint("buy_num", order.BuyNum),
				zap.Error(err))
		}
	}

	// 记录取消日志
	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusCancelled,
		Operator:  0,
		Remark:    "订单取消",
	})

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

// GetUserOrderListWithPage 分页获取用户订单
func (o *OrderService) GetUserOrderListWithPage(userID uint, page, size int, status *int) ([]model.Order, int64, error) {
	orders, total, err := o.orderRepo.GetUserOrderListWithPage(userID, page, size, status)
	if err != nil {
		logger.Log.Error("分页获取用户订单列表失败",
			zap.Uint("user_id", userID),
			zap.Error(err))
	}
	return orders, total, err
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

// ---------- 订单状态流转 ----------

// ShipOrder 已支付 → 已发货（需物流信息）
func (o *OrderService) ShipOrder(orderNo, shipCompany, shipNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}
	if !model.CanTransition(order.Status, model.OrderStatusShipped) {
		return fmt.Errorf("%w：当前状态【%s】不允许发货", errs.OrderStatusInvalid, model.OrderStatusNames[order.Status])
	}

	if err := o.orderRepo.UpdateStatusWithShip(orderNo, shipCompany, shipNo); err != nil {
		return err
	}
	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusShipped,
		Operator:  0,
		Remark:    fmt.Sprintf("物流公司:%s 单号:%s", shipCompany, shipNo),
	})
	return nil
}

// ConfirmReceipt 已发货 → 已收货（用户确认）
func (o *OrderService) ConfirmReceipt(userID uint, orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}
	if order.UserID != userID {
		return fmt.Errorf("%w", errs.OrderNotOwned)
	}
	if !model.CanTransition(order.Status, model.OrderStatusReceived) {
		return fmt.Errorf("%w：当前状态【%s】不允许确认收货", errs.OrderStatusInvalid, model.OrderStatusNames[order.Status])
	}

	if err := o.orderRepo.UpdateStatus(orderNo, order.Status, model.OrderStatusReceived); err != nil {
		return err
	}
	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusReceived,
		Operator:  userID,
		Remark:    "用户确认收货",
	})
	return nil
}

// CompleteOrder 已收货 → 已完成
func (o *OrderService) CompleteOrder(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}
	if !model.CanTransition(order.Status, model.OrderStatusCompleted) {
		return fmt.Errorf("%w：当前状态【%s】不允许完成", errs.OrderStatusInvalid, model.OrderStatusNames[order.Status])
	}

	if err := o.orderRepo.UpdateStatus(orderNo, order.Status, model.OrderStatusCompleted); err != nil {
		return err
	}
	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusCompleted,
		Operator:  0,
		Remark:    "订单已完成",
	})
	return nil
}

// GetOrderLogs 查询订单操作日志
func (o *OrderService) GetOrderLogs(orderNo string) ([]model.OrderLog, error) {
	return o.orderRepo.GetOrderLogs(orderNo)
}

// ---------- 退款流程 ----------

// RequestRefund 用户申请退款（Paid/Shipped/Received → Refunding）
func (o *OrderService) RequestRefund(userID uint, orderNo, reason string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return fmt.Errorf("%w：%s", errs.OrderNotFound, orderNo)
	}
	if order.UserID != userID {
		return fmt.Errorf("%w", errs.OrderNotOwned)
	}
	if !model.CanTransition(order.Status, model.OrderStatusRefunding) {
		return fmt.Errorf("%w：当前状态【%s】不允许申请退款", errs.RefundNotAllowed, model.OrderStatusNames[order.Status])
	}

	if err := o.orderRepo.UpdateStatus(orderNo, order.Status, model.OrderStatusRefunding); err != nil {
		return fmt.Errorf("%w：%v", errs.RefundFailed, err)
	}

	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusRefunding,
		Operator:  userID,
		Remark:    fmt.Sprintf("用户申请退款，原因:%s", reason),
	})
	return nil
}

// ProcessRefund 商家/管理员同意退款（Refunding → Refunded，回滚库存）
func (o *OrderService) ProcessRefund(orderNo string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return fmt.Errorf("%w：%s", errs.OrderNotFound, orderNo)
	}
	if !model.CanTransition(order.Status, model.OrderStatusRefunded) {
		return fmt.Errorf("%w：当前状态【%s】不允许退款", errs.RefundNotAllowed, model.OrderStatusNames[order.Status])
	}

	// 回滚库存
	if len(order.OrderItems) > 0 {
		for _, item := range order.OrderItems {
			if err := o.goodsRepo.IncrementStock(item.GoodsID, int64(item.Quantity)); err != nil {
				logger.Log.Error("退款回滚库存失败",
					zap.String("order_no", orderNo),
					zap.Uint("goods_id", item.GoodsID),
					zap.Error(err))
				return fmt.Errorf("回滚库存失败: %w", err)
			}
		}
	} else if order.GoodsID > 0 {
		if err := o.goodsRepo.IncrementStock(order.GoodsID, int64(order.BuyNum)); err != nil {
			logger.Log.Error("退款回滚库存失败",
				zap.String("order_no", orderNo),
				zap.Uint("goods_id", order.GoodsID),
				zap.Error(err))
			return fmt.Errorf("回滚库存失败: %w", err)
		}
	}

	if err := o.orderRepo.UpdateStatus(orderNo, order.Status, model.OrderStatusRefunded); err != nil {
		return fmt.Errorf("%w：%v", errs.RefundFailed, err)
	}

	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusRefunded,
		Operator:  0,
		Remark:    "退款审批通过，已退款",
	})
	return nil
}

// RejectRefund 商家/管理员拒绝退款（Refunding → Paid）
func (o *OrderService) RejectRefund(orderNo, remark string) error {
	order, err := o.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return fmt.Errorf("%w：%s", errs.OrderNotFound, orderNo)
	}
	if !model.CanTransition(order.Status, model.OrderStatusPaid) {
		return fmt.Errorf("%w：当前状态【%s】不允许拒绝退款", errs.RefundNotAllowed, model.OrderStatusNames[order.Status])
	}

	if err := o.orderRepo.UpdateStatus(orderNo, order.Status, model.OrderStatusPaid); err != nil {
		return fmt.Errorf("%w：%v", errs.RefundFailed, err)
	}

	_ = o.orderRepo.CreateOrderLog(&model.OrderLog{
		OrderNo:   orderNo,
		OldStatus: order.Status,
		NewStatus: model.OrderStatusPaid,
		Operator:  0,
		Remark:    fmt.Sprintf("退款审批拒绝，原因:%s", remark),
	})
	return nil
}
