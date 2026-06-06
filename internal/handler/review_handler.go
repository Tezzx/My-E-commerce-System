package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// @Summary      创建评价
// @Tags         评价
// @Accept       json
// @Produce      json
// @Param        body  body      types.ReviewReq  true  "评价内容"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/review [post]
func (h *ReviewHandler) Create(c *gin.Context) {
	var req types.ReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	if err := h.svc.Create(userID, &req); err != nil {
		logger.Log.Error("创建评价失败", zap.Error(err))
		response.Error(c, 500, 1, "评价失败")
		return
	}
	response.Success(c, "评价成功，等待审核")
}

// @Summary      评价列表
// @Tags         评价
// @Accept       json
// @Produce      json
// @Param        goods_id  query     int  true  "商品ID"
// @Param        page      query     int  false "页码"  default(1)
// @Param        size      query     int  false "每页数量"  default(10)
// @Success      200       {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/review/list [get]
func (h *ReviewHandler) List(c *gin.Context) {
	goodsIDStr := c.Query("goods_id")
	goodsID, err := strconv.ParseUint(goodsIDStr, 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	reviews, total, err := h.svc.ListByGoods(uint(goodsID), page, size)
	if err != nil {
		logger.Log.Error("获取评价列表失败", zap.Error(err))
		response.Error(c, 500, 1, "获取失败")
		return
	}
	response.Success(c, gin.H{
		"list":  reviews,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ==================== 管理员审核 ====================

// @Summary      待审核评价列表（管理员）
// @Tags         评价管理
// @Accept       json
// @Produce      json
// @Param        page  query     int  false "页码"  default(1)
// @Param        size  query     int  false "每页数量"  default(10)
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/review/pending [get]
func (h *ReviewHandler) ListPending(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	reviews, total, err := h.svc.ListPending(page, size)
	if err != nil {
		logger.Log.Error("获取待审核评价失败", zap.Error(err))
		response.Error(c, 500, 1, "获取失败")
		return
	}
	response.Success(c, gin.H{
		"list":  reviews,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// @Summary      通过评价审核（管理员）
// @Tags         评价管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "评价ID"
// @Success      200  {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/review/{id}/approve [post]
func (h *ReviewHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	if err := h.svc.Approve(uint(id)); err != nil {
		logger.Log.Error("通过评价审核失败", zap.Error(err))
		response.Error(c, 500, 1, "操作失败")
		return
	}
	response.Success(c, "审核已通过")
}

// @Summary      驳回评价（管理员）
// @Tags         评价管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "评价ID"
// @Success      200  {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/review/{id}/reject [post]
func (h *ReviewHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	if err := h.svc.Reject(uint(id)); err != nil {
		logger.Log.Error("驳回评价失败", zap.Error(err))
		response.Error(c, 500, 1, "操作失败")
		return
	}
	response.Success(c, "评价已驳回")
}
