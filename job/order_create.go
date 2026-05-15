package job

import (
	"context"
	"encoding/json"
	"order-payment-system/internal/model"
	"order-payment-system/internal/service"
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

func (c *OrderCreateConsumer) Start() {
	ctx := context.Background()
	go func() {
		for {
			err := c.consumeMessages(ctx)
			if err != nil {
				time.Sleep(1 * time.Second)
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

	q, err := ch.QueueDeclare(
		"order_create_queue",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
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

			var order model.Order
			if err := json.Unmarshal(d.Body, &order); err != nil {
				d.Nack(false, false)
				continue
			}

			if err := c.orderService.SaveOrder(&order); err != nil {
				d.Nack(false, true)
				continue
			}

			if err := c.orderService.AddQueue(order.OrderNo); err != nil {
				d.Nack(false, true)
				continue
			}

			if err := d.Ack(false); err != nil {
				continue
			}
		}
	}
}
