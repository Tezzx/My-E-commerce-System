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
	alipayKey      string
}

func NewPaymentHandler(paymentService *service.PaymentService, alipayKey string) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		alipayKey:      alipayKey,
	}
}

// Create 提出支付请求
// @Summary      创建支付
// @Description  生成支付宝支付链接
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        orderNo  query     string  true  "订单号"
// @Success      200      {object}  response.Resp{data=string}
// @Failure      404      {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/pay/create [post]
func (p *PaymentHandler) CreatePay(c *gin.Context) {

	orderNo := c.Query("orderNo")
	order, err := p.paymentService.GetOrder(orderNo)
	if err != nil || order == nil {
		response.Error(c, 404, 1, "订单不存在")
		return
	}

	payUrl, err := p.paymentService.GenerateAlipayUrl(order)
	if err != nil {
		response.Error(c, 500, 1, "生成支付链接失败")
		return
	}

	response.Success(c, payUrl)
}

// @Summary      支付宝回调（内部接口）
// @Tags         支付
// @Accept       json
// @Produce      plain
// @Param        body  body  object  true  "支付宝回调参数"
// @Success      200   {string}  string  "success"
// @Router       /api/notify/alipay [post]
func (p *PaymentHandler) AlipayNotify(c *gin.Context) {

	bm, err := alipay.ParseNotifyToBodyMap(c.Request)
	if err != nil {
		response.Error(c, 500, 1, "支付错误")
		return
	}

	ok, err := alipay.VerifySign(p.alipayKey, bm)
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
