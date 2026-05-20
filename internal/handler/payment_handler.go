package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay/alipay"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

var AlipayKey string

// Create 提出支付请求
func (p *PaymentHandler) CreatePay(c *gin.Context) {

	orderNo := c.Query("orderNo")
	order, _ := p.paymentService.GetOrder(orderNo)

	payUrl, err := p.paymentService.GenerateAlipayUrl(order)
	if err != nil {
		response.Error(c, 500, 1, "生成支付链接失败")
		return
	}

	response.Success(c, payUrl)
}

func (p *PaymentHandler) AlipayNotify(c *gin.Context) {

	bm, err := alipay.ParseNotifyToBodyMap(c.Request)
	if err != nil {
		response.Error(c, 500, 1, "支付错误")
		return
	}

	ok, err := alipay.VerifySign(AlipayKey, bm)
	if !ok || err != nil {
		logger.Log.Error("支付宝回调验签失败")
		c.String(200, "fail")
		return
	}

	tradeStatus := bm.GetString("trade_status")
	orderNo := bm.GetString("out_trade_no")
	tradeNo := bm.GetString("trade_no")

	if tradeStatus == "TRADE_SUCCESS" {
		err = p.paymentService.HandlePaySuccess(orderNo, tradeNo)
		if err != nil {
			c.String(200, "fail")
			return
		}
	}
	c.String(200, "success")
}
