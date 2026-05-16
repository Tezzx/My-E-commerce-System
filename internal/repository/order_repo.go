package repository

import (
	"context"
	"order-payment-system/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 订单数据访问层
type OrderRepo struct {
	db  *gorm.DB
	rdb *redis.Client
}

// 构造函数
func NewOrderRepo(db *gorm.DB, rdb *redis.Client) *OrderRepo {
	return &OrderRepo{
		db:  db,
		rdb: rdb,
	}
}

//----------------业务---------------------

// CreateOrder 创建订单
func (o *OrderRepo) SaveOrder(order *model.Order) error {
	err := o.db.Create(order).Error
	return err
}

// 超时队列
func (o *OrderRepo) AddQueue(orderNo string) error {

	timeout := 30 * time.Minute
	expireAt := time.Now().Add(timeout).Unix()

	ctx := context.Background()
	if err := o.rdb.ZAdd(ctx, "order:timeout:queue", redis.Z{
		Score:  float64(expireAt),
		Member: orderNo,
	}).Err(); err != nil {
		return err
	}
	return nil
}

// 从超时队列删除
func (o *OrderRepo) DelQueue(orderNo string) error {
	ctx := context.Background()
	o.rdb.ZRem(ctx, "order:timeout:queue", orderNo)
	return nil
}

// 取消订单(改变订单状态)
func (o *OrderRepo) CancelOrderStatus(orderNo string) error {
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, 0).
		Update("status", 2)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// --------------------辅助功能-------------------
// 根据订单编号查询订单
func (o *OrderRepo) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := o.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 查询用户的所有订单
func (o *OrderRepo) GetUserOrderList(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := o.db.Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	return orders, err
}

func (o *OrderRepo) ReturnDB() *gorm.DB {
	return o.db
}

func (o *OrderRepo) ChangeStatusToPayed(orderNo string) error {
	now := time.Now()
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, 0). // 仅允许待支付订单
		Updates(map[string]interface{}{
			"status":   1,
			"pay_time": now,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
