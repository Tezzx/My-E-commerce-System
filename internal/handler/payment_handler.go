package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// Settle 处理支付请求
func (p *PaymentHandler) Settle(c *gin.Context) {
	var req types.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warn("支付请求 - 参数绑定失败", zap.Error(err))
		response.Error(c, 400, 1, "无订单号")
		return
	}

	order, err := p.paymentService.GetOrder(req.OrderNo)
	if err != nil {
		logger.Log.Warn("支付请求 - 订单不存在",
			zap.String("order_no", req.OrderNo))
		response.Error(c, 400, 1, "订单号错误")
		return
	}

	if order.Status != 0 {
		logger.Log.Info("支付请求 - 订单状态异常，拒绝重复支付",
			zap.String("order_no", req.OrderNo),
			zap.Int("status", order.Status))
		response.Error(c, 400, 1, "订单已支付/已取消，无需重复支付")
		return
	}

	logger.Log.Info("开始执行支付",
		zap.String("order_no", req.OrderNo))

	err = p.paymentService.Settling(order, req.Password)
	if err != nil {
		logger.Log.Error("支付失败",
			zap.String("order_no", req.OrderNo),
			zap.Error(err))
		response.Error(c, 400, 1, err.Error())
		return
	}

	logger.Log.Info("支付成功",
		zap.String("order_no", req.OrderNo))
	response.Success(c, "支付成功，余额已扣减")
}
