package types

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// UserInfo 当前用户信息响应
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
