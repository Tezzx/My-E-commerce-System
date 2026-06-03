package types

type AddCartReq struct {
	GoodsID  uint `json:"goods_id" binding:"required"`
	Quantity uint `json:"quantity" binding:"required,min=1"`
}

type UpdateCartReq struct {
	GoodsID  uint `json:"goods_id" binding:"required"`
	Quantity uint `json:"quantity" binding:"required,min=1"`
}

type DeleteCartReq struct {
	GoodsID uint `json:"goods_id" binding:"required"`
}

type ToggleSelectReq struct {
	GoodsID  uint `json:"goods_id" binding:"required"`
	Selected bool `json:"selected"`
}

type CartItemResp struct {
	ID        uint   `json:"id"`
	GoodsID   uint   `json:"goods_id"`
	GoodsName string `json:"goods_name"`
	Price     uint   `json:"price"`
	Quantity  uint   `json:"quantity"`
	Selected  bool   `json:"selected"`
}

type CartListResp struct {
	Items      []CartItemResp `json:"items"`
	TotalPrice uint           `json:"total_price"` // 仅计算selected的
}
