#!/bin/bash

# 完整启动脚本 - 解决空白页问题
# 作者：AI Assistant
# 日期：2025-11-05

set -e

echo "============================================"
echo "🚀 Zervigo 完整启动流程"
echo "============================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 步骤 1: 停止所有旧进程
echo -e "${BLUE}步骤 1/5: 停止所有旧服务${NC}"
echo "-------------------------------------------"
pkill -f "central-brain" 2>/dev/null || echo "Central Brain 已停止"
pkill -f "auth.*8207" 2>/dev/null || echo "Auth Service 已停止"
lsof -ti :8207 | xargs kill -9 2>/dev/null || true
lsof -ti :9000 | xargs kill -9 2>/dev/null || true
sleep 2
echo -e "${GREEN}✅ 旧服务已停止${NC}"
echo ""

# 步骤 2: 启动 Auth Service
echo -e "${BLUE}步骤 2/5: 启动 Auth Service (端口 8207)${NC}"
echo "-------------------------------------------"
cd /Users/szjason72/gozervi/zervigo.demo/services/core/auth
DATABASE_URL="postgres://vuecmf:vuecmf@localhost:5432/zervigo_mvp?sslmode=disable" \
JWT_SECRET="zervigo-mvp-secret-key-2025" \
nohup go run main.go > /tmp/auth-service.log 2>&1 &
AUTH_PID=$!
echo $AUTH_PID > /tmp/auth-service.pid
echo "Auth Service PID: $AUTH_PID"
sleep 3

# 验证 Auth Service
if curl -s http://localhost:8207/health > /dev/null; then
    echo -e "${GREEN}✅ Auth Service 启动成功${NC}"
else
    echo -e "${RED}❌ Auth Service 启动失败${NC}"
    echo "查看日志: tail -50 /tmp/auth-service.log"
    exit 1
fi
echo ""

# 步骤 3: 启动 Central Brain
echo -e "${BLUE}步骤 3/5: 启动 Central Brain (端口 9000)${NC}"
echo "-------------------------------------------"
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
nohup go run . > /tmp/central-brain.log 2>&1 &
CB_PID=$!
echo $CB_PID > /tmp/central-brain.pid
echo "Central Brain PID: $CB_PID"
sleep 5

# 验证 Central Brain
if curl -s http://localhost:9000/health > /dev/null; then
    echo -e "${GREEN}✅ Central Brain 启动成功${NC}"
else
    echo -e "${RED}❌ Central Brain 启动失败${NC}"
    echo "查看日志: tail -50 /tmp/central-brain.log"
    exit 1
fi
echo ""

# 步骤 4: 测试关键接口
echo -e "${BLUE}步骤 4/5: 测试关键接口${NC}"
echo "-------------------------------------------"

# 测试登录
echo "测试登录接口..."
LOGIN_CODE=$(curl -s -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"data":{"login_name":"admin","password":"admin123"}}' \
  | jq -r '.code')

if [ "$LOGIN_CODE" == "0" ]; then
    echo -e "${GREEN}✅ 登录接口正常 (code: 0)${NC}"
else
    echo -e "${YELLOW}⚠️  登录接口返回 code: $LOGIN_CODE${NC}"
fi

# 测试菜单
echo "测试菜单接口..."
MENU_CODE=$(curl -s http://localhost:9000/api/v1/menu/nav | jq -r '.code')

if [ "$MENU_CODE" == "0" ]; then
    echo -e "${GREEN}✅ 菜单接口正常 (code: 0)${NC}"
    MENU_COUNT=$(curl -s http://localhost:9000/api/v1/menu/nav | jq '.data | length')
    echo "   菜单数量: $MENU_COUNT 条"
else
    echo -e "${RED}❌ 菜单接口错误 (code: $MENU_CODE)${NC}"
fi
echo ""

# 步骤 5: 显示服务信息
echo -e "${BLUE}步骤 5/5: 服务信息总结${NC}"
echo "============================================"
echo ""
echo -e "${GREEN}✅ 所有服务已启动！${NC}"
echo ""
echo "📊 服务列表："
echo "  • Auth Service:    http://localhost:8207 (PID: $AUTH_PID)"
echo "  • Central Brain:   http://localhost:9000 (PID: $CB_PID)"
echo ""
echo "🌐 前端访问："
echo "  • 登录页面:        http://localhost:8081"
echo "  • 用户名:          admin"
echo "  • 密码:            admin123"
echo ""
echo "📝 日志文件："
echo "  • Auth Service:    tail -f /tmp/auth-service.log"
echo "  • Central Brain:   tail -f /tmp/central-brain.log"
echo ""
echo "🛑 停止服务："
echo "  • kill \$(cat /tmp/auth-service.pid)"
echo "  • kill \$(cat /tmp/central-brain.pid)"
echo ""
echo -e "${YELLOW}⚠️  重要提示：${NC}"
echo "1. 请在浏览器中清除缓存："
echo "   localStorage.clear()"
echo "   sessionStorage.clear()"
echo "   location.reload()"
echo ""
echo "2. 然后使用 admin/admin123 登录"
echo ""
echo "============================================"
echo -e "${GREEN}🎉 启动完成！${NC}"
echo "============================================"

