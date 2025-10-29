#!/bin/bash

echo "🛑 停止Zervigo MVP微服务集群..."

# 进入MVP目录
cd "$(dirname "$0")/.."

# 停止所有服务
echo "📦 停止所有服务..."
docker-compose -f docker/docker-compose.yml down

# 清理容器
echo "🧹 清理容器..."
docker-compose -f docker/docker-compose.yml rm -f

# 清理镜像（可选）
if [ "$1" = "--clean-images" ]; then
    echo "🗑️ 清理镜像..."
    docker-compose -f docker/docker-compose.yml down --rmi all
fi

# 清理数据卷（可选）
if [ "$1" = "--clean-volumes" ]; then
    echo "🗑️ 清理数据卷..."
    docker-compose -f docker/docker-compose.yml down -v
fi

echo "✅ Zervigo MVP微服务集群已停止"
echo ""
echo "💡 使用说明："
echo "   ./scripts/stop-mvp.sh                    # 停止服务"
echo "   ./scripts/stop-mvp.sh --clean-images     # 停止服务并清理镜像"
echo "   ./scripts/stop-mvp.sh --clean-volumes     # 停止服务并清理数据卷"
