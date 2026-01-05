#!/bin/bash
# Central Brain (API Gateway) 启动脚本
# 用途: 启动Zervigo中央大脑服务，作为API网关统一入口
# 支持: 环境变量配置、自动加载.env文件

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

# 检测并加载环境配置文件
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
    # 过滤掉注释行和空行
    source <(cat "$ENV_FILE" | grep "^[^#]" | grep -v "^$" | sed 's/#.*$//')
    set +a
    
    # 如果使用PostgreSQL，清理MySQL环境变量（避免优先级冲突）
    if [ -n "$POSTGRESQL_HOST" ] && [ -n "$POSTGRESQL_PORT" ]; then
        # 检查MySQL配置是否被注释（不在配置文件中启用）
        if ! grep -q "^[^#]*MYSQL_HOST=" "$ENV_FILE"; then
            echo -e "${YELLOW}🔄 清理MySQL环境变量以使用PostgreSQL${NC}"
            unset MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD MYSQL_DATABASE 2>/dev/null
        fi
    fi
else
    echo -e "${YELLOW}⚠️  未找到环境配置文件，使用默认配置${NC}"
fi

# 设置默认环境变量（如果未设置）
export SERVICE_SECRET="${SERVICE_SECRET:-central-brain-secret-2025}"
export SERVICE_ID="${SERVICE_ID:-central-brain}"
export SERVICE_HOST="${SERVICE_HOST:-localhost}"
export CENTRAL_BRAIN_PORT="${CENTRAL_BRAIN_PORT:-9000}"
export AUTH_SERVICE_PORT="${AUTH_SERVICE_PORT:-8207}"

# 显示配置信息
echo ""
echo -e "${GREEN}📊 配置信息:${NC}"
echo "  服务ID: ${SERVICE_ID}"
echo "  服务主机: ${SERVICE_HOST}"
echo "  Central Brain端口: ${CENTRAL_BRAIN_PORT}"
echo "  Auth Service端口: ${AUTH_SERVICE_PORT}"
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
    echo "⚠️  Central Brain已经在运行中"
    echo "   停止现有进程..."
    pkill -f "central-brain" || true
    sleep 2
fi

# 构建Central Brain
echo "🔨 构建Central Brain..."
if go build -o central-brain *.go 2>&1 | tee /tmp/cb_build.log; then
    echo "✅ 构建成功"
else
    echo "❌ 构建失败，查看错误:"
    cat /tmp/cb_build.log
    exit 1
fi

# 检查依赖
if [ ! -f "central-brain" ]; then
    echo "❌ 构建产物不存在"
    exit 1
fi

# 启动Central Brain
echo "🚀 启动Central Brain..."
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
