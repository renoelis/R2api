package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
	"r2api/service"
)

// TokenController 处理令牌相关的API请求
type TokenController struct {
	tokenService *service.TokenService
}

// NewTokenController 创建令牌控制器实例
func NewTokenController() *TokenController {
	return &TokenController{
		tokenService: service.NewTokenService(),
	}
}

// GetToken 获取系统令牌信息
func (c *TokenController) GetToken(ctx *gin.Context) {
	tokenInfo := c.tokenService.GetTokenInfo()
	
	ctx.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"token":       tokenInfo.Token,
		"is_permanent": tokenInfo.IsPermanent,
	})
}

// InitToken 初始化系统API令牌
func (c *TokenController) InitToken(ctx *gin.Context) {
	token, err := c.tokenService.GenerateDefaultToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "生成令牌失败: " + err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API令牌已初始化",
		"token":   token,
	})
}

// ResetToken 重置API令牌（强制生成新令牌）
func (c *TokenController) ResetToken(ctx *gin.Context) {
	token, err := c.tokenService.ResetToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "重置令牌失败: " + err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API令牌已成功重置",
		"token":   token,
	})
} 