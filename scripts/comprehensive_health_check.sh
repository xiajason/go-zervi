#!/bin/bash
# 腾讯云服务器全面健康检查脚本
# 用途: 检查系统、服务、容器、数据库的健康状况
# 部署位置: /opt/scripts/comprehensive_health_check.sh
# Cron: */10 * * * * /opt/scripts/comprehensive_health_check.sh >> /var/log/health_check.log 2>&1

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log() { echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}[✅]${NC} $1"; }
warning() { echo -e "${YELLOW}[⚠️ ]${NC} $1"; }
error() { echo -e "${RED}[❌]${NC} $1"; }

echo ""
echo "=========================================="
echo "数据库集群健康检查"
echo "检查时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "=========================================="
echo ""

# 初始化健康分数
SCORE=100

# ==================== 1. 系统资源检查 ====================
log "1. 系统资源状态"

# CPU检查
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
CPU_INT=${CPU_USAGE%.*}
echo "  CPU使用率: ${CPU_USAGE}%"
if [ "$CPU_INT" -gt 80 ]; then
    warning "CPU使用率过高 (-10分)"
    SCORE=$((SCORE - 10))
fi

# 内存检查
MEM_TOTAL=$(free -m | awk 'NR==2{print $2}')
MEM_USED=$(free -m | awk 'NR==2{print $3}')
MEM_AVAILABLE=$(free -m | awk 'NR==2{print $7}')
MEM_PERCENT=$(awk "BEGIN {printf \"%.1f\", ($MEM_USED/$MEM_TOTAL)*100}")

echo "  内存使用: ${MEM_USED}MB / ${MEM_TOTAL}MB (${MEM_PERCENT}%)"
echo "  可用内存: ${MEM_AVAILABLE}MB"

MEM_PERCENT_INT=${MEM_PERCENT%.*}
if [ "$MEM_PERCENT_INT" -gt 85 ]; then
    error "内存使用率超过85% (-20分)"
    SCORE=$((SCORE - 20))
elif [ "$MEM_PERCENT_INT" -gt 70 ]; then
    warning "内存使用率超过70% (-10分)"
    SCORE=$((SCORE - 10))
fi

if [ "$MEM_AVAILABLE" -lt 200 ]; then
    error "可用内存不足200MB，系统存在OOM风险 (-15分)"
    SCORE=$((SCORE - 15))
fi

# 磁盘检查
DISK_USAGE=$(df / | awk 'NR==2{print $5}' | sed 's/%//')
DISK_AVAILABLE=$(df -h / | awk 'NR==2{print $4}')

echo "  磁盘使用: ${DISK_USAGE}%"
echo "  磁盘可用: ${DISK_AVAILABLE}"

if [ "$DISK_USAGE" -gt 80 ]; then
    warning "磁盘使用率超过80% (-10分)"
    SCORE=$((SCORE - 10))
fi

# 系统负载
LOAD_1MIN=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
echo "  系统负载: ${LOAD_1MIN} (1分钟)"

echo ""

# ==================== 2. 容器状态检查 ====================
log "2. 容器运行状态"

# 期望的容器列表
EXPECTED_CONTAINERS=(
    "test-mysql"
    "test-postgres"
    "test-redis"
    "test-mongodb"
    "test-neo4j"
    "test-elasticsearch"
    "test-weaviate"
)

RUNNING_COUNT=0
for container in "${EXPECTED_CONTAINERS[@]}"; do
    if docker ps | grep -q "$container"; then
        success "$container 运行中"
        RUNNING_COUNT=$((RUNNING_COUNT + 1))
    else
        error "$container 未运行"
        SCORE=$((SCORE - 10))
    fi
done

echo "  运行中容器: ${RUNNING_COUNT}/${#EXPECTED_CONTAINERS[@]}"

if [ "$RUNNING_COUNT" -eq "${#EXPECTED_CONTAINERS[@]}" ]; then
    success "所有数据库容器正常运行"
else
    error "部分容器未运行 (-${#EXPECTED_CONTAINERS[@]}0分)"
fi

echo ""

# ==================== 3. 服务状态检查 ====================
log "3. 应用服务状态"

# Zervigo检查
if pgrep -f "unified-auth" > /dev/null; then
    ZERVIGO_PID=$(pgrep -f "unified-auth")
    success "Zervigo运行中 (PID: $ZERVIGO_PID)"
    
    # 健康检查
    ZERVIGO_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8207/health)
    if [ "$ZERVIGO_HEALTH" = "200" ]; then
        success "Zervigo健康检查通过"
    else
        warning "Zervigo健康检查异常 (HTTP $ZERVIGO_HEALTH)"
        SCORE=$((SCORE - 5))
    fi
else
    error "Zervigo未运行 (-15分)"
    SCORE=$((SCORE - 15))
fi

# AI Service检查
if pgrep -f "ai_service_with_zervigo" > /dev/null; then
    AI_PID=$(pgrep -f "ai_service_with_zervigo")
    success "AI Service运行中 (PID: $AI_PID)"
    
    # 健康检查
    AI_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8100/health)
    if [ "$AI_HEALTH" = "200" ]; then
        success "AI Service健康检查通过"
    else
        warning "AI Service健康检查异常 (HTTP $AI_HEALTH)"
        SCORE=$((SCORE - 5))
    fi
else
    error "AI Service未运行 (-15分)"
    SCORE=$((SCORE - 15))
fi

echo ""

# ==================== 4. MySQL状态检查 ====================
log "4. MySQL数据库状态"

if docker exec test-mysql mysqladmin -uroot -ptest_mysql_password ping 2>/dev/null | grep -q "alive"; then
    success "MySQL服务正常"
    
    # 检查三个数据库
    DBS=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW DATABASES LIKE 'jobfirst%';" 2>/dev/null | tail -n +2)
    DB_COUNT=$(echo "$DBS" | wc -l)
    echo "  数据库数量: $DB_COUNT"
    echo "$DBS" | sed 's/^/    - /'
    
    if [ "$DB_COUNT" -ne 3 ]; then
        warning "数据库数量不正确（应为3个）"
        SCORE=$((SCORE - 5))
    fi
    
    # 连接数检查
    CONNECTIONS=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW STATUS LIKE 'Threads_connected';" 2>/dev/null | awk 'NR==2{print $2}')
    MAX_CONNECTIONS=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW VARIABLES LIKE 'max_connections';" 2>/dev/null | awk 'NR==2{print $2}')
    
    echo "  连接数: $CONNECTIONS / $MAX_CONNECTIONS"
    
    if [ "$CONNECTIONS" -gt 250 ]; then
        warning "MySQL连接数过高"
        SCORE=$((SCORE - 5))
    fi
    
    # 主从复制状态
    SLAVE_IO=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Slave_IO_Running:" | awk '{print $2}')
    SLAVE_SQL=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Slave_SQL_Running:" | awk '{print $2}')
    SECONDS_BEHIND=$(docker exec test-mysql mysql -uroot -ptest_mysql_password -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Seconds_Behind_Master:" | awk '{print $2}')
    
    if [ "$SLAVE_IO" = "Yes" ] && [ "$SLAVE_SQL" = "Yes" ]; then
        success "主从复制正常 (延迟: ${SECONDS_BEHIND}秒)"
    else
        error "主从复制异常 (IO: $SLAVE_IO, SQL: $SLAVE_SQL) (-20分)"
        SCORE=$((SCORE - 20))
    fi
else
    error "MySQL服务异常 (-20分)"
    SCORE=$((SCORE - 20))
fi

echo ""

# ==================== 5. Redis状态检查 ====================
log "5. Redis缓存状态"

if docker exec test-redis redis-cli -a test_redis_password PING 2>/dev/null | grep -q "PONG"; then
    success "Redis服务正常"
    
    # 内存使用
    REDIS_MEM=$(docker exec test-redis redis-cli -a test_redis_password INFO memory 2>/dev/null | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r')
    echo "  内存使用: $REDIS_MEM"
    
    # 连接数
    REDIS_CLIENTS=$(docker exec test-redis redis-cli -a test_redis_password INFO clients 2>/dev/null | grep "connected_clients:" | cut -d: -f2 | tr -d '\r')
    echo "  连接数: $REDIS_CLIENTS"
else
    error "Redis服务异常 (-10分)"
    SCORE=$((SCORE - 10))
fi

echo ""

# ==================== 6. Neo4j状态检查 ====================
log "6. Neo4j图数据库状态 (DAO/区块链)"

NEO4J_HTTP=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:7474)
if [ "$NEO4J_HTTP" = "200" ]; then
    success "Neo4j HTTP接口正常"
    
    # CPU使用率
    NEO4J_CPU=$(docker stats --no-stream --format "{{.CPUPerc}}" test-neo4j)
    echo "  CPU使用: $NEO4J_CPU"
    
    NEO4J_CPU_INT=$(echo $NEO4J_CPU | sed 's/%//' | cut -d'.' -f1)
    if [ "$NEO4J_CPU_INT" -gt 50 ]; then
        warning "Neo4j CPU使用率较高"
        SCORE=$((SCORE - 5))
    fi
else
    warning "Neo4j HTTP接口异常 (HTTP $NEO4J_HTTP)"
    SCORE=$((SCORE - 5))
fi

echo ""

# ==================== 7. 容器资源使用检查 ====================
log "7. 容器资源使用"

echo "  容器内存使用排行:"
docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}" | \
    grep test- | sort -k3 -rn | head -3 | sed 's/^/    /'

echo ""

# ==================== 8. 磁盘空间检查 ====================
log "8. 磁盘空间检查"

DOCKER_DIR_SIZE=$(du -sh /var/lib/docker 2>/dev/null | awk '{print $1}')
echo "  Docker目录大小: $DOCKER_DIR_SIZE"

SERVICES_SIZE=$(du -sh /opt/services 2>/dev/null | awk '{print $1}')
echo "  服务目录大小: $SERVICES_SIZE"

# 检查日志大小
LARGE_LOGS=$(find /opt/services -name "*.log" -size +100M 2>/dev/null)
if [ -n "$LARGE_LOGS" ]; then
    warning "发现大日志文件:"
    echo "$LARGE_LOGS" | sed 's/^/    /'
fi

echo ""

# ==================== 9. SSH隧道检查 ====================
log "9. SSH隧道状态 (跨云同步)"

# MySQL隧道
if pgrep -f "ssh.*13306" > /dev/null; then
    success "MySQL SSH隧道运行中"
else
    error "MySQL SSH隧道未运行 (-10分)"
    SCORE=$((SCORE - 10))
fi

# Redis隧道
if pgrep -f "ssh.*16379" > /dev/null; then
    success "Redis SSH隧道运行中"
else
    error "Redis SSH隧道未运行 (-10分)"
    SCORE=$((SCORE - 10))
fi

echo ""

# ==================== 10. 端口监听检查 ====================
log "10. 关键端口监听"

CRITICAL_PORTS=(8207 8100 3306 6379 7687 9200)
for port in "${CRITICAL_PORTS[@]}"; do
    if ss -tlnp | grep -q ":$port "; then
        success "端口 $port 正常监听"
    else
        error "端口 $port 未监听 (-5分)"
        SCORE=$((SCORE - 5))
    fi
done

echo ""

# ==================== 健康分数评估 ====================
echo "=========================================="
echo "健康分数: ${SCORE}/100"

if [ "$SCORE" -ge 90 ]; then
    success "系统状态: 优秀 ⭐⭐⭐⭐⭐"
    STATUS="EXCELLENT"
elif [ "$SCORE" -ge 70 ]; then
    warning "系统状态: 良好（建议优化）⭐⭐⭐⭐"
    STATUS="GOOD"
elif [ "$SCORE" -ge 50 ]; then
    warning "系统状态: 警告（需要处理）⭐⭐⭐"
    STATUS="WARNING"
else
    error "系统状态: 危险（立即处理）⭐"
    STATUS="CRITICAL"
fi
echo "=========================================="

# ==================== 生成建议 ====================
echo ""
log "优化建议:"

if [ "$MEM_AVAILABLE" -lt 500 ]; then
    warning "建议: 设置容器内存限制或添加Swap空间"
fi

if [ "$RUNNING_COUNT" -lt 7 ]; then
    warning "建议: 检查并重启停止的容器"
fi

if [ "$SLAVE_IO" != "Yes" ] || [ "$SLAVE_SQL" != "Yes" ]; then
    warning "建议: 检查并修复MySQL主从复制"
fi

# ==================== 保存健康报告 ====================
REPORT_FILE="/opt/monitoring/health_reports/health_$(date +%Y%m%d_%H%M%S).json"
mkdir -p /opt/monitoring/health_reports

cat > "$REPORT_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "score": $SCORE,
  "status": "$STATUS",
  "system": {
    "cpu_usage": $CPU_USAGE,
    "memory_percent": $MEM_PERCENT,
    "memory_available_mb": $MEM_AVAILABLE,
    "disk_usage": $DISK_USAGE
  },
  "containers": {
    "running": $RUNNING_COUNT,
    "expected": ${#EXPECTED_CONTAINERS[@]}
  },
  "services": {
    "zervigo_running": $(pgrep -f unified-auth > /dev/null && echo "true" || echo "false"),
    "ai_service_running": $(pgrep -f ai_service > /dev/null && echo "true" || echo "false")
  },
  "databases": {
    "mysql_healthy": $(docker exec test-mysql mysqladmin -uroot -ptest_mysql_password ping 2>/dev/null | grep -q alive && echo "true" || echo "false"),
    "redis_healthy": $(docker exec test-redis redis-cli -a test_redis_password PING 2>/dev/null | grep -q PONG && echo "true" || echo "false"),
    "replication_lag_seconds": $SECONDS_BEHIND
  }
}
EOF

log "健康报告已保存: $REPORT_FILE"

# ==================== 告警通知（如果分数过低）====================
if [ "$SCORE" -lt 70 ]; then
    echo ""
    error "⚠️  系统健康分数低于70，请立即检查！"
    
    # 这里可以添加邮件或微信通知
    # send_alert "系统健康分数: $SCORE/100"
fi

echo ""
log "健康检查完成"
echo ""

# 返回状态码（用于监控系统）
if [ "$SCORE" -ge 90 ]; then
    exit 0  # 优秀
elif [ "$SCORE" -ge 70 ]; then
    exit 0  # 良好
elif [ "$SCORE" -ge 50 ]; then
    exit 1  # 警告
else
    exit 2  # 危险
fi

