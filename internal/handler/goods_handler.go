package handler

import (
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
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
		response.Error(c, 500, "获取商品列表失败")
		return
	}
	response.Success(c, goods)
}

func (g *GoodsHandler) GetGoodsDetail(c *gin.Context) {
	goodsID := c.Query("id")

	id, err := parseUint(goodsID)
	if err != nil {
		response.Error(c, 400, "商品ID格式错误")
		return
	}

	price, goodsNum, goodsName, err := g.goodsService.GetGoodsInfoByID(id)
	if err != nil {
		response.Error(c, 404, "商品不存在")
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
