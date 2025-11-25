#!/bin/bash

# 启动P1核心服务
# Router Service (8087)
# Permission Service (8086)  
# User Service (8082)

set -e

echo "============================================"
echo "🚀 启动P1核心服务"
echo "============================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 工作目录
BASE_DIR="/Users/szjason72/gozervi/zervigo.demo"
cd $BASE_DIR

# 加载环境变量
if [ -f "configs/local.env" ]; then
    echo -e "${BLUE}加载环境变量...${NC}"
    set -a
    source configs/local.env
    set +a
    echo -e "${GREEN}✅ 环境变量已加载${NC}"
else
    echo -e "${RED}❌ 找不到 configs/local.env${NC}"
    exit 1
fi

# 确保日志目录存在
mkdir -p logs

# 停止旧进程
echo -e "${BLUE}停止旧服务...${NC}"
pkill -f "router-service" 2>/dev/null || true
pkill -f "permission-service" 2>/dev/null || true
pkill -f "user-service" 2>/dev/null || true
lsof -ti :8087 | xargs kill -9 2>/dev/null || true
lsof -ti :8086 | xargs kill -9 2>/dev/null || true
lsof -ti :8082 | xargs kill -9 2>/dev/null || true
sleep 2
echo -e "${GREEN}✅ 旧服务已停止${NC}"
echo ""

# 步骤 1: 启动 Permission Service (8086)
echo -e "${BLUE}步骤 1/3: 启动 Permission Service (8086)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/infrastructure/permission
nohup go run main.go > $BASE_DIR/logs/permission-service.log 2>&1 &
PERM_PID=$!
echo $PERM_PID > $BASE_DIR/logs/permission-service.pid
echo "Permission Service PID: $PERM_PID"
sleep 3

# 验证
if curl -s http://localhost:8086/health > /dev/null; then
    echo -e "${GREEN}✅ Permission Service 启动成功${NC}"
else
    echo -e "${RED}❌ Permission Service 启动失败${NC}"
    echo "查看日志: tail -50 $BASE_DIR/logs/permission-service.log"
    exit 1
fi
echo ""

# 步骤 2: 启动 Router Service (8087)
echo -e "${BLUE}步骤 2/3: 启动 Router Service (8087)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/infrastructure/router
nohup go run main.go > $BASE_DIR/logs/router-service.log 2>&1 &
ROUTER_PID=$!
echo $ROUTER_PID > $BASE_DIR/logs/router-service.pid
echo "Router Service PID: $ROUTER_PID"
sleep 3

# 验证
if curl -s http://localhost:8087/health > /dev/null; then
    echo -e "${GREEN}✅ Router Service 启动成功${NC}"
else
    echo -e "${RED}❌ Router Service 启动失败${NC}"
    echo "查看日志: tail -50 $BASE_DIR/logs/router-service.log"
    exit 1
fi
echo ""

# 步骤 3: 启动 User Service (8082)
echo -e "${BLUE}步骤 3/3: 启动 User Service (8082)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/core/user
nohup go run main.go > $BASE_DIR/logs/user-service.log 2>&1 &
USER_PID=$!
echo $USER_PID > $BASE_DIR/logs/user-service.pid
echo "User Service PID: $USER_PID"
sleep 3

# 验证
if curl -s http://localhost:8082/health > /dev/null; then
    echo -e "${GREEN}✅ User Service 启动成功${NC}"
else
    echo -e "${RED}❌ User Service 启动失败${NC}"
    echo "查看日志: tail -50 $BASE_DIR/logs/user-service.log"
    exit 1
fi
echo ""

# 服务总结
echo -e "${BLUE}服务信息总结${NC}"
echo "============================================"
echo ""
echo -e "${GREEN}✅ P1核心服务已全部启动！${NC}"
echo ""
echo "📊 服务列表："
echo "  • Permission Service:  http://localhost:8086 (PID: $PERM_PID)"
echo "  • Router Service:      http://localhost:8087 (PID: $ROUTER_PID)"
echo "  • User Service:        http://localhost:8082 (PID: $USER_PID)"
echo ""
echo "📝 日志文件："
echo "  • Permission Service:  tail -f logs/permission-service.log"
echo "  • Router Service:      tail -f logs/router-service.log"
echo "  • User Service:        tail -f logs/user-service.log"
echo ""
echo "🛑 停止服务："
echo "  • kill \$(cat logs/permission-service.pid)"
echo "  • kill \$(cat logs/router-service.pid)"
echo "  • kill \$(cat logs/user-service.pid)"
echo ""
echo "🔍 测试服务："
echo "  curl http://localhost:8086/health  # Permission Service"
echo "  curl http://localhost:8087/health  # Router Service"
echo "  curl http://localhost:8082/health  # User Service"
echo ""
echo "============================================"
echo -e "${GREEN}🎉 启动完成！${NC}"
echo "============================================"

