#!/bin/bash

# Router Service 和 Permission Service 测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Router & Permission Service 测试${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# 测试函数
test_service() {
    local service_name=$1
    local url=$2
    local health_check=$3
    
    echo -e "${YELLOW}测试 $service_name...${NC}"
    
    if curl -s "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $service_name 可达${NC}"
        echo "  URL: $url"
        
        if [ ! -z "$health_check" ]; then
            eval "$health_check"
        fi
    else
        echo -e "${RED}❌ $service_name 不可达${NC}"
        echo "  URL: $url"
    fi
    echo ""
}

# 1. 测试 Router Service
test_service "Router Service" "http://localhost:8087/health" '
STATUS=$(curl -s http://localhost:8087/health | jq -r ".status // .service" 2>/dev/null)
echo "  状态: $STATUS"
'

# 2. 测试 Permission Service
test_service "Permission Service" "http://localhost:8086/health" '
STATUS=$(curl -s http://localhost:8086/health | jq -r ".status" 2>/dev/null)
DB_INFO=$(curl -s http://localhost:8086/health | jq -r ".core_health.database.postgresql | \"\\(.database)@\\(.host):\\(.port)\"" 2>/dev/null)
echo "  状态: $STATUS"
echo "  数据库: $DB_INFO"
'

# 3. 测试 Router Service API
echo -e "${YELLOW}测试 Router Service API...${NC}"
if curl -s "http://localhost:8087/api/v1/router/routes" > /dev/null 2>&1; then
    RESPONSE=$(curl -s "http://localhost:8087/api/v1/router/routes")
    CODE=$(echo "$RESPONSE" | jq -r '.code' 2>/dev/null)
    MESSAGE=$(echo "$RESPONSE" | jq -r '.message' 2>/dev/null)
    ROUTE_COUNT=$(echo "$RESPONSE" | jq -r '.data | length' 2>/dev/null)
    
    if [ "$CODE" = "0" ]; then
        echo -e "${GREEN}✅ 路由配置获取成功${NC}"
        echo "  路由数量: $ROUTE_COUNT"
    else
        echo -e "${RED}❌ 路由配置获取失败${NC}"
        echo "  错误: $MESSAGE"
    fi
else
    echo -e "${RED}❌ API 不可达${NC}"
fi
echo ""

# 4. 测试 Permission Service API
echo -e "${YELLOW}测试 Permission Service API...${NC}"
if curl -s "http://localhost:8086/api/v1/roles" > /dev/null 2>&1; then
    RESPONSE=$(curl -s "http://localhost:8086/api/v1/roles")
    CODE=$(echo "$RESPONSE" | jq -r '.code' 2>/dev/null)
    MESSAGE=$(echo "$RESPONSE" | jq -r '.message' 2>/dev/null)
    
    if [ "$CODE" = "0" ]; then
        echo -e "${GREEN}✅ 角色列表获取成功${NC}"
    else
        echo -e "${RED}❌ 角色列表获取失败${NC}"
        echo "  错误: $MESSAGE"
    fi
else
    echo -e "${RED}❌ API 不可达${NC}"
fi
echo ""

# 5. 测试 Central Brain 代理
echo -e "${YELLOW}测试 Central Brain 代理...${NC}"
if curl -s "http://localhost:9000/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Central Brain 可达${NC}"
    
    # 测试 Router 代理
    if curl -s "http://localhost:9000/api/v1/router/routes" > /dev/null 2>&1; then
        RESPONSE=$(curl -s "http://localhost:9000/api/v1/router/routes")
        CODE=$(echo "$RESPONSE" | jq -r '.code' 2>/dev/null)
        
        if [ "$CODE" = "0" ]; then
            echo -e "${GREEN}✅ Router 代理正常工作${NC}"
        else
            echo -e "${RED}❌ Router 代理失败${NC}"
        fi
    else
        echo -e "${RED}❌ Router 代理不可达${NC}"
    fi
    
    # 测试 Permission 代理
    if curl -s "http://localhost:9000/api/v1/permission/roles" > /dev/null 2>&1; then
        RESPONSE=$(curl -s "http://localhost:9000/api/v1/permission/roles")
        CODE=$(echo "$RESPONSE" | jq -r '.code' 2>/dev/null)
        
        if [ "$CODE" = "0" ]; then
            echo -e "${GREEN}✅ Permission 代理正常工作${NC}"
        else
            echo -e "${RED}❌ Permission 代理失败${NC}"
        fi
    else
        echo -e "${RED}❌ Permission 代理不可达${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Central Brain 未运行，跳过代理测试${NC}"
fi
echo ""

# 总结
echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}测试完成${NC}"
echo -e "${BLUE}================================${NC}"
echo ""
echo "💡 提示："
echo "  - 所有测试通过表示服务运行正常"
echo "  - 如有失败，请查看日志: tail -f logs/router-service.log"
echo "  - 重启服务: ./scripts/start-router-permission.sh"
echo ""

