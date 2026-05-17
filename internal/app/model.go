package app

import (
	"order-payment-system/internal/handler"
	"order-payment-system/job"
)

type App struct {
	UserHandler    *handler.UserHandler
	GoodsHandler   *handler.GoodsHandler
	OrderHandler   *handler.OrderHandler
	PaymentHandler *handler.PaymentHandler

	orderTimeout *job.OrderTimeoutJob
	orderCreate  *job.OrderCreateConsumer
	cachePreheat *job.GoodsCacheWarmJob
}
