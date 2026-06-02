package model

import "gorm.io/gorm"

type Address struct {
	gorm.Model
	UserID    uint   `gorm:"index;comment:用户ID"`
	Receiver  string `gorm:"type:varchar(50);not null;comment:收件人姓名"`
	Phone     string `gorm:"type:varchar(20);not null;comment:联系电话"`
	Province  string `gorm:"type:varchar(20);comment:省份"`
	City      string `gorm:"type:varchar(20);comment:城市"`
	District  string `gorm:"type:varchar(20);comment:区/县"`
	Detail    string `gorm:"type:varchar(200);not null;comment:详细地址"`
	IsDefault bool   `gorm:"default:false;comment:是否默认地址"`
}
