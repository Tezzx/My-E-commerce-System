package types

type Goods struct {
	ID         uint   `json:"id"`
	GoodsName  string `json:"goodsname"`
	GoodsNum   uint   `json:"goodsnum"`
	Price      uint   `json:"price"`
	CategoryID *uint  `json:"category_id"`
	ImageURL   string `json:"image_url"`
	Sales      uint   `json:"sales"`
}

// CreateGoodsReq 创建商品请求
type CreateGoodsReq struct {
	GoodsName  string `json:"goodsname" binding:"required"`
	GoodsNum   uint   `json:"goodsnum" binding:"required"`
	Price      uint   `json:"price" binding:"required"`
	CategoryID *uint  `json:"category_id"`
	ImageURL   string `json:"image_url"`
}

// UpdateGoodsReq 更新商品请求（全部字段可选）
type UpdateGoodsReq struct {
	GoodsName  *string `json:"goodsname"`
	GoodsNum   *uint   `json:"goodsnum"`
	Price      *uint   `json:"price"`
	CategoryID *uint   `json:"category_id"`
	ImageURL   *string `json:"image_url"`
	Status     *int    `json:"status"`
}
