package service

import (
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
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
	return p.orderRepo.GetOrderByOrderNo(orderNo)
}

func (p *PaymentService) Settling(order *model.Order) error {
	// 先做状态检查
	if order.Status != 0 {
		return errors.New("订单已支付或失效")
	}

	// 开启事务
	return p.db.Transaction(func(tx *gorm.DB) error {
		// 创建使用事务 tx 的临时 repo 实例
		var x *amqp.Connection
		userRepoTx := repository.NewUserRepo(tx, p.rdb)
		orderRepoTx := repository.NewOrderRepo(tx, p.rdb, x)

		// 扣款
		err := userRepoTx.Deduct(order.UserID, order.TotalPrice)
		if err != nil {
			return err // 自动回滚
		}

		// 更新状态
		err = orderRepoTx.ChangeStatusToPayed(order.OrderNo)
		if err != nil {
			return err
		}

		// 更新时间
		err = orderRepoTx.ChangeTime(order.OrderNo)
		if err != nil {
			return err
		}

		return nil // 提交事务
	})
}
