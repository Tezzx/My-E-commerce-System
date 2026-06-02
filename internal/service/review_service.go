package service

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/types"
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepo
	orderRepo  *repository.OrderRepo
}

func NewReviewService(reviewRepo *repository.ReviewRepo, orderRepo *repository.OrderRepo) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		orderRepo:  orderRepo,
	}
}

func (s *ReviewService) Create(userID uint, req *types.ReviewReq) error {
	// 校验订单归属
	order, err := s.orderRepo.GetOrderByOrderNo(req.OrderNo)
	if err != nil || order.UserID != userID {
		return err
	}
	// 必须已支付才能评价
	if order.Status != 1 {
		return nil
	}
	// 防止重复评价
	exists, _ := s.reviewRepo.Exists(userID, req.GoodsID, req.OrderNo)
	if exists {
		return nil
	}
	return s.reviewRepo.Create(&model.Review{
		UserID:  userID,
		GoodsID: req.GoodsID,
		OrderNo: req.OrderNo,
		Rating:  req.Rating,
		Content: req.Content,
		IsAnon:  req.IsAnon,
	})
}

func (s *ReviewService) ListByGoods(goodsID uint, page, size int) ([]types.ReviewResp, int64, error) {
	reviews, total, err := s.reviewRepo.ListByGoods(goodsID, page, size)
	if err != nil {
		return nil, 0, err
	}
	var resp []types.ReviewResp
	for _, r := range reviews {
		username := "用户" // 简化：实际应从 userRepo 获取
		if r.IsAnon {
			username = "匿名用户"
		}
		resp = append(resp, types.ReviewResp{
			ID:       r.ID,
			UserID:   r.UserID,
			Username: username,
			GoodsID:  r.GoodsID,
			Rating:   r.Rating,
			Content:  r.Content,
			IsAnon:   r.IsAnon,
		})
	}
	return resp, total, nil
}

func (s *ReviewService) AvgRating(goodsID uint) (float64, error) {
	return s.reviewRepo.AvgRating(goodsID)
}
