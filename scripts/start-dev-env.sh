#!/bin/bash

# Zervigo 二次开发环境启动脚本

echo "🚀 启动 Zervigo 二次开发环境"
echo "================================"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 创建必要的目录
echo "📁 创建开发目录..."
mkdir -p src/auth-service-go/tmp
mkdir -p src/ai-service-python/logs

# 启动数据库集群
echo "🗄️  启动数据库集群..."
docker-compose -f docker-compose.dev.yml up -d mysql postgres redis

# 等待数据库启动
echo "⏳ 等待数据库启动..."
sleep 10

# 检查数据库健康状态
echo "🔍 检查数据库健康状态..."
docker-compose -f docker-compose.dev.yml ps

# 启动开发服务
echo "🔧 启动开发服务..."
docker-compose -f docker-compose.dev.yml up -d auth-service-dev ai-service-dev

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 检查服务状态
echo "📊 服务状态检查:"
echo "认证服务 (Go):"
curl -s http://localhost:8207/health | jq . || echo "认证服务未响应"

echo ""
echo "AI服务 (Python):"
curl -s http://localhost:8100/health | jq . || echo "AI服务未响应"

echo ""
echo "🎉 开发环境启动完成！"
echo "================================"
echo "认证服务: http://localhost:8207"
echo "AI服务:   http://localhost:8100"
echo "MySQL:    localhost:3306"
echo "PostgreSQL: localhost:5432"
echo "Redis:    localhost:6379"
echo ""
echo "📝 开发命令:"
echo "查看日志: docker-compose -f docker-compose.dev.yml logs -f"
echo "停止服务: docker-compose -f docker-compose.dev.yml down"
echo "重启服务: docker-compose -f docker-compose.dev.yml restart"
