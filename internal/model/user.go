package model

import "gorm.io/gorm"

const (
	RoleUser     = "user"
	RoleMerchant = "merchant"
	RoleAdmin    = "admin"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);uniqueIndex"`
	Password string `gorm:"type:varchar(255)"`
	Role     string `gorm:"type:varchar(20);default:user;comment:角色(user/merchant/admin)"`
}
