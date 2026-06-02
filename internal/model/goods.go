package model

import "gorm.io/gorm"

type Goods struct {
	gorm.Model
	Goodsname  string `gorm:"type:varchar(50);uniqueIndex"`
	Goodsnum   uint   `gorm:"type:int unsigned"`
	Price      uint   `gorm:"type:int unsigned"`
	CategoryID *uint  `gorm:"index;comment:所属分类ID"`
	ImageURL   string `gorm:"type:varchar(255);comment:商品主图"`
	Status     int    `gorm:"default:1;comment:状态 1-上架 0-下架"`
	Sales      uint   `gorm:"default:0;comment:销量"`
}
