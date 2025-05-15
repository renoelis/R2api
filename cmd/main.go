package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"r2api/config"
	"r2api/router"
	"r2api/service"
)

func main() {
	// 加载环境变量
	config.LoadEnv()
	
	// 初始化TokenService，确保有默认令牌
	tokenService := service.NewTokenService()
	token, err := tokenService.GenerateDefaultToken()
	if err != nil {
		log.Printf("警告: 无法生成默认令牌: %v\n", err)
	} else if token != "" {
		log.Printf("API令牌已设置。请记得在应用程序调用API时使用此令牌。\n")
	}
	
	// 获取端口
	port := config.GetPort()
	
	// 设置路由
	r := router.SetupRouter()
	
	// 启动服务器
	go func() {
		log.Printf("开始监听端口 %s...\n", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("启动服务器失败: %v\n", err)
		}
	}()
	
	// 等待信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("正在关闭服务器...")
} 