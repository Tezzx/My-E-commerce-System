package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/logger" // 导入 logger 包
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		logger.Log.Warn("创建订单 - 参数绑定失败", zap.Error(err))
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userIDany, exists := c.Get("userID")
	if !exists {
		logger.Log.Warn("创建订单 - 用户未登录")
		response.Error(c, 401, 1, "请先登录")
		return
	}

	userID, ok := userIDany.(uint)
	if !ok {
		logger.Log.Warn("创建订单 - 用户ID类型转换失败")
		response.Error(c, 500, 1, "用户ID类型错误")
		return
	}

	orderNo, err := o.orderService.CreateOrder(c.Request.Context(), userID, uint(req.GoodsID), uint(req.BuyNum))
	if err != nil {
		logger.Ctx(c.Request.Context()).Error("创建订单失败", zap.Uint("user_id", userID), zap.Uint("goods_id", uint(req.GoodsID)), zap.Uint("buy_num", uint(req.BuyNum)), zap.Error(err))
		response.Error(c, 500, 1, "订单创建失败")
		return
	}

	logger.Ctx(c.Request.Context()).Info("订单创建成功", zap.String("order_no", orderNo))
	response.Success(c, orderNo)
}

// 取消订单
func (o *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Query("orderNo")
	err := o.orderService.CancelOrder(orderNo)
	if err != nil {
		logger.Log.Error("取消订单失败", zap.String("order_no", orderNo), zap.Error(err))
		response.Error(c, 500, 1, "取消订单失败")
		return
	}

	logger.Log.Debug("订单取消成功", zap.String("order_no", orderNo)) // Debug 级别，可根据环境调整是否输出
	response.Success(c, "取消成功")
}

func (o *OrderHandler) GetUserOrders(c *gin.Context) {
	userIDany, exists := c.Get("userID")
	if !exists {
		logger.Log.Warn("获取用户订单 - 用户未登录")
		response.Error(c, 401, 1, "请先登录")
		return
	}

	userID, ok := userIDany.(uint)
	if !ok {
		logger.Log.Warn("获取用户订单 - 用户ID类型转换失败")
		response.Error(c, 500, 1, "用户ID类型错误")
		return
	}

	orders, err := o.orderService.GetUserOrderList(userID)
	if err != nil {
		logger.Log.Error("获取用户订单列表失败", zap.Uint("user_id", userID), zap.Error(err))
		response.Error(c, 500, 1, "获取订单失败")
		return
	}

	logger.Log.Debug("获取用户订单列表成功", zap.Uint("user_id", userID), zap.Int("orders_count", len(orders))) // 根据需要调整级别
	response.Success(c, orders)
}
