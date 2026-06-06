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

// ListByGoods 只返回已通过审核的评价
func (r *ReviewRepo) ListByGoods(goodsID uint, page, size int) ([]model.Review, int64, error) {
	var total int64
	var reviews []model.Review
	query := r.db.Where("goods_id = ? AND status = ?", goodsID, model.ReviewStatusApproved)
	query.Count(&total)
	err := query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&reviews).Error
	return reviews, total, err
}

// ListPending 查询待审核的评价（管理员）
func (r *ReviewRepo) ListPending(page, size int) ([]model.Review, int64, error) {
	var total int64
	var reviews []model.Review
	query := r.db.Where("status = ?", model.ReviewStatusPending)
	query.Count(&total)
	err := query.Order("created_at asc").Offset((page - 1) * size).Limit(size).Find(&reviews).Error
	return reviews, total, err
}

// Approve 通过审核
func (r *ReviewRepo) Approve(reviewID uint) error {
	return r.db.Model(&model.Review{}).Where("id = ? AND status = ?", reviewID, model.ReviewStatusPending).
		Update("status", model.ReviewStatusApproved).Error
}

// Reject 驳回评价
func (r *ReviewRepo) Reject(reviewID uint) error {
	return r.db.Model(&model.Review{}).Where("id = ? AND status = ?", reviewID, model.ReviewStatusPending).
		Update("status", model.ReviewStatusRejected).Error
}

func (r *ReviewRepo) AvgRating(goodsID uint) (float64, error) {
	var avg float64
	err := r.db.Model(&model.Review{}).Where("goods_id = ? AND status = ?", goodsID, model.ReviewStatusApproved).
		Select("COALESCE(AVG(rating), 0)").Scan(&avg).Error
	return avg, err
}

func (r *ReviewRepo) Exists(userID, goodsID uint, orderNo string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Review{}).Where("user_id = ? AND goods_id = ? AND order_no = ?", userID, goodsID, orderNo).Count(&count).Error
	return count > 0, err
}
