package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
	"r2api/model"
	"r2api/service"
)

// UploadController 处理文件上传相关的API请求
type UploadController struct {
	fileService *service.FileService
}

// NewUploadController 创建上传控制器实例
func NewUploadController() *UploadController {
	return &UploadController{
		fileService: service.NewFileService(),
	}
}

// UploadFromURL 从URL下载文件并上传到R2
func (c *UploadController) UploadFromURL(ctx *gin.Context) {
	var req model.UploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	// 从URL下载文件
	fileData, contentType, _, err := c.fileService.DownloadFile(req.FileURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "下载文件失败: " + err.Error(),
		})
		return
	}
	
	// 上传到R2
	result, err := c.fileService.UploadToR2(
		fileData,
		contentType,
		req.BucketName,
		req.ObjectKey,
		req.Endpoint,
		req.AccessKeyID,
		req.SecretAccessKey,
		req.CustomDomain,
	)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "上传到R2失败: " + err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, model.Response{
		Status:  "success",
		Message: "文件上传成功",
		Data:    result,
	})
}

// UploadDirectly 直接上传文件到R2
func (c *UploadController) UploadDirectly(ctx *gin.Context) {
	// 获取表单参数
	bucketName := ctx.PostForm("bucket_name")
	if bucketName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bucket_name参数必填",
		})
		return
	}
	
	objectKey := ctx.PostForm("object_key")
	endpoint := ctx.PostForm("endpoint")
	if endpoint == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "endpoint参数必填",
		})
		return
	}
	
	accessKeyID := ctx.PostForm("access_key_id")
	if accessKeyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "access_key_id参数必填",
		})
		return
	}
	
	secretAccessKey := ctx.PostForm("secret_access_key")
	if secretAccessKey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "secret_access_key参数必填",
		})
		return
	}
	
	customDomain := ctx.PostForm("custom_domain")
	
	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "无法获取上传文件: " + err.Error(),
		})
		return
	}
	
	// 打开文件
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "无法打开上传的文件: " + err.Error(),
		})
		return
	}
	defer src.Close()
	
	// 如果未提供objectKey，则使用原文件名
	if objectKey == "" {
		objectKey = file.Filename
	}
	
	// 上传文件
	result, err := c.fileService.UploadDirectly(
		src,
		file.Filename,
		file.Header.Get("Content-Type"),
		bucketName,
		objectKey,
		endpoint,
		accessKeyID,
		secretAccessKey,
		customDomain,
	)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "上传到R2失败: " + err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, model.Response{
		Status:  "success",
		Message: "文件上传成功",
		Data:    result,
	})
} 