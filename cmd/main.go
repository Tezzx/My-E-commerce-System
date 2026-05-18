package main

import (
	"order-payment-system/internal/app"
	"order-payment-system/pkg/logger"

	"go.uber.org/zap"
)

func main() {

	logger.Init()
	defer logger.Log.Sync()
	logger.Log.Info("app starting...")

	appl, port := app.InitializeApp()

	r := app.SetupRoutes(appl)

	app.Start(appl)

	logger.Log.Info("http server starting", zap.String("port", port))
	r.Run(":" + port)

}
