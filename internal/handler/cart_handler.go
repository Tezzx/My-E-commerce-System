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

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// @Summary      添加到购物车
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Param        body  body      types.AddCartReq  true  "商品信息"
// @Success      200   {object}  response.Resp
// @Failure      400   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/cart/add [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	var req types.AddCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userIDany, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, 1, "请先登录")
		return
	}
	userID := userIDany.(uint)

	if err := h.cartService.AddToCart(userID, req.GoodsID, req.Quantity); err != nil {
		logger.Log.Error("添加到购物车失败", zap.Error(err))
		response.Error(c, 500, 1, err.Error())
		return
	}

	response.Success(c, "添加成功")
}

// @Summary      更新购物车数量
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Param        body  body      types.UpdateCartReq  true  "更新信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/cart/update [put]
func (h *CartHandler) UpdateCart(c *gin.Context) {
	var req types.UpdateCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userIDany, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, 1, "请先登录")
		return
	}
	userID := userIDany.(uint)

	if err := h.cartService.UpdateCart(userID, req.GoodsID, req.Quantity); err != nil {
		logger.Log.Error("更新购物车失败", zap.Error(err))
		response.Error(c, 500, 1, err.Error())
		return
	}

	response.Success(c, "更新成功")
}

// @Summary      删除购物车商品
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Param        goods_id  query     int  true  "商品ID"
// @Success      200       {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/cart/delete [delete]
func (h *CartHandler) DeleteCart(c *gin.Context) {
	goodsIDStr := c.Query("goods_id")
	goodsID, err := strconv.ParseUint(goodsIDStr, 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userIDany, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, 1, "请先登录")
		return
	}
	userID := userIDany.(uint)

	if err := h.cartService.DeleteCart(userID, uint(goodsID)); err != nil {
		logger.Log.Error("删除购物车失败", zap.Error(err))
		response.Error(c, 500, 1, err.Error())
		return
	}

	response.Success(c, "删除成功")
}

// @Summary      切换购物车勾选
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Param        body  body      types.ToggleSelectReq  true  "勾选信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/cart/select [patch]
func (h *CartHandler) ToggleSelect(c *gin.Context) {
	var req types.ToggleSelectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userIDany, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, 1, "请先登录")
		return
	}
	userID := userIDany.(uint)

	if err := h.cartService.ToggleSelect(userID, req.GoodsID, req.Selected); err != nil {
		logger.Log.Error("切换购物车勾选失败", zap.Error(err))
		response.Error(c, 500, 1, err.Error())
		return
	}

	response.Success(c, "操作成功")
}

// @Summary      购物车列表
// @Tags         购物车
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Resp{data=types.CartListResp}
// @Security     BearerAuth
// @Router       /home/cart/list [get]
func (h *CartHandler) GetCartList(c *gin.Context) {
	userIDany, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, 1, "请先登录")
		return
	}
	userID := userIDany.(uint)

	resp, err := h.cartService.GetCartList(userID)
	if err != nil {
		logger.Log.Error("获取购物车列表失败", zap.Error(err))
		response.Error(c, 500, 1, "获取失败")
		return
	}

	response.Success(c, resp)
}
