package main

/*
package main

import (
	"fmt"
	"log"
	"strconv"

	"order-payment-system/config"
	"order-payment-system/internal/handler"
	"order-payment-system/pkg/database"
	"order-payment-system/pkg/middleware"

	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/service"

	"github.com/gin-gonic/gin"
)

// App 包含所有需要的依赖
type App struct {
	UserHandler    *handler.UserHandler
	GoodsHandler   *handler.GoodsHandler
	OrderHandler   *handler.OrderHandler
	PaymentHandler *handler.PaymentHandler
}

func main() {

	app,port := initializeApp()

	r:=setupRoutes(app)

	r.Run(port)

}

// initializeApp 初始化所有依赖并返回 App
func initializeApp() (*App,string) {
	// 读取配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("配置文件读取失败: %v", err)
	}
	port := strconv.Itoa(cfg.Server.Port)

	// 连接数据库
	db, err := database.InitMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")

	// 自动建表
	if err := db.AutoMigrate(&model.User{}, &model.Goods{}, &model.Order{}); err != nil {
		log.Fatalf("数据表创建失败: %v", err)
	}

	// 依赖注入
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

	// 商品数据库初始化
	goodsHandler.GoodsInitial()

	return &App{
		UserHandler:    userHandler,
		GoodsHandler:   goodsHandler,
		OrderHandler:   orderHandler,
		PaymentHandler: paymentHandler,

	},port
}

// setupRoutes 注册所有路由
func setupRoutes(app *App) *gin.Engine {

	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CorsMiddleware())

	// 首页路由
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

	// 用户路由
	r.POST("/register", app.UserHandler.RegisterUser)
	r.POST("/login", app.UserHandler.LoginUser)

	// 订单路由
	order := r.Group("/order")
	{
		order.POST("/create", middleware.TokenIdentify(), app.OrderHandler.CreateOrder)
		order.GET("/topay", app.OrderHandler.ToPay)
	}

	// 支付路由
	payment := r.Group("/payment")
	{
		payment.POST("/ensure", middleware.TokenIdentify(), app.PaymentHandler.MakeSure)
		payment.POST("/settle", middleware.TokenIdentify(), app.PaymentHandler.Settle)
	}

	return r
}
*/
