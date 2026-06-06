package repository

import (
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

// 取消订单(改变订单状态)
func (o *OrderRepo) CancelOrderStatus(orderNo string) error {
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, model.OrderStatusPending).
		Update("status", model.OrderStatusCancelled)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// --------------------辅助功能-------------------
// 根据订单编号查询订单
func (o *OrderRepo) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := o.db.Preload("OrderItems").Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 查询用户的所有订单
func (o *OrderRepo) GetUserOrderList(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := o.db.Preload("OrderItems").Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	return orders, err
}

// GetUserOrderListWithPage 分页查询用户订单
func (o *OrderRepo) GetUserOrderListWithPage(userID uint, page, size int, status *int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := o.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Preload("OrderItems").Order("created_at desc").Offset(offset).Limit(size).Find(&orders).Error
	return orders, total, err
}

func (o *OrderRepo) ReturnDB() *gorm.DB {
	return o.db
}

// UpdateStatus 通用状态流转（带乐观锁校验 oldStatus）
func (o *OrderRepo) UpdateStatus(orderNo string, oldStatus, newStatus int) error {
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, oldStatus).
		Update("status", newStatus)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateStatusWithShip 已支付 → 已发货，更新物流信息
func (o *OrderRepo) UpdateStatusWithShip(orderNo string, shipCompany, shipNo string) error {
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, model.OrderStatusPaid).
		Updates(map[string]interface{}{
			"status":       model.OrderStatusShipped,
			"ship_company": shipCompany,
			"ship_no":      shipNo,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (o *OrderRepo) ChangeStatusToPaid(orderNo string) error {
	now := time.Now()
	result := o.db.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":   model.OrderStatusPaid,
			"pay_time": now,
		})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ---------- OrderLog ----------

func (o *OrderRepo) CreateOrderLog(log *model.OrderLog) error {
	return o.db.Create(log).Error
}

func (o *OrderRepo) GetOrderLogs(orderNo string) ([]model.OrderLog, error) {
	var logs []model.OrderLog
	err := o.db.Where("order_no = ?", orderNo).Order("created_at asc").Find(&logs).Error
	return logs, err
}
