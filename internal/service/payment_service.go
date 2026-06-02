package service

import (
	"context"
	"fmt"
	"order-payment-system/config"
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
	tm        *repository.TransactionManager
	orderRepo *repository.OrderRepo
	userRepo  *repository.UserRepo
	goodsRepo *repository.GoodsRepo
	db        *gorm.DB
	rdb       *redis.Client
	aliClient *alipay.Client
	notifyUrl string
}

func NewPaymentService(tm *repository.TransactionManager, orderRepo *repository.OrderRepo, userRepo *repository.UserRepo, goodsRepo *repository.GoodsRepo, db *gorm.DB, rdb *redis.Client, cfg *config.AliPay) *PaymentService {
	isProd := false
	aliClient, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, isProd)
	if err != nil {
		logger.Log.Error("初始化支付宝客户端失败", zap.Error(err))
	}

	return &PaymentService{
		tm:        tm,
		orderRepo: orderRepo,
		userRepo:  userRepo,
		goodsRepo: goodsRepo,
		db:        db,
		rdb:       rdb,
		aliClient: aliClient,
		notifyUrl: cfg.NotifyUrl,
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
	bm.Set("notify_url", p.notifyUrl)
	bm.Set("return_url", "http://localhost:8080/pay/success")

	payUrl, err := p.aliClient.TradePagePay(context.Background(), bm)
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

	return p.tm.Transaction(func(txManager *repository.TransactionManager) error {

		goodsRepoTx := txManager.GoodsRepo()

		res := txManager.OrderRepo().ReturnDB().Model(&model.Order{}).Where("order_no = ? AND status = ?", orderNo, 0).Updates(map[string]interface{}{
			"status":      1,
			"trade_no":    tradeNo,
			"pay_channel": "alipay",
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// 说明已经被其他并发请求处理，直接返回成功以保证幂等性
			return nil
		}

		// 扣减 MySQL 库存
		if len(order.OrderItems) > 0 {
			for _, item := range order.OrderItems {
				err = goodsRepoTx.DeductStockSQL(item.GoodsID, int64(item.Quantity))
				if err != nil {
					return err
				}
			}
		} else {
			// 兼容旧的单商品下单流程
			err = goodsRepoTx.DeductStockSQL(order.GoodsID, int64(order.BuyNum))
			if err != nil {
				return err
			}
		}

		logger.Log.Info("第三方支付成功并处理完成", zap.String("order_no", orderNo))
		return nil
	})
}
