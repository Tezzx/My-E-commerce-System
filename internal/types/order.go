package types

import "time"

// ----- 请求 ----

type OrderRequest struct {
	GoodsID   int   `json:"goodsId"`
	BuyNum    int   `json:"buyNum"`
	AddressID *uint `json:"address_id"` // 收货地址ID
}

type CartCheckoutReq struct {
	AddressID uint `json:"address_id" binding:"required"` // 必填
}

type ShipRequest struct {
	OrderNo     string `json:"order_no" binding:"required"`
	ShipCompany string `json:"ship_company" binding:"required"`
	ShipNo      string `json:"ship_no" binding:"required"`
}

// ----- 响应 ----

type OrderResp struct {
	ID              uint            `json:"id"`
	OrderNo         string          `json:"order_no"`
	UserID          uint            `json:"user_id"`
	TotalPrice      uint            `json:"total_price"`
	Status          int             `json:"status"`
	StatusName      string          `json:"status_name"`
	PayTime         *time.Time      `json:"pay_time,omitempty"`
	AddressSnapshot *AddressResp    `json:"address_snapshot,omitempty"`
	ShipCompany     string          `json:"ship_company,omitempty"`
	ShipNo          string          `json:"ship_no,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	Items           []OrderItemResp `json:"items,omitempty"`
}

type OrderItemResp struct {
	ID        uint   `json:"id"`
	GoodsID   uint   `json:"goods_id"`
	GoodsName string `json:"goods_name"`
	Price     uint   `json:"price"`
	Quantity  uint   `json:"quantity"`
}

type OrderLogResp struct {
	ID        uint      `json:"id"`
	OrderNo   string    `json:"order_no"`
	OldStatus int       `json:"old_status"`
	NewStatus int       `json:"new_status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}
