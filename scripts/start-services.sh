#!/bin/bash

# Zervi 智能服务编排脚本
# 支持7种服务组合自动编排

set -e

echo "🧠 智能中央大脑 - 服务编排系统"
echo "======================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 项目配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/configs/local.env"

# 加载环境变量
load_environment() {
    if [ -n "$ENV_FILE" ] && [ -f "$ENV_FILE" ]; then
        log_step "加载环境变量: $ENV_FILE"
        set -a  # 自动导出所有变量
        # 创建临时文件加载环境变量
        TEMP_ENV=$(mktemp)
        while IFS= read -r line; do
            # 跳过以#开头的注释行
            [[ "$line" =~ ^[[:space:]]*# ]] && continue
            # 跳过空行
            [[ -z "$line" ]] && continue
            # 移除行内注释（=后面的#注释）
            line=$(echo "$line" | sed 's/^\([^=]*=[^#]*\)#.*/\1/')
            # 移除首尾空白
            line=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
            # 输出非空行
            [[ -n "$line" ]] && echo "$line" >> "$TEMP_ENV"
        done < "$ENV_FILE"
        source "$TEMP_ENV"
        rm -f "$TEMP_ENV"
        set +a
        
        # 检测配置并给出明确提示
        HAS_MYSQL=0
        HAS_POSTGRESQL=0
        
        if [ -n "$MYSQL_HOST" ] && [ "$MYSQL_HOST" != "" ]; then
            HAS_MYSQL=1
        fi
        
        if [ -n "$POSTGRESQL_HOST" ] && [ "$POSTGRESQL_HOST" != "" ]; then
            HAS_POSTGRESQL=1
        fi
        
        # 根据检测结果决定使用哪个数据库
        if [ $HAS_MYSQL -eq 1 ] && [ $HAS_POSTGRESQL -eq 0 ]; then
            log_info "🔄 检测到 MySQL 配置，使用 MySQL 数据库"
        elif [ $HAS_POSTGRESQL -eq 1 ] && [ $HAS_MYSQL -eq 0 ]; then
            log_info "🔄 检测到 PostgreSQL 配置，使用 PostgreSQL 数据库"
        elif [ $HAS_MYSQL -eq 1 ] && [ $HAS_POSTGRESQL -eq 1 ]; then
            log_error "❌ 检测到 MySQL 和 PostgreSQL 配置同时存在！请只启用其中一个"
            exit 1
        else
            log_warning "⚠️  未检测到任何数据库配置，使用默认配置"
        fi
    else
        log_warning "⚠️  未找到环境配置文件，使用默认配置"
    fi
}

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# 显示使用说明
show_usage() {
    echo ""
    echo "使用方法:"
    echo "  ./start-services.sh <composition_name>"
    echo ""
    echo "支持的服务组合:"
    echo "  1. job_only          - 只启动job服务"
    echo "  2. resume_only       - 只启动resume服务"
    echo "  3. company_only      - 只启动company服务"
    echo "  4. job_resume        - job + resume服务"
    echo "  5. job_company       - job + company服务"
    echo "  6. resume_company    - resume + company服务"
    echo "  7. all_services      - 所有服务"
    echo ""
}

# 解析配置
parse_composition() {
    local composition=$1
    
    case $composition in
        "job_only")
            echo "auth-service user-service job-service"
            ;;
        "resume_only")
            echo "auth-service user-service resume-service"
            ;;
        "company_only")
            echo "auth-service user-service company-service"
            ;;
        "job_resume")
            echo "auth-service user-service job-service resume-service"
            ;;
        "job_company")
            echo "auth-service user-service job-service company-service"
            ;;
        "resume_company")
            echo "auth-service user-service resume-service company-service"
            ;;
        "all_services")
            echo "auth-service user-service job-service resume-service company-service"
            ;;
        *)
            log_error "未知的服务组合: $composition"
            show_usage
            exit 1
            ;;
    esac
}

# 启动服务
start_service() {
    local service=$1
    local port=$2
    
    log_step "启动服务: $service (端口: $port)"
    
    # 检查端口是否被占用
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        log_info "端口 $port 已被占用，服务可能已在运行"
        return 0
    fi
    
    # 根据服务名启动
    case $service in
        "auth-service")
            cd "$PROJECT_ROOT/services/core/auth"
            ;;
        "user-service")
            cd "$PROJECT_ROOT/services/core/user"
            ;;
        "job-service")
            cd "$PROJECT_ROOT/services/business/job"
            ;;
        "resume-service")
            cd "$PROJECT_ROOT/services/business/resume"
            ;;
        "company-service")
            cd "$PROJECT_ROOT/services/business/company"
            ;;
        *)
            log_error "未知服务: $service"
            return 1
            ;;
    esac
    
    # 启动服务（环境变量已在load_environment中加载）
    nohup env \
        MYSQL_HOST="$MYSQL_HOST" \
        MYSQL_PORT="$MYSQL_PORT" \
        MYSQL_USER="$MYSQL_USER" \
        MYSQL_PASSWORD="$MYSQL_PASSWORD" \
        MYSQL_DATABASE="$MYSQL_DATABASE" \
        POSTGRESQL_HOST="$POSTGRESQL_HOST" \
        POSTGRESQL_PORT="$POSTGRESQL_PORT" \
        POSTGRESQL_USER="$POSTGRESQL_USER" \
        POSTGRESQL_PASSWORD="$POSTGRESQL_PASSWORD" \
        POSTGRESQL_DATABASE="$POSTGRESQL_DATABASE" \
        POSTGRESQL_SSL_MODE="$POSTGRESQL_SSL_MODE" \
        go run main.go > "$PROJECT_ROOT/logs/$service.log" 2>&1 &
    echo $! > "$PROJECT_ROOT/logs/$service.pid"
    
    sleep 3
    
    # 健康检查
    if curl -s http://localhost:$port/health > /dev/null 2>&1; then
        log_success "$service 启动成功"
    else
        log_error "$service 启动失败"
        return 1
    fi
}

# 健康检查
health_check() {
    log_step "执行健康检查..."
    
    local services=$1
    local failed=0
    
    for service in $services; do
        case $service in
            "auth-service")
                if curl -s http://localhost:8207/health > /dev/null 2>&1; then
                    log_success "auth-service 健康检查通过"
                else
                    log_error "auth-service 健康检查失败"
                    failed=1
                fi
                ;;
            "user-service")
                if curl -s http://localhost:8082/health > /dev/null 2>&1; then
                    log_success "user-service 健康检查通过"
                else
                    log_error "user-service 健康检查失败"
                    failed=1
                fi
                ;;
            "job-service")
                if curl -s http://localhost:8084/health > /dev/null 2>&1; then
                    log_success "job-service 健康检查通过"
                else
                    log_error "job-service 健康检查失败"
                    failed=1
                fi
                ;;
            "resume-service")
                if curl -s http://localhost:8085/health > /dev/null 2>&1; then
                    log_success "resume-service 健康检查通过"
                else
                    log_error "resume-service 健康检查失败"
                    failed=1
                fi
                ;;
            "company-service")
                if curl -s http://localhost:8083/health > /dev/null 2>&1; then
                    log_success "company-service 健康检查通过"
                else
                    log_error "company-service 健康检查失败"
                    failed=1
                fi
                ;;
        esac
    done
    
    return $failed
}

# 主函数
main() {
    # 首先加载环境变量
    load_environment
    
    if [ $# -eq 0 ]; then
        log_error "请指定服务组合"
        show_usage
        exit 1
    fi
    
    local composition=$1
    log_info "选择的组合: $composition"
    
    # 解析组合
    local services=$(parse_composition "$composition")
    log_info "需要启动的服务: $services"
    
    # 启动服务
    for service in $services; do
        case $service in
            "auth-service")
                start_service "auth-service" 8207
                ;;
            "user-service")
                start_service "user-service" 8082
                ;;
            "job-service")
                start_service "job-service" 8084
                ;;
            "resume-service")
                start_service "resume-service" 8085
                ;;
            "company-service")
                start_service "company-service" 8083
                ;;
        esac
    done
    
    # 等待服务启动
    log_info "等待服务启动完成..."
    sleep 5
    
    # 健康检查
    health_check "$services"
    
    if [ $? -eq 0 ]; then
        log_success "所有服务启动成功！"
        echo ""
        echo "服务访问地址："
        for service in $services; do
            case $service in
                "auth-service")
                    echo "  认证服务: http://localhost:8207"
                    ;;
                "user-service")
                    echo "  用户服务: http://localhost:8082"
                    ;;
                "job-service")
                    echo "  职位服务: http://localhost:8084"
                    ;;
                "resume-service")
                    echo "  简历服务: http://localhost:8085"
                    ;;
                "company-service")
                    echo "  企业服务: http://localhost:8083"
                    ;;
            esac
        done
    else
        log_error "部分服务启动失败"
        exit 1
    fi
}

# 执行主函数
main "$@"

