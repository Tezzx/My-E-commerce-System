package model

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID  uint   `gorm:"index;comment:用户ID"`
	GoodsID uint   `gorm:"index;comment:商品ID"`
	OrderNo string `gorm:"type:varchar(32);index;comment:关联订单号"`
	Rating  int    `gorm:"not null;comment:评分 1-5"`
	Content string `gorm:"type:text;comment:评价内容"`
	IsAnon  bool   `gorm:"default:false;comment:是否匿名"`
}
