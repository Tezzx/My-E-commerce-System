package middleware

import (
	"order-payment-system/pkg/jwt"
	"order-payment-system/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// 验证token
func TokenIdentify() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, 1, "没有token")
			c.Abort()
			return
		}
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			response.Error(c, 401, 1, "Authorization 格式错误，应为 'Bearer <token>'")
			c.Abort()
			return
		}
		//为了方便调用接口，所以绕过token检验
		tokenStr := authHeader[len(bearerPrefix):]
		if tokenStr == "yes" {
			c.Set("userID", uint(2))
			c.Next()
			return
		}
		claims, err := jwt.ValidateJWT(tokenStr)
		if err != nil {
			response.Error(c, 401, 1, "token过期或错误")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
