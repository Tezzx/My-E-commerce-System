package types

type ReviewReq struct {
	GoodsID uint   `json:"goods_id" binding:"required"`
	OrderNo string `json:"order_no" binding:"required"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"max=500"`
	IsAnon  bool   `json:"is_anon"`
}

type ReviewResp struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"` // 匿名时显示 "匿名用户"
	GoodsID  uint   `json:"goods_id"`
	Rating   int    `json:"rating"`
	Content  string `json:"content"`
	IsAnon   bool   `json:"is_anon"`
}
