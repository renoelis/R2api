package service

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"r2api/config"
	"r2api/model"
)

// FileService 处理文件操作
type FileService struct {
	maxFileSize int64
}

// NewFileService 创建文件服务实例
func NewFileService() *FileService {
	return &FileService{
		maxFileSize: config.GetMaxFileSize(),
	}
}

// DownloadFile 从URL下载文件
func (s *FileService) DownloadFile(fileURL string) ([]byte, string, int64, error) {
	// 创建HTTP请求
	client := &http.Client{}
	
	// 先发送HEAD请求获取文件信息
	headReq, err := http.NewRequest("HEAD", fileURL, nil)
	if err != nil {
		return nil, "", 0, err
	}
	
	headResp, err := client.Do(headReq)
	if err != nil {
		return nil, "", 0, err
	}
	defer headResp.Body.Close()
	
	// 获取内容类型和文件大小
	contentType := headResp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	
	contentLength := headResp.ContentLength
	if contentLength > s.maxFileSize {
		return nil, "", 0, fmt.Errorf("文件大小超过限制：%d > %d", contentLength, s.maxFileSize)
	}
	
	// 发送GET请求下载文件
	getReq, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, "", 0, err
	}
	
	getResp, err := client.Do(getReq)
	if err != nil {
		return nil, "", 0, err
	}
	defer getResp.Body.Close()
	
	if getResp.StatusCode != http.StatusOK {
		return nil, "", 0, fmt.Errorf("下载文件失败，HTTP状态码: %d", getResp.StatusCode)
	}
	
	// 读取文件内容
	var buffer bytes.Buffer
	var fileSize int64
	
	// 使用io.Copy读取内容并计算大小
	fileSize, err = io.Copy(&buffer, io.LimitReader(getResp.Body, s.maxFileSize+1))
	if err != nil {
		return nil, "", 0, err
	}
	
	if fileSize > s.maxFileSize {
		return nil, "", 0, fmt.Errorf("文件大小超过限制：%d > %d", fileSize, s.maxFileSize)
	}
	
	return buffer.Bytes(), contentType, fileSize, nil
}

// UploadToR2 将文件上传到R2存储桶
func (s *FileService) UploadToR2(fileData []byte, contentType, bucketName, objectKey, endpoint, accessKeyID, secretAccessKey, customDomain string) (model.UploadResult, error) {
	// 验证objectKey不能以/开头
	if strings.HasPrefix(objectKey, "/") {
		return model.UploadResult{}, fmt.Errorf("objectKey 不能以 '/' 开头")
	}
	
	// 创建AWS会话
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("auto"), // Cloudflare R2 使用 "auto" 区域
		S3ForcePathStyle: aws.Bool(true),     // 必须为R2启用
	}
	
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return model.UploadResult{}, err
	}
	
	// 创建S3/R2客户端
	svc := s3.New(sess)
	
	// 上传文件
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(fileData),
		ContentLength: aws.Int64(int64(len(fileData))),
		ContentType:   aws.String(contentType),
	})
	
	if err != nil {
		return model.UploadResult{}, err
	}
	
	// 构建公共URL
	var publicURL string
	if customDomain != "" {
		// 使用自定义域名
		publicURL = fmt.Sprintf("%s/%s", strings.TrimRight(customDomain, "/"), objectKey)
	} else {
		// 使用默认R2 URL
		publicURL = fmt.Sprintf("%s/%s/%s", strings.TrimRight(endpoint, "/"), bucketName, objectKey)
	}
	
	// 提取文件名
	fileName := filepath.Base(objectKey)
	
	// 返回结果
	result := model.UploadResult{
		PublicURL:   publicURL,
		Size:        int64(len(fileData)),
		ContentType: contentType,
		FileName:    fileName,
	}
	
	return result, nil
}

// UploadDirectly 直接上传文件到R2存储桶（非URL下载）
func (s *FileService) UploadDirectly(file io.Reader, fileName, contentType, bucketName, objectKey, endpoint, accessKeyID, secretAccessKey, customDomain string) (model.UploadResult, error) {
	// 如果未提供objectKey，则使用原文件名
	if objectKey == "" {
		objectKey = fileName
	}
	
	// 验证objectKey不能以/开头
	if strings.HasPrefix(objectKey, "/") {
		return model.UploadResult{}, fmt.Errorf("objectKey 不能以 '/' 开头")
	}
	
	// 读取文件内容到内存
	var buffer bytes.Buffer
	fileSize, err := io.Copy(&buffer, io.LimitReader(file, s.maxFileSize+1))
	if err != nil {
		return model.UploadResult{}, err
	}
	
	if fileSize > s.maxFileSize {
		return model.UploadResult{}, fmt.Errorf("文件大小超过限制：%d > %d", fileSize, s.maxFileSize)
	}
	
	// 如果未提供contentType，尝试根据文件名推断
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(fileName))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}
	
	// 调用上传方法
	return s.UploadToR2(buffer.Bytes(), contentType, bucketName, objectKey, endpoint, accessKeyID, secretAccessKey, customDomain)
} 