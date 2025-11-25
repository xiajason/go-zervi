#!/bin/bash

# 批量更新Go文件中的导入路径
# 将 github.com/jobfirst/jobfirst-core 替换为 github.com/szjason72/zervigo/shared/core

set -e

echo "🔄 开始批量更新Go文件中的导入路径..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 更新shared/core目录中的导入路径
update_shared_core() {
    log_info "更新 shared/core 目录中的导入路径..."
    
    # 查找所有Go文件
    find shared/core -name "*.go" -type f | while read -r file; do
        if grep -q "github.com/jobfirst/jobfirst-core" "$file"; then
            log_info "更新文件: $file"
            # 使用sed替换导入路径
            sed -i '' 's|github.com/jobfirst/jobfirst-core|github.com/szjason72/zervigo/shared/core|g' "$file"
        fi
    done
    
    log_success "shared/core 目录更新完成"
}

# 更新所有服务目录中的导入路径
update_services() {
    log_info "更新所有服务目录中的导入路径..."
    
    # 查找所有Go文件
    find services -name "*.go" -type f | while read -r file; do
        if grep -q "github.com/jobfirst/jobfirst-core" "$file"; then
            log_info "更新文件: $file"
            # 使用sed替换导入路径
            sed -i '' 's|github.com/jobfirst/jobfirst-core|github.com/szjason72/zervigo/shared/core|g' "$file"
        fi
    done
    
    log_success "所有服务目录更新完成"
}

# 更新shared/central-brain目录中的导入路径
update_central_brain() {
    log_info "更新 shared/central-brain 目录中的导入路径..."
    
    # 查找所有Go文件
    find shared/central-brain -name "*.go" -type f | while read -r file; do
        if grep -q "github.com/jobfirst/jobfirst-core" "$file"; then
            log_info "更新文件: $file"
            # 使用sed替换导入路径
            sed -i '' 's|github.com/jobfirst/jobfirst-core|github.com/szjason72/zervigo/shared/core|g' "$file"
        fi
    done
    
    log_success "shared/central-brain 目录更新完成"
}

# 主函数
main() {
    log_info "开始批量更新导入路径..."
    
    update_shared_core
    update_services
    update_central_brain
    
    log_success "🎉 所有导入路径更新完成！"
    
    # 显示更新统计
    echo ""
    echo "📊 更新统计："
    echo "   ✅ shared/core 目录"
    echo "   ✅ services 目录"
    echo "   ✅ shared/central-brain 目录"
    echo ""
    echo "🚀 现在可以尝试编译服务了！"
}

# 执行主函数
main "$@"
