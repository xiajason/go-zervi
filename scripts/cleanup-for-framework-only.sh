#!/bin/bash

# Go-Zervi 框架代码整理脚本
# 用途：清理 Git 历史，只保留核心框架代码

set -e

echo "🚀 Go-Zervi 框架代码整理脚本"
echo "=========================================="
echo ""
echo "⚠️  警告：此脚本将从 Git 历史中移除以下目录："
echo "  - prototypes/"
echo "  - cleanup-backup/"
echo "  - services/"
echo "  - src/"
echo ""
read -p "是否继续？(y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "已取消"
    exit 1
fi

echo ""
echo "📋 步骤 1: 从 Git 历史中移除大文件和目录..."

# 移除原型文件
git filter-branch --force --index-filter \
    'git rm -rf --cached --ignore-unmatch prototypes cleanup-backup services src bin' \
    --prune-empty --tag-name-filter cat -- --all || true

echo ""
echo "📋 步骤 2: 清理 Git 历史..."
git for-each-ref --format='delete %(refname)' refs/original | git update-ref --stdin
git reflog expire --expire=now --all
git gc --prune=now --aggressive

echo ""
echo "✅ 清理完成！"
echo ""
echo "检查仓库大小："
git count-objects -vH

echo ""
echo "下一步："
echo "  1. 检查 git status 确认更改"
echo "  2. 如果满意，运行: git push -u origin main --force"
echo ""
echo "⚠️  注意：--force 推送会覆盖远程仓库，请确保这是您想要的！"

