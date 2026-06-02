package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name     string     `gorm:"type:varchar(50);not null;comment:分类名称"`
	ParentID *uint      `gorm:"index;comment:父分类ID,nil表示一级分类"`
	Sort     int        `gorm:"default:0;comment:排序权重"`
	Children []Category `gorm:"foreignKey:ParentID"`
}
