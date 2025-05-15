package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	
	"r2api/config"
	"r2api/model"
)

// TokenService 处理令牌的生成和验证
type TokenService struct{}

// NewTokenService 创建令牌服务实例
func NewTokenService() *TokenService {
	return &TokenService{}
}

// GenerateDefaultToken 生成默认的API令牌
func (s *TokenService) GenerateDefaultToken() (string, error) {
	// 检查是否已存在令牌
	existingToken := config.GetAPIToken()
	if existingToken != "" {
		return existingToken, nil
	}

	// 生成新令牌
	token, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	// 保存令牌到文件
	err = config.SaveAPIToken(token)
	if err != nil {
		return "", err
	}

	log.Println("已生成并保存默认API令牌")
	return token, nil
}

// ResetToken 强制重置API令牌
func (s *TokenService) ResetToken() (string, error) {
	// 生成新令牌
	token, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	// 保存令牌到文件
	err = config.SaveAPIToken(token)
	if err != nil {
		return "", err
	}

	log.Println("已重置并保存新的API令牌")
	return token, nil
}

// ValidateToken 验证API令牌
func (s *TokenService) ValidateToken(token string) (bool, error) {
	validToken := config.GetAPIToken()
	if validToken == "" {
		return false, errors.New("系统未设置API令牌")
	}

	// 简单比较令牌
	return strings.TrimSpace(token) == strings.TrimSpace(validToken), nil
}

// GetTokenInfo 获取令牌信息
func (s *TokenService) GetTokenInfo() model.TokenInfo {
	token := config.GetAPIToken()
	
	// 创建令牌信息
	tokenInfo := model.TokenInfo{
		Token:       token,
		CreatedAt:   time.Now(), // 我们不存储创建时间，所以使用当前时间
		IsPermanent: true,       // 默认令牌是永久的
		Username:    "system",   // 默认用户名
	}

	return tokenInfo
}

// generateSecureToken 生成安全的随机令牌
func generateSecureToken(length int) (string, error) {
	// 首先生成UUID作为基础
	id := uuid.New().String()
	
	// 然后生成随机字节
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	// 将UUID和随机字节组合并编码
	return strings.ReplaceAll(id, "-", "") + base64.URLEncoding.EncodeToString(b)[:20], nil
} 