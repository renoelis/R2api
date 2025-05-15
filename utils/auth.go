package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	
	"r2api/service"
)

// AuthMiddleware 验证API令牌的中间件
func AuthMiddleware() gin.HandlerFunc {
	tokenService := service.NewTokenService()
	
	return func(c *gin.Context) {
		// 从Authorization头部获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "未提供认证令牌",
			})
			return
		}
		
		// 提取Bearer令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "认证格式无效，应使用Bearer令牌",
			})
			return
		}
		
		token := parts[1]
		
		// 验证令牌
		valid, err := tokenService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("令牌验证错误: %v", err),
			})
			return
		}
		
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "无效的认证令牌",
			})
			return
		}
		
		// 令牌有效，继续处理请求
		c.Next()
	}
} 