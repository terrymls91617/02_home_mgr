# 构建阶段
FROM docker.m.daocloud.io/library/golang:1.23-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 配置国内镜像
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖
RUN go mod download

# 复制所有文件
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o home-mgr main.go

# 运行阶段
FROM docker.m.daocloud.io/library/alpine:latest

# 安装必要的包
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/home-mgr .

# 暴露端口
EXPOSE 8000

# 启动服务
CMD ["./home-mgr"]
