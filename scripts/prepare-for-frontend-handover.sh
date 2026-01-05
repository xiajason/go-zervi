#!/usr/bin/env bash

# ========================================================================================
# 后端前端交付准备脚本
# 用途: 自动化检查后端服务是否准备好交付给前端团队
# 运行: ./scripts/prepare-for-frontend-handover.sh
# ========================================================================================

# 禁用set -e，手动处理错误
set +e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
log() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[✅]${NC} $1"; }
warning() { echo -e "${YELLOW}[⚠️ ]${NC} $1"; }
error() { echo -e "${RED}[❌]${NC} $1"; }
section() { echo -e "\n${MAGENTA}==========================================${NC}"; echo -e "${CYAN}$1${NC}"; echo -e "${MAGENTA}==========================================${NC}\n"; }

# 初始化检查结果
CHECK_SCORE=0
TOTAL_CHECKS=0
FAILED_CHECKS=0

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║      后端前端交付准备检查脚本                              ║"
echo "║      Zervigo MVP - Backend Frontend Handover Check        ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "项目路径: $PROJECT_ROOT"
echo "检查时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 定义检查函数
check() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    if [ $? -eq 0 ]; then
        CHECK_SCORE=$((CHECK_SCORE + 1))
        success "$1"
        return 0
    else
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        error "$1"
        return 1
    fi
}

# ========================================================================================
# 一、基础环境检查
# ========================================================================================
section "一、基础环境检查"

# 检查PostgreSQL是否运行
log "检查PostgreSQL服务..."
if pgrep -x "postgres" > /dev/null 2>&1; then
    check "PostgreSQL 服务正在运行"
else
    warning "PostgreSQL 服务未运行"
fi

# 检查Redis是否运行
log "检查Redis服务..."
if pgrep -x "redis-server" > /dev/null 2>&1; then
    check "Redis 服务正在运行"
else
    warning "Redis 服务未运行"
fi

# 检查配置文件是否存在
log "检查配置文件..."
if [ -f "configs/dev.env" ]; then
    check "开发环境配置文件存在: configs/dev.env"
else
    error "开发环境配置文件不存在: configs/dev.env"
fi

# ========================================================================================
# 二、API文档检查
# ========================================================================================
section "二、API文档检查"

# 检查API定义文件
api_files=(
    "api/auth.api"
    "api/user.api"
    "api/job.api"
    "api/resume.api"
    "api/company.api"
    "api/ai.api"
    "api/blockchain.api"
)

log "检查API定义文件..."
for api_file in "${api_files[@]}"; do
    if [ -f "$api_file" ]; then
        success "API定义文件存在: $api_file"
    else
        error "API定义文件不存在: $api_file"
    fi
done

# 检查交付文档
if [ -f "docs/BACKEND_FRONTEND_HANDOVER_CHECKLIST.md" ]; then
    check "前端交付检查清单文档存在"
else
    error "前端交付检查清单文档不存在"
fi

# ========================================================================================
# 三、中间件配置检查
# ========================================================================================
section "三、中间件配置检查"

# 检查中间件文件
middleware_files=(
    "shared/core/middleware/auth.go"
    "shared/core/middleware/error_handler.go"
)

log "检查中间件文件..."
for mw_file in "${middleware_files[@]}"; do
    if [ -f "$mw_file" ]; then
        success "中间件文件存在: $mw_file"
    else
        error "中间件文件不存在: $mw_file"
    fi
done

# 检查CORS配置
log "检查CORS配置..."
if grep -q "Access-Control-Allow-Origin" "shared/core/middleware/error_handler.go" 2>/dev/null; then
    check "CORS中间件已配置"
else
    error "CORS中间件未配置"
fi

# 检查Security配置
log "检查Security配置..."
if grep -q "X-Content-Type-Options" "shared/core/middleware/error_handler.go" 2>/dev/null; then
    check "Security中间件已配置"
else
    error "Security中间件未配置"
fi

# ========================================================================================
# 四、服务健康检查
# ========================================================================================
section "四、服务健康检查"

# 定义服务列表（使用并行数组避免关联数组兼容性问题）
ports=(9000 8207 8082 8083 8084 8085 8100 8208)
service_names=(
    "Central Brain (API Gateway)"
    "Auth Service (认证服务)"
    "User Service (用户服务)"
    "Company Service (企业服务)"
    "Job Service (职位服务)"
    "Resume Service (简历服务)"
    "AI Service (AI服务)"
    "Blockchain Service (区块链服务)"
)

log "检查服务健康状态..."

for i in "${!ports[@]}"; do
    port="${ports[$i]}"
    service_name="${service_names[$i]}"
    
    # 检查端口是否监听
    if lsof -i :$port > /dev/null 2>&1; then
        success "端口 $port 正在监听 ($service_name)"
        
        # 尝试健康检查
        health_response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port/health" 2>/dev/null || echo "000")
        if [ "$health_response" == "200" ]; then
            check "$service_name 健康检查通过"
        else
            warning "$service_name 健康检查失败 (HTTP $health_response)"
        fi
    else
        warning "端口 $port 未监听 ($service_name)"
    fi
done

# ========================================================================================
# 五、数据库检查
# ========================================================================================
section "五、数据库检查"

# 检查数据库连接
log "检查数据库连接..."
if command -v psql &> /dev/null; then
    if PGPASSWORD=dev_password psql -h localhost -U postgres -d zervigo_mvp -c "SELECT 1;" > /dev/null 2>&1; then
        check "PostgreSQL 数据库连接成功"
    else
        error "PostgreSQL 数据库连接失败"
    fi
else
    warning "psql 命令未找到，跳过数据库连接检查"
fi

# 检查数据库表
log "检查数据库表..."
if command -v psql &> /dev/null; then
    table_count=$(PGPASSWORD=dev_password psql -h localhost -U postgres -d zervigo_mvp -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')
    if [ "$table_count" -gt 0 ]; then
        check "数据库表数量: $table_count"
    else
        error "数据库表为空"
    fi
else
    warning "psql 命令未找到，跳过数据库表检查"
fi

# ========================================================================================
# 六、测试数据检查
# ========================================================================================
section "六、测试数据检查"

# 检查测试数据生成脚本
if [ -f "scripts/quick_test_data_generator.py" ]; then
    check "测试数据生成脚本存在"
else
    error "测试数据生成脚本不存在"
fi

# ========================================================================================
# 七、脚本和工具检查
# ========================================================================================
section "七、脚本和工具检查"

# 检查启动脚本
startup_scripts=(
    "scripts/start-local-services.sh"
    "scripts/start-consul.sh"
    "scripts/stop-local-services.sh"
    "scripts/test-mvp.sh"
    "scripts/comprehensive_health_check.sh"
)

log "检查启动和测试脚本..."
for script in "${startup_scripts[@]}"; do
    if [ -f "$script" ]; then
        if [ -x "$script" ]; then
            success "脚本可执行: $script"
        else
            warning "脚本不可执行: $script"
            chmod +x "$script"
            success "已添加执行权限: $script"
        fi
    else
        error "脚本不存在: $script"
    fi
done

# ========================================================================================
# 八、日志检查
# ========================================================================================
section "八、日志检查"

# 检查日志目录
if [ -d "logs" ]; then
    check "日志目录存在: logs/"
    
    # 检查是否有大日志文件
    log_files=$(find logs -name "*.log" -size +10M 2>/dev/null)
    if [ -n "$log_files" ]; then
        warning "发现大日志文件（>10MB）:"
        echo "$log_files"
    fi
else
    error "日志目录不存在: logs/"
    mkdir -p logs
    success "已创建日志目录"
fi

# ========================================================================================
# 九、依赖检查
# ========================================================================================
section "九、依赖检查"

# 检查Go环境
log "检查Go环境..."
if command -v go &> /dev/null; then
    go_version=$(go version | cut -d' ' -f3)
    check "Go 环境: $go_version"
else
    error "Go 环境未安装"
fi

# 检查Python环境
log "检查Python环境..."
if command -v python3 &> /dev/null; then
    python_version=$(python3 --version)
    check "$python_version"
else
    warning "Python3 环境未安装（AI服务需要）"
fi

# ========================================================================================
# 十、文档检查
# ========================================================================================
section "十、文档检查"

# 检查关键文档
doc_files=(
    "README.md"
    "docs/BACKEND_FRONTEND_HANDOVER_CHECKLIST.md"
    "docs/MICROSERVICE_DATABASE_DESIGN.md"
    "CHANGELOG.md"
)

log "检查关键文档..."
for doc in "${doc_files[@]}"; do
    if [ -f "$doc" ]; then
        success "文档存在: $doc"
    else
        error "文档不存在: $doc"
    fi
done

# ========================================================================================
# 十一、生成检查报告
# ========================================================================================
section "十一、检查报告"

# 计算检查通过率
PASS_RATE=$(awk "BEGIN {printf \"%.2f\", ($CHECK_SCORE/$TOTAL_CHECKS)*100}")

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "                  检查结果汇总"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "总检查项:     $TOTAL_CHECKS"
echo "通过项:       $CHECK_SCORE"
echo "失败项:       $FAILED_CHECKS"
echo "通过率:       $PASS_RATE%"
echo ""

# 评分等级
if [ "$FAILED_CHECKS" -eq 0 ]; then
    success "🎉 完美！所有检查项通过，可以交付给前端团队"
    EXIT_CODE=0
elif [ "$FAILED_CHECKS" -le 5 ]; then
    warning "⚠️  基本通过，但有一些问题需要修复"
    EXIT_CODE=0
elif [ "$FAILED_CHECKS" -le 10 ]; then
    warning "⚠️  部分通过，建议修复关键问题后再交付"
    EXIT_CODE=1
else
    error "❌ 大量问题未解决，不建议交付"
    EXIT_CODE=2
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ========================================================================================
# 十二、建议
# ========================================================================================
if [ "$FAILED_CHECKS" -gt 0 ]; then
    section "十二、建议的后续步骤"
    echo ""
    log "建议执行以下操作来解决问题："
    echo ""
    echo "1. 启动所有服务："
    echo "   $ ./scripts/start-local-services.sh"
    echo ""
    echo "2. 运行健康检查："
    echo "   $ ./scripts/comprehensive_health_check.sh"
    echo ""
    echo "3. 生成测试数据："
    echo "   $ python3 scripts/quick_test_data_generator.py"
    echo ""
    echo "4. 检查API文档："
    echo "   $ cat docs/BACKEND_FRONTEND_HANDOVER_CHECKLIST.md"
    echo ""
    echo "5. 查看日志："
    echo "   $ tail -f logs/*.log"
    echo ""
fi

# ========================================================================================
# 十三、生成报告文件
# ========================================================================================
REPORT_FILE="logs/handover_check_$(date +%Y%m%d_%H%M%S).txt"

cat > "$REPORT_FILE" << EOF
================================================================================
后端前端交付准备检查报告
Zervigo MVP - Backend Frontend Handover Check Report
================================================================================

生成时间:    $(date '+%Y-%m-%d %H:%M:%S')
检查项总数:  $TOTAL_CHECKS
通过项:      $CHECK_SCORE
失败项:      $FAILED_CHECKS
通过率:      $PASS_RATE%

检查结论:
$(if [ "$EXIT_CODE" -eq 0 ]; then echo "✅ 建议可以交付"; else echo "❌ 建议修复问题后再交付"; fi)

详细信息请参考: docs/BACKEND_FRONTEND_HANDOVER_CHECKLIST.md

EOF

success "检查报告已保存: $REPORT_FILE"

echo ""
section "检查完成"
echo ""
log "查看详细检查清单: cat docs/BACKEND_FRONTEND_HANDOVER_CHECKLIST.md"
echo ""

exit $EXIT_CODE

