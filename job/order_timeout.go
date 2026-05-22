package job

import (
	"context"
	"fmt"
	"order-payment-system/internal/service"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// OrderTimeoutJob 超时订单任务
type OrderTimeoutJob struct {
	orderService *service.OrderService
	rdb          *redis.Client
}

func NewOrderTimeoutJob(orderService *service.OrderService, rdb *redis.Client) *OrderTimeoutJob {
	return &OrderTimeoutJob{
		orderService: orderService,
		rdb:          rdb,
	}
}

// Start 开始死循环监听超时订单
func (j *OrderTimeoutJob) Start(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("启动超时订单监听任务...")
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			now := time.Now().Unix()

			orderNos, err := j.rdb.ZRangeByScore(ctx, "order:timeout:queue", &redis.ZRangeBy{
				Min: "-inf",
				Max: fmt.Sprintf("%d", now),
			}).Result()

			if err != nil {
				fmt.Println("获取超时订单失败：", err)
				time.Sleep(1 * time.Second)
				continue
			}

			//取消超时订单
			for _, orderNo := range orderNos {
				err := j.orderService.CancelOrder(orderNo)
				if err != nil {
					fmt.Println("取消订单失败：", orderNo)
					continue
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}
