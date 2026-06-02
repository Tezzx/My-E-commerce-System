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
	//r.Use(middleware.RateLimit(100, 10, app.Rdb))
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
			order.POST("/cart_checkout", app.orderHandler.CheckoutCart)
			order.GET("/cancel", app.orderHandler.CancelOrder)
		}
		pay := home.Group("/pay")
		{
			pay.POST("/create", app.paymentHandler.CreatePay)
		}
		cart := home.Group("/cart")
		{
			cart.POST("/add", app.cartHandler.AddToCart)
			cart.PUT("/update", app.cartHandler.UpdateCart)
			cart.DELETE("/delete", app.cartHandler.DeleteCart)
			cart.GET("/list", app.cartHandler.GetCartList)
		}
		addr := home.Group("/address")
		{
			addr.POST("/", app.addressHandler.Create)
			addr.PUT("/:id", app.addressHandler.Update)
			addr.DELETE("/:id", app.addressHandler.Delete)
			addr.GET("/list", app.addressHandler.List)
		}
		category := home.Group("/category")
		{
			category.POST("/", app.categoryHandler.Create)
			category.PUT("/:id", app.categoryHandler.Update)
			category.DELETE("/:id", app.categoryHandler.Delete)
			category.GET("/tree", app.categoryHandler.GetTree)
		}
		review := home.Group("/review")
		{
			review.POST("/", app.reviewHandler.Create)
			review.GET("/list", app.reviewHandler.List)
		}
	}
	//辅助功能
	r.GET("help/users", app.userHandler.GetAllUsers)

	return r
}
