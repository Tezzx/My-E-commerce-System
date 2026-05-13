package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// 注册
func (u *UserHandler) RegisterUser(c *gin.Context) {
	var req types.RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	userID, err := u.userService.RegisterUser(req.Username, req.Password)
	if err != nil {
		response.Error(c, 400, "新用户创建失败")
		return
	}

	token, err := u.userService.TokenCreate(userID)
	if err != nil {
		response.Error(c, 500, "服务器无法生成token")
		return
	}
	response.Success(c, token)

}

// 登录
func (u *UserHandler) LoginUser(c *gin.Context) {
	var req types.LoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	userID, err := u.userService.LoginUser(req.Username, req.Password)
	if err != nil {
		response.Error(c, 400, "账户或密码错误")
		return
	}
	token, err := u.userService.TokenCreate(userID)
	if err != nil {
		response.Error(c, 500, "服务器无法生成token")
		return
	}
	response.Success(c, token)

}
func (u *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := u.userService.GetAllUsers()
	if err != nil {
		response.Error(c, 500, "获取用户列表失败")
		return
	}
	response.Success(c, users)
}
