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
