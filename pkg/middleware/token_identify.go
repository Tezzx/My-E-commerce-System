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
		tokenStr := authHeader[len(bearerPrefix):]
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
