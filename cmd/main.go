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
	"sync"
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

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	app.Start(ctx, &wg, appl)
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

	// 通知 Job 和后台协程停止
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server Shutdown Err:", err)
	}

	// 等待后台 Job 处理完毕
	wg.Wait()

	// 清理连接池等资源
	if appl.Db != nil {
		if sqlDB, err := appl.Db.DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}
	if appl.Rdb != nil {
		appl.Rdb.Close()
	}
	if appl.Mq != nil {
		appl.Mq.Close()
	}
	fmt.Println("Server exiting")
}
