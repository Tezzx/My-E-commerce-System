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

type GoodsHandler struct {
	goodsService *service.GoodsService
}

func NewGoodsHandler(goodsService *service.GoodsService) *GoodsHandler {
	return &GoodsHandler{
		goodsService: goodsService,
	}
}

func (g *GoodsHandler) GetGoodsList(c *gin.Context) {
	goods, err := g.goodsService.GetGoodsList()
	if err != nil {
		logger.Log.Error("获取商品列表失败", zap.Error(err))
		response.Error(c, 500, 2, "获取商品列表失败")
		return
	}
	response.Success(c, goods)
}

func (g *GoodsHandler) GetGoodsDetail(c *gin.Context) {
	goodsID := c.Query("id")

	id, err := parseUint(goodsID)
	if err != nil {
		logger.Log.Warn("商品ID格式错误", zap.String("input", goodsID))
		response.Error(c, 400, 1, "商品ID格式错误")
		return
	}

	price, goodsNum, goodsName, err := g.goodsService.GetGoodsInfoByID(id)
	if err != nil {
		logger.Log.Error("查询商品信息失败", zap.Uint("goods_id", id), zap.Error(err))
		response.Error(c, 404, 1, "商品不存在")
		return
	}

	dto := types.Goods{
		ID:        id,
		GoodsName: goodsName,
		GoodsNum:  goodsNum,
		Price:     price,
	}
	response.Success(c, dto)
}

func parseUint(s string) (uint, error) {
	num, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}
