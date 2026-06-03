package model

import (
	"time"

	"gorm.io/gorm"
)

// ---------- 订单状态常量 ----------
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付
	OrderStatusCancelled = 2 // 已取消
	OrderStatusShipped   = 3 // 已发货
	OrderStatusReceived  = 4 // 已收货
	OrderStatusCompleted = 5 // 已完成
	OrderStatusRefunding = 6 // 退款中
	OrderStatusRefunded  = 7 // 已退款
)

// OrderStatusNames 状态中文名映射
var OrderStatusNames = map[int]string{
	OrderStatusPending:   "待支付",
	OrderStatusPaid:      "已支付",
	OrderStatusCancelled: "已取消",
	OrderStatusShipped:   "已发货",
	OrderStatusReceived:  "已收货",
	OrderStatusCompleted: "已完成",
	OrderStatusRefunding: "退款中",
	OrderStatusRefunded:  "已退款",
}

// validTransitions 合法的状态流转
var validTransitions = map[int][]int{
	OrderStatusPending:   {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPaid:      {OrderStatusShipped, OrderStatusRefunding},
	OrderStatusShipped:   {OrderStatusReceived, OrderStatusRefunding},
	OrderStatusReceived:  {OrderStatusCompleted},
	OrderStatusRefunding: {OrderStatusRefunded, OrderStatusPaid}, // 退款成功 / 退款拒绝
}

// CanTransition 校验状态流转是否合法
func CanTransition(from, to int) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

type Order struct {
	gorm.Model

	OrderNo string `gorm:"type:varchar(32);uniqueIndex;comment:订单编号" json:"order_no"`
	UserID  uint   `gorm:"index;comment:用户ID" json:"user_id"`

	// 单商品（兼容旧逻辑），多商品时设为0
	GoodsID   uint   `gorm:"index;comment:商品ID" json:"goods_id"`
	GoodsName string `gorm:"type:varchar(50);comment:商品名称" json:"goods_name"`
	Price     uint   `gorm:"comment:商品单价(下单时)" json:"price"`
	BuyNum    uint   `gorm:"comment:购买数量" json:"buy_num"`

	TotalPrice uint       `gorm:"comment:订单总价" json:"total_price"`
	Status     int        `gorm:"default:0;comment:订单状态" json:"status"`
	PayTime    *time.Time `gorm:"comment:支付时间" json:"pay_time"`
	PayChannel string     `gorm:"type:varchar(20);comment:支付渠道(alipay/wechat)" json:"pay_channel"`
	TradeNo    string     `gorm:"type:varchar(100);comment:第三方交易流水号" json:"trade_no"`

	// 收货地址
	AddressID       *uint  `gorm:"comment:收货地址ID" json:"address_id"`
	AddressSnapshot string `gorm:"type:text;comment:收货地址快照JSON" json:"address_snapshot"`

	// 物流
	ShipCompany string `gorm:"type:varchar(30);comment:物流公司" json:"ship_company"`
	ShipNo      string `gorm:"type:varchar(50);comment:物流单号" json:"ship_no"`

	// 关联订单明细
	OrderItems []OrderItem `gorm:"foreignKey:OrderID;references:ID" json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint   `gorm:"index;comment:关联主订单ID" json:"order_id"`
	GoodsID   uint   `gorm:"index;comment:商品ID" json:"goods_id"`
	GoodsName string `gorm:"type:varchar(50);comment:商品名称" json:"goods_name"`
	Price     uint   `gorm:"comment:商品单价(下单时)" json:"price"`
	Quantity  uint   `gorm:"comment:购买数量" json:"quantity"`
}
