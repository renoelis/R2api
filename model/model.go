package model

import "time"

// 通用API响应结构
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 上传请求参数
type UploadRequest struct {
	FileURL         string `json:"fileUrl" binding:"required,url"`
	BucketName      string `json:"bucketName" binding:"required"`
	ObjectKey       string `json:"objectKey" binding:"required"`
	Endpoint        string `json:"endpoint" binding:"required,url"`
	AccessKeyID     string `json:"accessKeyId" binding:"required"`
	SecretAccessKey string `json:"secretAccessKey" binding:"required"`
	CustomDomain    string `json:"customdomain,omitempty"`
}

// 文件上传结果
type UploadResult struct {
	PublicURL   string `json:"public_url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name,omitempty"`
}

// 令牌元数据
type TokenInfo struct {
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	IsPermanent bool      `json:"is_permanent"`
	Username    string    `json:"username,omitempty"`
	Email       string    `json:"email,omitempty"`
}

// 令牌生成响应
type TokenResponse struct {
	Status      string    `json:"status"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	IsPermanent bool      `json:"is_permanent"`
} 