#!/bin/bash

# GoZervi GitHub 推送脚本
# 使用方法: ./scripts/push-to-github.sh <github-repo-url> [branch-name]
# 例如: ./scripts/push-to-github.sh https://github.com/szjason72/GoZervi.git main

set -e

if [ -z "$1" ]; then
    echo "❌ 错误: 请提供 GitHub 仓库 URL"
    echo ""
    echo "使用方法:"
    echo "  ./scripts/push-to-github.sh <github-repo-url> [branch-name]"
    echo ""
    echo "示例:"
    echo "  ./scripts/push-to-github.sh https://github.com/szjason72/GoZervi.git main"
    echo "  ./scripts/push-to-github.sh git@github.com:szjason72/GoZervi.git feature/oauth2-provider"
    exit 1
fi

REPO_URL=$1
BRANCH_NAME=${2:-$(git branch --show-current)}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🚀 GoZervi 推送到 GitHub (xiajason 账号)"
echo "📦 仓库地址: $REPO_URL"
echo "📌 分支: $BRANCH_NAME"
echo "📁 项目路径: $PROJECT_ROOT"
echo ""
echo "⚠️  注意: 此项目使用 xiajason 账号推送"
echo "   推送时需要 xiajason 的 Personal Access Token"
echo ""

# 检查是否有未提交的更改
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  检测到未提交的更改："
    git status --short
    echo ""
    read -p "是否先提交这些更改? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "请输入提交信息: " COMMIT_MSG
        if [ -z "$COMMIT_MSG" ]; then
            COMMIT_MSG="chore: prepare for GitHub push"
        fi
        git add .
        git commit -m "$COMMIT_MSG"
        echo "✅ 更改已提交"
    else
        echo "⚠️  继续推送未提交的更改..."
    fi
    echo ""
fi

# 检查是否已经设置了远程仓库
if git remote | grep -q "^origin$"; then
    echo "⚠️  检测到已存在的 origin 远程仓库"
    CURRENT_URL=$(git remote get-url origin)
    echo "   当前地址: $CURRENT_URL"
    read -p "是否要更新为新的地址? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git remote set-url origin "$REPO_URL"
        echo "✅ 已更新远程仓库地址"
    else
        echo "❌ 操作已取消"
        exit 1
    fi
else
    echo "➕ 添加远程仓库..."
    git remote add origin "$REPO_URL"
    echo "✅ 远程仓库已添加"
fi

# 检查当前分支
CURRENT_BRANCH=$(git branch --show-current)
echo "📌 当前分支: $CURRENT_BRANCH"

if [ "$CURRENT_BRANCH" != "$BRANCH_NAME" ]; then
    echo "⚠️  当前分支 ($CURRENT_BRANCH) 与指定分支 ($BRANCH_NAME) 不同"
    read -p "是否切换到 $BRANCH_NAME 分支? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git checkout "$BRANCH_NAME" 2>/dev/null || git checkout -b "$BRANCH_NAME"
        CURRENT_BRANCH="$BRANCH_NAME"
    fi
fi

# 推送代码
echo ""
echo "📤 推送代码到 GitHub..."
git push -u origin "$CURRENT_BRANCH"

echo ""
echo "✅ 推送完成！"
echo "🌐 你可以在 GitHub 上查看: $REPO_URL"

