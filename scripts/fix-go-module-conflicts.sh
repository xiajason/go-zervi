#!/bin/bash

# Go模块冲突清理脚本
# 解决多个auth模块冲突的问题

set -e

echo "🔧 开始清理Go模块冲突..."

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

# 1. 备份当前的go.work文件
log_info "备份当前的go.work文件..."
if [ -f "go.work" ]; then
    cp go.work go.work.backup
    log_success "go.work已备份为go.work.backup"
fi

# 2. 创建新的go.work文件，只包含我们需要的模块
log_info "创建新的go.work文件..."
cat > go.work << 'EOF'
go 1.25.0

use (
	./src/auth-service-go
	./src/microservices/user-service
	./src/microservices/job-service
	./src/microservices/resume-service
	./src/microservices/company-service
	./src/microservices/blockchain-service
	./src/central-brain
	./src/shared
)
EOF

log_success "新的go.work文件已创建"

# 3. 检查冲突的模块
log_info "检查冲突的模块..."
CONFLICT_MODULES=(
    "./tools/rpc/auth"
    "./service/auth" 
    "./rpc/auth"
)

for module in "${CONFLICT_MODULES[@]}"; do
    if [ -d "$module" ]; then
        log_warning "发现冲突模块: $module"
        # 重命名冲突的go.mod文件
        if [ -f "$module/go.mod" ]; then
            mv "$module/go.mod" "$module/go.mod.disabled"
            log_info "已禁用 $module/go.mod"
        fi
    fi
done

# 4. 清理Go模块缓存
log_info "清理Go模块缓存..."
go clean -modcache 2>/dev/null || true
go mod tidy 2>/dev/null || true

log_success "Go模块缓存已清理"

# 5. 验证auth-service-go模块
log_info "验证auth-service-go模块..."
cd src/auth-service-go
if go mod tidy; then
    log_success "auth-service-go模块验证成功"
else
    log_error "auth-service-go模块验证失败"
    exit 1
fi

cd ../..

# 6. 测试编译
log_info "测试编译auth-service-go..."
cd src/auth-service-go
if go build -o auth-service main.go; then
    log_success "auth-service-go编译成功"
    rm -f auth-service  # 清理编译产物
else
    log_error "auth-service-go编译失败"
    exit 1
fi

cd ../..

log_success "🎉 Go模块冲突清理完成！"
echo ""
echo "📋 清理总结："
echo "   ✅ 备份了原始go.work文件"
echo "   ✅ 创建了新的go.work文件，只包含需要的模块"
echo "   ✅ 禁用了冲突的模块定义"
echo "   ✅ 清理了Go模块缓存"
echo "   ✅ 验证了auth-service-go模块"
echo ""
echo "🚀 现在可以尝试启动认证服务了！"
