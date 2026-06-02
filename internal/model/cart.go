package model

import "gorm.io/gorm"

// CartItem 购物车项模型
type CartItem struct {
	gorm.Model
	UserID   uint `gorm:"index;comment:用户ID"`
	GoodsID  uint `gorm:"index;comment:商品ID"`
	Quantity uint `gorm:"type:int unsigned;default:1;comment:商品数量"`
	Selected bool `gorm:"default:true;comment:是否勾选"`
}
