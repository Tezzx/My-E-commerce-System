package types

type Goods struct {
	ID        uint   `json:"id"`
	GoodsName string `json:"goodsname"`
	GoodsNum  uint   `json:"goodsnum"`
	Price     uint   `json:"price"`
}
