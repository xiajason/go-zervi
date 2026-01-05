#!/bin/bash
# Central Brain基础设施层优化验证测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CB_DIR="$PROJECT_ROOT/shared/central-brain"

echo "=========================================="
echo "🧠 Central Brain基础设施层优化验证测试"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查环境
echo "📋 检查环境..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go未安装${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Go环境就绪${NC}"

# 检查配置文件
if [ ! -f "$PROJECT_ROOT/configs/local.env" ]; then
    echo -e "${RED}❌ 配置文件不存在: configs/local.env${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 配置文件存在${NC}"

# 加载环境变量
echo ""
echo "📂 加载环境变量..."
set -a
source <(cat "$PROJECT_ROOT/configs/local.env" | grep "^[^#]" | grep -v "^$")
set +a
echo -e "${GREEN}✅ 环境变量已加载${NC}"

# 检查端口占用
echo ""
echo "🔍 检查端口占用..."
if lsof -ti:9000 &> /dev/null; then
    echo -e "${YELLOW}⚠️  端口9000已被占用${NC}"
    PID=$(lsof -ti:9000)
    echo "   进程ID: $PID"
    read -p "   是否停止现有进程? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kill $PID 2>/dev/null || true
        sleep 2
        echo -e "${GREEN}✅ 已停止现有进程${NC}"
    else
        echo -e "${YELLOW}⚠️  使用现有进程进行测试${NC}"
    fi
else
    echo -e "${GREEN}✅ 端口9000可用${NC}"
fi

# 编译Central Brain
echo ""
echo "🔨 编译Central Brain..."
cd "$CB_DIR"
if go build -o "$PROJECT_ROOT/bin/central-brain" *.go; then
    echo -e "${GREEN}✅ 编译成功${NC}"
else
    echo -e "${RED}❌ 编译失败${NC}"
    exit 1
fi

# 启动服务（后台运行）
echo ""
echo "🚀 启动Central Brain服务..."
cd "$PROJECT_ROOT"
nohup ./bin/central-brain > /tmp/central-brain.log 2>&1 &
CB_PID=$!
echo "   进程ID: $CB_PID"
echo "   日志文件: /tmp/central-brain.log"

# 等待服务启动
echo ""
echo "⏳ 等待服务启动..."
sleep 5

# 检查服务是否启动
if ! ps -p $CB_PID > /dev/null; then
    echo -e "${RED}❌ 服务启动失败${NC}"
    echo "   查看日志: tail -f /tmp/central-brain.log"
    exit 1
fi

# 健康检查
echo ""
echo "🏥 健康检查..."
for i in {1..10}; do
    if curl -s http://localhost:9000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 服务运行正常${NC}"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${RED}❌ 服务未响应${NC}"
        echo "   查看日志: tail -f /tmp/central-brain.log"
        exit 1
    fi
    sleep 1
done

# 运行测试
echo ""
echo "=========================================="
echo "🧪 开始测试验证"
echo "=========================================="
echo ""

# 测试1: 健康检查
echo "1️⃣  测试健康检查端点..."
HEALTH_RESPONSE=$(curl -s http://localhost:9000/health)
if echo "$HEALTH_RESPONSE" | grep -q '"code":0' || echo "$HEALTH_RESPONSE" | grep -q '"code":200'; then
    echo -e "${GREEN}✅ 健康检查通过${NC}"
    echo "   响应: $(echo $HEALTH_RESPONSE | jq -c . 2>/dev/null || echo $HEALTH_RESPONSE)"
else
    echo -e "${RED}❌ 健康检查失败${NC}"
    echo "   响应: $HEALTH_RESPONSE"
fi
echo ""

# 测试2: 性能指标
echo "2️⃣  测试性能指标端点..."
METRICS_RESPONSE=$(curl -s http://localhost:9000/api/v1/metrics)
if echo "$METRICS_RESPONSE" | grep -q '"code":0'; then
    echo -e "${GREEN}✅ 性能指标查询成功${NC}"
    TOTAL_REQUESTS=$(echo "$METRICS_RESPONSE" | jq -r '.data.total_requests // 0' 2>/dev/null || echo "0")
    echo "   总请求数: $TOTAL_REQUESTS"
else
    echo -e "${YELLOW}⚠️  性能指标查询失败或未授权${NC}"
    echo "   响应: $METRICS_RESPONSE"
fi
echo ""

# 测试3: 熔断器状态
echo "3️⃣  测试熔断器状态端点..."
BREAKER_RESPONSE=$(curl -s http://localhost:9000/api/v1/circuit-breakers)
if echo "$BREAKER_RESPONSE" | grep -q '"code":0'; then
    echo -e "${GREEN}✅ 熔断器状态查询成功${NC}"
    AUTH_STATE=$(echo "$BREAKER_RESPONSE" | jq -r '.data.auth.state // "unknown"' 2>/dev/null || echo "unknown")
    echo "   Auth服务熔断器状态: $AUTH_STATE"
else
    echo -e "${YELLOW}⚠️  熔断器状态查询失败或未授权${NC}"
    echo "   响应: $BREAKER_RESPONSE"
fi
echo ""

# 测试4: TraceID生成
echo "4️⃣  测试TraceID生成..."
TRACE_RESPONSE=$(curl -s -H "X-Trace-ID: test-trace-123" http://localhost:9000/health)
TRACE_ID=$(echo "$TRACE_RESPONSE" | jq -r '.trace_id // empty' 2>/dev/null || echo "")
if [ -n "$TRACE_ID" ]; then
    echo -e "${GREEN}✅ TraceID生成成功${NC}"
    echo "   TraceID: $TRACE_ID"
else
    echo -e "${YELLOW}⚠️  TraceID未在响应中返回（可能需要检查日志）${NC}"
fi
echo ""

# 测试5: 限流测试（可选）
echo "5️⃣  测试限流机制（快速请求）..."
LIMIT_TEST_COUNT=0
for i in {1..150}; do
    RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:9000/health)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" == "429" ]; then
        LIMIT_TEST_COUNT=$((LIMIT_TEST_COUNT + 1))
    fi
done
if [ $LIMIT_TEST_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ 限流机制正常工作（触发次数: $LIMIT_TEST_COUNT）${NC}"
else
    echo -e "${YELLOW}⚠️  限流未触发（可能配置较高或测试不够）${NC}"
fi
echo ""

# 显示日志
echo "=========================================="
echo "📊 服务日志（最近20行）"
echo "=========================================="
tail -n 20 /tmp/central-brain.log 2>/dev/null || echo "日志文件不存在"
echo ""

# 总结
echo "=========================================="
echo "✅ 测试验证完成"
echo "=========================================="
echo ""
echo "📋 服务信息:"
echo "   进程ID: $CB_PID"
echo "   端口: 9000"
echo "   日志: /tmp/central-brain.log"
echo ""
echo "📡 API端点:"
echo "   GET /health                    - 健康检查"
echo "   GET /api/v1/metrics            - 性能指标"
echo "   GET /api/v1/circuit-breakers   - 熔断器状态"
echo ""
echo "🛑 停止服务:"
echo "   kill $CB_PID"
echo ""
echo "📝 查看实时日志:"
echo "   tail -f /tmp/central-brain.log"
echo ""

