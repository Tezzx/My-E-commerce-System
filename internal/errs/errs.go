package errs

import "errors"

// 自定义业务错误
var (
	// 通用
	UnknowError = errors.New("服务器运行错误")
	ParamError  = errors.New("参数错误")

	// 用户模块
	UserExists          = errors.New("用户名已存在")
	UserNotFound        = errors.New("用户不存在")
	PasswordWrong       = errors.New("密码错误")
	InsufficientBalance = errors.New("余额不足")

	// 商品模块
	GoodsNotFound     = errors.New("商品不存在")
	InsufficientStock = errors.New("库存不足")
	CreateGoodsError  = errors.New("商品创建失败")
	GetGoodsError     = errors.New("商品列表读取失败")

	// 订单模块
	OrderNotFound      = errors.New("订单不存在")
	OrderPaid          = errors.New("订单已支付")
	OrderStatusInvalid = errors.New("订单状态不允许此操作")
	OrderCreateFailed  = errors.New("订单创建失败")
	OrderNotOwned      = errors.New("无权操作此订单")
	OrderMQAckTimeout  = errors.New("等待MQ确认超时")

	// 支付模块
	PaymentFailed    = errors.New("支付失败")
	PaymentSignError = errors.New("支付签名验证失败")

	// 退款模块
	RefundNotAllowed = errors.New("当前状态不允许退款")
	RefundFailed     = errors.New("退款处理失败")

	// 认证鉴权
	Unauthorized = errors.New("未登录或Token无效")
	Forbidden    = errors.New("权限不足")
)

// ErrCode 错误码映射
var ErrCode = map[error]int{
	// 1xxx 参数/鉴权
	ParamError:   1001,
	Unauthorized: 1002,
	Forbidden:    1003,

	// 2xxx 用户
	UserExists:          2001,
	UserNotFound:        2002,
	PasswordWrong:       2003,
	InsufficientBalance: 2004,

	// 3xxx 商品
	GoodsNotFound:     3001,
	InsufficientStock: 3002,
	CreateGoodsError:  3003,
	GetGoodsError:     3004,

	// 4xxx 订单
	OrderNotFound:      4001,
	OrderPaid:          4002,
	OrderStatusInvalid: 4003,
	OrderCreateFailed:  4004,
	OrderNotOwned:      4005,
	OrderMQAckTimeout:  4006,

	// 5xxx 支付
	PaymentFailed:    5001,
	PaymentSignError: 5002,

	// 6xxx 退款
	RefundNotAllowed: 6001,
	RefundFailed:     6002,
}

// GetCode 根据 error 获取业务错误码，找不到返回 1
func GetCode(err error) int {
	for target, code := range ErrCode {
		if errors.Is(err, target) {
			return code
		}
	}
	return 1
}
