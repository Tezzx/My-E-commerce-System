package app

import (
	"fmt"
	"log"
	"order-payment-system/config"
	"order-payment-system/internal/handler"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/service"
	"order-payment-system/job"
	"order-payment-system/pkg/database"
	"order-payment-system/pkg/jwt"
	"order-payment-system/pkg/logger"
	"strconv"
)

func InitializeApp() (*App, string) {
	// 读取配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("配置文件读取失败: %v", err)
	}
	port := strconv.Itoa(cfg.Server.Port)
	jwt.InitJwtKey(string(cfg.Jwt.Key))
	// 连接数据库
	db, err := database.InitMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")

	//连接redis
	rdb := database.InitRedis(&cfg.Redis)

	//连接rabbitmq
	mq, err := database.InitRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("消息队列连接失败: %v", err)
	}
	fmt.Println("消息队列连接成功")

	// 自动建表
	if err := db.AutoMigrate(&model.User{}, &model.Goods{}, &model.Order{}); err != nil {
		log.Fatalf("数据表创建失败: %v", err)
	}

	//设置队列
	ch, _ := mq.Channel()
	defer ch.Close()
	database.DeclareQueueWithDLX(ch, "order_create_queue")
	database.DeclareDelayTimeoutQueue(ch)

	// 依赖注入
	//internal部分
	userRepo := repository.NewUserRepo(db, rdb)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	goodsRepo := repository.NewGoodsRepo(db, rdb)
	goodsService := service.NewGoodsService(goodsRepo)
	goodsHandler := handler.NewGoodsHandler(goodsService)

	orderRepo := repository.NewOrderRepo(db, rdb)
	orderService := service.NewOrderService(orderRepo, goodsRepo, mq)
	orderHandler := handler.NewOrderHandler(orderService)

	paymentService := service.NewPaymentService(repository.NewTransactionManager(db, rdb), orderRepo, userRepo, goodsRepo, db, rdb, &cfg.AliPay)
	paymentHandler := handler.NewPaymentHandler(paymentService, cfg.AliPay.AliPayKey)

	//job部分
	orderTimeout := job.NewOrderTimeoutJob(orderService, mq)
	orderCreate := job.NewOrderCreateConsumer(orderService, mq)
	cachePreheat := job.NewGoodsCacheWarmJob(goodsService, rdb, db)

	logger.Log.Info("初始化成功")

	return &App{
		userHandler:    userHandler,
		goodsHandler:   goodsHandler,
		orderHandler:   orderHandler,
		paymentHandler: paymentHandler,

		orderTimeout: orderTimeout,
		orderCreate:  orderCreate,
		cachePreheat: cachePreheat,

		Db:  db,
		Rdb: rdb,
		Mq:  mq,
	}, port
}
