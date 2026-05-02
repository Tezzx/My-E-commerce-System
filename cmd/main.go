package main

import (
	"fmt"
	"order-payment-system/config"
	"order-payment-system/internal/handler"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/service"
	"order-payment-system/pkg/database"
	"order-payment-system/pkg/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	//读取配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("配置文件读取失败")
		return
	}
	port := strconv.Itoa(cfg.Server.Port)

	//连接数据库
	db, err := database.InitMySQL(&cfg.Database)
	if err != nil {
		fmt.Println("数据库连接失败:", err)
		return
	}
	fmt.Println("数据库连接成功")

	//自动建表
	err = db.AutoMigrate(&model.User{}, &model.Goods{}, &model.Order{})
	if err != nil {
		fmt.Println("数据表创建失败:", err)
		return
	}

	//依赖注入
	userRepo := repository.NewUserRepo(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	goodsRepo := repository.NewGoodsRepo(db)
	goodsService := service.NewGoodsService(goodsRepo)
	goodsHandler := handler.NewGoodsHandler(goodsService)

	orderRepo := repository.NewOrderRepo(db)
	orderService := service.NewOrderService(orderRepo, goodsRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	paymentService := service.NewPaymentService(orderRepo, userRepo, goodsRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	//商品数据库初始化
	goodsHandler.GoodsInitial()

	r := gin.Default()

	r.Use(middleware.CorsMiddleware())
	//路径是根据工作路径来算而不是main.go所在路径
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		type GoodsItem struct {
			ID        string
			GoodsName string
			Price     uint
		}

		goodsList := []GoodsItem{
			{ID: "1", GoodsName: "雪影娃娃", Price: 1200},
			{ID: "2", GoodsName: "恶魔狼", Price: 600},
			{ID: "3", GoodsName: "治愈兔", Price: 1800},
			{ID: "4", GoodsName: "月牙雪熊", Price: 1800},
		}

		c.HTML(200, "index.html", gin.H{
			"goodsList": goodsList,
		})
	})

	r.GET("/auth", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})

	r.POST("/register", userHandler.RegisterUser)
	r.POST("/login", userHandler.LoginUser)

	order := r.Group("/order")
	{
		order.POST("/create", middleware.TokenIdentify(), orderHandler.CreateOrder)

		order.GET("/topay", orderHandler.ToPay)

	}
	payment := r.Group("/payment")
	{
		payment.POST("/ensure", middleware.TokenIdentify(), paymentHandler.MakeSure)

		payment.POST("/settle", middleware.TokenIdentify(), paymentHandler.Settle)
	}
	r.Run(":" + port)
}
