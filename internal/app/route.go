package app

import (
	"order-payment-system/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *App) *gin.Engine {

	r := gin.Default()

	r.Use(middleware.CorsMiddleware())
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

	r.POST("/login", app.UserHandler.LoginUser)
	r.POST("/register", app.UserHandler.RegisterUser)

	home := r.Group("/home")
	home.Use(middleware.TokenIdentify())
	{
		home.GET("/", app.GoodsHandler.GetGoodsList)
		home.GET("/search/goods", app.GoodsHandler.GetGoodsDetail)
		home.GET("/search/orders", app.OrderHandler.GetUserOrders)
		order := home.Group("/order")
		{
			order.POST("/", app.OrderHandler.CreateOrder)
			order.POST("/cancel", app.OrderHandler.CancelOrder)
		}
		pay := home.Group("/pay")
		{
			pay.POST("/", app.PaymentHandler.Settle)
		}
	}
	//辅助功能
	r.GET("help/users", app.UserHandler.GetAllUsers)

	return r
}
