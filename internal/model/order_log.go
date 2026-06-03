package model

import "gorm.io/gorm"

// OrderLog 订单状态变更日志
type OrderLog struct {
	gorm.Model
	OrderNo   string `gorm:"type:varchar(32);index;comment:订单编号" json:"order_no"`
	OldStatus int    `gorm:"comment:变更前状态" json:"old_status"`
	NewStatus int    `gorm:"comment:变更后状态" json:"new_status"`
	Operator  uint   `gorm:"comment:操作人用户ID,0表示系统" json:"operator"`
	Remark    string `gorm:"type:varchar(200);comment:备注" json:"remark"`
}
