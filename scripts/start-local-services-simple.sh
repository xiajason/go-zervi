#!/bin/bash

# Zervigo 本地开发环境启动脚本
# 简化版本，兼容macOS bash

set -e

echo "🚀 启动 Zervigo 本地开发环境"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$PROJECT_ROOT/logs"
mkdir -p "$LOG_DIR"

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

# 检查本地服务
check_local_services() {
    log_info "检查本地服务状态..."
    
    # 检查PostgreSQL
    if brew services list | grep postgresql@14 | grep -q "started"; then
        log_success "PostgreSQL 14 运行正常"
    else
        log_warning "PostgreSQL 14 未运行，正在启动..."
        brew services start postgresql@14
        sleep 3
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
    if psql -U $(whoami) -d zervigo_mvp -c "SELECT 1;" > /dev/null 2>&1; then
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
    log_info "启动认证服务..."
    
    if check_port 8207 "auth-service"; then
        cd "$PROJECT_ROOT/src/auth-service-go"
        nohup go run main.go > "$LOG_DIR/auth-service.log" 2>&1 &
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
    log_info "启动用户服务..."
    
    if check_port 8082 "user-service"; then
        cd "$PROJECT_ROOT/src/microservices/user-service"
        nohup go run main.go > "$LOG_DIR/user-service.log" 2>&1 &
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
    log_info "启动职位服务..."
    
    if check_port 8084 "job-service"; then
        cd "$PROJECT_ROOT/src/microservices/job-service"
        nohup go run main.go > "$LOG_DIR/job-service.log" 2>&1 &
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

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    local services=("auth-service:8207" "user-service:8082" "job-service:8084")
    
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
    log_info "服务状态总览"
    echo "================================"
    echo "📊 服务访问地址："
    echo "   统一认证服务: http://localhost:8207"
    echo "   用户服务: http://localhost:8082"
    echo "   职位服务: http://localhost:8084"
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
}

# 主函数
main() {
    log_info "开始启动 Zervigo 本地开发环境..."
    
    # 检查本地服务
    check_local_services
    
    # 启动核心微服务
    start_auth_service
    start_user_service
    start_job_service
    
    # 等待服务启动
    log_info "等待所有服务启动完成..."
    sleep 5
    
    # 健康检查
    health_check
    
    # 显示状态
    show_status
    
    log_success "🎉 Zervigo 核心服务启动完成！"
}

# 执行主函数
main "$@"
