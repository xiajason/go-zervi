#!/bin/bash

# Zervigo 本地服务停止脚本
# 停止所有本地运行的微服务

set -e

echo "🛑 停止 Zervigo 本地开发环境"
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

# PostgreSQL 控制工具配置
BREW_PREFIX=$(brew --prefix 2>/dev/null)
PG_VERSION="postgresql@14"
PG_CTL_BIN="${BREW_PREFIX}/opt/${PG_VERSION}/bin/pg_ctl"
PG_DATA_DIR="${BREW_PREFIX}/var/${PG_VERSION}"

if [ ! -x "$PG_CTL_BIN" ]; then
    PG_CTL_BIN="$(command -v pg_ctl 2>/dev/null)"
fi
# 停止 PostgreSQL
stop_postgresql() {
    local stopped=false

    if command -v brew >/dev/null 2>&1 && brew services list | grep "$PG_VERSION" | grep -q "started"; then
        log_info "停止 PostgreSQL 14 (brew services stop)..."
        if brew services stop "$PG_VERSION" >/dev/null 2>&1; then
            stopped=true
            sleep 1
        else
            log_warning "brew services 停止失败"
        fi
    fi

    if [ "$stopped" = false ] && [ -n "$PG_CTL_BIN" ] && [ -d "$PG_DATA_DIR" ]; then
        if "$PG_CTL_BIN" status -D "$PG_DATA_DIR" > /dev/null 2>&1; then
            log_info "使用 pg_ctl 停止 PostgreSQL 14..."
            "$PG_CTL_BIN" -D "$PG_DATA_DIR" stop -m fast >/dev/null 2>&1 || log_warning "pg_ctl 停止命令执行异常"
            stopped=true
        else
            stopped=true
        fi
    fi

    if [ "$stopped" = false ]; then
        log_warning "未能确认 PostgreSQL 14 是否已停止"
    fi
}


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

# 停止服务函数
stop_service() {
    local service_name=$1
    local pid_file="$LOG_DIR/${service_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_info "停止 $service_name (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 2
            
            # 强制杀死如果还在运行
            if ps -p "$pid" > /dev/null 2>&1; then
                log_warning "强制停止 $service_name..."
                kill -9 "$pid" 2>/dev/null || true
                sleep 1
            fi
            
            log_success "$service_name 已停止"
        else
            log_warning "$service_name 进程不存在"
        fi
        rm -f "$pid_file"
    else
        log_warning "$service_name PID文件不存在"
    fi
}

# 停止所有服务
stop_all_services() {
    log_info "停止所有微服务..."
    
    local services=("auth-service" "user-service" "job-service" "resume-service" "company-service" "ai-service" "blockchain-service" "central-brain")
    
    for service in "${services[@]}"; do
        stop_service "$service"
    done
}

# 清理端口占用
cleanup_ports() {
    log_info "清理端口占用..."
    
    local ports=(8207 8082 8084 8085 8083 8100 8208 9000)
    
    for port in "${ports[@]}"; do
        local pid=$(lsof -ti :$port 2>/dev/null)
        if [ ! -z "$pid" ]; then
            log_info "清理端口 $port (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 1
            # 强制清理（二次检查）
            if lsof -ti :$port > /dev/null 2>&1; then
                local retry_pid=$(lsof -ti :$port 2>/dev/null)
                if [ ! -z "$retry_pid" ]; then
                    kill -9 "$retry_pid" 2>/dev/null || true
                    sleep 1
                fi
            fi
        fi
    done
}

# 清理日志文件
cleanup_logs() {
    log_info "清理日志文件..."
    
    if [ -d "$LOG_DIR" ]; then
        rm -f "$LOG_DIR"/*.log
        rm -f "$LOG_DIR"/*.pid
        log_success "日志文件已清理"
    else
        log_warning "日志目录不存在"
    fi
}

# 显示状态
show_status() {
    log_info "检查服务状态..."
    
    local services=("auth-service:8207" "user-service:8082" "job-service:8084" "resume-service:8085" "company-service:8083" "ai-service:8100" "blockchain-service:8208" "central-brain:9000")
    
    local running_count=0
    local total_count=${#services[@]}
    
    for service in "${services[@]}"; do
        local name=$(echo $service | cut -d: -f1)
        local port=$(echo $service | cut -d: -f2)
        
        if lsof -i :$port > /dev/null 2>&1; then
            log_warning "$name 仍在运行 (端口: $port)"
            running_count=$((running_count + 1))
        else
            log_success "$name 已停止"
        fi
    done
    
    echo ""
    if [ $running_count -eq 0 ]; then
        log_success "所有服务已停止"
    else
        log_warning "$running_count/$total_count 服务仍在运行"
    fi
}

# 主函数
main() {
    log_info "开始停止 Zervigo 本地开发环境..."
    
    # 停止所有服务
    stop_all_services
    
    # 停止 PostgreSQL
    stop_postgresql

    # 清理端口占用
    cleanup_ports
    
    # 清理日志文件
    cleanup_logs
    
    # 显示状态
    show_status
    
    log_success "🎉 Zervigo 本地开发环境已停止！"
}

# 执行主函数
main "$@"
