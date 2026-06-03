package model

import "gorm.io/gorm"

type Goods struct {
	gorm.Model
	Goodsname  string `gorm:"type:varchar(50);uniqueIndex" json:"goodsname"`
	Goodsnum   uint   `gorm:"type:int unsigned" json:"goodsnum"`
	Price      uint   `gorm:"type:int unsigned" json:"price"`
	CategoryID *uint  `gorm:"index;comment:所属分类ID" json:"category_id"`
	ImageURL   string `gorm:"type:varchar(255);comment:商品主图" json:"image_url"`
	Status     int    `gorm:"default:1;comment:状态 1-上架 0-下架" json:"status"`
	Sales      uint   `gorm:"default:0;comment:销量" json:"sales"`
}
