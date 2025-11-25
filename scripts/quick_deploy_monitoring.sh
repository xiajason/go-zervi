#!/bin/bash
# 腾讯云服务器监控平台一键部署脚本
# 功能: 部署Grafana + Prometheus + Portainer + Exporters
# 适用: Ubuntu Server on 101.33.251.158
# 作者: AI Assistant
# 日期: 2025-10-10

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  腾讯云监控平台一键部署工具${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# 检查是否为root或有sudo权限
if [ "$EUID" -ne 0 ]; then 
    log_warning "建议使用sudo运行此脚本"
fi

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    log_error "Docker未安装，请先安装Docker"
    exit 1
fi

log_success "Docker已安装: $(docker --version)"

# 创建目录
log_info "创建配置目录..."
mkdir -p /opt/monitoring/prometheus
mkdir -p /opt/monitoring/prometheus/data
mkdir -p /opt/monitoring/grafana
chmod -R 777 /opt/monitoring

# ==================== 1. 部署Prometheus ====================
log_info "=== 1/6 部署Prometheus ==="

# 创建Prometheus配置
cat > /opt/monitoring/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'tencent-cloud'
    server: '101.33.251.158'

rule_files:
  - 'alert_rules.yml'

scrape_configs:
  - job_name: 'node'
    static_configs:
      - targets: ['172.17.0.1:9100']
  
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['172.17.0.1:8090']
  
  - job_name: 'mysql'
    static_configs:
      - targets: ['172.17.0.1:9104']
  
  - job_name: 'redis'
    static_configs:
      - targets: ['172.17.0.1:9121']
EOF

# 创建告警规则
cat > /opt/monitoring/prometheus/alert_rules.yml << 'EOF'
groups:
  - name: system_alerts
    rules:
      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "内存使用率超过85%"
      
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "CPU使用率超过80%"
      
      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) < 0.15
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "磁盘空间不足15%"
EOF

# 停止并删除旧容器（如果存在）
docker rm -f prometheus 2>/dev/null || true

# 部署Prometheus
docker run -d \
  --name prometheus \
  --restart=always \
  -p 9090:9090 \
  -v /opt/monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
  -v /opt/monitoring/prometheus/alert_rules.yml:/etc/prometheus/alert_rules.yml \
  -v /opt/monitoring/prometheus/data:/prometheus \
  prom/prometheus:latest \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --storage.tsdb.retention.time=30d

log_success "Prometheus已部署: http://101.33.251.158:9090"
sleep 2

# ==================== 2. 部署Node Exporter ====================
log_info "=== 2/6 部署Node Exporter ==="

docker rm -f node-exporter 2>/dev/null || true

docker run -d \
  --name node-exporter \
  --restart=always \
  --network=host \
  prom/node-exporter:latest

log_success "Node Exporter已部署: port 9100"
sleep 2

# ==================== 3. 部署cAdvisor ====================
log_info "=== 3/6 部署cAdvisor ==="

docker rm -f cadvisor 2>/dev/null || true

docker run -d \
  --name=cadvisor \
  --restart=always \
  --volume=/:/rootfs:ro \
  --volume=/var/run:/var/run:ro \
  --volume=/sys:/sys:ro \
  --volume=/var/lib/docker/:/var/lib/docker:ro \
  --publish=8090:8080 \
  --detach=true \
  gcr.io/cadvisor/cadvisor:latest 2>/dev/null || \
  docker run -d \
  --name=cadvisor \
  --restart=always \
  --volume=/:/rootfs:ro \
  --volume=/var/run:/var/run:ro \
  --volume=/sys:/sys:ro \
  --volume=/var/lib/docker/:/var/lib/docker:ro \
  --publish=8090:8080 \
  google/cadvisor:latest

log_success "cAdvisor已部署: http://101.33.251.158:8090"
sleep 2

# ==================== 4. 部署MySQL Exporter ====================
log_info "=== 4/6 部署MySQL Exporter ==="

docker rm -f mysql-exporter 2>/dev/null || true

docker run -d \
  --name mysql-exporter \
  --restart=always \
  --network=host \
  -e DATA_SOURCE_NAME="root:test_mysql_password@(localhost:3306)/" \
  prom/mysqld-exporter:latest

log_success "MySQL Exporter已部署: port 9104"
sleep 2

# ==================== 5. 部署Grafana ====================
log_info "=== 5/6 部署Grafana ==="

docker rm -f grafana 2>/dev/null || true
docker volume create grafana_data 2>/dev/null || true

docker run -d \
  --name=grafana \
  --restart=always \
  -p 3000:3000 \
  -e "GF_SECURITY_ADMIN_USER=admin" \
  -e "GF_SECURITY_ADMIN_PASSWORD=Admin@2025" \
  -e "GF_SERVER_ROOT_URL=http://101.33.251.158:3000" \
  -e "GF_INSTALL_PLUGINS=redis-datasource" \
  -v grafana_data:/var/lib/grafana \
  grafana/grafana:latest

log_success "Grafana已部署: http://101.33.251.158:3000"
log_info "默认账号: admin / Admin@2025"
sleep 3

# ==================== 6. 部署Portainer ====================
log_info "=== 6/6 部署Portainer ==="

docker rm -f portainer 2>/dev/null || true
docker volume create portainer_data 2>/dev/null || true

docker run -d \
  --name portainer \
  --restart=always \
  -p 8000:8000 \
  -p 9443:9443 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v portainer_data:/data \
  portainer/portainer-ce:latest

log_success "Portainer已部署: https://101.33.251.158:9443"
sleep 2

# ==================== 验证部署 ====================
echo ""
log_info "验证部署状态..."
sleep 5

echo ""
log_info "运行中的管理容器:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "NAMES|prometheus|grafana|portainer|exporter|cadvisor"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "访问地址:"
echo -e "  ${CYAN}Grafana:${NC}    http://101.33.251.158:3000  (admin / Admin@2025)"
echo -e "  ${CYAN}Portainer:${NC}  https://101.33.251.158:9443 (首次访问创建账号)"
echo -e "  ${CYAN}Prometheus:${NC} http://101.33.251.158:9090"
echo -e "  ${CYAN}cAdvisor:${NC}   http://101.33.251.158:8090"
echo ""
echo "下一步操作:"
echo "  1. 访问Grafana (http://101.33.251.158:3000)"
echo "  2. 添加Prometheus数据源 (URL: http://localhost:9090)"
echo "  3. 导入仪表盘模板:"
echo "     - Node Exporter: 1860"
echo "     - Docker: 193"
echo "     - MySQL: 7362"
echo "  4. 访问Portainer管理容器"
echo ""
echo -e "${YELLOW}重要提示:${NC}"
echo "  • 请修改Grafana默认密码"
echo "  • 建议配置防火墙限制访问IP"
echo "  • 定期检查告警规则"
echo ""

