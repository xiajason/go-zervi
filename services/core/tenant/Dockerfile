# GoZervi Tenant Service Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./
COPY shared/core/go.mod shared/core/go.sum ./shared/core/
COPY services/core/tenant/go.mod services/core/tenant/go.sum ./services/core/tenant/

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
WORKDIR /app/services/core/tenant
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tenant-service .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/services/core/tenant/tenant-service .

EXPOSE 8088

CMD ["./tenant-service"]




