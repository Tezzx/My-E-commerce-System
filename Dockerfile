# 阶段 1：构建阶段
FROM golang:1.25-alpine AS builder

# 设置环境变量，启用 Go Modules 并关闭 CGO (静态编译)
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 设置工作目录
WORKDIR /app

# 下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝项目源代码
COPY . .

# 编译 Go 项目并输出为 order-payment-system 二进制文件
RUN go build -o /app/order-payment-system ./cmd/main.go

# 阶段 2：运行阶段
FROM alpine:latest

WORKDIR /app

# 设置时区和安装根证书（供支付宝HTTPS请求等使用）
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 从 builder 阶段把可执行文件、配置文件以及静态文件拷贝过来
COPY --from=builder /app/order-payment-system .
# 初始把配置文件复制进去
COPY --from=builder /app/config ./config 
COPY --from=builder /app/templates ./templates

# 暴露服务端口
EXPOSE 8080

# 运行应用
ENTRYPOINT ["./order-payment-system"]