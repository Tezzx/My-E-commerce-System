package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"

	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// 创建订单
func (o *OrderHandler) CreateOrder(c *gin.Context) {

	var req types.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	userIDany, bol := c.Get("userID")

	if !bol {
		c.JSON(200, gin.H{"code": 401, "msg": "请先登录"})
		return
	}
	userID, ok := userIDany.(uint)
	if !ok {
		c.JSON(200, gin.H{"code": 401, "msg": "登录信息无效"})
		return
	}

	orderNo, err := o.orderService.CreateOrder(userID, uint(req.GoodsID), uint(req.BuyNum))
	if err != nil {
		response.Error(c, 500, "订单创建失败")
		return
	}

	response.Success(c, orderNo)
}

// 取消订单
func (o *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Query("orderNo")
	o.orderService.CancelOrder(orderNo)
	response.Success(c, "取消成功")
}

func (o *OrderHandler) GetUserOrders(c *gin.Context) {
	userIDany, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, "请先登录")
		return
	}

	userID, ok := userIDany.(uint)
	if !ok {
		response.Error(c, 500, "用户ID类型错误")
		return
	}

	orders, err := o.orderService.GetUserOrderList(userID)
	if err != nil {
		response.Error(c, 500, "获取订单失败")
		return
	}

	response.Success(c, orders)
}
