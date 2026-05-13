package types

type OrderRequest struct {
	GoodsID int `json:"goodsId"`
	BuyNum  int `json:"buyNum"`
}
