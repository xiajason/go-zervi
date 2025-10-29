#!/bin/bash

echo "🚀 启动Zervigo MVP微服务集群..."
echo "📋 MVP目标：验证核心价值假设 - 统一认证 + AI能力 + 区块链审计 + 完整业务闭环"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 进入MVP目录
cd "$(dirname "$0")/.."

# 停止现有服务
echo "📦 停止现有服务..."
docker-compose -f docker/docker-compose.yml down

# 构建MVP服务
echo "🔨 构建MVP微服务镜像..."
docker-compose -f docker/docker-compose.yml build --no-cache

# 启动MVP服务
echo "🎯 启动MVP微服务集群..."
docker-compose -f docker/docker-compose.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动完成..."
sleep 30

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker/docker-compose.yml ps

# 健康检查
echo "🏥 执行健康检查..."
echo ""

echo "📊 服务访问地址："
echo "   中央大脑: http://localhost:9000/health"
echo "   认证服务: http://localhost:8207/health"
echo "   AI服务: http://localhost:8100/health"
echo "   区块链服务: http://localhost:8208/health"
echo "   用户服务: http://localhost:8082/health"
echo "   职位服务: http://localhost:8084/health"
echo "   简历服务: http://localhost:8085/health"
echo "   企业服务: http://localhost:8083/health"
echo "   MySQL数据库: localhost:3306"
echo "   Redis缓存: localhost:6379"
echo "   Consul服务发现: http://localhost:8500"
echo ""

echo "🧪 快速验证命令："
echo "   # 中央大脑健康检查"
echo "   curl http://localhost:9000/health"
echo ""
echo "   # 认证服务健康检查"
echo "   curl http://localhost:8207/health"
echo ""
echo "   # AI服务健康检查"
echo "   curl http://localhost:8100/health"
echo ""
echo "   # 区块链服务健康检查"
echo "   curl http://localhost:8208/health"
echo ""

echo "💡 MVP验证重点："
echo "   1. 用户注册登录流程"
echo "   2. AI智能匹配功能"
echo "   3. 区块链数据记录"
echo "   4. 跨服务数据一致性"
echo "   5. 完整业务闭环验证"
echo ""

echo "✅ Zervigo MVP微服务集群启动完成！"
echo "🎯 下一步：验证核心价值假设，收集用户反馈，快速迭代"
