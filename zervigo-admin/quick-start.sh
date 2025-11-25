#!/bin/bash

# Zervigo Admin 快速启动脚本
# 用于快速启动开发环境

set -e

echo "=========================================="
echo "  🧠 Zervigo Admin 快速启动脚本"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查 Node.js
echo -e "${YELLOW}[1/6] 检查 Node.js 环境...${NC}"
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ 错误: Node.js 未安装${NC}"
    echo "请先安装 Node.js: https://nodejs.org/"
    exit 1
fi
echo -e "${GREEN}✅ Node.js 版本: $(node --version)${NC}"
echo ""

# 检查 npm
echo -e "${YELLOW}[2/6] 检查 npm 环境...${NC}"
if ! command -v npm &> /dev/null; then
    echo -e "${RED}❌ 错误: npm 未安装${NC}"
    exit 1
fi
echo -e "${GREEN}✅ npm 版本: $(npm --version)${NC}"
echo ""

# 检查并安装依赖
echo -e "${YELLOW}[3/6] 检查项目依赖...${NC}"
if [ ! -d "node_modules" ]; then
    echo -e "${BLUE}首次运行，正在安装依赖...${NC}"
    npm install
    echo -e "${GREEN}✅ 依赖安装完成${NC}"
else
    echo -e "${GREEN}✅ 依赖已安装${NC}"
fi
echo ""

# 检查后端服务 (Central Brain)
echo -e "${YELLOW}[4/6] 检查后端服务 (Central Brain)...${NC}"
BACKEND_URL="http://localhost:9000"
if curl -s -f -o /dev/null "$BACKEND_URL/health" 2>/dev/null; then
    echo -e "${GREEN}✅ 后端服务运行正常 ($BACKEND_URL)${NC}"
else
    echo -e "${RED}⚠️  警告: 后端服务未运行${NC}"
    echo ""
    echo -e "${YELLOW}后端服务需要先启动才能使用动态菜单功能${NC}"
    echo ""
    echo "启动后端服务的方法："
    echo "  cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain"
    echo "  go run ."
    echo ""
    read -p "是否继续启动前端? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}已取消启动${NC}"
        exit 1
    fi
fi
echo ""

# 显示使用指南
echo -e "${YELLOW}[5/6] 使用指南${NC}"
echo "=========================================="
echo ""
echo -e "${BLUE}📚 功能特性${NC}"
echo "  ✅ 动态菜单系统 - 从后端API加载"
echo "  ✅ 智能路由注册 - 自动注册动态路由"
echo "  ✅ 组件智能查找 - 借鉴VueCMF技术"
echo "  ✅ 优雅降级策略 - Common.vue通用模板"
echo "  ✅ 面包屑导航   - 自动追踪路径"
echo ""
echo -e "${BLUE}🔐 默认账号${NC}"
echo "  用户名: admin"
echo "  密码:   admin123"
echo ""
echo -e "${BLUE}📖 文档位置${NC}"
echo "  快速开始: cat QUICK_START.md"
echo "  详细指南: cat DYNAMIC_MENU_GUIDE.md"
echo "  技术总结: cat VUECMF_LESSONS_LEARNED.md"
echo ""
echo -e "${BLUE}🔍 调试技巧${NC}"
echo "  打开浏览器开发者工具 (F12)"
echo "  查看 Console 标签中的日志："
echo "    - [路由加载] 可用的模板文件"
echo "    - [路由注册] 菜单注册情况"
echo "    - ✅ 成功 / ⚠️ 使用通用模板"
echo ""
echo "=========================================="
echo ""

# 启动开发服务器
echo -e "${YELLOW}[6/6] 启动开发服务器...${NC}"
echo -e "${GREEN}✅ 准备启动前端服务${NC}"
echo ""
echo -e "${BLUE}Vite 正在启动...${NC}"
echo -e "${YELLOW}注意：实际访问地址以终端输出为准（通常是 http://localhost:3000）${NC}"
echo ""
echo "=========================================="
echo ""

# 启动 Vite 开发服务器
npm run dev

