package service

import (
	"errors"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/util"

	"github.com/redis/go-redis/v9"
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

func (p *PaymentService) Settling(order *model.Order, userPassword string) error {
	// 先做状态检查
	if order.Status != 0 {
		return errors.New("订单已支付或失效")
	}

	// 开启事务
	return p.db.Transaction(func(tx *gorm.DB) error {
		// 创建使用事务 tx 的临时 repo 实例
		userRepoTx := repository.NewUserRepo(tx, p.rdb)
		orderRepoTx := repository.NewOrderRepo(tx, p.rdb)

		//验证密码
		password, err := p.userRepo.GetByUserID(order.UserID)
		if err != nil {
			return err
		}
		err = util.VerifyPassword(userPassword, password)
		if err != nil {
			return err
		}
		// 扣款
		err = userRepoTx.Deduct(order.UserID, order.TotalPrice)
		if err != nil {
			return err
		}

		// 更新状态
		err = orderRepoTx.ChangeStatusToPayed(order.OrderNo)
		if err != nil {
			return err
		}

		err = orderRepoTx.DelQueue(order.OrderNo)
		if err != nil {
			return err
		}
		return nil // 提交事务
	})
}
