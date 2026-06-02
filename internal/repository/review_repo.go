package repository

import (
	"order-payment-system/internal/model"

	"gorm.io/gorm"
)

type ReviewRepo struct {
	db *gorm.DB
}

func NewReviewRepo(db *gorm.DB) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) Create(review *model.Review) error {
	return r.db.Create(review).Error
}

func (r *ReviewRepo) ListByGoods(goodsID uint, page, size int) ([]model.Review, int64, error) {
	var total int64
	var reviews []model.Review
	query := r.db.Where("goods_id = ?", goodsID)
	query.Count(&total)
	err := query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&reviews).Error
	return reviews, total, err
}

func (r *ReviewRepo) AvgRating(goodsID uint) (float64, error) {
	var avg float64
	err := r.db.Model(&model.Review{}).Where("goods_id = ?", goodsID).Select("COALESCE(AVG(rating), 0)").Scan(&avg).Error
	return avg, err
}

func (r *ReviewRepo) Exists(userID, goodsID uint, orderNo string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Review{}).Where("user_id = ? AND goods_id = ? AND order_no = ?", userID, goodsID, orderNo).Count(&count).Error
	return count > 0, err
}
