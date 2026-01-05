#!/bin/bash

# Router Service 和 Permission Service 停止脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}停止 Router & Permission Service${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# 停止 Router Service
echo -e "${YELLOW}🛑 停止 Router Service...${NC}"
if pkill -f router-service 2>/dev/null; then
    echo -e "${GREEN}✅ Router Service 已停止${NC}"
else
    echo -e "${YELLOW}⚠️  Router Service 未运行${NC}"
fi

# 停止 Permission Service
echo ""
echo -e "${YELLOW}🛑 停止 Permission Service...${NC}"
if pkill -f permission-service 2>/dev/null; then
    echo -e "${GREEN}✅ Permission Service 已停止${NC}"
else
    echo -e "${YELLOW}⚠️  Permission Service 未运行${NC}"
fi

# 释放端口
echo ""
echo -e "${YELLOW}🔓 释放端口...${NC}"

# 释放端口 8087
if lsof -ti:8087 > /dev/null 2>&1; then
    lsof -ti:8087 | xargs kill -9 2>/dev/null || true
    echo -e "${GREEN}✅ 端口 8087 已释放${NC}"
fi

# 释放端口 8086
if lsof -ti:8086 > /dev/null 2>&1; then
    lsof -ti:8086 | xargs kill -9 2>/dev/null || true
    echo -e "${GREEN}✅ 端口 8086 已释放${NC}"
fi

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}停止完成！${NC}"
echo -e "${GREEN}================================${NC}"

