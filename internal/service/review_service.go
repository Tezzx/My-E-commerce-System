package service

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/logger"

	"go.uber.org/zap"
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepo
	orderRepo  *repository.OrderRepo
	userRepo   *repository.UserRepo
}

func NewReviewService(reviewRepo *repository.ReviewRepo, orderRepo *repository.OrderRepo, userRepo *repository.UserRepo) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		orderRepo:  orderRepo,
		userRepo:   userRepo,
	}
}

func (s *ReviewService) Create(userID uint, req *types.ReviewReq) error {
	// 校验订单归属
	order, err := s.orderRepo.GetOrderByOrderNo(req.OrderNo)
	if err != nil || order.UserID != userID {
		return err
	}
	// 必须已收货或已完成才能评价
	if order.Status != model.OrderStatusReceived && order.Status != model.OrderStatusCompleted {
		logger.Log.Warn("评价创建 - 订单状态不允许评价",
			zap.String("order_no", req.OrderNo),
			zap.Int("status", order.Status))
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
		Status:  model.ReviewStatusPending, // 默认待审核
	})
}

func (s *ReviewService) ListByGoods(goodsID uint, page, size int) ([]types.ReviewResp, int64, error) {
	reviews, total, err := s.reviewRepo.ListByGoods(goodsID, page, size)
	if err != nil {
		return nil, 0, err
	}

	return s.buildReviewResp(reviews), total, nil
}

// ListPending 管理员查看待审核评价
func (s *ReviewService) ListPending(page, size int) ([]types.ReviewResp, int64, error) {
	reviews, total, err := s.reviewRepo.ListPending(page, size)
	if err != nil {
		return nil, 0, err
	}
	return s.buildReviewResp(reviews), total, nil
}

// Approve 通过评价审核
func (s *ReviewService) Approve(reviewID uint) error {
	return s.reviewRepo.Approve(reviewID)
}

// Reject 驳回评价
func (s *ReviewService) Reject(reviewID uint) error {
	return s.reviewRepo.Reject(reviewID)
}

func (s *ReviewService) AvgRating(goodsID uint) (float64, error) {
	return s.reviewRepo.AvgRating(goodsID)
}

// buildReviewResp 组装评价响应（批量获取用户名）
func (s *ReviewService) buildReviewResp(reviews []model.Review) []types.ReviewResp {
	userIDs := make([]uint, 0, len(reviews))
	for _, r := range reviews {
		if !r.IsAnon {
			userIDs = append(userIDs, r.UserID)
		}
	}
	usernameMap, _ := s.userRepo.GetUsernamesByIDs(userIDs)

	var resp []types.ReviewResp
	for _, r := range reviews {
		username := "匿名用户"
		if !r.IsAnon {
			if name, ok := usernameMap[r.UserID]; ok {
				username = name
			} else {
				username = "用户"
			}
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
	return resp
}
