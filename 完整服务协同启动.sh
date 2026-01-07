#!/bin/bash

# 完整服务协同启动脚本
# P0基础设施 + P1核心服务 + P2业务服务 + AI Service

set -e

echo "============================================"
echo "🚀 Zervigo完整服务协同启动"
echo "============================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

# 停止所有旧服务
echo -e "${BLUE}停止所有旧服务...${NC}"
pkill -f "central-brain" 2>/dev/null || true
pkill -f "auth.*8207" 2>/dev/null || true
pkill -f "router-service" 2>/dev/null || true
pkill -f "permission-service" 2>/dev/null || true
pkill -f "user-service" 2>/dev/null || true
pkill -f "ai_service_with_zervigo.py" 2>/dev/null || true
lsof -ti :9000 | xargs kill -9 2>/dev/null || true
lsof -ti :8207 | xargs kill -9 2>/dev/null || true
lsof -ti :8087 | xargs kill -9 2>/dev/null || true
lsof -ti :8086 | xargs kill -9 2>/dev/null || true
lsof -ti :8082 | xargs kill -9 2>/dev/null || true
lsof -ti :8100 | xargs kill -9 2>/dev/null || true
sleep 2
echo -e "${GREEN}✅ 旧服务已停止${NC}"
echo ""

echo "============================================"
echo "📊 P0 基础设施层启动"
echo "============================================"
echo ""

# P0-1: Auth Service (8207)
echo -e "${PURPLE}P0-1: 启动 Auth Service (8207)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/core/auth
DATABASE_URL="postgres://vuecmf:vuecmf@localhost:5432/zervigo_mvp?sslmode=disable" \
JWT_SECRET="zervigo-mvp-secret-key-2025" \
nohup go run main.go > $BASE_DIR/logs/auth-service.log 2>&1 &
AUTH_PID=$!
echo $AUTH_PID > $BASE_DIR/logs/auth-service.pid
echo "Auth Service PID: $AUTH_PID"
sleep 3

if curl -s http://localhost:8207/health > /dev/null; then
    echo -e "${GREEN}✅ Auth Service 启动成功${NC}"
else
    echo -e "${RED}❌ Auth Service 启动失败${NC}"
    tail -20 $BASE_DIR/logs/auth-service.log
    exit 1
fi
echo ""

# P0-2: Central Brain (9000)
echo -e "${PURPLE}P0-2: 启动 Central Brain (9000)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/shared/central-brain
nohup go run . > $BASE_DIR/logs/central-brain.log 2>&1 &
CB_PID=$!
echo $CB_PID > $BASE_DIR/logs/central-brain.pid
echo "Central Brain PID: $CB_PID"
sleep 5

if curl -s http://localhost:9000/health > /dev/null; then
    echo -e "${GREEN}✅ Central Brain 启动成功${NC}"
else
    echo -e "${RED}❌ Central Brain 启动失败${NC}"
    tail -20 $BASE_DIR/logs/central-brain.log
    exit 1
fi
echo ""

echo "============================================"
echo "📊 P1 核心服务层启动"
echo "============================================"
echo ""

# P1-1: Permission Service (8086)
echo -e "${PURPLE}P1-1: 启动 Permission Service (8086)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/infrastructure/permission
nohup go run main.go > $BASE_DIR/logs/permission-service.log 2>&1 &
PERM_PID=$!
echo $PERM_PID > $BASE_DIR/logs/permission-service.pid
echo "Permission Service PID: $PERM_PID"
sleep 3

if curl -s http://localhost:8086/health > /dev/null; then
    echo -e "${GREEN}✅ Permission Service 启动成功${NC}"
else
    echo -e "${RED}❌ Permission Service 启动失败${NC}"
    tail -20 $BASE_DIR/logs/permission-service.log
    exit 1
fi
echo ""

# P1-2: Router Service (8087)
echo -e "${PURPLE}P1-2: 启动 Router Service (8087)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/infrastructure/router
nohup go run main.go > $BASE_DIR/logs/router-service.log 2>&1 &
ROUTER_PID=$!
echo $ROUTER_PID > $BASE_DIR/logs/router-service.pid
echo "Router Service PID: $ROUTER_PID"
sleep 3

if curl -s http://localhost:8087/health > /dev/null; then
    echo -e "${GREEN}✅ Router Service 启动成功${NC}"
else
    echo -e "${RED}❌ Router Service 启动失败${NC}"
    tail -20 $BASE_DIR/logs/router-service.log
    exit 1
fi
echo ""

# P1-3: User Service (8082)
echo -e "${PURPLE}P1-3: 启动 User Service (8082)${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/services/core/user
nohup go run main.go > $BASE_DIR/logs/user-service.log 2>&1 &
USER_PID=$!
echo $USER_PID > $BASE_DIR/logs/user-service.pid
echo "User Service PID: $USER_PID"
sleep 5

if curl -s http://localhost:8082/health > /dev/null; then
    echo -e "${GREEN}✅ User Service 启动成功${NC}"
else
    echo -e "${RED}❌ User Service 启动失败${NC}"
    tail -20 $BASE_DIR/logs/user-service.log
    exit 1
fi
echo ""

echo "============================================"
echo "📊 P2 业务服务层启动（可选）"
echo "============================================"
echo ""

# 检查是否需要启动P2服务
START_P2=${START_P2:-"no"}

if [ "$START_P2" = "yes" ]; then
    # P2-1: Company Service (8083)
    echo -e "${PURPLE}P2-1: 启动 Company Service (8083)${NC}"
    cd $BASE_DIR/services/business/company
    nohup go run main.go > $BASE_DIR/logs/company-service.log 2>&1 &
    COMPANY_PID=$!
    echo $COMPANY_PID > $BASE_DIR/logs/company-service.pid
    sleep 3
    
    if curl -s http://localhost:8083/health > /dev/null; then
        echo -e "${GREEN}✅ Company Service 启动成功${NC}"
    fi
    echo ""
    
    # P2-2: Job Service (8084)
    echo -e "${PURPLE}P2-2: 启动 Job Service (8084)${NC}"
    cd $BASE_DIR/services/business/job
    nohup go run main.go > $BASE_DIR/logs/job-service.log 2>&1 &
    JOB_PID=$!
    echo $JOB_PID > $BASE_DIR/logs/job-service.pid
    sleep 3
    
    if curl -s http://localhost:8084/health > /dev/null; then
        echo -e "${GREEN}✅ Job Service 启动成功${NC}"
    fi
    echo ""
    
    # P2-3: Resume Service (8085)
    echo -e "${PURPLE}P2-3: 启动 Resume Service (8085)${NC}"
    cd $BASE_DIR/services/business/resume
    nohup go run main.go > $BASE_DIR/logs/resume-service.log 2>&1 &
    RESUME_PID=$!
    echo $RESUME_PID > $BASE_DIR/logs/resume-service.pid
    sleep 3
    
    if curl -s http://localhost:8085/health > /dev/null; then
        echo -e "${GREEN}✅ Resume Service 启动成功${NC}"
    fi
    echo ""
else
    echo -e "${YELLOW}⏭  跳过P2业务服务（使用 START_P2=yes 启用）${NC}"
    echo ""
fi

echo "============================================"
echo "📊 P4 AI服务层启动"
echo "============================================"
echo ""

# P4: AI Service (8100)
echo -e "${PURPLE}P4: 启动 AI Service (8100) - Python/Sanic${NC}"
echo "-------------------------------------------"
cd $BASE_DIR/src/ai-service-python

# 激活Python虚拟环境并启动
source venv/bin/activate
nohup python ai_service_with_zervigo.py > $BASE_DIR/logs/ai-service.log 2>&1 &
AI_PID=$!
echo $AI_PID > $BASE_DIR/logs/ai-service.pid
echo "AI Service PID: $AI_PID"
sleep 3

if curl -s http://localhost:8100/health > /dev/null; then
    echo -e "${GREEN}✅ AI Service 启动成功${NC}"
else
    echo -e "${RED}❌ AI Service 启动失败${NC}"
    tail -20 $BASE_DIR/logs/ai-service.log
    exit 1
fi
echo ""

echo "============================================"
echo "📊 服务启动完成总结"
echo "============================================"
echo ""
echo -e "${GREEN}✅ 所有服务已启动！${NC}"
echo ""
echo "🏗️  服务架构："
echo ""
echo -e "${PURPLE}【P0 基础设施层】${NC}"
echo "  • Auth Service:        http://localhost:8207 (PID: $AUTH_PID)"
echo "  • Central Brain:       http://localhost:9000 (PID: $CB_PID)"
echo ""
echo -e "${PURPLE}【P1 核心服务层】${NC}"
echo "  • Permission Service:  http://localhost:8086 (PID: $PERM_PID)"
echo "  • Router Service:      http://localhost:8087 (PID: $ROUTER_PID)"
echo "  • User Service:        http://localhost:8082 (PID: $USER_PID)"
echo ""

if [ "$START_P2" = "yes" ]; then
    echo -e "${PURPLE}【P2 业务服务层】${NC}"
    echo "  • Company Service:     http://localhost:8083 (PID: $COMPANY_PID)"
    echo "  • Job Service:         http://localhost:8084 (PID: $JOB_PID)"
    echo "  • Resume Service:      http://localhost:8085 (PID: $RESUME_PID)"
    echo ""
fi

echo -e "${PURPLE}【P4 AI服务层】${NC}"
echo "  • AI Service:          http://localhost:8100 (PID: $AI_PID)"
echo ""
echo "📝 日志文件："
echo "  tail -f logs/auth-service.log"
echo "  tail -f logs/central-brain.log"
echo "  tail -f logs/permission-service.log"
echo "  tail -f logs/router-service.log"
echo "  tail -f logs/user-service.log"
echo "  tail -f logs/ai-service.log"
echo ""
echo "🧪 测试命令："
echo ""
echo "# 测试登录"
echo 'curl -X POST http://localhost:9000/api/v1/auth/login \'
echo '  -H "Content-Type: application/json" \'
echo '  -d '"'"'{"data":{"login_name":"admin","password":"admin123"}}'"'"' | jq .'
echo ""
echo "# 测试AI聊天（通过Central Brain）"
echo 'curl -X POST http://localhost:9000/api/v1/ai/chat \'
echo '  -H "Content-Type: application/json" \'
echo '  -d '"'"'{"message":"Hello AI"}'"'"' | jq .'
echo ""
echo "# 测试权限服务"
echo 'curl http://localhost:9000/api/v1/permission/roles | jq .'
echo ""
echo "# 测试路由服务"
echo 'curl http://localhost:9000/api/v1/router/routes | jq . | head -50'
echo ""
echo "🛑 停止所有服务："
echo "  pkill -f 'central-brain|auth-service|router-service|permission-service|user-service|ai_service'"
echo ""
echo "============================================"
echo -e "${GREEN}🎉 完整服务链路已启动！${NC}"
echo "============================================"

