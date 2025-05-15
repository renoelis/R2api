# R2API - Cloudflare R2 文件上传服务

R2API 是一个基于Go语言开发的RESTful API服务，用于将URL或本地文件上传到Cloudflare R2存储桶。服务采用了Go语言和Gin框架，提供高性能、低资源占用的特性。

## 功能特点

- 从URL下载文件并上传到R2存储桶
- 直接上传本地文件到R2存储桶
- 支持自定义域名用于生成返回URL
- 简单的API令牌认证
- 首次部署自动生成API令牌
- 支持API令牌重置功能

## 安装与运行

### 前提条件

- Go 1.21+（开发时需要）
- Docker 和 Docker Compose（部署时需要）
- Cloudflare R2账户

### 本地开发

1. 克隆仓库
```bash
git clone https://your-repo-url/r2api.git
cd r2api
```

2. 安装依赖
```bash
go mod tidy
```

3. 运行服务
```bash
go run cmd/main.go
```

服务将在 `http://localhost:3009` 启动。

### Docker部署

1. 构建Docker镜像
```bash
docker-compose -f docker-compose-r2api.yml build
```

2. 启动服务
```bash
docker-compose -f docker-compose-r2api.yml up -d
```

3. 查看日志
```bash
docker-compose -f docker-compose-r2api.yml logs -f
```

## API文档

### 令牌管理

#### 初始化令牌

首次部署时，服务会自动生成一个默认API令牌。如果令牌丢失，您可以通过以下接口初始化：

```
POST /R2api/init-token
```

响应:
```json
{
  "status": "success",
  "message": "API令牌已初始化",
  "token": "your-api-token"
}
```

**注意**：此接口只在令牌不存在时才会生成新令牌。如果令牌已存在，将返回现有令牌。

#### 重置令牌

如果需要强制重置令牌（如令牌泄露时），可以使用以下接口：

```
POST /R2api/reset-token
```

响应:
```json
{
  "status": "success",
  "message": "API令牌已成功重置",
  "token": "new-api-token"
}
```

**注意**：此接口会强制生成新令牌，无论是否已存在旧令牌。使用此接口后，旧令牌将立即失效。

#### 获取令牌信息

```
GET /R2api/token
```

响应:
```json
{
  "status": "success",
  "token": "********", 
  "is_permanent": true
}
```

### 文件上传

#### 从URL上传文件

```
POST /R2api/upload
Authorization: Bearer your-api-token
Content-Type: application/json

{
  "fileUrl": "https://example.com/image.jpg",
  "bucketName": "your-bucket",
  "objectKey": "images/image.jpg",
  "endpoint": "https://xxx.r2.cloudflarestorage.com",
  "accessKeyId": "your-access-key",
  "secretAccessKey": "your-secret-key",
  "customdomain": "https://your-custom-domain.com" // 可选
}
```

响应:
```json
{
  "status": "success",
  "message": "文件上传成功",
  "data": {
    "public_url": "https://your-custom-domain.com/images/image.jpg",
    "size": 1024,
    "content_type": "image/jpeg",
    "file_name": "image.jpg"
  }
}
```

#### 直接上传文件

```
POST /R2api/upload-direct
Authorization: Bearer your-api-token
Content-Type: multipart/form-data

- file: (binary)
- bucket_name: your-bucket
- object_key: images/image.jpg (可选，如不提供则使用原文件名)
- endpoint: https://xxx.r2.cloudflarestorage.com
- access_key_id: your-access-key
- secret_access_key: your-secret-key
- custom_domain: https://your-custom-domain.com (可选)
```

响应:
```json
{
  "status": "success",
  "message": "文件上传成功",
  "data": {
    "public_url": "https://your-custom-domain.com/images/image.jpg",
    "size": 1024,
    "content_type": "image/jpeg",
    "file_name": "image.jpg"
  }
}
```

## 配置参数

通过环境变量配置:

- `PORT`: 服务监听端口 (默认: 3009)
- `MAX_FILE_SIZE`: 最大文件大小限制，单位字节 (默认: 200MB)
- `API_TOKEN`: 可以直接通过环境变量设置API令牌
- `TOKEN_FILE_PATH`: 自定义令牌文件路径 (默认: /var/lib/r2api/token_data.json)

## 数据持久化

令牌数据存储在Docker命名卷中，确保数据在容器重启后仍然保留：

```yaml
volumes:
  - r2api_data:/var/lib/r2api
```

查看存储的令牌：

```bash
# 在容器内查看
docker exec -it r2api cat /var/lib/r2api/token_data.json

# 或使用临时容器
docker run --rm -it -v r2api_r2api_data:/data alpine cat /data/token_data.json
```

## 注意事项

- 首次部署时自动生成的令牌存储在JSON文件中，通过Docker卷持久化
- 所有API请求需要通过Bearer令牌认证 (除了 `/health`, `/`, `/R2api/token`, `/R2api/init-token` 和 `/R2api/reset-token`) 