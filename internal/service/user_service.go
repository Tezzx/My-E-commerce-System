package service

import (
	"order-payment-system/internal/errs"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/jwt"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/util"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (u *UserService) RegisterUser(userName, userPassword string) (uint, error) {
	exists, err := u.userRepo.CheckUsernameExists(userName)
	if err != nil {
		logger.Log.Error("注册 - 检查用户名是否存在失败",
			zap.String("username", userName),
			zap.Error(err))
		return 0, errs.UnknowError
	}
	if exists {
		logger.Log.Warn("注册 - 用户名已存在",
			zap.String("username", userName))
		return 0, errs.UserExists
	}

	hashedPassword, err := util.HashPassword(userPassword)
	if err != nil {
		logger.Log.Error("注册 - 密码哈希失败",
			zap.String("username", userName),
			zap.Error(err))
		return 0, errs.UnknowError
	}

	user := model.User{
		Username: userName,
		Password: hashedPassword,
		Balance:  100000,
	}

	id, err := u.userRepo.CreateUser(&user)
	if err != nil {
		logger.Log.Error("注册 - 创建用户失败",
			zap.String("username", userName),
			zap.Error(err))
		return 0, errs.UnknowError
	}

	logger.Log.Info("用户注册成功",
		zap.Uint("user_id", id),
		zap.String("username", userName))
	return id, nil
}

func (u *UserService) LoginUser(userName, userPassword string) (uint, error) {
	storedPassword, err := u.userRepo.GetByUsername(userName)
	if err != nil {
		logger.Log.Warn("登录 - 用户不存在",
			zap.String("username", userName))
		return 0, errs.UserNotFound
	}

	userID, err := u.userRepo.GetID(userName)
	if err != nil {
		logger.Log.Error("登录 - 获取用户ID失败",
			zap.String("username", userName),
			zap.Error(err))
		return 0, errs.UnknowError
	}

	err = util.VerifyPassword(userPassword, storedPassword)
	if err != nil {
		logger.Log.Warn("登录 - 密码验证失败",
			zap.String("username", userName))
		return 0, err
	}

	logger.Log.Info("用户登录成功",
		zap.Uint("user_id", userID),
		zap.String("username", userName))
	return userID, nil
}

func (u *UserService) TokenCreate(userID uint) (string, error) {
	token, err := jwt.GenerateJWT(userID)
	if err != nil {
		logger.Log.Error("生成Token失败",
			zap.Uint("user_id", userID),
			zap.Error(err))
	}
	return token, err
}

func (u *UserService) GetAllUsers() ([]model.User, error) {
	users, err := u.userRepo.GetAllUsers()
	if err != nil {
		logger.Log.Error("获取所有用户失败", zap.Error(err))
	}
	return users, err
}
