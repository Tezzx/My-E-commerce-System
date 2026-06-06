package app

import (
	"order-payment-system/pkg/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(app *App) *gin.Engine {

	r := gin.Default()

	// HTML 页面
	r.LoadHTMLGlob("templates/*")

	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.RequestLogger())
	// 增加令牌桶限流
	//r.Use(middleware.RateLimit(100, 10, app.Rdb))

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/auth", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/pay", func(c *gin.Context) {
		c.HTML(200, "pay.html", nil)
	})

	r.POST("/login", app.userHandler.LoginUser)
	r.POST("/register", app.userHandler.RegisterUser)
	r.POST("/api/notify/alipay", app.paymentHandler.AlipayNotify)

	home := r.Group("/home")
	home.Use(middleware.TokenIdentify())
	{
		home.GET("/me", app.userHandler.Me)
		home.GET("/", app.goodsHandler.GetGoodsList)
		home.GET("/search/goods", app.goodsHandler.GetGoodsDetail)
		home.GET("/search/orders", app.orderHandler.GetUserOrders)
		order := home.Group("/order")
		{
			order.POST("/", app.orderHandler.CreateOrder)
			order.POST("/cart_checkout", app.orderHandler.CheckoutCart)
			order.GET("/cancel", app.orderHandler.CancelOrder)
			order.POST("/ship", middleware.MerchantOnly(app.UserRepo), app.orderHandler.ShipOrder) // 商家发货
			order.POST("/confirm", app.orderHandler.ConfirmReceipt)                                // 用户确认收货
			order.POST("/complete", app.orderHandler.CompleteOrder)                                // 订单完成
			order.GET("/logs", app.orderHandler.GetOrderLogs)                                      // 操作日志
			// 退款
			refund := order.Group("/refund")
			{
				refund.POST("/request", app.orderHandler.RequestRefund)                                        // 用户申请退款
				refund.POST("/process", middleware.MerchantOnly(app.UserRepo), app.orderHandler.ProcessRefund) // 商家/管理员审批退款
			}
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
			cart.PATCH("/select", app.cartHandler.ToggleSelect)
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
			category.POST("/", middleware.AdminOnly(app.UserRepo), app.categoryHandler.Create)
			category.PUT("/:id", middleware.AdminOnly(app.UserRepo), app.categoryHandler.Update)
			category.DELETE("/:id", middleware.AdminOnly(app.UserRepo), app.categoryHandler.Delete)
			category.GET("/tree", app.categoryHandler.GetTree)
		}
		// 商品管理（商家/管理员）
		goods := home.Group("/goods")
		{
			goods.POST("/", middleware.MerchantOnly(app.UserRepo), app.goodsHandler.CreateGoods)
			goods.PUT("/:id", middleware.MerchantOnly(app.UserRepo), app.goodsHandler.UpdateGoods)
			goods.DELETE("/:id", middleware.MerchantOnly(app.UserRepo), app.goodsHandler.DeleteGoods)
			goods.GET("/list", middleware.MerchantOnly(app.UserRepo), app.goodsHandler.GetGoodsPageList)
		}
		review := home.Group("/review")
		{
			review.POST("/", app.reviewHandler.Create)
			review.GET("/list", app.reviewHandler.List)
			// 管理员审核
			adminReview := review.Group("/")
			adminReview.Use(middleware.AdminOnly(app.UserRepo))
			{
				adminReview.GET("/pending", app.reviewHandler.ListPending)
				adminReview.POST("/:id/approve", app.reviewHandler.Approve)
				adminReview.POST("/:id/reject", app.reviewHandler.Reject)
			}
		}
	}
	//辅助功能
	r.GET("help/users", middleware.TokenIdentify(), middleware.AdminOnly(app.UserRepo), app.userHandler.GetAllUsers)

	return r
}
