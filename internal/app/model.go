package app

import (
	"order-payment-system/internal/handler"
	"order-payment-system/internal/repository"
	"order-payment-system/job"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type App struct {
	userHandler     *handler.UserHandler
	goodsHandler    *handler.GoodsHandler
	orderHandler    *handler.OrderHandler
	paymentHandler  *handler.PaymentHandler
	cartHandler     *handler.CartHandler
	addressHandler  *handler.AddressHandler
	categoryHandler *handler.CategoryHandler
	reviewHandler   *handler.ReviewHandler

	UserRepo *repository.UserRepo

	orderTimeout *job.OrderTimeoutJob
	orderCreate  *job.OrderCreateConsumer
	cachePreheat *job.GoodsCacheWarmJob

	Db  *gorm.DB
	Rdb *redis.Client
	Mq  *amqp.Connection
}
