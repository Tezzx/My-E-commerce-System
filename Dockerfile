# ---- 构建阶段 ----
FROM golang:1.25-alpine AS builder

# 设置 Go 代理，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译为静态链接的二进制文件
RUN go build -ldflags="-s -w" -o /app/server ./cmd/main.go

# ---- 运行阶段 ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
COPY --from=builder /app/templates ./templates

EXPOSE 8081

CMD ["./server"]
