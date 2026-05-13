package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// 支付
func (p *PaymentHandler) Settle(c *gin.Context) {
	var req types.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "无订单号")
		return
	}
	order, err := p.paymentService.GetOrder(req.OrderNo)
	if err != nil {
		response.Error(c, 400, "订单号错误")
		return
	}
	if order.Status != 0 {
		response.Error(c, 400, "订单已支付/已取消，无需重复支付")
		return
	}
	err = p.paymentService.Settling(order)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Success(c, "支付成功，余额已扣减")
}
