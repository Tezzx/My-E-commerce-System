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

type AddressHandler struct {
	svc *service.AddressService
}

func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

// @Summary      创建地址
// @Tags         地址
// @Accept       json
// @Produce      json
// @Param        body  body      types.AddressReq  true  "地址信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/address [post]
func (h *AddressHandler) Create(c *gin.Context) {
	var req types.AddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	if err := h.svc.Create(userID, &req); err != nil {
		logger.Log.Error("创建地址失败", zap.Error(err))
		response.Error(c, 500, 1, "创建失败")
		return
	}
	response.Success(c, "创建成功")
}

// @Summary      更新地址
// @Tags         地址
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "地址ID"
// @Param        body  body      types.AddressReq  true  "地址信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/address/{id} [put]
func (h *AddressHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	var req types.AddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	if err := h.svc.Update(userID, uint(id), &req); err != nil {
		logger.Log.Error("更新地址失败", zap.Error(err))
		response.Error(c, 500, 1, "更新失败")
		return
	}
	response.Success(c, "更新成功")
}

// @Summary      删除地址
// @Tags         地址
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "地址ID"
// @Success      200  {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/address/{id} [delete]
func (h *AddressHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	if err := h.svc.Delete(userID, uint(id)); err != nil {
		logger.Log.Error("删除地址失败", zap.Error(err))
		response.Error(c, 500, 1, "删除失败")
		return
	}
	response.Success(c, "删除成功")
}

// @Summary      地址列表
// @Tags         地址
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Resp{data=[]types.AddressResp}
// @Security     BearerAuth
// @Router       /home/address/list [get]
func (h *AddressHandler) List(c *gin.Context) {
	userID := c.GetUint("userID")
	addrs, err := h.svc.List(userID)
	if err != nil {
		logger.Log.Error("获取地址列表失败", zap.Error(err))
		response.Error(c, 500, 1, "获取失败")
		return
	}
	response.Success(c, addrs)
}
