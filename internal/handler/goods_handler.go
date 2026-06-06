package handler

import (
	"order-payment-system/internal/model"
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

// ==================== 商家/管理员 商品管理 ====================

// @Summary      创建商品（管理员/商家）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        body  body      types.CreateGoodsReq  true  "商品信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/goods [post]
func (g *GoodsHandler) CreateGoods(c *gin.Context) {
	var req types.CreateGoodsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	goods := &model.Goods{
		Goodsname:  req.GoodsName,
		Goodsnum:   req.GoodsNum,
		Price:      req.Price,
		CategoryID: req.CategoryID,
		ImageURL:   req.ImageURL,
		Status:     1,
	}
	if err := g.goodsService.CreateGoods(goods); err != nil {
		response.Error(c, 500, 1, "创建商品失败")
		return
	}
	response.Success(c, goods)
}

// @Summary      更新商品（管理员/商家）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "商品ID"
// @Param        body  body      types.UpdateGoodsReq  true  "更新字段"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/goods/{id} [put]
func (g *GoodsHandler) UpdateGoods(c *gin.Context) {
	goodsID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "商品ID格式错误")
		return
	}

	var req types.UpdateGoodsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.GoodsName != nil {
		updates["goodsname"] = *req.GoodsName
	}
	if req.GoodsNum != nil {
		updates["goodsnum"] = *req.GoodsNum
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}
	if req.ImageURL != nil {
		updates["image_url"] = *req.ImageURL
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		response.Error(c, 400, 1, "无更新字段")
		return
	}

	if err := g.goodsService.UpdateGoods(uint(goodsID), updates); err != nil {
		response.Error(c, 500, 1, "更新商品失败")
		return
	}
	response.Success(c, "更新成功")
}

// @Summary      删除商品（管理员/商家）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "商品ID"
// @Success      200  {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/goods/{id} [delete]
func (g *GoodsHandler) DeleteGoods(c *gin.Context) {
	goodsID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, 1, "商品ID格式错误")
		return
	}

	if err := g.goodsService.DeleteGoods(uint(goodsID)); err != nil {
		response.Error(c, 500, 1, "删除商品失败")
		return
	}
	response.Success(c, "删除成功")
}

// @Summary      商品分页列表（后台管理）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        page        query     int  false "页码"        default(1)
// @Param        size        query     int  false "每页数量"    default(10)
// @Param        category_id query     int  false "分类ID"
// @Param        status      query     int  false "状态 1-上架 0-下架"
// @Success      200         {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/goods/list [get]
func (g *GoodsHandler) GetGoodsPageList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	var categoryID *uint
	if cidStr := c.Query("category_id"); cidStr != "" {
		cid, _ := strconv.ParseUint(cidStr, 10, 64)
		cidUint := uint(cid)
		categoryID = &cidUint
	}

	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	goods, total, err := g.goodsService.GetGoodsListWithPage(page, size, categoryID, status)
	if err != nil {
		response.Error(c, 500, 1, "获取商品列表失败")
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

	response.Success(c, gin.H{
		"list":  dto,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
