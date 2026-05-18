package errs

import "errors"

// 自定义业务错误
var (
	UnknowError = errors.New("服务器运行错误")
	// 用户模块
	UserExists          = errors.New("用户名已存在") // 新增
	UserNotFound        = errors.New("用户不存在")
	PasswordWrong       = errors.New("密码错误")
	InsufficientBalance = errors.New("余额不足")

	// 商品模块
	GoodsNotFound     = errors.New("商品不存在")
	InsufficientStock = errors.New("库存不足")
	CreateGoodsError  = errors.New("商品创建失败")
	GetGoodsError     = errors.New("商品列表读取失败")

	// 订单模块
	OrderNotFound = errors.New("订单不存在")
	OrderPaid     = errors.New("订单已支付")
)
