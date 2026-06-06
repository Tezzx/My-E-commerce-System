package model

import "gorm.io/gorm"

// 评价审核状态
const (
	ReviewStatusPending  = 0 // 待审核
	ReviewStatusApproved = 1 // 已通过
	ReviewStatusRejected = 2 // 已驳回
)

type Review struct {
	gorm.Model
	UserID  uint   `gorm:"index;comment:用户ID"`
	GoodsID uint   `gorm:"index;comment:商品ID"`
	OrderNo string `gorm:"type:varchar(32);index;comment:关联订单号"`
	Rating  int    `gorm:"not null;comment:评分 1-5"`
	Content string `gorm:"type:text;comment:评价内容"`
	IsAnon  bool   `gorm:"default:false;comment:是否匿名"`
	Status  int    `gorm:"default:0;comment:审核状态 0-待审核 1-已通过 2-已驳回"`
}
