package types

type CategoryReq struct {
	Name     string `json:"name" binding:"required,max=50"`
	ParentID *uint  `json:"parent_id"` // nil=一级分类
	Sort     int    `json:"sort"`
}

type CategoryResp struct {
	ID       uint           `json:"id"`
	Name     string         `json:"name"`
	ParentID *uint          `json:"parent_id"`
	Sort     int            `json:"sort"`
	Children []CategoryResp `json:"children,omitempty"`
}
