package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterUser 用户注册
// @Summary      用户注册
// @Description  创建新账号
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      types.RegisterRequest  true  "注册信息"
// @Success      200   {object}  response.Resp{data=string}
// @Failure      400   {object}  response.Resp
// @Router       /register [post]
func (u *UserHandler) RegisterUser(c *gin.Context) {
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warn("注册 - 参数绑定失败", zap.Error(err))
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userID, err := u.userService.RegisterUser(req.Username, req.Password)
	if err != nil {
		logger.Log.Warn("注册失败",
			zap.String("username", req.Username),
			zap.Error(err))
		response.Error(c, 400, 1, "新用户创建失败")
		return
	}

	token, err := u.userService.TokenCreate(userID)
	if err != nil {
		logger.Log.Error("注册成功但生成 token 失败",
			zap.Uint("user_id", userID),
			zap.Error(err))
		response.Error(c, 500, 1, "服务器无法生成token")
		return
	}

	logger.Log.Info("用户注册成功",
		zap.Uint("user_id", userID),
		zap.String("username", req.Username))
	response.Success(c, token)
}

// LoginUser 用户登录
// @Summary      用户登录
// @Description  登录获取 Token
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      types.LoginRequest  true  "登录信息"
// @Success      200   {object}  response.Resp{data=string}
// @Failure      400   {object}  response.Resp
// @Router       /login [post]
func (u *UserHandler) LoginUser(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warn("登录 - 参数绑定失败", zap.Error(err))
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userID, err := u.userService.LoginUser(req.Username, req.Password)
	if err != nil {
		logger.Log.Warn("登录失败",
			zap.String("username", req.Username))
		response.Error(c, 400, 1, "账户或密码错误")
		return
	}

	token, err := u.userService.TokenCreate(userID)
	if err != nil {
		logger.Log.Error("登录成功但生成 token 失败",
			zap.Uint("user_id", userID),
			zap.Error(err))
		response.Error(c, 500, 1, "服务器无法生成token")
		return
	}

	logger.Log.Info("用户登录成功",
		zap.Uint("user_id", userID),
		zap.String("username", req.Username))
	response.Success(c, token)
}

// GetAllUsers 获取所有用户
// @Summary      获取所有用户
// @Tags         用户
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /help/users [get]
func (u *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := u.userService.GetAllUsers()
	if err != nil {
		logger.Log.Error("获取用户列表失败", zap.Error(err))
		response.Error(c, 500, 1, "获取用户列表失败")
		return
	}

	response.Success(c, users)
}

// Me 获取当前登录用户信息（含角色）
// @Summary      当前用户信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Resp{data=types.UserInfo}
// @Security     BearerAuth
// @Router       /home/me [get]
func (u *UserHandler) Me(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := u.userService.GetCurrentUser(userID)
	if err != nil {
		response.Error(c, 404, 1, "用户不存在")
		return
	}
	response.Success(c, types.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
}
