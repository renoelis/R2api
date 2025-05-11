# R2 文件上传服务

这是一个用于将文件上传到Cloudflare R2存储桶的API服务。

## 功能特点

- **多种上传方式**：
  - 从URL下载文件并上传到R2存储桶
  - 直接上传本地文件到R2存储桶（form-data格式）
- **灵活的Token管理**：
  - 创建普通有效期Token或永久有效Token
  - 续期Token或修改Token有效期
  - 支持将普通Token转为永久Token，或将永久Token转为普通Token
- **其他特性**：
  - 支持文件大小限制检查（最大200MB）
  - 支持自定义域名
  - 支持文件夹路径
  - 内置重试机制，提高稳定性
  - 与轻流平台集成的Token管理

## 技术栈

- FastAPI + Gunicorn + UvicornWorker
- Httpx 用于异步HTTP请求
- Boto3 用于S3/R2交互
- Docker 用于部署

## 安装与运行

### 本地开发

1. 克隆此仓库
2. 创建并激活虚拟环境

```bash
python3 -m venv venv
source venv/bin/activate  # Linux/macOS
# 或
venv\Scripts\activate  # Windows
```

3. 安装依赖

```bash
pip install -r requirements.txt
```

4. 运行开发服务器

```bash
uvicorn app.main:app --reload --port 3009
```

### Docker部署

1. 构建Docker镜像

```bash
docker build -t r2-uploader:latest .
```

2. 使用docker-compose运行

```bash
docker-compose -f docker-compose-r2-uploader.yml up -d
```

## API文档

启动服务后，访问 http://localhost:3009/docs 查看完整的API文档。

### 主要端点

#### 1. 注册Token

```
POST /R2api/register
```

请求体:
```json
{
  "username": "用户名",
  "email": "邮箱@example.com",
  "expires_in_days": 30  // 可选，默认30天，-99表示永久有效
}
```

响应:
```json
{
  "status": "success",
  "token": "generated_api_token",
  "expires_at": "2023-12-31 10:30:45",  // 如果不是永久有效
  "is_permanent": false
}
```

#### 2. 续期Token

```
POST /R2api/renew
```

请求体:
```json
{
  "token": "your_token_here",
  "extend_days": 30  // -99表示设为永久有效
}
```

响应:
```json
{
  "status": "success",
  "message": "Token有效期已延长",
  "token": "your_token_here",
  "old_status": "有效期至 2023-12-01 10:30:45",
  "new_expires_at": "2023-12-31 10:30:45",
  "extended_days": 30,
  "is_permanent": false
}
```

#### 3. URL文件上传

```
POST /R2api/upload
```

请求头:
```
Authorization: Bearer {token}
```

请求体:
```json
{
  "fileUrl": "https://example.com/image.jpg",
  "bucketName": "my-bucket",
  "objectKey": "folder/image.jpg",
  "endpoint": "https://xxx.r2.cloudflarestorage.com",
  "accessKeyId": "your_access_key",
  "secretAccessKey": "your_secret_key",
  "customdomain": "https://bucket.example.com"  // 可选
}
```

响应:
```json
{
  "status": "success",
  "message": "文件上传成功",
  "data": {
    "public_url": "https://bucket.example.com/folder/image.jpg",
    "size": 1024000,
    "content_type": "image/jpeg"
  }
}
```

#### 4. 直接文件上传

```
POST /R2api/upload-direct
```

请求头:
```
Authorization: Bearer {token}
Content-Type: multipart/form-data
```

表单字段:
```
file: 要上传的文件（必需）
bucket_name: 存储桶名称（必需）
object_key: 对象键名，包含路径（可选，默认使用原文件名）
endpoint: R2存储桶端点URL（必需）
access_key_id: 访问密钥ID（必需）
secret_access_key: 访问密钥（必需）
custom_domain: 自定义域名（可选）
```

响应:
```json
{
  "status": "success",
  "message": "文件上传成功",
  "data": {
    "public_url": "https://bucket.example.com/folder/image.jpg",
    "size": 1024000,
    "content_type": "image/jpeg",
    "file_name": "image.jpg"
  }
}
```

## 使用示例

### 注册Token

```bash
curl -X POST http://localhost:3009/R2api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","expires_in_days":30}'
```

### 续期Token

```bash
curl -X POST http://localhost:3009/R2api/renew \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token_here",
    "extend_days": 30
  }'
```

### 将Token设为永久有效

```bash
curl -X POST http://localhost:3009/R2api/renew \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token_here",
    "extend_days": -99
  }'
```

### 通过URL上传文件

```bash
curl -X POST http://localhost:3009/R2api/upload \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "fileUrl": "https://example.com/image.jpg",
    "bucketName": "my-bucket",
    "objectKey": "folder/image.jpg",
    "endpoint": "https://xxx.r2.cloudflarestorage.com",
    "accessKeyId": "your_access_key",
    "secretAccessKey": "your_secret_key",
    "customdomain": "https://bucket.example.com"
  }'
```

### 直接上传文件

```bash
curl -X POST http://localhost:3009/R2api/upload-direct \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/your/file.jpg" \
  -F "bucket_name=my-bucket" \
  -F "object_key=folder/image.jpg" \
  -F "endpoint=https://xxx.r2.cloudflarestorage.com" \
  -F "access_key_id=your_access_key" \
  -F "secret_access_key=your_secret_key" \
  -F "custom_domain=https://bucket.example.com"
```

## 注意事项

- 文件大小限制为200MB
- 确保您的R2存储桶已正确配置权限
- Token一旦生成无法修改用户名和邮箱，只能通过API修改其有效期
- 上传文件时如不指定`objectKey`，直接上传会使用原始文件名
- 系统内置重试机制，提高网络波动时的稳定性