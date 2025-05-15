package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	// 应用配置
	ServiceName = "r2-uploader"
	APIVersion  = "v1"
	DefaultPort = "3009"
	
	// 默认文件大小限制 (200MB)
	DefaultMaxFileSize = 200 * 1024 * 1024
	
	// 默认令牌文件
	DefaultTokenFile = "config/token_data.json"
)

// 令牌数据结构
type TokenData struct {
	Token string `json:"token"`
}

// 加载环境变量
func LoadEnv() {
	// 尝试加载.env文件
	err := godotenv.Load()
	if err != nil {
		log.Println("警告: .env文件未找到，将使用默认配置或环境变量")
	}
}

// 获取服务端口
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return DefaultPort
	}
	return port
}

// 获取最大文件大小限制（字节）
func GetMaxFileSize() int64 {
	sizeStr := os.Getenv("MAX_FILE_SIZE")
	if sizeStr == "" {
		return DefaultMaxFileSize
	}
	
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		log.Printf("警告: 无法解析MAX_FILE_SIZE环境变量，使用默认值: %d\n", DefaultMaxFileSize)
		return DefaultMaxFileSize
	}
	
	return size
}

// 获取API令牌
func GetAPIToken() string {
	// 首先检查环境变量
	token := os.Getenv("API_TOKEN")
	if token != "" {
		return token
	}
	
	// 然后检查令牌文件
	tokenFile := getTokenFilePath()
	
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		log.Printf("警告: 读取令牌文件失败: %v\n", err)
		return ""
	}
	
	if len(data) == 0 {
		return ""
	}
	
	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		log.Printf("警告: 解析令牌文件失败: %v\n", err)
		return ""
	}
	
	return tokenData.Token
}

// 保存API令牌到文件
func SaveAPIToken(token string) error {
	tokenFile := getTokenFilePath()
	
	// 确保目录存在
	dir := filepath.Dir(tokenFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// 创建令牌数据
	tokenData := TokenData{
		Token: token,
	}
	
	// 序列化为JSON
	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return err
	}
	
	// 写入令牌文件
	err = os.WriteFile(tokenFile, data, 0600)
	if err != nil {
		log.Printf("警告: 保存令牌文件失败: %v\n", err)
		return err
	}
	
	log.Printf("成功保存令牌到: %s\n", tokenFile)
	return nil
}

// 获取令牌文件路径
func getTokenFilePath() string {
	// 首先检查环境变量指定的路径
	tokenPath := os.Getenv("TOKEN_FILE_PATH")
	if tokenPath != "" {
		// 确保环境变量指定的目录存在
		dir := filepath.Dir(tokenPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("警告: 无法创建TOKEN_FILE_PATH指定的目录: %v\n", err)
		} else {
			log.Printf("使用TOKEN_FILE_PATH指定的路径: %s\n", tokenPath)
			return tokenPath
		}
	}
	
	// 环境变量未指定或指定的目录不可用，尝试几个备选目录
	log.Println("尝试寻找可写目录...")
	possibleDirs := []string{
		"/var/lib/r2api", // Docker卷挂载目录
		"/app/config",    // 应用程序配置目录
		"/tmp",           // 临时目录
	}
	
	for _, dir := range possibleDirs {
		// 确保目录存在
		if err := os.MkdirAll(dir, 0755); err == nil {
			// 测试目录是否可写
			testFile := filepath.Join(dir, "write_test")
			if err := os.WriteFile(testFile, []byte("test"), 0600); err == nil {
				os.Remove(testFile) // 删除测试文件
				path := filepath.Join(dir, "token_data.json")
				log.Printf("使用可写目录: %s\n", path)
				return path
			}
		}
	}
	
	// 备选目录都不可用，使用当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v\n", err)
		dir = "."
	}
	
	path := filepath.Join(dir, DefaultTokenFile)
	log.Printf("回退使用工作目录: %s\n", path)
	return path
} 