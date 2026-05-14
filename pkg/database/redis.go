package database

import (
	"order-payment-system/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.RedisConfig) *redis.Client {
	addr := cfg.Host + ":" + cfg.Port
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
		//设置超时
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	return rdb
}
