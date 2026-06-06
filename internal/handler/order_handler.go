package handler

import (
	"encoding/json"
	"order-payment-system/internal/model"
	"order-payment-system/internal/service"
	"order-payment-system/internal/types"
	"order-payment-system/pkg/logger"
	"order-payment-system/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderService   *service.OrderService
	cartService    *service.CartService
	addressService *service.AddressService
}

func NewOrderHandler(orderService *service.OrderService, cartService *service.CartService, addressService *service.AddressService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		cartService:    cartService,
		addressService: addressService,
	}
}

// snapshotAddress 将地址模型序列化为 JSON 快照
func snapshotAddress(addr *model.Address) string {
	if addr == nil {
		return "{}"
	}
	b, _ := json.Marshal(types.AddressResp{
		ID:       addr.ID,
		Receiver: addr.Receiver,
		Phone:    addr.Phone,
		Province: addr.Province,
		City:     addr.City,
		District: addr.District,
		Detail:   addr.Detail,
	})
	return string(b)
}

// @Summary      创建订单
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        body  body      types.OrderRequest  true  "订单信息"
// @Success      200   {object}  response.Resp{data=string}
// @Failure      400   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order [post]
func (o *OrderHandler) CreateOrder(c *gin.Context) {
	var req types.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warn("创建订单 - 参数绑定失败", zap.Error(err))
		response.Error(c, 400, 1, "参数错误")
		return
	}

	userID := c.GetUint("userID")

	// 快照收货地址
	var snapshot string
	if req.AddressID != nil {
		addr, err := o.addressService.GetByID(userID, *req.AddressID)
		if err == nil {
			snapshot = snapshotAddress(addr)
		}
	}

	orderNo, err := o.orderService.CreateOrder(c.Request.Context(), userID, uint(req.GoodsID), uint(req.BuyNum), req.AddressID, snapshot)
	if err != nil {
		logger.Ctx(c.Request.Context()).Error("创建订单失败",
			zap.Uint("user_id", userID), zap.Uint("goods_id", uint(req.GoodsID)), zap.Uint("buy_num", uint(req.BuyNum)), zap.Error(err))
		response.Error(c, 500, 1, "订单创建失败")
		return
	}

	logger.Ctx(c.Request.Context()).Info("订单创建成功", zap.String("order_no", orderNo))
	response.Success(c, orderNo)
}

// @Summary      购物车结算
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        body  body      types.CartCheckoutReq  true  "结算信息"
// @Success      200   {object}  response.Resp{data=string}
// @Security     BearerAuth
// @Router       /home/order/cart_checkout [post]
func (o *OrderHandler) CheckoutCart(c *gin.Context) {
	var req types.CartCheckoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "请选择收货地址")
		return
	}
	userID := c.GetUint("userID")

	// 校验收货地址并快照
	addr, err := o.addressService.GetByID(userID, req.AddressID)
	if err != nil {
		response.Error(c, 400, 1, "收货地址不存在")
		return
	}
	snapshot := snapshotAddress(addr)

	// 获取用户购物车所有商品
	cartResp, err := o.cartService.GetCartList(userID)
	if err != nil {
		response.Error(c, 500, 1, "获取购物车异常")
		return
	}

	var cartItems []model.CartItem
	for _, item := range cartResp.Items {
		cartItems = append(cartItems, model.CartItem{
			GoodsID:  item.GoodsID,
			Quantity: item.Quantity,
			Selected: item.Selected,
		})
	}

	orderNo, err := o.orderService.CreateCartOrder(c.Request.Context(), userID, cartItems, req.AddressID, snapshot)
	if err != nil {
		logger.Log.Error("购物车下单失败", zap.Error(err))
		response.Error(c, 500, 1, err.Error())
		return
	}

	// 下单成功后清空已勾选的购物车商品
	for _, item := range cartItems {
		if item.Selected {
			_ = o.cartService.DeleteCart(userID, item.GoodsID)
		}
	}

	response.Success(c, orderNo)
}

// @Summary      取消订单
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        orderNo  query     string  true  "订单号"
// @Success      200      {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order/cancel [get]
func (o *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Query("orderNo")
	err := o.orderService.CancelOrder(orderNo)
	if err != nil {
		logger.Log.Error("取消订单失败", zap.String("order_no", orderNo), zap.Error(err))
		response.Error(c, 500, 1, "取消订单失败")
		return
	}
	response.Success(c, "取消成功")
}

// @Summary      商家发货
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        body  body      types.ShipRequest  true  "物流信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order/ship [post]
func (o *OrderHandler) ShipOrder(c *gin.Context) {
	var req types.ShipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	if err := o.orderService.ShipOrder(req.OrderNo, req.ShipCompany, req.ShipNo); err != nil {
		response.Error(c, 500, 1, err.Error())
		return
	}
	response.Success(c, "发货成功")
}

// @Summary      确认收货
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        order_no  query     string  true  "订单号"
// @Success      200       {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order/confirm [post]
func (o *OrderHandler) ConfirmReceipt(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, 400, 1, "缺少订单号")
		return
	}
	userID := c.GetUint("userID")
	if err := o.orderService.ConfirmReceipt(userID, orderNo); err != nil {
		response.Error(c, 500, 1, err.Error())
		return
	}
	response.Success(c, "确认收货成功")
}

// @Summary      完成订单
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        order_no  query     string  true  "订单号"
// @Success      200       {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order/complete [post]
func (o *OrderHandler) CompleteOrder(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, 400, 1, "缺少订单号")
		return
	}
	if err := o.orderService.CompleteOrder(orderNo); err != nil {
		response.Error(c, 500, 1, err.Error())
		return
	}
	response.Success(c, "订单已完成")
}

// @Summary      订单日志
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        order_no  query     string  true  "订单号"
// @Success      200       {object}  response.Resp{data=[]types.OrderLogResp}
// @Security     BearerAuth
// @Router       /home/order/logs [get]
func (o *OrderHandler) GetOrderLogs(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, 400, 1, "缺少订单号")
		return
	}
	logs, err := o.orderService.GetOrderLogs(orderNo)
	if err != nil {
		response.Error(c, 500, 1, "获取失败")
		return
	}
	var resp []types.OrderLogResp
	for _, l := range logs {
		resp = append(resp, types.OrderLogResp{
			ID:        l.ID,
			OrderNo:   l.OrderNo,
			OldStatus: l.OldStatus,
			NewStatus: l.NewStatus,
			Remark:    l.Remark,
			CreatedAt: l.CreatedAt,
		})
	}
	response.Success(c, resp)
}

// @Summary      用户订单列表
// @Summary      用户订单列表
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        page    query     int  false "页码"      default(1)
// @Param        size    query     int  false "每页数量"  default(10)
// @Param        status  query     int  false "订单状态筛选"
// @Success      200     {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/search/orders [get]
func (o *OrderHandler) GetUserOrders(c *gin.Context) {
	userIDany, exists := c.Get("userID")
	if !exists {
		logger.Log.Warn("获取用户订单 - 用户未登录")
		response.Error(c, 401, 1, "请先登录")
		return
	}

	userID, ok := userIDany.(uint)
	if !ok {
		logger.Log.Warn("获取用户订单 - 用户ID类型转换失败")
		response.Error(c, 500, 1, "用户ID类型错误")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	orders, total, err := o.orderService.GetUserOrderListWithPage(userID, page, size, status)
	if err != nil {
		logger.Log.Error("获取用户订单列表失败", zap.Uint("user_id", userID), zap.Error(err))
		response.Error(c, 500, 1, "获取订单失败")
		return
	}

	// 组装 DTO
	var dto []types.OrderResp
	for _, order := range orders {
		statusName := model.OrderStatusNames[order.Status]

		var items []types.OrderItemResp
		for _, item := range order.OrderItems {
			items = append(items, types.OrderItemResp{
				ID:        item.ID,
				GoodsID:   item.GoodsID,
				GoodsName: item.GoodsName,
				Price:     item.Price,
				Quantity:  item.Quantity,
			})
		}

		resp := types.OrderResp{
			ID:         order.ID,
			OrderNo:    order.OrderNo,
			UserID:     order.UserID,
			TotalPrice: order.TotalPrice,
			Status:     order.Status,
			StatusName: statusName,
			PayTime:    order.PayTime,
			CreatedAt:  order.CreatedAt,
			Items:      items,
		}
		// 快递信息
		if order.ShipCompany != "" {
			resp.ShipCompany = order.ShipCompany
			resp.ShipNo = order.ShipNo
		}
		dto = append(dto, resp)
	}

	logger.Log.Debug("获取用户订单列表成功", zap.Uint("user_id", userID), zap.Int("count", len(dto)))
	response.Success(c, gin.H{
		"list":  dto,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ==================== 退款相关 ====================

// @Summary      申请退款
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        body  body      types.RefundRequest  true  "退款申请"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order/refund/request [post]
func (o *OrderHandler) RequestRefund(c *gin.Context) {
	var req types.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	if err := o.orderService.RequestRefund(userID, req.OrderNo, req.Reason); err != nil {
		logger.Log.Error("申请退款失败", zap.String("order_no", req.OrderNo), zap.Error(err))
		response.Error(c, 500, 1, err.Error())
		return
	}
	response.Success(c, "退款申请已提交")
}

// @Summary      处理退款（商家/管理员审批）
// @Tags         订单
// @Accept       json
// @Produce      json
// @Param        body  body      types.RefundProcessReq  true  "审批信息"
// @Success      200   {object}  response.Resp
// @Security     BearerAuth
// @Router       /home/order/refund/process [post]
func (o *OrderHandler) ProcessRefund(c *gin.Context) {
	var req types.RefundProcessReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 1, "参数错误")
		return
	}

	if req.Approved {
		if err := o.orderService.ProcessRefund(req.OrderNo); err != nil {
			logger.Log.Error("退款审批通过失败", zap.String("order_no", req.OrderNo), zap.Error(err))
			response.Error(c, 500, 1, err.Error())
			return
		}
		response.Success(c, "退款已通过")
	} else {
		if err := o.orderService.RejectRefund(req.OrderNo, req.Remark); err != nil {
			logger.Log.Error("退款审批拒绝失败", zap.String("order_no", req.OrderNo), zap.Error(err))
			response.Error(c, 500, 1, err.Error())
			return
		}
		response.Success(c, "退款已拒绝")
	}
}
