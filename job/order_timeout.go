package job

import (
	"context"
	"fmt"
	"order-payment-system/internal/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

// OrderTimeoutJob 超时订单任务
type OrderTimeoutJob struct {
	orderRepo *repository.OrderRepo
	rdb       *redis.Client
}

func NewOrderTimeoutJob(orderRepo *repository.OrderRepo, rdb *redis.Client) *OrderTimeoutJob {
	return &OrderTimeoutJob{
		orderRepo: orderRepo,
		rdb:       rdb,
	}
}

// Start 开始死循环监听超时订单
func (j *OrderTimeoutJob) Start() {
	fmt.Println("启动超时订单监听任务...")

	ctx := context.Background()

	go func() {
		for {
			now := time.Now().Unix()

			orderIDs, err := j.rdb.ZRangeByScore(ctx, "order:timeout:queue", &redis.ZRangeBy{
				Min: "-inf",
				Max: fmt.Sprintf("%d", now),
			}).Result()

			if err != nil {
				fmt.Println("获取超时订单失败：", err)
				time.Sleep(1 * time.Second)
				continue
			}

			//取消超时订单
			for _, orderID := range orderIDs {
				err := j.orderRepo.ChangeStatusToConceled(orderID)
				if err != nil {
					fmt.Println("取消订单失败：", orderID)
					continue
				}

				j.rdb.ZRem(ctx, "order:timeout:queue", orderID)
			}

			time.Sleep(1 * time.Second)
		}
	}()
}
