package service

import (
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/util"

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

func NewPaymentService(orderRepo *repository.OrderRepo, userRepo *repository.UserRepo, goodsRepo *repository.GoodsRepo, db *gorm.DB) *PaymentService {
	return &PaymentService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		goodsRepo: goodsRepo,
		db:        db,
	}
}

func (p *PaymentService) GetOrder(orderNo string) (*model.Order, error) {
	order, err := p.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		logger.Log.Warn("支付 - 订单不存在", zap.String("order_no", orderNo))
	}
	return order, err
}

func (p *PaymentService) Settling(order *model.Order, userPassword string) error {
	if order.Status != 0 {
		logger.Log.Warn("支付 - 订单状态异常，拒绝支付",
			zap.String("order_no", order.OrderNo),
			zap.Int("status", order.Status))
		return errors.New("订单已支付或失效")
	}

	return p.db.Transaction(func(tx *gorm.DB) error {
		userRepoTx := repository.NewUserRepo(tx, p.rdb)
		orderRepoTx := repository.NewOrderRepo(tx, p.rdb)

		password, err := userRepoTx.GetByUserID(order.UserID)
		if err != nil {
			logger.Log.Error("支付 - 获取用户密码失败",
				zap.Uint("user_id", order.UserID),
				zap.String("order_no", order.OrderNo),
				zap.Error(err))
			return err
		}

		err = util.VerifyPassword(userPassword, password)
		if err != nil {
			logger.Log.Warn("支付 - 密码验证失败",
				zap.Uint("user_id", order.UserID),
				zap.String("order_no", order.OrderNo))
			return err
		}

		err = userRepoTx.Deduct(order.UserID, order.TotalPrice)
		if err != nil {
			logger.Log.Error("支付 - 扣款失败",
				zap.Uint("user_id", order.UserID),
				zap.String("order_no", order.OrderNo),
				zap.Uint("amount", order.TotalPrice),
				zap.Error(err))
			return err
		}

		err = orderRepoTx.ChangeStatusToPayed(order.OrderNo)
		if err != nil {
			logger.Log.Error("支付 - 更新订单状态失败",
				zap.String("order_no", order.OrderNo),
				zap.Error(err))
			return err
		}

		err = orderRepoTx.DelQueue(order.OrderNo)
		if err != nil {
			logger.Log.Warn("支付 - 从超时队列删除失败（可容忍）",
				zap.String("order_no", order.OrderNo),
				zap.Error(err))
		}

		logger.Log.Info("支付成功",
			zap.String("order_no", order.OrderNo),
			zap.Uint("user_id", order.UserID),
			zap.Uint("amount", order.TotalPrice))

		return nil
	})
}
