package app

import (
	"context"
	"order-payment-system/pkg/logger"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, app *App) {
	app.cachePreheat.Warm()
	app.orderTimeout.Start(ctx, wg)
	app.orderCreate.Start(ctx, wg)
	logger.Log.Info("后台服务启动成功")
}
