#!/bin/bash
# Router Service集成到Central Brain测试验证脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=========================================="
echo "🔗 Router Service集成测试验证"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 加载环境变量
echo "📂 加载环境变量..."
set -a
source <(cat "$PROJECT_ROOT/configs/local.env" | grep "^[^#]" | grep -v "^$")
set +a
echo -e "${GREEN}✅ 环境变量已加载${NC}"

# 检查Router Service是否运行
echo ""
echo "🔍 检查Router Service状态..."
if curl -s http://localhost:${ROUTER_SERVICE_PORT:-8087}/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Router Service运行正常 (端口: ${ROUTER_SERVICE_PORT:-8087})${NC}"
    ROUTER_RUNNING=true
else
    echo -e "${YELLOW}⚠️  Router Service未运行 (端口: ${ROUTER_SERVICE_PORT:-8087})${NC}"
    echo "   提示: 需要先启动Router Service才能测试集成"
    ROUTER_RUNNING=false
fi

# 编译Central Brain
echo ""
echo "🔨 编译Central Brain..."
cd "$PROJECT_ROOT/shared/central-brain"
if go build -o "$PROJECT_ROOT/bin/central-brain" *.go; then
    echo -e "${GREEN}✅ 编译成功${NC}"
else
    echo -e "${RED}❌ 编译失败${NC}"
    exit 1
fi

# 停止旧进程
echo ""
echo "🛑 停止旧进程..."
lsof -ti:9000 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 2

# 启动Central Brain
echo ""
echo "🚀 启动Central Brain..."
cd "$PROJECT_ROOT"
nohup ./bin/central-brain > /tmp/central-brain-router-test.log 2>&1 &
CB_PID=$!
echo "   进程ID: $CB_PID"
echo "   日志文件: /tmp/central-brain-router-test.log"

# 等待启动
echo ""
echo "⏳ 等待服务启动..."
sleep 5

# 检查服务是否启动
if ! ps -p $CB_PID > /dev/null; then
    echo -e "${RED}❌ Central Brain启动失败${NC}"
    echo "   查看日志: tail -f /tmp/central-brain-router-test.log"
    exit 1
fi

# 健康检查
echo ""
echo "🏥 健康检查..."
for i in {1..10}; do
    if curl -s http://localhost:9000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Central Brain运行正常${NC}"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${RED}❌ Central Brain未响应${NC}"
        exit 1
    fi
    sleep 1
done

# 运行测试
echo ""
echo "=========================================="
echo "🧪 开始集成测试"
echo "=========================================="
echo ""

# 测试1: 获取所有路由配置（公开）
echo "1️⃣  测试获取所有路由配置..."
ROUTES_RESPONSE=$(curl -s http://localhost:9000/api/v1/router/routes)
if echo "$ROUTES_RESPONSE" | grep -q '"code":0' || echo "$ROUTES_RESPONSE" | grep -q '"code":200'; then
    echo -e "${GREEN}✅ 路由配置查询成功${NC}"
    ROUTE_COUNT=$(echo "$ROUTES_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(len(d.get('data', [])))" 2>/dev/null || echo "N/A")
    echo "   路由数量: $ROUTE_COUNT"
else
    echo -e "${YELLOW}⚠️  路由配置查询失败或Router Service未运行${NC}"
    echo "   响应: $(echo $ROUTES_RESPONSE | head -c 200)"
fi
echo ""

# 测试2: 获取所有页面配置（公开）
echo "2️⃣  测试获取所有页面配置..."
PAGES_RESPONSE=$(curl -s http://localhost:9000/api/v1/router/pages)
if echo "$PAGES_RESPONSE" | grep -q '"code":0' || echo "$PAGES_RESPONSE" | grep -q '"code":200'; then
    echo -e "${GREEN}✅ 页面配置查询成功${NC}"
    PAGE_COUNT=$(echo "$PAGES_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(len(d.get('data', [])))" 2>/dev/null || echo "N/A")
    echo "   页面数量: $PAGE_COUNT"
else
    echo -e "${YELLOW}⚠️  页面配置查询失败或Router Service未运行${NC}"
    echo "   响应: $(echo $PAGES_RESPONSE | head -c 200)"
fi
echo ""

# 测试3: 获取用户路由（需要认证）
echo "3️⃣  测试获取用户路由（需要认证）..."
USER_ROUTES_RESPONSE=$(curl -s http://localhost:9000/api/v1/router/user-routes)
if echo "$USER_ROUTES_RESPONSE" | grep -q '"code":401' || echo "$USER_ROUTES_RESPONSE" | grep -q '"code":0'; then
    if echo "$USER_ROUTES_RESPONSE" | grep -q '"code":401'; then
        echo -e "${GREEN}✅ 认证检查正常（返回401未授权）${NC}"
    else
        echo -e "${GREEN}✅ 用户路由查询成功${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  用户路由查询异常${NC}"
    echo "   响应: $(echo $USER_ROUTES_RESPONSE | head -c 200)"
fi
echo ""

# 测试4: 获取用户页面（需要认证）
echo "4️⃣  测试获取用户页面（需要认证）..."
USER_PAGES_RESPONSE=$(curl -s http://localhost:9000/api/v1/router/user-pages)
if echo "$USER_PAGES_RESPONSE" | grep -q '"code":401' || echo "$USER_PAGES_RESPONSE" | grep -q '"code":0'; then
    if echo "$USER_PAGES_RESPONSE" | grep -q '"code":401'; then
        echo -e "${GREEN}✅ 认证检查正常（返回401未授权）${NC}"
    else
        echo -e "${GREEN}✅ 用户页面查询成功${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  用户页面查询异常${NC}"
    echo "   响应: $(echo $USER_PAGES_RESPONSE | head -c 200)"
fi
echo ""

# 显示日志
echo "=========================================="
echo "📊 服务日志（最近10行）"
echo "=========================================="
tail -n 10 /tmp/central-brain-router-test.log 2>/dev/null || echo "日志文件不存在"
echo ""

# 总结
echo "=========================================="
echo "✅ Router Service集成测试完成"
echo "=========================================="
echo ""
echo "📋 测试结果:"
if [ "$ROUTER_RUNNING" = true ]; then
    echo "   Router Service: ✅ 运行中"
else
    echo "   Router Service: ⚠️  未运行（需要先启动）"
fi
echo "   Central Brain: ✅ 运行中"
echo "   集成代码: ✅ 编译成功"
echo ""
echo "📡 新增API端点:"
echo "   GET /api/v1/router/routes       - 获取所有路由配置（公开）"
echo "   GET /api/v1/router/pages        - 获取所有页面配置（公开）"
echo "   GET /api/v1/router/user-routes  - 获取用户路由（需认证）"
echo "   GET /api/v1/router/user-pages   - 获取用户页面（需认证）"
echo ""
echo "🛑 停止服务:"
echo "   kill $CB_PID"
echo ""
echo "📝 查看实时日志:"
echo "   tail -f /tmp/central-brain-router-test.log"
echo ""

