package app

import "order-payment-system/pkg/logger"

func Start(app *App) {
	app.cachePreheat.Warm()
	app.orderTimeout.Start()
	app.orderCreate.Start()
	logger.Log.Info("后台服务启动成功")
}
