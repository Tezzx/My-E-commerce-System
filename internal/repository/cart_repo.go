package repository

import (
	"order-payment-system/internal/model"

	"gorm.io/gorm"
)

type CartRepo struct {
	db *gorm.DB
}

func NewCartRepo(db *gorm.DB) *CartRepo {
	return &CartRepo{db: db}
}

func (r *CartRepo) AddItem(item *model.CartItem) error {
	return r.db.Create(item).Error
}

func (r *CartRepo) UpdateItem(item *model.CartItem) error {
	return r.db.Save(item).Error
}

func (r *CartRepo) DeleteItem(userID uint, goodsID uint) error {
	return r.db.Where("user_id = ? AND goods_id = ?", userID, goodsID).Delete(&model.CartItem{}).Error
}

func (r *CartRepo) GetCartByUserID(userID uint) ([]*model.CartItem, error) {
	var items []*model.CartItem
	err := r.db.Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *CartRepo) GetItem(userID uint, goodsID uint) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.Where("user_id = ? AND goods_id = ?", userID, goodsID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // not found
		}
		return nil, err
	}
	return &item, nil
}
