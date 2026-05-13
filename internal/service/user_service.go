package service

import (
	"order-payment-system/internal/errs"
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/jwt"
	"order-payment-system/pkg/util"
)

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// 注册
func (u *UserService) RegisterUser(userName, userPassword string) (uint, error) {
	bol, err := u.userRepo.CheckUsernameExists(userName)
	if err != nil || bol == true {
		return 0, errs.UserExists
	}
	userPassword, err = util.HashPassword(userPassword)
	if err != nil {
		return 0, errs.UnknowError
	}
	user := model.User{
		Username: userName,
		Password: userPassword,
		Balance:  100000,
	}
	id, err := u.userRepo.CreateUser(&user)
	if err != nil {
		return 0, errs.UnknowError
	}
	return id, nil
}

// 登录
func (u *UserService) LoginUser(userName, userPassword string) (uint, error) {
	password, err := u.userRepo.GetByUsername(userName)
	if err != nil {
		return 0, errs.UserNotFound
	}
	userID, err := u.userRepo.GetID(userName)
	if err != nil {
		return 0, err
	}
	return userID, util.VerifyPassword(userPassword, password)
}

func (u *UserService) TokenCreate(userID uint) (string, error) {
	token, err := jwt.GenerateJWT(userID)
	return token, err
}

func (u *UserService) GetAllUsers() ([]model.User, error) {
	return u.userRepo.GetAllUsers()
}
