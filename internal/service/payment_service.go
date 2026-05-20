package service

import (
	"context"
	"fmt"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/logger"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentService struct {
	orderRepo *repository.OrderRepo
	userRepo  *repository.UserRepo
	goodsRepo *repository.GoodsRepo
	db        *gorm.DB
	rdb       *redis.Client
}

var appID string
var privatekey string

func NewPaymentService(orderRepo *repository.OrderRepo, userRepo *repository.UserRepo, goodsRepo *repository.GoodsRepo, db *gorm.DB, rdb *redis.Client) *PaymentService {
	return &PaymentService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		goodsRepo: goodsRepo,
		db:        db,
		rdb:       rdb,
	}
}

var aliClient *alipay.Client

func InitialPay(appId, privateKey string) {
	isProd := false
	appID := appId
	privatekey := privateKey
	var err error
	aliClient, err = alipay.NewClient(appID, privatekey, isProd)
	if err != nil {
		logger.Log.Error("初始化支付宝客户端失败", zap.Error(err))
	}
}

func (p *PaymentService) GetOrder(orderNo string) (*model.Order, error) {
	order, err := p.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		logger.Log.Warn("支付 - 订单不存在", zap.String("order_no", orderNo))
	}
	return order, err
}

// GenerateAlipayUrl 调用支付宝SDK生成支付链接
func (p *PaymentService) GenerateAlipayUrl(order *model.Order) (string, error) {

	bm := make(gopay.BodyMap)
	bm.Set("subject", order.GoodsName)
	bm.Set("out_trade_no", order.OrderNo)
	bm.Set("product_code", "FAST_INSTANT_TRADE_PAY")
	amount := float64(order.TotalPrice)
	bm.Set("total_amount", fmt.Sprintf("%.2f", amount))
	bm.Set("notify_url", "https://公网域名/api/notify/alipay")
	bm.Set("return_url", "http://localhost:8080/pay/success")

	payUrl, err := aliClient.TradePagePay(context.Background(), bm)
	if err != nil {
		logger.Log.Error("生成支付宝链接失败", zap.Error(err))
		return "", err
	}

	return payUrl, nil
}

func (p *PaymentService) HandlePaySuccess(orderNo, tradeNo string) error {
	order, err := p.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return err
	}

	if order.Status == 1 {
		return nil
	}

	return p.db.Transaction(func(tx *gorm.DB) error {
		orderRepoTx := repository.NewOrderRepo(tx, p.rdb)

		err := tx.Model(&model.Order{}).Where("order_no = ?", orderNo).Updates(map[string]interface{}{
			"status":      1,
			"trade_no":    tradeNo,
			"pay_channel": "alipay",
		}).Error
		if err != nil {
			return err
		}

		_ = orderRepoTx.DelQueue(orderNo)

		logger.Log.Info("第三方支付成功并处理完成", zap.String("order_no", orderNo))
		return nil
	})
}
