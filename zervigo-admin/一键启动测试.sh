#!/bin/bash

# Zervigo Admin 一键启动和测试脚本
# 包含前端、后端的完整启动和验证

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "  🧠 Zervigo Admin 一键启动测试"
echo "=========================================="
echo ""

# 检查后端是否已启动
echo -e "${YELLOW}[1/5] 检查后端服务状态...${NC}"
if curl -s -f http://localhost:9000/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 后端服务已运行 (http://localhost:9000)${NC}"
else
    echo -e "${RED}❌ 后端服务未运行${NC}"
    echo ""
    echo -e "${YELLOW}准备启动后端服务...${NC}"
    echo ""
    
    # 检查后端目录是否存在
    BACKEND_DIR="/Users/szjason72/gozervi/zervigo.demo/shared/central-brain"
    if [ ! -d "$BACKEND_DIR" ]; then
        echo -e "${RED}错误: 找不到后端目录 $BACKEND_DIR${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}在新终端窗口启动后端...${NC}"
    echo ""
    echo "请在新终端窗口执行以下命令："
    echo ""
    echo -e "${YELLOW}  cd $BACKEND_DIR${NC}"
    echo -e "${YELLOW}  go run .${NC}"
    echo ""
    echo "等待后端启动完成后，回到本终端按任意键继续..."
    read -n 1 -s
    
    # 再次检查
    if curl -s -f http://localhost:9000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 后端服务已启动${NC}"
    else
        echo -e "${RED}❌ 后端服务仍未运行，将继续（功能受限）${NC}"
    fi
fi
echo ""

# 测试后端登录接口
echo -e "${YELLOW}[2/5] 测试后端登录接口...${NC}"
LOGIN_TEST=$(curl -s -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' 2>&1)

if echo "$LOGIN_TEST" | grep -q "token"; then
    echo -e "${GREEN}✅ 后端登录接口正常${NC}"
    echo -e "${BLUE}响应示例:${NC}"
    echo "$LOGIN_TEST" | jq . 2>/dev/null || echo "$LOGIN_TEST"
else
    echo -e "${RED}⚠️  后端登录接口可能有问题${NC}"
    echo -e "${YELLOW}响应内容:${NC}"
    echo "$LOGIN_TEST"
fi
echo ""

# 检查前端依赖
echo -e "${YELLOW}[3/5] 检查前端依赖...${NC}"
if [ ! -d "node_modules" ]; then
    echo -e "${BLUE}安装依赖...${NC}"
    npm install
fi
echo -e "${GREEN}✅ 依赖就绪${NC}"
echo ""

# 显示使用指南
echo -e "${YELLOW}[4/5] 使用指南${NC}"
echo "=========================================="
echo ""
echo -e "${BLUE}🔐 默认账号${NC}"
echo "  用户名: admin"
echo "  密码:   admin123"
echo ""
echo -e "${BLUE}🌐 访问地址${NC}"
echo "  前端: http://localhost:3000 (以终端输出为准)"
echo "  后端: http://localhost:9000"
echo ""
echo -e "${BLUE}📖 文档${NC}"
echo "  API格式说明: cat API响应格式说明.md"
echo "  登录修复说明: cat 登录修复说明.md"
echo "  完整启动指南: cat 完整启动指南.md"
echo ""
echo -e "${BLUE}🔍 调试技巧${NC}"
echo "  1. 打开浏览器开发者工具 (F12)"
echo "  2. 查看 Console 标签 - 登录响应日志"
echo "  3. 查看 Network 标签 - API 请求详情"
echo ""
echo "=========================================="
echo ""

# 启动前端
echo -e "${YELLOW}[5/5] 启动前端开发服务器...${NC}"
echo -e "${GREEN}✅ 准备启动${NC}"
echo ""
echo -e "${YELLOW}注意：访问地址以下方 Vite 输出为准！${NC}"
echo ""

npm run dev

