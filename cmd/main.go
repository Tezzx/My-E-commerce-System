package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"order-payment-system/internal/app"
	"order-payment-system/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	go func() {
		logger.Log.Info("http server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown Err:", err)
	}

}
