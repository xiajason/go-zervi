#!/bin/bash

# Router Service 和 Permission Service 启动脚本
# 用法: ./scripts/start-router-permission.sh [router-mode]
# router-mode: standalone (默认) 或 database

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Router & Permission Service 启动${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# 检查配置文件
if [ ! -f "$PROJECT_ROOT/configs/local.env" ]; then
    echo -e "${RED}❌ 配置文件不存在: configs/local.env${NC}"
    exit 1
fi

# 加载环境变量
echo -e "${GREEN}📋 加载配置文件: configs/local.env${NC}"
export $(cat "$PROJECT_ROOT/configs/local.env" | grep "^[^#]" | grep -v "^$" | xargs)

# 获取Router Service启动模式
ROUTER_MODE=${1:-standalone}

echo -e "${GREEN}🔧 Router Service 模式: $ROUTER_MODE${NC}"
echo ""

# 停止现有服务
echo -e "${YELLOW}🛑 停止现有服务...${NC}"
pkill -f router-service 2>/dev/null || true
pkill -f permission-service 2>/dev/null || true
lsof -ti:8087 | xargs kill -9 2>/dev/null || true
lsof -ti:8086 | xargs kill -9 2>/dev/null || true
sleep 1

# 启动 Router Service
echo -e "${GREEN}🚀 启动 Router Service (模式: $ROUTER_MODE)...${NC}"
cd "$PROJECT_ROOT/services/infrastructure/router"

if [ "$ROUTER_MODE" = "standalone" ]; then
    echo "  使用 standalone 模式 (不需要数据库)"
    nohup go run standalone_main.go > "$PROJECT_ROOT/logs/router-service.log" 2>&1 &
else
    echo "  使用 database 模式 (需要数据库)"
    nohup go run main.go > "$PROJECT_ROOT/logs/router-service.log" 2>&1 &
fi

echo "  PID: $!"
echo "  日志: logs/router-service.log"

sleep 3

# 启动 Permission Service
echo ""
echo -e "${GREEN}🚀 启动 Permission Service...${NC}"
cd "$PROJECT_ROOT/services/infrastructure/permission"
echo "  使用 database 模式 (必须使用数据库)"

# 重新加载环境变量（确保在permission目录中）
export $(cat "$PROJECT_ROOT/configs/local.env" | grep "^[^#]" | grep -v "^$" | xargs)
nohup go run main.go > "$PROJECT_ROOT/logs/permission-service.log" 2>&1 &

echo "  PID: $!"
echo "  日志: logs/permission-service.log"

sleep 5

# 检查服务状态
echo ""
echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}服务状态检查${NC}"
echo -e "${BLUE}================================${NC}"

# 检查 Router Service
echo -e "\n${YELLOW}1. Router Service${NC}"
if curl -s http://localhost:8087/health > /dev/null 2>&1; then
    ROUTER_STATUS=$(curl -s http://localhost:8087/health | jq -r '.status // .service' 2>/dev/null)
    echo -e "${GREEN}✅ Router Service 运行正常 (状态: $ROUTER_STATUS)${NC}"
    echo "   URL: http://localhost:8087/health"
else
    echo -e "${RED}❌ Router Service 未响应${NC}"
    echo "   查看日志: tail -f logs/router-service.log"
fi

# 检查 Permission Service
echo -e "\n${YELLOW}2. Permission Service${NC}"
if curl -s http://localhost:8086/health > /dev/null 2>&1; then
    PERM_STATUS=$(curl -s http://localhost:8086/health | jq -r '.status // .service' 2>/dev/null)
    DB_INFO=$(curl -s http://localhost:8086/health | jq -r '.core_health.database.postgresql | "\(.database)@\(.host):\(.port)"' 2>/dev/null)
    echo -e "${GREEN}✅ Permission Service 运行正常 (状态: $PERM_STATUS)${NC}"
    echo "   URL: http://localhost:8086/health"
    echo "   数据库: $DB_INFO"
else
    echo -e "${RED}❌ Permission Service 未响应${NC}"
    echo "   查看日志: tail -f logs/permission-service.log"
fi

# 显示数据库配置
echo -e "\n${YELLOW}📊 数据库配置${NC}"
echo "  数据库: $POSTGRESQL_DATABASE"
echo "  主机: $POSTGRESQL_HOST"
echo "  端口: $POSTGRESQL_PORT"
echo "  用户: $POSTGRESQL_USER"

# 测试 API
echo -e "\n${BLUE}================================${NC}"
echo -e "${BLUE}API 测试${NC}"
echo -e "${BLUE}================================${NC}"

# 测试 Router Service API
echo -e "\n${YELLOW}测试 Router Service API:${NC}"
echo "  获取路由配置:"
curl -s http://localhost:8087/api/v1/router/routes | jq -r 'if .code == 0 then "✅ 路由配置获取成功 (\(.data | length) 个路由)" else "❌ 路由配置获取失败: \(.message)" end' 2>/dev/null || echo "  ⚠️ 无法访问 API"

# 测试 Permission Service API
echo -e "\n${YELLOW}测试 Permission Service API:${NC}"
echo "  获取角色列表:"
curl -s http://localhost:8086/api/v1/roles | jq -r 'if .code == 0 then "✅ 角色列表获取成功" else "❌ 角色列表获取失败: \(.message)" end' 2>/dev/null || echo "  ⚠️ 无法访问 API"

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}启动完成！${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "📝 日志位置:"
echo "  - Router Service: logs/router-service.log"
echo "  - Permission Service: logs/permission-service.log"
echo ""
echo "🔍 查看日志:"
echo "  tail -f logs/router-service.log"
echo "  tail -f logs/permission-service.log"
echo ""
echo "🛑 停止服务:"
echo "  ./scripts/stop-router-permission.sh"
echo ""

