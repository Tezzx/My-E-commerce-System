package repository

import (
	"order-payment-system/internal/model"

	"gorm.io/gorm"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(cat *model.Category) error {
	return r.db.Create(cat).Error
}

func (r *CategoryRepo) Update(cat *model.Category) error {
	return r.db.Save(cat).Error
}

func (r *CategoryRepo) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *CategoryRepo) GetByID(id uint) (*model.Category, error) {
	var cat model.Category
	err := r.db.First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// GetTree 递归加载完整分类树
func (r *CategoryRepo) GetTree() ([]model.Category, error) {
	var roots []model.Category
	// 查一级分类
	if err := r.db.Where("parent_id IS NULL").Order("sort asc").Find(&roots).Error; err != nil {
		return nil, err
	}
	for i := range roots {
		r.loadChildren(&roots[i])
	}
	return roots, nil
}

func (r *CategoryRepo) loadChildren(cat *model.Category) {
	r.db.Where("parent_id = ?", cat.ID).Order("sort asc").Find(&cat.Children)
	for i := range cat.Children {
		r.loadChildren(&cat.Children[i])
	}
}
