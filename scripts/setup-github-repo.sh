#!/bin/bash

# GitHub 仓库设置脚本
# 用途：帮助将项目上传到 GitHub

set -e

echo "🚀 Go-Zervi 框架 - GitHub 仓库设置脚本"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否在正确的目录
if [ ! -f "README.md" ] || [ ! -d "shared/core" ]; then
    echo -e "${RED}❌ 错误: 请在项目根目录运行此脚本${NC}"
    exit 1
fi

# 步骤1: 检查 Git 状态
echo -e "${YELLOW}📋 步骤 1: 检查 Git 状态...${NC}"
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}❌ 错误: 当前目录不是 Git 仓库${NC}"
    exit 1
fi

# 检查是否已有远程仓库
if git remote -v | grep -q origin; then
    echo -e "${YELLOW}⚠️  警告: 已存在远程仓库 'origin'${NC}"
    git remote -v
    echo ""
    read -p "是否要更新远程仓库 URL? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${GREEN}✓ 跳过远程仓库设置${NC}"
        exit 0
    fi
fi

# 步骤2: 获取 GitHub 仓库信息
echo -e "${YELLOW}📋 步骤 2: 获取 GitHub 仓库信息...${NC}"
echo ""
echo "请输入 GitHub 仓库信息："
echo "  格式: https://github.com/用户名/仓库名.git"
echo "  或: git@github.com:用户名/仓库名.git"
echo ""
read -p "GitHub 仓库 URL: " REPO_URL

if [ -z "$REPO_URL" ]; then
    echo -e "${RED}❌ 错误: 仓库 URL 不能为空${NC}"
    exit 1
fi

# 步骤3: 添加或更新远程仓库
echo ""
echo -e "${YELLOW}📋 步骤 3: 配置远程仓库...${NC}"
if git remote | grep -q origin; then
    git remote set-url origin "$REPO_URL"
    echo -e "${GREEN}✓ 更新远程仓库 URL${NC}"
else
    git remote add origin "$REPO_URL"
    echo -e "${GREEN}✓ 添加远程仓库${NC}"
fi

# 步骤4: 检查并提交更改
echo ""
echo -e "${YELLOW}📋 步骤 4: 检查未提交的更改...${NC}"
if [ -n "$(git status --porcelain)" ]; then
    echo "发现未提交的更改："
    git status --short | head -20
    echo ""
    read -p "是否要提交这些更改? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "请输入提交信息 (默认: 'chore: update project files'): " COMMIT_MSG
        COMMIT_MSG=${COMMIT_MSG:-"chore: update project files"}
        
        git add .
        git commit -m "$COMMIT_MSG"
        echo -e "${GREEN}✓ 已提交更改${NC}"
    else
        echo -e "${YELLOW}⚠️  跳过提交，您可以在稍后手动提交${NC}"
    fi
else
    echo -e "${GREEN}✓ 没有未提交的更改${NC}"
fi

# 步骤5: 确定默认分支
echo ""
echo -e "${YELLOW}📋 步骤 5: 确定默认分支...${NC}"
CURRENT_BRANCH=$(git branch --show-current)
echo "当前分支: $CURRENT_BRANCH"

# 步骤6: 推送到 GitHub
echo ""
echo -e "${YELLOW}📋 步骤 6: 准备推送到 GitHub...${NC}"
echo ""
echo "⚠️  注意: 推送可能需要一些时间，取决于项目大小"
echo ""
read -p "是否现在推送到 GitHub? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}⚠️  已跳过推送。您可以稍后运行: git push -u origin $CURRENT_BRANCH${NC}"
    exit 0
fi

# 推送到 GitHub
echo ""
echo -e "${YELLOW}正在推送到 GitHub...${NC}"
if git push -u origin "$CURRENT_BRANCH"; then
    echo ""
    echo -e "${GREEN}🎉 成功！项目已推送到 GitHub${NC}"
    echo ""
    echo "您的仓库地址:"
    echo "  $REPO_URL"
    echo ""
    echo "下一步："
    echo "  1. 访问 GitHub 查看您的仓库"
    echo "  2. 设置仓库描述和 README"
    echo "  3. 配置 GitHub Actions (如果需要)"
else
    echo ""
    echo -e "${RED}❌ 推送失败${NC}"
    echo ""
    echo "可能的原因："
    echo "  1. 仓库不存在或没有访问权限"
    echo "  2. 需要先创建 GitHub 仓库"
    echo "  3. 认证失败（需要配置 SSH 密钥或 Personal Access Token）"
    echo ""
    echo "解决方法："
    echo "  1. 在 GitHub 上创建新仓库"
    echo "  2. 配置 SSH 密钥: https://docs.github.com/en/authentication/connecting-to-github-with-ssh"
    echo "  3. 或使用 HTTPS 并提供 Personal Access Token"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ 完成！${NC}"
