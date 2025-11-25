#!/bin/bash

# 测试脚本：用于验证start和stop脚本的功能
# 测试目标：auth-service, user-service, central-brain

set -e

echo "🧪 测试 start 和 stop 脚本功能"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$PROJECT_ROOT/logs"
mkdir -p "$LOG_DIR"

# 日志函数
log_info() { echo -e "${BLUE}[TEST]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

# 测试服务
TEST_SERVICES=(
    "auth-service:8207:services/core/auth"
    "user-service:8082:services/core/user"
    "central-brain:9000:shared/central-brain"
)

# 加载环境变量
load_env() {
    local env_file="$PROJECT_ROOT/configs/local.env"
    if [ -f "$env_file" ]; then
        log_info "加载环境变量: $env_file"
        set -a
        source <(cat "$env_file" | grep "^[^#]" | grep -v "^$" | sed 's/#.*$//')
        set +a
    fi
}

# 启动服务函数
start_test_service() {
    local service_name=$1
    local port=$2
    local service_dir=$3
    
    log_info "启动 $service_name..."
    
    # 检查端口
    if lsof -i :$port > /dev/null 2>&1; then
        log_error "端口 $port 已被占用"
        return 1
    fi
    
    # 加载环境变量（在后台启动前）
    load_env
    
    # 进入服务目录
    cd "$PROJECT_ROOT/$service_dir"
    
    # 启动服务（使用环境变量）
    nohup go run main.go > "$LOG_DIR/${service_name}.log" 2>&1 &
    local pid=$!
    echo $pid > "$LOG_DIR/${service_name}.pid"
    
    # 等待启动
    sleep 3
    
    # 健康检查
    if curl -s http://localhost:$port/health > /dev/null 2>&1; then
        log_success "$service_name 启动成功 (PID: $pid, Port: $port)"
        return 0
    else
        log_error "$service_name 启动失败"
        tail -20 "$LOG_DIR/${service_name}.log"
        return 1
    fi
}

# 停止服务函数
stop_test_service() {
    local service_name=$1
    local pid_file="$LOG_DIR/${service_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_info "停止 $service_name (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 2
            
            # 强制停止
            if ps -p "$pid" > /dev/null 2>&1; then
                kill -9 "$pid" 2>/dev/null || true
            fi
            
            log_success "$service_name 已停止"
        fi
        rm -f "$pid_file"
    fi
}

# 测试1: 启动服务
test_start() {
    echo ""
    echo "📦 测试1: 启动核心服务"
    echo "--------------------------------"
    
    for service in "${TEST_SERVICES[@]}"; do
        IFS=':' read -r name port dir <<< "$service"
        if ! start_test_service "$name" "$port" "$dir"; then
            return 1
        fi
    done
    
    echo ""
    log_success "所有服务启动成功！"
    return 0
}

# 测试2: 健康检查
test_health() {
    echo ""
    echo "🏥 测试2: 健康检查"
    echo "--------------------------------"
    
    for service in "${TEST_SERVICES[@]}"; do
        IFS=':' read -r name port dir <<< "$service"
        if curl -s http://localhost:$port/health > /dev/null 2>&1; then
            log_success "$name 健康检查通过"
        else
            log_error "$name 健康检查失败"
            return 1
        fi
    done
    
    return 0
}

# 测试3: 停止服务
test_stop() {
    echo ""
    echo "🛑 测试3: 停止服务"
    echo "--------------------------------"
    
    for service in "${TEST_SERVICES[@]}"; do
        IFS=':' read -r name port dir <<< "$service"
        stop_test_service "$name"
    done
    
    # 验证端口已释放
    sleep 2
    for service in "${TEST_SERVICES[@]}"; do
        IFS=':' read -r name port dir <<< "$service"
        if lsof -i :$port > /dev/null 2>&1; then
            log_error "端口 $port 仍被占用"
            return 1
        else
            log_success "$name 端口已释放"
        fi
    done
    
    return 0
}

# 主测试流程
main() {
    echo "开始测试..."
    echo ""
    
    # 清理旧状态
    log_info "清理旧状态..."
    for service in "${TEST_SERVICES[@]}"; do
        IFS=':' read -r name port dir <<< "$service"
        stop_test_service "$name"
    done
    rm -f "$LOG_DIR"/*.log
    rm -f "$LOG_DIR"/*.pid
    sleep 2
    
    # 执行测试
    if test_start && test_health && test_stop; then
        echo ""
        log_success "🎉 所有测试通过！"
        echo ""
        echo "测试总结："
        echo "  ✓ 服务启动功能正常"
        echo "  ✓ 健康检查功能正常"
        echo "  ✓ 服务停止功能正常"
        return 0
    else
        echo ""
        log_error "❌ 测试失败"
        return 1
    fi
}

# 执行测试
main "$@"
