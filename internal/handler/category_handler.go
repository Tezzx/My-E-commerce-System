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

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// @Summary      创建分类
// @Tags         分类
// @Accept       json
// @Produce      json
// @Param        body  body      types.CategoryReq  true  "分类信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/category [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req types.CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	if err := h.svc.Create(&req); err != nil {
		logger.Log.Error("创建分类失败", zap.Error(err))
		response.Error(c, 500, 1, "创建失败")
		return
	}
	response.Success(c, "创建成功")
}

// @Summary      更新分类
// @Tags         分类
// @Accept       json
// @Produce      json
// @Param        id    path      int               true  "分类ID"
// @Param        body  body      types.CategoryReq  true  "分类信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/category/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	var req types.CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		logger.Log.Error("更新分类失败", zap.Error(err))
		response.Error(c, 500, 1, "更新失败")
		return
	}
	response.Success(c, "更新成功")
}

// @Summary      删除分类
// @Tags         分类
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "分类ID"
// @Success      200  {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/category/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		logger.Log.Error("删除分类失败", zap.Error(err))
		response.Error(c, 500, 1, "删除失败")
		return
	}
	response.Success(c, "删除成功")
}

// @Summary      分类树
// @Tags         分类
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Resp{data=[]types.CategoryResp}
// @Security     BearerAuth
// @Router       /home/category/tree [get]
func (h *CategoryHandler) GetTree(c *gin.Context) {
	tree, err := h.svc.GetTree()
	if err != nil {
		logger.Log.Error("获取分类树失败", zap.Error(err))
		response.Error(c, 500, 1, "获取失败")
		return
	}
	response.Success(c, tree)
}
