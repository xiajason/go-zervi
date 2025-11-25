#!/bin/bash
# Central Brain 启动脚本（支持环境变量配置）
# 用途: 加载环境变量并启动Central Brain服务

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🧠 启动Zervigo中央大脑 (API Gateway)...${NC}"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 检测环境配置文件
ENV_FILE=""
if [ -f "$PROJECT_ROOT/configs/local.env" ]; then
    ENV_FILE="$PROJECT_ROOT/configs/local.env"
    echo -e "${GREEN}📋 使用配置文件: configs/local.env${NC}"
elif [ -f "$PROJECT_ROOT/configs/dev.env" ]; then
    ENV_FILE="$PROJECT_ROOT/configs/dev.env"
    echo -e "${GREEN}📋 使用配置文件: configs/dev.env${NC}"
elif [ -f "$PROJECT_ROOT/.env" ]; then
    ENV_FILE="$PROJECT_ROOT/.env"
    echo -e "${GREEN}📋 使用配置文件: .env${NC}"
fi

# 加载环境变量
if [ -n "$ENV_FILE" ]; then
    echo -e "${YELLOW}📂 加载环境变量: $ENV_FILE${NC}"
    set -a  # 自动导出所有变量
    source "$ENV_FILE"
    set +a
else
    echo -e "${YELLOW}⚠️  未找到环境配置文件，使用默认配置${NC}"
fi

# 检查必需的环境变量
if [ -z "$SERVICE_SECRET" ]; then
    echo -e "${YELLOW}⚠️  SERVICE_SECRET未设置，使用默认值${NC}"
    export SERVICE_SECRET="${SERVICE_SECRET:-central-brain-secret-2025}"
fi

if [ -z "$SERVICE_ID" ]; then
    export SERVICE_ID="${SERVICE_ID:-central-brain}"
fi

if [ -z "$SERVICE_HOST" ]; then
    export SERVICE_HOST="${SERVICE_HOST:-localhost}"
fi

if [ -z "$CENTRAL_BRAIN_PORT" ]; then
    export CENTRAL_BRAIN_PORT="${CENTRAL_BRAIN_PORT:-9000}"
fi

if [ -z "$AUTH_SERVICE_PORT" ]; then
    export AUTH_SERVICE_PORT="${AUTH_SERVICE_PORT:-8207}"
fi

# 显示配置信息
echo ""
echo -e "${GREEN}📊 配置信息:${NC}"
echo "  服务ID: ${SERVICE_ID}"
echo "  服务主机: ${SERVICE_HOST}"
echo "  Central Brain端口: ${CENTRAL_BRAIN_PORT}"
echo "  Auth Service端口: ${AUTH_SERVICE_PORT}"
echo "  服务发现: ${SERVICE_DISCOVERY_ENABLED:-false}"
echo "  Consul URL: ${CONSUL_AGENT_URL:-http://localhost:8500}"
echo ""

# 进入Central Brain目录
CENTRAL_BRAIN_DIR="$PROJECT_ROOT/shared/central-brain"

if [ ! -d "$CENTRAL_BRAIN_DIR" ]; then
    echo -e "${RED}❌ Central Brain目录不存在: $CENTRAL_BRAIN_DIR${NC}"
    exit 1
fi

cd "$CENTRAL_BRAIN_DIR"

# 检查是否已经运行
if pgrep -f "central-brain" > /dev/null; then
    echo -e "${YELLOW}⚠️  Central Brain已经在运行中${NC}"
    echo "   停止现有进程..."
    pkill -f "central-brain" || true
    sleep 2
fi

# 构建Central Brain
echo -e "${GREEN}🔨 构建Central Brain...${NC}"
if go build -o central-brain *.go 2>&1 | tee /tmp/cb_build.log; then
    echo -e "${GREEN}✅ 构建成功${NC}"
else
    echo -e "${RED}❌ 构建失败，查看错误:${NC}"
    cat /tmp/cb_build.log
    exit 1
fi

# 检查构建产物
if [ ! -f "central-brain" ]; then
    echo -e "${RED}❌ 构建产物不存在${NC}"
    exit 1
fi

# 启动Central Brain
echo -e "${GREEN}🚀 启动Central Brain...${NC}"
nohup ./central-brain > /tmp/central-brain.log 2>&1 &
CENTRAL_BRAIN_PID=$!

# 等待启动
echo "⏳ 等待服务启动..."
sleep 3

# 检查是否启动成功
if ps -p $CENTRAL_BRAIN_PID > /dev/null; then
    echo -e "${GREEN}✅ Central Brain已启动 (PID: $CENTRAL_BRAIN_PID)${NC}"
    echo "   📊 日志文件: /tmp/central-brain.log"
    echo "   🌐 访问地址: http://localhost:${CENTRAL_BRAIN_PORT}"
    echo "   🏥 健康检查: http://localhost:${CENTRAL_BRAIN_PORT}/health"
    
    # 等待一下然后测试健康检查
    sleep 2
    if curl -s http://localhost:${CENTRAL_BRAIN_PORT}/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 健康检查通过${NC}"
    else
        echo -e "${YELLOW}⚠️  健康检查失败，请查看日志: tail -f /tmp/central-brain.log${NC}"
    fi
else
    echo -e "${RED}❌ Central Brain启动失败${NC}"
    echo "   查看日志: tail -f /tmp/central-brain.log"
    exit 1
fi

echo ""
echo -e "${GREEN}📋 服务路由配置:${NC}"
echo "   /api/v1/auth/**      → Auth Service (${SERVICE_HOST}:${AUTH_SERVICE_PORT})"
echo "   /api/v1/ai/**        → AI Service (${SERVICE_HOST}:${AI_SERVICE_PORT:-8100})"
echo "   /api/v1/blockchain/** → Blockchain Service (${SERVICE_HOST}:${BLOCKCHAIN_SERVICE_PORT:-8208})"
echo "   /api/v1/user/**      → User Service (${SERVICE_HOST}:${USER_SERVICE_PORT:-8082})"
echo "   /api/v1/job/**       → Job Service (${SERVICE_HOST}:${JOB_SERVICE_PORT:-8084})"
echo "   /api/v1/resume/**    → Resume Service (${SERVICE_HOST}:${RESUME_SERVICE_PORT:-8085})"
echo "   /api/v1/company/**   → Company Service (${SERVICE_HOST}:${COMPANY_SERVICE_PORT:-8083})"
echo ""
echo -e "${GREEN}✅ Central Brain启动完成！${NC}"

