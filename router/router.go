package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	"r2api/controller"
	"r2api/utils"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	// 创建Gin引擎
	router := gin.Default()
	
	// 设置信任的代理
	router.SetTrustedProxies([]string{"0.0.0.0/0"})
	
	// 配置CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	
	// 创建控制器实例
	tokenController := controller.NewTokenController()
	uploadController := controller.NewUploadController()
	
	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "r2-uploader",
		})
	})
	
	// 根路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "欢迎使用 r2-uploader API",
			"docs_url": "/docs",
		})
	})
	
	// API分组
	api := router.Group("/R2api")
	{
		// 不需要认证的路由
		api.GET("/token", tokenController.GetToken) // 只返回令牌是否已设置信息，不返回实际令牌
		api.POST("/init-token", tokenController.InitToken) // 令牌初始化端点（只应在部署时使用一次）
		api.POST("/reset-token", tokenController.ResetToken) // 令牌重置端点（强制生成新令牌）
		
		// 需要认证的路由
		protected := api.Group("")
		protected.Use(utils.AuthMiddleware())
		{
			// 上传路由
			protected.POST("/upload", uploadController.UploadFromURL)
			protected.POST("/upload-direct", uploadController.UploadDirectly)
		}
	}
	
	return router
} 