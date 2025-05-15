FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod .
COPY go.sum .

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o r2api ./cmd/main.go

# 使用小型镜像
FROM alpine:latest

WORKDIR /root/

# 从builder阶段复制编译好的应用
COPY --from=builder /app/r2api .

# 设置权限
RUN chmod +x ./r2api

# 暴露端口
EXPOSE 3009

# 设置启动命令
CMD ["./r2api"] 