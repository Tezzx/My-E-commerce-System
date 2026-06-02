package repository

import (
	"order-payment-system/internal/model"

	"gorm.io/gorm"
)

type AddressRepo struct {
	db *gorm.DB
}

func NewAddressRepo(db *gorm.DB) *AddressRepo {
	return &AddressRepo{db: db}
}

func (r *AddressRepo) Create(addr *model.Address) error {
	return r.db.Create(addr).Error
}

func (r *AddressRepo) Update(addr *model.Address) error {
	return r.db.Save(addr).Error
}

func (r *AddressRepo) Delete(userID, addrID uint) error {
	return r.db.Where("id = ? AND user_id = ?", addrID, userID).Delete(&model.Address{}).Error
}

func (r *AddressRepo) GetByID(userID, addrID uint) (*model.Address, error) {
	var addr model.Address
	err := r.db.Where("id = ? AND user_id = ?", addrID, userID).First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *AddressRepo) ListByUser(userID uint) ([]model.Address, error) {
	var addrs []model.Address
	err := r.db.Where("user_id = ?", userID).Order("is_default desc, updated_at desc").Find(&addrs).Error
	return addrs, err
}

func (r *AddressRepo) ClearDefault(userID uint) error {
	return r.db.Model(&model.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error
}
