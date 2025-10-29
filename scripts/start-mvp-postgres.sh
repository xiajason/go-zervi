#!/bin/bash

echo "🚀 启动Zervigo MVP微服务集群 (PostgreSQL版本)..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 停止并移除旧服务
echo "📦 停止并移除现有MVP服务..."
docker-compose -f docker/docker-compose-postgres.yml down --remove-orphans

# 构建服务
echo "🔨 构建MVP微服务镜像..."
docker-compose -f docker/docker-compose-postgres.yml build --no-cache

# 启动服务
echo "🎯 启动MVP微服务集群..."
docker-compose -f docker/docker-compose-postgres.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动完成 (30秒)..."
sleep 30

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker/docker-compose-postgres.yml ps

echo ""
echo "✅ Zervigo MVP微服务集群启动完成！"
echo ""
echo "📊 服务访问地址："
echo "   中央大脑 (API Gateway): http://localhost:9000"
echo "   统一认证服务: http://localhost:8207"
echo "   AI服务: http://localhost:8100"
echo "   区块链服务: http://localhost:8208"
echo "   用户服务: http://localhost:8082"
echo "   职位服务: http://localhost:8084"
echo "   简历服务: http://localhost:8085"
echo "   企业服务: http://localhost:8083"
echo "   PostgreSQL: localhost:5432"
echo "   Redis: localhost:6379"
echo "   Consul UI: http://localhost:8500/ui"
echo ""
echo "💡 提示："
echo "   - 数据库: postgres://postgres:dev_password@localhost:5432/zervigo_mvp"
echo "   - 默认管理员: admin / admin123"
echo "   - 可通过中央大脑测试接口连通性"
