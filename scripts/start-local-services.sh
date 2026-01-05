#!/bin/bash

# Zervigo 本地开发环境启动脚本
# 完全使用本地服务，避免Docker依赖

set -e

echo "🚀 启动 Zervigo 本地开发环境"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$PROJECT_ROOT/logs"
mkdir -p "$LOG_DIR"

# PostgreSQL 控制配置
BREW_PREFIX=$(brew --prefix 2>/dev/null)
PG_VERSION="postgresql@14"
PG_CTL_BIN="${BREW_PREFIX}/opt/${PG_VERSION}/bin/pg_ctl"
PG_DATA_DIR="${BREW_PREFIX}/var/${PG_VERSION}"
PG_LOG_DIR="${BREW_PREFIX}/var/log"
PG_LOG_FILE="${PG_LOG_DIR}/${PG_VERSION}.log"

if [ ! -x "$PG_CTL_BIN" ]; then
    PG_CTL_BIN="$(command -v pg_ctl 2>/dev/null)"
fi

mkdir -p "$PG_LOG_DIR"

# 服务端口配置 (使用普通数组)
SERVICES=(
    "auth-service:8207"
    "user-service:8082"
    "job-service:8084"
    "resume-service:8085"
    "company-service:8083"
    "ai-service:8100"
    "blockchain-service:8208"
    "central-brain:9000"
)

# 日志函数
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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 检测并加载环境配置文件
load_environment() {
    # 先清理所有数据库相关环境变量，避免继承之前的值
    unset POSTGRESQL_HOST POSTGRESQL_PORT POSTGRESQL_USER POSTGRESQL_PASSWORD POSTGRESQL_DATABASE POSTGRESQL_SSL_MODE 2>/dev/null
    unset MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD MYSQL_DATABASE 2>/dev/null
    
    ENV_FILE=""
    if [ -f "$PROJECT_ROOT/configs/local.env" ]; then
        ENV_FILE="$PROJECT_ROOT/configs/local.env"
        log_info "📋 使用配置文件: configs/local.env"
    elif [ -f "$PROJECT_ROOT/configs/dev.env" ]; then
        ENV_FILE="$PROJECT_ROOT/configs/dev.env"
        log_info "📋 使用配置文件: configs/dev.env"
    fi

    # 加载环境变量
    if [ -n "$ENV_FILE" ]; then
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
            unset POSTGRESQL_HOST POSTGRESQL_PORT POSTGRESQL_USER POSTGRESQL_PASSWORD POSTGRESQL_DATABASE POSTGRESQL_SSL_MODE 2>/dev/null
        elif [ $HAS_POSTGRESQL -eq 1 ] && [ $HAS_MYSQL -eq 0 ]; then
            log_info "🔄 检测到 PostgreSQL 配置，使用 PostgreSQL 数据库"
            unset MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD MYSQL_DATABASE 2>/dev/null
        elif [ $HAS_MYSQL -eq 1 ] && [ $HAS_POSTGRESQL -eq 1 ]; then
            log_error "❌ 检测到 MySQL 和 PostgreSQL 配置同时存在！请只启用其中一个"
            log_error "建议: 注释掉不需要的数据库配置"
            exit 1
        else
            log_error "❌ 未检测到任何数据库配置！"
            log_error "请在 configs/local.env 中配置 MySQL 或 PostgreSQL"
            exit 1
        fi
    else
        log_warning "⚠️  未找到环境配置文件，使用默认配置"
    fi
}

# 加载环境变量（在主函数之前调用）
load_environment

# 检查本地服务
check_local_services() {
    log_step "检查本地服务状态..."
    
    # 检查PostgreSQL
    local started="false"

    if command -v brew >/dev/null 2>&1; then
        if brew services list | grep "$PG_VERSION" | grep -q "started"; then
            log_success "PostgreSQL 14 运行正常"
            started="true"
        else
            log_warning "PostgreSQL 14 未运行，正在通过 brew services 启动..."
            if brew services start "$PG_VERSION" >/dev/null 2>&1; then
                started="true"
                sleep 3
            else
                log_warning "brew services 启动失败"
            fi
        fi
    fi

    if [ "$started" = "false" ] && [ -n "$PG_CTL_BIN" ] && [ -d "$PG_DATA_DIR" ]; then
        if "$PG_CTL_BIN" status -D "$PG_DATA_DIR" > /dev/null 2>&1; then
            log_success "PostgreSQL 14 运行正常"
            started="true"
        else
            log_warning "PostgreSQL 14 未运行，正在通过 pg_ctl 启动..."
            if "$PG_CTL_BIN" -D "$PG_DATA_DIR" -l "$PG_LOG_FILE" start; then
                started="true"
                sleep 3
            else
                log_warning "pg_ctl 启动失败"
            fi
        fi
    fi

    if [ "$started" = "false" ]; then
        log_error "未能启动 PostgreSQL 14，请检查环境配置"
        exit 1
    fi
    
    # 检查Redis
    if brew services list | grep redis | grep -q "started"; then
        log_success "Redis 运行正常"
    else
        log_warning "Redis 未运行，正在启动..."
        brew services start redis
        sleep 2
    fi
    
    # 验证数据库连接
    local pg_host=${POSTGRESQL_HOST:-${POSTGRES_HOST:-localhost}}
    local pg_port=${POSTGRESQL_PORT:-${POSTGRES_PORT:-5432}}
    local pg_user=${POSTGRESQL_USER:-${POSTGRES_USER:-postgres}}
    local pg_database=${POSTGRESQL_DATABASE:-${POSTGRES_DB:-zervigo_mvp}}

    if PGPASSWORD="${POSTGRESQL_PASSWORD:-${POSTGRES_PASSWORD:-}}" psql -h "$pg_host" -p "$pg_port" -U "$pg_user" -d "$pg_database" -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "PostgreSQL 数据库连接正常"
    else
        log_error "PostgreSQL 数据库连接失败！"
        exit 1
    fi
    
    # 验证Redis连接
    if redis-cli ping > /dev/null 2>&1; then
        log_success "Redis 连接正常"
    else
        log_error "Redis 连接失败！"
        exit 1
    fi
}

# 检查端口是否被占用
check_port() {
    local port=$1
    local service_name=$2
    
    if lsof -i :$port > /dev/null 2>&1; then
        log_warning "端口 $port ($service_name) 已被占用"
        return 1
    else
        log_success "端口 $port ($service_name) 可用"
        return 0
    fi
}

# 启动认证服务
start_auth_service() {
    log_step "启动认证服务..."
    
    if check_port 8207 "auth-service"; then
        cd "$PROJECT_ROOT/services/core/auth"
        nohup go run . > "$LOG_DIR/auth-service.log" 2>&1 &
        echo $! > "$LOG_DIR/auth-service.pid"
        sleep 3
        
        if curl -s http://localhost:8207/health > /dev/null 2>&1; then
            log_success "认证服务启动成功 (端口: 8207)"
        else
            log_error "认证服务启动失败"
            return 1
        fi
    else
        log_warning "认证服务端口被占用，跳过启动"
    fi
}

# 启动用户服务
start_user_service() {
    log_step "启动用户服务..."
    
    if check_port 8082 "user-service"; then
        cd "$PROJECT_ROOT/services/core/user"
        
        # 编译服务
        if [ ! -f "user-service" ] || [ "main.go" -nt "user-service" ]; then
            log_info "编译用户服务..."
            go build -o user-service .
        fi
        
        # 导出环境变量到后台进程
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
            ./user-service > "$LOG_DIR/user-service.log" 2>&1 &
        echo $! > "$LOG_DIR/user-service.pid"
        sleep 3
        
        if curl -s http://localhost:8082/health > /dev/null 2>&1; then
            log_success "用户服务启动成功 (端口: 8082)"
        else
            log_error "用户服务启动失败"
            return 1
        fi
    else
        log_warning "用户服务端口被占用，跳过启动"
    fi
}

# 启动职位服务
start_job_service() {
    log_step "启动职位服务..."
    
    if check_port 8084 "job-service"; then
        cd "$PROJECT_ROOT/services/business/job"
        # 导出环境变量到后台进程
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
            go run . > "$LOG_DIR/job-service.log" 2>&1 &
        echo $! > "$LOG_DIR/job-service.pid"
        sleep 3
        
        if curl -s http://localhost:8084/health > /dev/null 2>&1; then
            log_success "职位服务启动成功 (端口: 8084)"
        else
            log_error "职位服务启动失败"
            return 1
        fi
    else
        log_warning "职位服务端口被占用，跳过启动"
    fi
}

# 启动简历服务
start_resume_service() {
    log_step "启动简历服务..."
    
    if check_port 8085 "resume-service"; then
        cd "$PROJECT_ROOT/services/business/resume"
        # 导出环境变量到后台进程
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
            go run $(find . -maxdepth 1 -name '*.go' ! -name 'simple_main.go' -print | tr '\n' ' ') > "$LOG_DIR/resume-service.log" 2>&1 &
        echo $! > "$LOG_DIR/resume-service.pid"
        sleep 3
        
        if curl -s http://localhost:8085/health > /dev/null 2>&1; then
            log_success "简历服务启动成功 (端口: 8085)"
        else
            log_error "简历服务启动失败"
            return 1
        fi
    else
        log_warning "简历服务端口被占用，跳过启动"
    fi
}

# 启动企业服务
start_company_service() {
    log_step "启动企业服务..."
    
    if check_port 8083 "company-service"; then
        cd "$PROJECT_ROOT/services/business/company"
        # 导出环境变量到后台进程
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
            go run $(find . -maxdepth 1 -name '*.go' ! -name 'simple_main.go' -print | tr '\n' ' ') > "$LOG_DIR/company-service.log" 2>&1 &
        echo $! > "$LOG_DIR/company-service.pid"
        sleep 3
        
        if curl -s http://localhost:8083/health > /dev/null 2>&1; then
            log_success "企业服务启动成功 (端口: 8083)"
        else
            log_error "企业服务启动失败"
            return 1
        fi
    else
        log_warning "企业服务端口被占用，跳过启动"
    fi
}

# 启动AI服务
start_ai_service() {
    log_step "启动AI服务..."
    
    if check_port 8100 "ai-service"; then
        cd "$PROJECT_ROOT/src/ai-service-python"
        
        # 检查Python虚拟环境
        if [ ! -d "venv" ]; then
            log_info "创建Python虚拟环境..."
            python3 -m venv venv
        fi
        
        # 激活虚拟环境并启动服务
        source venv/bin/activate
        pip install -r requirements.txt > /dev/null 2>&1
        
        nohup python ai_service_with_zervigo.py > "$LOG_DIR/ai-service.log" 2>&1 &
        echo $! > "$LOG_DIR/ai-service.pid"
        sleep 5
        
        if curl -s http://localhost:8100/health > /dev/null 2>&1; then
            log_success "AI服务启动成功 (端口: 8100)"
        else
            log_error "AI服务启动失败"
            return 1
        fi
    else
        log_warning "AI服务端口被占用，跳过启动"
    fi
}

# 启动区块链服务
start_blockchain_service() {
    log_step "启动区块链服务..."
    
    if check_port 8208 "blockchain-service"; then
        cd "$PROJECT_ROOT/services/infrastructure/blockchain"
        # 导出环境变量到后台进程
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
            go run . > "$LOG_DIR/blockchain-service.log" 2>&1 &
        echo $! > "$LOG_DIR/blockchain-service.pid"
        sleep 3
        
        if curl -s http://localhost:8208/health > /dev/null 2>&1; then
            log_success "区块链服务启动成功 (端口: 8208)"
        else
            log_error "区块链服务启动失败"
            return 1
        fi
    else
        log_warning "区块链服务端口被占用，跳过启动"
    fi
}

# 启动中央大脑
start_central_brain() {
    log_step "启动中央大脑 (API Gateway)..."
    
    if check_port 9000 "central-brain"; then
        cd "$PROJECT_ROOT/shared/central-brain"
        # 导出环境变量到后台进程
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
            go run . > "$LOG_DIR/central-brain.log" 2>&1 &
        echo $! > "$LOG_DIR/central-brain.pid"
        sleep 3
        
        if curl -s http://localhost:9000/health > /dev/null 2>&1; then
            log_success "中央大脑启动成功 (端口: 9000)"
        else
            log_error "中央大脑启动失败"
            return 1
        fi
    else
        log_warning "中央大脑端口被占用，跳过启动"
    fi
}

# 健康检查
health_check() {
    log_step "执行健康检查..."
    
    local services=("auth-service:8207" "user-service:8082" "job-service:8084" "resume-service:8085" "company-service:8083" "ai-service:8100" "blockchain-service:8208" "central-brain:9000")
    
    for service in "${services[@]}"; do
        local name=$(echo $service | cut -d: -f1)
        local port=$(echo $service | cut -d: -f2)
        
        if curl -s http://localhost:$port/health > /dev/null 2>&1; then
            log_success "$name 健康检查通过"
        else
            log_warning "$name 健康检查失败"
        fi
    done
}

# 显示服务状态
show_status() {
    log_step "服务状态总览"
    echo "================================"
    echo "📊 服务访问地址："
    echo "   中央大脑 (API Gateway): http://localhost:9000"
    echo "   统一认证服务: http://localhost:8207"
    echo "   用户服务: http://localhost:8082"
    echo "   职位服务: http://localhost:8084"
    echo "   简历服务: http://localhost:8085"
    echo "   企业服务: http://localhost:8083"
    echo "   AI服务: http://localhost:8100"
    echo "   区块链服务: http://localhost:8208"
    echo ""
    echo "🗄️ 数据库连接："
    echo "   PostgreSQL: postgres://$(whoami)@localhost:5432/zervigo_mvp"
    echo "   Redis: redis://localhost:6379"
    echo ""
    echo "👤 默认管理员账号："
    echo "   用户名: admin"
    echo "   密码: admin123"
    echo "   邮箱: admin@zervigo.com"
    echo ""
    echo "🔧 管理命令："
    echo "   查看日志: tail -f logs/[service-name].log"
    echo "   停止服务: ./scripts/stop-local-services.sh"
    echo "   重启服务: ./scripts/restart-local-services.sh"
}

# 主函数
main() {
    log_info "开始启动 Zervigo 本地开发环境..."
    
    # 检查本地服务
    check_local_services
    
    # 启动微服务
    start_auth_service
    start_user_service
    start_job_service
    start_resume_service
    start_company_service
    start_ai_service
    start_blockchain_service
    start_central_brain
    
    # 等待服务启动
    log_info "等待所有服务启动完成..."
    sleep 10
    
    # 健康检查
    health_check
    
    # 显示状态
    show_status
    
    log_success "🎉 Zervigo 本地开发环境启动完成！"
}

# 执行主函数
main "$@"
