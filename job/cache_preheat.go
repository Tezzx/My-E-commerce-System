package job

import (
	"context"
	"encoding/json"
	"fmt"
	"order-payment-system/internal/model"
	"order-payment-system/internal/service"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// GoodsCacheWarmJob 商品缓存预热任务
type GoodsCacheWarmJob struct {
	goodsService *service.GoodsService
	rdb          *redis.Client
	db           *gorm.DB // 字段，不是方法
}

func NewGoodsCacheWarmJob(goodsService *service.GoodsService, rdb *redis.Client, db *gorm.DB) *GoodsCacheWarmJob {
	return &GoodsCacheWarmJob{
		goodsService: goodsService,
		rdb:          rdb,
		db:           db,
	}
}

func (j *GoodsCacheWarmJob) Warm() error {
	fmt.Println("开始商品缓存预热...")

	var goodsList []model.Goods
	if err := j.db.Find(&goodsList).Error; err != nil {
		return err
	}

	ctx := context.Background()
	pipe := j.rdb.Pipeline()

	for _, g := range goodsList {
		goodsIDStr := strconv.FormatUint(uint64(g.ID), 10)

		data, err := json.Marshal(g)
		if err != nil {
			return err
		}
		pipe.Set(ctx, "goods:info:"+goodsIDStr, data, 0)

		pipe.HSet(ctx, "goods", goodsIDStr, g.Goodsnum)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
