package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code    int         `json:"code"`    // 业务码
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 数据
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// 失败响应
func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Resp{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}

/*
0	成功	所有成功响应

1xxx 参数/鉴权

1001	参数错误
1002	未登录 / Token 无效
1003	权限不足

2xxx 用户相关

2001	用户已存在
2002	用户不存在
2003	密码错误
2004	余额不足

3xxx 商品/库存

3001	商品不存在
3002	库存不足

4xxx 订单

4001	订单不存在
4002	订单状态不允许（已支付/已取消）
4003	订单创建失败

5xxx 支付

5001	支付失败
5002	重复支付

9xxx 系统

9001	系统内部错误
9002	数据库/缓存错误
*/
