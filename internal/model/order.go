package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model

	OrderNo   string `gorm:"type:varchar(32);uniqueIndex;comment:订单编号"`
	UserID    uint   `gorm:"index;comment:用户ID"`
	GoodsID   uint   `gorm:"index;comment:商品ID"`
	GoodsName string `gorm:"type:varchar(50);comment:商品名称"`
	Price     uint   `gorm:"comment:商品单价(下单时)"`

	BuyNum     uint       `gorm:"comment:购买数量"`
	TotalPrice uint       `gorm:"comment:订单总价"`
	Status     int        `gorm:"default:0;comment:订单状态 0-未支付 1-已支付 2-已取消"`
	PayTime    *time.Time `gorm:"comment:支付时间"`
	PayChannel string     `gorm:"type:varchar(20);comment:支付渠道(alipay/wechat)"`
	TradeNo    string     `gorm:"type:varchar(100);comment:第三方交易流水号"`
}
