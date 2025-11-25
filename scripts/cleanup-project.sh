#!/bin/bash

# Zervigo 项目清理脚本
# 清理废弃的文件和目录，保持项目结构清晰

set -e

echo "🧹 开始清理 Zervigo 项目..."

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

# 1. 清理备份文件
cleanup_backup_files() {
    log_info "清理备份文件..."
    
    local backup_files=(
        "go.mod.bak"
        "CHANGELOG.md.bak"
        "go.work.backup"
        "*.disabled"
    )
    
    for pattern in "${backup_files[@]}"; do
        find . -name "$pattern" -type f | while read -r file; do
            log_info "删除备份文件: $file"
            rm -f "$file"
        done
    done
    
    log_success "备份文件清理完成"
}

# 2. 清理旧的src目录（保留有用的文件）
cleanup_old_src() {
    log_info "清理旧的src目录..."
    
    if [ -d "src" ]; then
        # 检查src目录内容
        log_info "src目录内容:"
        ls -la src/
        
        # 保留有用的文件，删除重复的服务
        log_warning "src目录包含以下内容，需要手动确认删除："
        echo "  - src/auth-service-go/ (已移动到 services/core/auth/)"
        echo "  - src/microservices/ (已移动到 services/)"
        echo "  - src/central-brain/ (已移动到 shared/central-brain/)"
        echo "  - src/shared/ (已移动到 shared/core/)"
        
        # 创建备份目录
        mkdir -p .cleanup-backup
        log_info "将src目录移动到 .cleanup-backup/src-backup"
        mv src .cleanup-backup/src-backup
        
        log_success "src目录已备份到 .cleanup-backup/src-backup"
    else
        log_info "src目录不存在，跳过清理"
    fi
}

# 3. 清理Go-Zero生成的冲突模块
cleanup_gozero_conflicts() {
    log_info "清理Go-Zero生成的冲突模块..."
    
    local conflict_dirs=(
        "service/auth"
        "service/user" 
        "service/job"
        "service/resume"
        "service/company"
        "service/blockchain"
        "service/ai"
        "tools/rpc/auth"
        "rpc/auth"
    )
    
    for dir in "${conflict_dirs[@]}"; do
        if [ -d "$dir" ]; then
            log_warning "发现冲突目录: $dir"
            # 检查是否有go.mod.disabled文件
            if [ -f "$dir/go.mod.disabled" ]; then
                log_info "删除已禁用的模块: $dir"
                rm -rf "$dir"
            else
                log_warning "目录 $dir 存在但未禁用，需要手动确认"
            fi
        fi
    done
    
    log_success "Go-Zero冲突模块清理完成"
}

# 4. 清理重复的依赖和缓存文件
cleanup_duplicates() {
    log_info "清理重复的依赖和缓存文件..."
    
    # 清理Go模块缓存
    log_info "清理Go模块缓存..."
    go clean -modcache 2>/dev/null || true
    
    # 清理编译产物
    log_info "清理编译产物..."
    find . -name "*.exe" -o -name "*.out" -o -name "auth-service" -o -name "unified-auth" | while read -r file; do
        log_info "删除编译产物: $file"
        rm -f "$file"
    done
    
    # 清理临时文件
    log_info "清理临时文件..."
    find . -name "*.tmp" -o -name "*.temp" -o -name ".DS_Store" | while read -r file; do
        log_info "删除临时文件: $file"
        rm -f "$file"
    done
    
    log_success "重复依赖和缓存文件清理完成"
}

# 5. 清理空的目录
cleanup_empty_dirs() {
    log_info "清理空目录..."
    
    # 查找并删除空目录（除了重要的目录）
    find . -type d -empty | grep -v -E "(\.git|\.cleanup-backup)" | while read -r dir; do
        log_info "删除空目录: $dir"
        rmdir "$dir" 2>/dev/null || true
    done
    
    log_success "空目录清理完成"
}

# 6. 显示清理后的项目结构
show_clean_structure() {
    log_info "清理后的项目结构:"
    echo ""
    echo "📁 主要目录结构:"
    tree -L 2 -I '.git|.cleanup-backup|node_modules' 2>/dev/null || {
        echo "services/"
        echo "├── core/"
        echo "├── business/"
        echo "└── infrastructure/"
        echo "shared/"
        echo "├── core/"
        echo "└── central-brain/"
        echo "api/"
        echo "rpc/"
        echo "tools/"
        echo "frontend/"
        echo "configs/"
        echo "scripts/"
        echo "docs/"
    }
}

# 主函数
main() {
    log_info "开始清理 Zervigo 项目..."
    
    cleanup_backup_files
    cleanup_old_src
    cleanup_gozero_conflicts
    cleanup_duplicates
    cleanup_empty_dirs
    
    log_success "🎉 项目清理完成！"
    
    show_clean_structure
    
    echo ""
    echo "📋 清理总结："
    echo "   ✅ 删除了备份文件"
    echo "   ✅ 备份了旧的src目录"
    echo "   ✅ 清理了Go-Zero冲突模块"
    echo "   ✅ 清理了重复依赖和缓存"
    echo "   ✅ 删除了空目录"
    echo ""
    echo "🚀 项目结构现在更加清晰！"
    echo "💡 如果需要恢复文件，请查看 .cleanup-backup/ 目录"
}

# 执行主函数
main "$@"
