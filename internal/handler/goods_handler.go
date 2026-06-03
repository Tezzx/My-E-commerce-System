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

// @Summary      商品列表
// @Description  获取所有商品
// @Tags         商品
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Resp{data=[]types.Goods}
// @Security     BearerAuth
// @Router       /home [get]
func (g *GoodsHandler) GetGoodsList(c *gin.Context) {
	goods, err := g.goodsService.GetGoodsList()
	if err != nil {
		logger.Log.Error("获取商品列表失败", zap.Error(err))
		response.Error(c, 500, 2, "获取商品列表失败")
		return
	}

	var dto []types.Goods
	for _, v := range goods {
		dto = append(dto, types.Goods{
			ID:         v.ID,
			GoodsName:  v.Goodsname,
			GoodsNum:   v.Goodsnum,
			Price:      v.Price,
			CategoryID: v.CategoryID,
			ImageURL:   v.ImageURL,
			Sales:      v.Sales,
		})
	}
	response.Success(c, dto)
}

// @Summary      商品详情
// @Description  根据商品ID查询详情
// @Tags         商品
// @Accept       json
// @Produce      json
// @Param        id   query     int  true  "商品ID"
// @Success      200  {object}  response.Resp{data=types.Goods}
// @Failure      404  {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/search/goods [get]
func (g *GoodsHandler) GetGoodsDetail(c *gin.Context) {
	goodsID := c.Query("id")

	id, err := strconv.ParseUint(goodsID, 10, 64)
	if err != nil {
		logger.Log.Warn("商品ID格式错误", zap.String("input", goodsID))
		response.Error(c, 400, 1, "商品ID格式错误")
		return
	}

	price, goodsNum, goodsName, err := g.goodsService.GetGoodsInfoByID(uint(id))
	if err != nil {
		logger.Log.Error("查询商品信息失败", zap.Uint("goods_id", uint(id)), zap.Error(err))
		response.Error(c, 404, 1, "商品不存在")
		return
	}

	dto := types.Goods{
		ID:        uint(id),
		GoodsName: goodsName,
		GoodsNum:  goodsNum,
		Price:     price,
	}
	response.Success(c, dto)
}
