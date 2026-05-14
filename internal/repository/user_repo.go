package repository

import (
	"errors"
	"order-payment-system/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepo struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewUserRepo(db *gorm.DB, rdb *redis.Client) *UserRepo {
	return &UserRepo{
		db:  db,
		rdb: rdb,
	}
}

// 创建用户
func (u *UserRepo) CreateUser(user *model.User) (uint, error) {
	err := u.db.Create(user).Error
	id, err := u.GetID(user.Username)
	return id, err
}

func (u *UserRepo) GetByUsername(username string) (string, error) {
	var user model.User
	err := u.db.Where("username=?", username).Select("password").First(&user).Error
	if err != nil {
		return "", err
	}
	return user.Password, nil
}
func (u *UserRepo) CheckUsernameExists(username string) (bool, error) {
	var count int64
	err := u.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (u *UserRepo) GetID(username string) (uint, error) {
	var userID uint
	err := u.db.Model(&model.User{}).
		Select("id").
		Where("username = ?", username).
		Scan(&userID).
		Error
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (u *UserRepo) Deduct(userId, balance uint) error {
	var user model.User
	err := u.db.Where("id=?", userId).Select("balance").First(&user).Error
	if err != nil {
		return err
	}
	if balance > user.Balance {
		return errors.New("余额不足")
	} else {
		newBalance := user.Balance - balance
		err = u.db.Model(&model.User{}).Where("id = ?", userId).Update("balance", newBalance).Error
		return err
	}
}

func (u *UserRepo) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := u.db.Find(&users).Error
	return users, err
}
