package middleware

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequireRole 角色鉴权中间件，roles 为允许的角色列表
func RequireRole(userRepo *repository.UserRepo, roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		userIDAny, exists := c.Get("userID")
		if !exists {
			response.Error(c, 401, 1, "请先登录")
			c.Abort()
			return
		}
		userID := userIDAny.(uint)

		role, err := userRepo.GetRoleByID(userID)
		if err != nil || !roleSet[role] {
			response.Error(c, 403, 1, "权限不足")
			c.Abort()
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

// AdminOnly 仅管理员
func AdminOnly(userRepo *repository.UserRepo) gin.HandlerFunc {
	return RequireRole(userRepo, model.RoleAdmin)
}

// MerchantOnly 商家及以上（含管理员）
func MerchantOnly(userRepo *repository.UserRepo) gin.HandlerFunc {
	return RequireRole(userRepo, model.RoleMerchant, model.RoleAdmin)
}
