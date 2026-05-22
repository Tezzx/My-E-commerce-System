package app

import (
	"order-payment-system/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *App) *gin.Engine {

	r := gin.Default()

	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.RequestLogger())
	// 增加令牌桶限流
	r.Use(middleware.RateLimit(100, 10, app.Rdb))
	//暂时忽略前端界面
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg":  "success",
			"data": "index",
		})
	})

	r.GET("/auth", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg":  "success",
			"data": "login",
		})
	})

	r.POST("/login", app.userHandler.LoginUser)
	r.POST("/register", app.userHandler.RegisterUser)
	r.POST("/api/notify/alipay", app.paymentHandler.AlipayNotify)

	home := r.Group("/home")
	home.Use(middleware.TokenIdentify())
	{
		home.GET("/", app.goodsHandler.GetGoodsList)
		home.GET("/search/goods", app.goodsHandler.GetGoodsDetail)
		home.GET("/search/orders", app.orderHandler.GetUserOrders)
		order := home.Group("/order")
		{
			order.POST("/", app.orderHandler.CreateOrder)
			order.GET("/cancel", app.orderHandler.CancelOrder)
		}
		pay := home.Group("/pay")
		{
			pay.POST("/create", app.paymentHandler.CreatePay)
		}
	}
	//辅助功能
	r.GET("help/users", app.userHandler.GetAllUsers)

	return r
}
