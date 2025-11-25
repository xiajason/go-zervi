#!/bin/bash
# Consul服务发现启动脚本
# 用途: 启动Consul服务发现，用于微服务注册和发现

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🔍 启动Consul服务发现...${NC}"

# 检查Consul是否已安装
if ! command -v consul &> /dev/null; then
    echo -e "${RED}❌ Consul未安装${NC}"
    echo ""
    echo "请安装Consul:"
    echo "  macOS: brew install consul"
    echo "  Linux: 请访问 https://www.consul.io/downloads"
    exit 1
fi

# 检查是否已经运行
if pgrep -f "consul agent" > /dev/null; then
    echo "⚠️  Consul已经在运行中"
    echo "   停止现有进程..."
    pkill -f "consul agent" || true
    sleep 2
fi

# 创建Consul数据目录
CONSUL_DATA_DIR="/tmp/consul-data"
mkdir -p "$CONSUL_DATA_DIR"

# 启动Consul (开发模式)
echo "🚀 启动Consul (开发模式)..."
nohup consul agent -dev -ui -client 0.0.0.0 \
    -data-dir="$CONSUL_DATA_DIR" \
    -log-file=/tmp/consul.log \
    -log-level=INFO \
    > /tmp/consul_startup.log 2>&1 &
CONSUL_PID=$!

# 等待启动
echo "⏳ 等待Consul启动..."
sleep 5

# 检查是否启动成功
if ps -p $CONSUL_PID > /dev/null; then
    echo "✅ Consul已启动 (PID: $CONSUL_PID)"
    echo "   📊 日志文件: /tmp/consul.log"
    echo "   🌐 Web UI: http://localhost:8500/ui"
    echo "   🔌 API地址: http://localhost:8500"
    
    # 等待一下然后测试
    sleep 2
    if curl -s http://localhost:8500/v1/status/leader > /dev/null 2>&1; then
        echo "✅ Consul服务正常"
        
        # 显示Consul状态
        echo ""
        echo "📋 Consul状态:"
        curl -s http://localhost:8500/v1/status/leader | head -1
        echo ""
        
        echo "✅ Consul启动完成！"
    else
        echo "⚠️  Consul API测试失败，请查看日志: tail -f /tmp/consul.log"
    fi
else
    echo "❌ Consul启动失败"
    echo "   查看日志: tail -f /tmp/consul.log"
    exit 1
fi

echo ""
echo "💡 使用提示:"
echo "   查看服务列表: curl http://localhost:8500/v1/agent/services"
echo "   查看服务健康: curl http://localhost:8500/v1/health/service/auth-service"
echo "   停止Consul: pkill -f consul"
echo ""
