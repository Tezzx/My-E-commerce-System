package job

import (
	"context"
	"fmt"
	"order-payment-system/internal/service"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// OrderTimeoutJob 超时订单任务
type OrderTimeoutJob struct {
	orderService *service.OrderService
	mq           *amqp.Connection
}

func NewOrderTimeoutJob(orderService *service.OrderService, mq *amqp.Connection) *OrderTimeoutJob {
	return &OrderTimeoutJob{
		orderService: orderService,
		mq:           mq,
	}
}

// Start 监听超时订单队列
func (j *OrderTimeoutJob) Start(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("启动超时订单监听任务...")
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			err := j.consumeTimeoutMessages(ctx)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(5 * time.Second)
				}
			} else {
				return
			}
		}
	}()
}

func (j *OrderTimeoutJob) consumeTimeoutMessages(ctx context.Context) error {
	ch, err := j.mq.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		"order_timeout_queue",
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
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			orderNo := string(d.Body)
			err := j.orderService.CancelOrder(orderNo)
			if err != nil {
				fmt.Println("取消订单失败，可能稍后重试：", orderNo)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}
}
