package repository

import (
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

func (u *UserRepo) GetByUserID(userid uint) (string, error) {
	var user model.User
	err := u.db.Where("id=?", userid).Select("password").First(&user).Error
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

func (u *UserRepo) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := u.db.Find(&users).Error
	return users, err
}

// GetRoleByID 获取用户角色
func (u *UserRepo) GetRoleByID(userID uint) (string, error) {
	var role string
	err := u.db.Model(&model.User{}).Select("role").Where("id = ?", userID).Scan(&role).Error
	return role, err
}

// GetUsernamesByIDs 批量获取用户名，返回 map[userID]username
func (u *UserRepo) GetUsernamesByIDs(ids []uint) (map[uint]string, error) {
	if len(ids) == 0 {
		return map[uint]string{}, nil
	}
	var users []model.User
	err := u.db.Select("id, username").Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(users))
	for _, user := range users {
		result[user.ID] = user.Username
	}
	return result, nil
}
