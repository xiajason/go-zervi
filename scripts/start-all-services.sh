#!/bin/bash

# Zervigo 微服务架构启动脚本

echo "🚀 启动 Zervigo 微服务架构"
echo "================================"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 创建必要的目录
echo "📁 创建微服务目录..."
mkdir -p src/microservices/{api-gateway,basic-server,user-service,company-service,job-service,resume-service,notification-service,banner-service,template-service,statistics-service,dev-team-service,multi-database-service}/{tmp,logs,uploads}

# 启动基础设施
echo "🗄️  启动基础设施服务..."
docker-compose -f docker-compose.microservices.yml up -d mysql postgres redis

# 等待数据库启动
echo "⏳ 等待数据库启动..."
sleep 15

# 检查数据库健康状态
echo "🔍 检查数据库健康状态..."
docker-compose -f docker-compose.microservices.yml ps

# 启动核心服务
echo "🔧 启动核心服务..."
docker-compose -f docker-compose.microservices.yml up -d auth-service api-gateway basic-server

# 等待核心服务启动
echo "⏳ 等待核心服务启动..."
sleep 10

# 启动业务微服务
echo "🏢 启动业务微服务..."
docker-compose -f docker-compose.microservices.yml up -d user-service company-service job-service resume-service

# 启动支持服务
echo "📋 启动支持服务..."
docker-compose -f docker-compose.microservices.yml up -d notification-service banner-service template-service statistics-service

# 启动专业服务
echo "👨‍💻 启动专业服务..."
docker-compose -f docker-compose.microservices.yml up -d dev-team-service multi-database-service

# 启动AI服务
echo "🤖 启动AI服务..."
docker-compose -f docker-compose.microservices.yml up -d ai-service

# 等待所有服务启动
echo "⏳ 等待所有服务启动..."
sleep 20

# 检查服务状态
echo "📊 微服务状态检查:"
echo "================================"

# 检查核心服务
echo "🔐 核心服务:"
curl -s http://localhost:8207/health | jq . || echo "❌ 认证服务未响应"
curl -s http://localhost:8080/health | jq . || echo "❌ API网关未响应"
curl -s http://localhost:8081/health | jq . || echo "❌ 基础服务未响应"

echo ""
echo "🏢 业务服务:"
curl -s http://localhost:8082/health | jq . || echo "❌ 用户服务未响应"
curl -s http://localhost:8083/health | jq . || echo "❌ 公司服务未响应"
curl -s http://localhost:8084/health | jq . || echo "❌ 职位服务未响应"
curl -s http://localhost:8085/health | jq . || echo "❌ 简历服务未响应"

echo ""
echo "📋 支持服务:"
curl -s http://localhost:8086/health | jq . || echo "❌ 通知服务未响应"
curl -s http://localhost:8087/health | jq . || echo "❌ 横幅服务未响应"
curl -s http://localhost:8088/health | jq . || echo "❌ 模板服务未响应"
curl -s http://localhost:8089/health | jq . || echo "❌ 统计服务未响应"

echo ""
echo "👨‍💻 专业服务:"
curl -s http://localhost:8090/health | jq . || echo "❌ 开发团队服务未响应"
curl -s http://localhost:8091/health | jq . || echo "❌ 多数据库服务未响应"

echo ""
echo "🤖 AI服务:"
curl -s http://localhost:8100/health | jq . || echo "❌ AI服务未响应"

echo ""
echo "🎉 微服务架构启动完成！"
echo "================================"
echo "API网关:     http://localhost:8080"
echo "认证服务:    http://localhost:8207"
echo "基础服务:    http://localhost:8081"
echo "用户服务:    http://localhost:8082"
echo "公司服务:    http://localhost:8083"
echo "职位服务:    http://localhost:8084"
echo "简历服务:    http://localhost:8085"
echo "通知服务:    http://localhost:8086"
echo "横幅服务:    http://localhost:8087"
echo "模板服务:    http://localhost:8088"
echo "统计服务:    http://localhost:8089"
echo "开发团队服务: http://localhost:8090"
echo "多数据库服务: http://localhost:8091"
echo "AI服务:      http://localhost:8100"
echo ""
echo "📝 管理命令:"
echo "查看日志: docker-compose -f docker-compose.microservices.yml logs -f"
echo "停止服务: docker-compose -f docker-compose.microservices.yml down"
echo "重启服务: docker-compose -f docker-compose.microservices.yml restart"
echo "查看状态: docker-compose -f docker-compose.microservices.yml ps"
