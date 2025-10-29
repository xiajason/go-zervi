#!/bin/bash
# 从腾讯云服务器下载所有数据到本地
# 用途: 在本地Mac上执行，自动化下载流程
# 日期: 2025-10-10

SERVER_IP="101.33.251.158"
SSH_KEY="$HOME/Downloads/basic.pem"
SSH_USER="ubuntu"

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# Zervigo项目根目录（scripts的上级目录）
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
# 下载到项目根目录下的downloaded文件夹
LOCAL_DIR="$PROJECT_ROOT/downloaded"
BACKUP_SCRIPT="$SCRIPT_DIR/server_full_backup.sh"

echo "=========================================="
echo "腾讯云服务器数据下载工具"
echo "=========================================="
echo ""
echo "服务器: $SERVER_IP"
echo "本地目录: $LOCAL_DIR"
echo "项目根目录: $PROJECT_ROOT"
echo ""

# 检查SSH密钥
if [ ! -f "$SSH_KEY" ]; then
    echo "❌ SSH密钥不存在: $SSH_KEY"
    echo ""
    echo "提示: 请确认SSH密钥位置，或修改脚本中的SSH_KEY变量"
    echo "当前查找路径: $SSH_KEY"
    exit 1
fi

echo "✅ SSH密钥已找到: $SSH_KEY"

# 创建本地目录结构
echo "创建本地目录结构..."
mkdir -p "$LOCAL_DIR"/{backups,services,configs,scripts,docs}
cd "$LOCAL_DIR"

# 步骤1: 上传备份脚本到服务器
echo ""
echo "=== 步骤1: 上传备份脚本到服务器 ==="

backup_script="$BACKUP_SCRIPT"

if [ ! -f "$backup_script" ]; then
    echo "❌ 备份脚本不存在: $backup_script"
    exit 1
fi

echo "上传备份脚本..."
scp -i "$SSH_KEY" "$backup_script" "$SSH_USER@$SERVER_IP:/tmp/full_backup.sh"

if [ $? -eq 0 ]; then
    echo "  ✅ 脚本上传成功"
else
    echo "  ❌ 脚本上传失败"
    exit 1
fi

# 步骤2: 在服务器上执行备份
echo ""
echo "=== 步骤2: 在服务器上执行完整备份 ==="
echo "这可能需要几分钟，请耐心等待..."
echo ""

ssh -i "$SSH_KEY" "$SSH_USER@$SERVER_IP" << 'EOSSH'
chmod +x /tmp/full_backup.sh

# 创建备份目录
sudo mkdir -p /opt/backups
sudo chown ubuntu:ubuntu /opt/backups

# 执行备份
/tmp/full_backup.sh

# 获取最新备份文件名
LATEST_BACKUP=$(ls -t /opt/backups/backup_*.tar.gz 2>/dev/null | head -1)

if [ -n "$LATEST_BACKUP" ]; then
    echo ""
    echo "最新备份: $LATEST_BACKUP"
    echo "$LATEST_BACKUP" > /tmp/latest_backup.txt
else
    echo "❌ 没有找到备份文件"
    exit 1
fi
EOSSH

if [ $? -ne 0 ]; then
    echo "❌ 备份执行失败"
    exit 1
fi

# 获取备份文件名
BACKUP_FILE=$(ssh -i "$SSH_KEY" "$SSH_USER@$SERVER_IP" 'cat /tmp/latest_backup.txt')

if [ -z "$BACKUP_FILE" ]; then
    echo "❌ 无法获取备份文件名"
    exit 1
fi

echo "备份文件: $BACKUP_FILE"

# 步骤3: 下载备份到本地
echo ""
echo "=== 步骤3: 下载备份到本地 ==="
echo "使用rsync支持断点续传..."

rsync -avz --progress -e "ssh -i $SSH_KEY" \
    "$SSH_USER@$SERVER_IP:$BACKUP_FILE" \
    "$LOCAL_DIR/backups/"

if [ $? -eq 0 ]; then
    echo "  ✅ 备份下载完成"
    
    # 获取本地备份文件名
    LOCAL_BACKUP=$(basename "$BACKUP_FILE")
    
    echo ""
    echo "=== 步骤4: 解压备份文件 ==="
    cd "$LOCAL_DIR/backups"
    tar -xzf "$LOCAL_BACKUP"
    
    if [ $? -eq 0 ]; then
        echo "  ✅ 备份解压完成"
        
        # 获取解压目录名
        EXTRACT_DIR=$(tar -tzf "$LOCAL_BACKUP" | head -1 | cut -f1 -d"/")
        
        # 移动databases目录内容到正确位置
        if [ -d "$EXTRACT_DIR/databases" ]; then
            # 如果databases目录已存在，先删除
            rm -rf databases
            # 移动解压的databases目录
            mv "$EXTRACT_DIR/databases" .
            # 删除临时目录
            rm -rf "$EXTRACT_DIR"
            echo "  ✅ 数据库备份准备就绪: $LOCAL_DIR/backups/databases"
        elif [ -d "$EXTRACT_DIR" ]; then
            # 如果没有databases子目录，整个目录就是数据
            mv "$EXTRACT_DIR" "databases"
            echo "  ✅ 数据库备份准备就绪: $LOCAL_DIR/backups/databases"
        fi
    else
        echo "  ❌ 备份解压失败"
        exit 1
    fi
else
    echo "  ❌ 备份下载失败"
    exit 1
fi

# 步骤5: 下载服务代码
echo ""
echo "=== 步骤5: 下载服务代码 ==="

echo "下载Zervigo服务..."
rsync -avz --progress -e "ssh -i $SSH_KEY" \
    --exclude='*.log' --exclude='logs/*' \
    "$SSH_USER@$SERVER_IP:/opt/services/zervigo/" \
    "$LOCAL_DIR/services/zervigo/"

echo ""
echo "下载AI Service..."
ssh -i "$SSH_KEY" "$SSH_USER@$SERVER_IP" << 'EOSSH'
cd /opt/services/ai-service-1/current
tar --exclude="venv" --exclude="__pycache__" --exclude="*.pyc" \
    --exclude="temp/*" --exclude="uploads/*" --exclude="*.log" \
    -czf /tmp/ai-service-1.tar.gz .
EOSSH

scp -i "$SSH_KEY" "$SSH_USER@$SERVER_IP:/tmp/ai-service-1.tar.gz" "$LOCAL_DIR/services/"

mkdir -p "$LOCAL_DIR/services/ai-service-1"
tar -xzf "$LOCAL_DIR/services/ai-service-1.tar.gz" -C "$LOCAL_DIR/services/ai-service-1/"
rm "$LOCAL_DIR/services/ai-service-1.tar.gz"

echo "  ✅ 服务代码下载完成"

# 步骤6: 下载配置文件和脚本
echo ""
echo "=== 步骤6: 下载配置文件和脚本 ==="

rsync -avz --progress -e "ssh -i $SSH_KEY" \
    "$SSH_USER@$SERVER_IP:/opt/services/configs/" \
    "$LOCAL_DIR/configs/"

rsync -avz --progress -e "ssh -i $SSH_KEY" \
    "$SSH_USER@$SERVER_IP:/opt/services/scripts/" \
    "$LOCAL_DIR/scripts/server/"

scp -i "$SSH_KEY" "$SSH_USER@$SERVER_IP:/opt/services/docker-compose.yml" \
    "$LOCAL_DIR/configs/"

scp -i "$SSH_KEY" "$SSH_USER@$SERVER_IP:~/manage_tencent_services.sh" \
    "$LOCAL_DIR/scripts/server/"

echo "  ✅ 配置和脚本下载完成"

# 步骤7: 导出Docker配置
echo ""
echo "=== 步骤7: 导出Docker配置 ==="

ssh -i "$SSH_KEY" "$SSH_USER@$SERVER_IP" << 'EOSSH'
mkdir -p /tmp/docker_configs

for container in test-mysql test-postgres test-redis test-mongodb test-neo4j test-elasticsearch test-weaviate; do
    docker inspect $container > /tmp/docker_configs/${container}_config.json 2>/dev/null
done

cp /opt/services/docker-compose.yml /tmp/docker_configs/ 2>/dev/null
cd /tmp
tar -czf docker_configs.tar.gz docker_configs/
EOSSH

scp -i "$SSH_KEY" "$SSH_USER@$SERVER_IP:/tmp/docker_configs.tar.gz" \
    "$LOCAL_DIR/configs/"

cd "$LOCAL_DIR/configs"
tar -xzf docker_configs.tar.gz
rm docker_configs.tar.gz

echo "  ✅ Docker配置导出完成"

# 步骤8: 复制本地环境配置文件
echo ""
echo "=== 步骤8: 复制本地环境配置 ==="

# 复制docker-compose.local.yml
if [ -f "$PROJECT_ROOT/docker-compose.local.yml" ]; then
    cp "$PROJECT_ROOT/docker-compose.local.yml" "$LOCAL_DIR/"
    echo "  ✅ docker-compose.local.yml已复制"
fi

# 复制恢复脚本
if [ -f "$SCRIPT_DIR/local_restore.sh" ]; then
    cp "$SCRIPT_DIR/local_restore.sh" "$LOCAL_DIR/"
    chmod +x "$LOCAL_DIR/local_restore.sh"
    echo "  ✅ local_restore.sh已复制"
fi

# 步骤9: 生成下载清单
echo ""
echo "=== 步骤9: 生成下载清单 ==="

cat > "$LOCAL_DIR/DOWNLOAD_MANIFEST.txt" << MANIFEST
腾讯云服务器数据下载清单
========================================
下载时间: $(date '+%Y-%m-%d %H:%M:%S')
服务器: $SERVER_IP
本地目录: $LOCAL_DIR

目录结构:
MANIFEST

tree -L 2 "$LOCAL_DIR" >> "$LOCAL_DIR/DOWNLOAD_MANIFEST.txt" 2>/dev/null || \
    find "$LOCAL_DIR" -maxdepth 2 -type d >> "$LOCAL_DIR/DOWNLOAD_MANIFEST.txt"

echo "" >> "$LOCAL_DIR/DOWNLOAD_MANIFEST.txt"
echo "磁盘使用:" >> "$LOCAL_DIR/DOWNLOAD_MANIFEST.txt"
du -sh "$LOCAL_DIR"/* >> "$LOCAL_DIR/DOWNLOAD_MANIFEST.txt"

# 创建README
cat > "$LOCAL_DIR/README.md" << 'EOFREADME'
# JobFirst 本地开发环境

## 📦 已下载的内容

✅ 所有数据库备份（MySQL、PostgreSQL、MongoDB、Redis、Neo4j）
✅ 服务代码（Zervigo、AI Service）  
✅ 配置文件和脚本
✅ Docker配置

## 🚀 快速开始

### 步骤1: 启动Docker环境

```bash
docker-compose -f docker-compose.local.yml up -d
```

### 步骤2: 等待容器启动（约30秒）

```bash
sleep 30
```

### 步骤3: 恢复数据库

```bash
./local_restore.sh
```

## 📊 访问信息

### 数据库连接

- **MySQL**: localhost:3306 (root / local_dev_password)
- **PostgreSQL**: localhost:5432 (postgres / local_dev_password)
- **MongoDB**: localhost:27017 (admin / local_dev_password)
- **Redis**: localhost:6379 (密码: local_dev_password)
- **Neo4j**: localhost:7474, 7687 (neo4j / local_dev_password)

### 管理界面

- **Adminer** (数据库管理): http://localhost:8888
- **Neo4j Browser**: http://localhost:7474

## 📁 目录结构

```
downloaded/
├── backups/databases/     # 数据库备份
├── services/              # 服务代码
├── configs/               # 配置文件
├── docker-compose.local.yml  # Docker环境配置
└── local_restore.sh       # 数据恢复脚本
```

## 🔧 常用命令

```bash
# 查看容器状态
docker-compose -f docker-compose.local.yml ps

# 查看日志
docker-compose -f docker-compose.local.yml logs -f

# 停止所有服务
docker-compose -f docker-compose.local.yml down
```
EOFREADME

echo "  ✅ 下载清单生成完成"

# 步骤10: 清理服务器临时文件
echo ""
echo "=== 步骤10: 清理服务器临时文件 ==="

ssh -i "$SSH_KEY" "$SSH_USER@$SERVER_IP" << 'EOSSH'
rm -f /tmp/full_backup.sh
rm -f /tmp/latest_backup.txt
rm -f /tmp/ai-service-1.tar.gz
rm -rf /tmp/docker_configs
rm -f /tmp/docker_configs.tar.gz
EOSSH

echo "  ✅ 临时文件清理完成"

# 完成
echo ""
echo "=========================================="
echo "✅ 所有数据下载完成！"
echo "=========================================="
echo ""
echo "下载位置: $LOCAL_DIR"
echo ""
echo "目录结构:"
echo "  $LOCAL_DIR/"
echo "  ├── backups/databases/      # 数据库备份"
echo "  ├── services/               # 服务代码"
echo "  │   ├── zervigo/           # Zervigo服务"
echo "  │   └── ai-service-1/      # AI服务"
echo "  ├── configs/                # 配置文件"
echo "  ├── scripts/                # 管理脚本"
echo "  ├── docker-compose.local.yml  # Docker配置"
echo "  ├── local_restore.sh        # 恢复脚本"
echo "  └── README.md               # 使用指南"
echo ""
echo "下一步:"
echo "  1. 进入下载目录: cd $LOCAL_DIR"
echo "  2. 查看README: cat README.md"
echo "  3. 启动环境: docker-compose -f docker-compose.local.yml up -d"
echo "  4. 恢复数据: ./local_restore.sh"
echo ""
echo "详细指南: $PROJECT_ROOT/QUICK_MIGRATION_GUIDE.md"
echo ""
