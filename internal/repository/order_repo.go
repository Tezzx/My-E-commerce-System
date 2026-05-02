package repository

import (
	"errors"
	"order-payment-system/internal/model"
	"time"

	"gorm.io/gorm"
)

// 订单数据访问层
type OrderRepo struct {
	db *gorm.DB
}

// 构造函数
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{
		db: db,
	}
}

// CreateOrder 创建订单
func (o *OrderRepo) CreateOrder(order *model.Order) error {
	err := o.db.Create(order).Error
	return err
}

// 根据订单ID查询订单
func (o *OrderRepo) GetOrderByID(orderID uint) (*model.Order, error) {
	var order model.Order
	err := o.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 根据订单编号查询订单
func (o *OrderRepo) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := o.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 更新订单支付状态（支付成功调用）
func (o *OrderRepo) UpdateOrderPayStatus(orderID uint) error {
	now := time.Now()
	err := o.db.Model(&model.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":   1,
		"pay_time": now,
	}).Error
	return err
}

// 查询用户的所有订单
func (o *OrderRepo) GetUserOrderList(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := o.db.Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	return orders, err
}

func (o *OrderRepo) ChangeStatus(orderId string) error {
	// 只更新状态为 0 的订单，避免重复更新
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderId, 0).
		Update("status", 1)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (o *OrderRepo) ChangeTime(orderId string) error {
	now := time.Now()
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? ", orderId).
		Update("pay_time", &now)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("order not found or already paid")
	}

	return nil
}
