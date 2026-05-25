package job

import (
	"context"
	"encoding/json"
	"order-payment-system/internal/model"
	"order-payment-system/internal/service"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type OrderCreateConsumer struct {
	orderService *service.OrderService
	mq           *amqp.Connection
}

func NewOrderCreateConsumer(orderService *service.OrderService, mq *amqp.Connection) *OrderCreateConsumer {
	return &OrderCreateConsumer{
		orderService: orderService,
		mq:           mq,
	}
}

func (c *OrderCreateConsumer) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err := c.consumeMessages(ctx)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(1 * time.Second)
				}
			} else {
				return
			}
		}
	}()
}

func (c *OrderCreateConsumer) consumeMessages(ctx context.Context) error {
	ch, err := c.mq.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(
		"order_create_queue",
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil // 正常退出，不报错

		case d, ok := <-msgs:
			if !ok {
				return err
			}

			// 提取 TraceID 并放入 context
			msgCtx := context.Background()
			if d.Headers != nil {
				if spanID, ok := d.Headers["x-trace-id"].(string); ok && spanID != "" {
					msgCtx = context.WithValue(msgCtx, "trace_id", spanID)
				}
			}

			var order model.Order
			if err := json.Unmarshal(d.Body, &order); err != nil {
				d.Nack(false, false)
				continue
			}

			if err := c.orderService.SaveOrder(&order); err != nil {
				if c.orderService.IsDuplicateKeyError(err) {
					d.Ack(false)
				} else {
					c.handleFail(ch, d)
				}
				continue
			}

			// 将订单发送至延迟队列，等待超时处理
			if err := ch.Publish(
				"",
				"order_delay_queue",
				false,
				false,
				amqp.Publishing{
					ContentType:  "text/plain",
					Body:         []byte(order.OrderNo),
					DeliveryMode: amqp.Persistent,
				},
			); err != nil {
				c.handleFail(ch, d)
				continue
			}

			if err := d.Ack(false); err != nil {
				continue
			}
		}
	}
}

func (c *OrderCreateConsumer) handleFail(ch *amqp.Channel, d amqp.Delivery) {
	var retryCount int32
	if d.Headers != nil {
		if c, ok := d.Headers["x-retry-count"].(int32); ok {
			retryCount = c
		}
	}

	if retryCount >= 3 {
		// 超过最大重试次数，转入死信队列
		d.Nack(false, false)
		return
	}

	// 增加重试次数记录并重新投递
	headers := amqp.Table{}
	if d.Headers != nil {
		for k, v := range d.Headers {
			headers[k] = v
		}
	}
	headers["x-retry-count"] = retryCount + 1

	err := ch.Publish(
		d.Exchange,
		d.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Headers:      headers,
			ContentType:  d.ContentType,
			Body:         d.Body,
			DeliveryMode: amqp.Persistent, // 或者看原本的设置
		},
	)
	if err != nil {
		// 如果重投失败，只能要求MQ再此发送本条
		d.Nack(false, true)
		return
	}
	// 重投成功，确认旧消息
	d.Ack(false)
}
