#!/bin/bash

# 一键修复启动脚本
# 解决 VueCMF 登录和菜单加载问题

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "============================================"
echo -e "${BLUE}🚀 一键修复并启动所有服务${NC}"
echo "============================================"
echo ""

# 步骤 1: 停止所有旧服务
echo -e "${YELLOW}步骤 1/4: 停止所有旧服务${NC}"
echo "-------------------------------------------"
pkill -9 -f "central-brain" 2>/dev/null || true
pkill -9 -f "auth.*8207" 2>/dev/null || true
lsof -ti :8207 | xargs kill -9 2>/dev/null || true
lsof -ti :9000 | xargs kill -9 2>/dev/null || true
killall -9 main 2>/dev/null || true
sleep 3
echo -e "${GREEN}✅ 旧服务已停止${NC}"
echo ""

# 步骤 2: 启动 Auth Service
echo -e "${YELLOW}步骤 2/4: 启动 Auth Service (8207)${NC}"
echo "-------------------------------------------"
cd /Users/szjason72/gozervi/zervigo.demo/services/core/auth

DATABASE_URL="postgres://vuecmf:vuecmf@localhost:5432/zervigo_mvp?sslmode=disable" \
JWT_SECRET="zervigo-mvp-secret-key-2025" \
nohup go run main.go > /tmp/auth-service.log 2>&1 &

AUTH_PID=$!
echo $AUTH_PID > /tmp/auth-service.pid
echo "Auth Service PID: $AUTH_PID"

sleep 5

# 验证 Auth Service
if curl -s http://localhost:8207/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Auth Service 启动成功${NC}"
else
    echo -e "${RED}❌ Auth Service 启动失败${NC}"
    echo "日志: tail -30 /tmp/auth-service.log"
    tail -30 /tmp/auth-service.log
    exit 1
fi
echo ""

# 步骤 3: 启动 Central Brain  
echo -e "${YELLOW}步骤 3/4: 启动 Central Brain (9000)${NC}"
echo "-------------------------------------------"
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain

nohup go run . > /tmp/central-brain.log 2>&1 &

CB_PID=$!
echo $CB_PID > /tmp/central-brain.pid
echo "Central Brain PID: $CB_PID"

echo "等待服务启动..."
sleep 10

# 验证 Central Brain
if curl -s http://localhost:9000/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Central Brain 启动成功${NC}"
else
    echo -e "${RED}❌ Central Brain 启动失败${NC}"
    echo "日志:"
    tail -50 /tmp/central-brain.log
    exit 1
fi
echo ""

# 步骤 4: 测试所有关键接口
echo -e "${YELLOW}步骤 4/4: 测试所有接口${NC}"
echo "-------------------------------------------"

# 测试登录
echo "测试登录..."
LOGIN_CODE=$(curl -s -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"data":{"login_name":"admin","password":"admin123"}}' \
  | jq -r '.code')

if [ "$LOGIN_CODE" == "0" ]; then
    echo -e "${GREEN}✅ 登录接口正常${NC}"
else
    echo -e "${RED}❌ 登录失败 (code: $LOGIN_CODE)${NC}"
    exit 1
fi

# 测试菜单
echo "测试菜单..."
MENU_CODE=$(curl -s http://localhost:9000/api/v1/menu/nav | jq -r '.code')

if [ "$MENU_CODE" == "0" ]; then
    MENU_COUNT=$(curl -s http://localhost:9000/api/v1/menu/nav | jq '.data.nav_menu | length')
    API_MAPS_COUNT=$(curl -s http://localhost:9000/api/v1/menu/nav | jq '.data.api_maps | length')
    echo -e "${GREEN}✅ 菜单接口正常${NC}"
    echo "   菜单数量: $MENU_COUNT"
    echo "   API映射: $API_MAPS_COUNT 个表"
else
    echo -e "${RED}❌ 菜单失败 (code: $MENU_CODE)${NC}"
    exit 1
fi

echo ""
echo "============================================"
echo -e "${GREEN}🎉 所有服务启动成功！${NC}"
echo "============================================"
echo ""
echo -e "${BLUE}📊 服务信息：${NC}"
echo "  • Auth Service:    http://localhost:8207 (PID: $AUTH_PID)"
echo "  • Central Brain:   http://localhost:9000 (PID: $CB_PID)"
echo ""
echo -e "${BLUE}🌐 前端访问：${NC}"
echo "  • URL:     http://localhost:8081"
echo "  • 用户名:  admin"
echo "  • 密码:    admin123"
echo "  • 角色:    super_admin"
echo "  • 数据库:  postgresql"
echo ""
echo -e "${BLUE}📝 日志文件：${NC}"
echo "  • Auth:    tail -f /tmp/auth-service.log"
echo "  • Brain:   tail -f /tmp/central-brain.log"
echo ""
echo -e "${YELLOW}⚠️  重要提示：${NC}"
echo "1. 在浏览器中清除缓存:"
echo "   localStorage.clear(); sessionStorage.clear(); location.reload();"
echo ""
echo "2. 或使用无痕模式访问：http://localhost:8081"
echo ""
echo "============================================"

