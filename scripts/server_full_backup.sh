#!/bin/bash
# 腾讯云服务器完整备份脚本
# 用途: 备份所有数据库、代码和配置
# 日期: 2025-10-10
# 使用: 在腾讯云服务器上执行

BACKUP_DIR="/opt/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_ROOT="$BACKUP_DIR/backup_$TIMESTAMP"

echo "=========================================="
echo "开始完整备份: $TIMESTAMP"
echo "服务器: $(hostname)"
echo "=========================================="

# 创建备份目录结构
mkdir -p "$BACKUP_ROOT"/{databases,configs,services,scripts}

# 1. 备份MySQL数据库
echo ""
echo "=== 1. 备份MySQL数据库 ==="
mkdir -p "$BACKUP_ROOT/databases/mysql"

for db in jobfirst_basic jobfirst_professional jobfirst_future; do
    echo "备份数据库: $db"
    docker exec test-mysql mysqldump \
        -uroot -ptest_mysql_password \
        --single-transaction \
        --routines --triggers --events \
        --set-gtid-purged=OFF \
        "$db" 2>/dev/null | gzip > "$BACKUP_ROOT/databases/mysql/${db}.sql.gz"
    
    if [ $? -eq 0 ]; then
        size=$(du -sh "$BACKUP_ROOT/databases/mysql/${db}.sql.gz" | cut -f1)
        echo "  ✅ $db 备份完成 ($size)"
    else
        echo "  ❌ $db 备份失败"
    fi
done

# 备份MySQL配置
docker exec test-mysql cat /etc/mysql/my.cnf > "$BACKUP_ROOT/databases/mysql/my.cnf" 2>/dev/null || echo "  ⚠️  MySQL配置文件不存在"

# 2. 备份PostgreSQL
echo ""
echo "=== 2. 备份PostgreSQL ==="
mkdir -p "$BACKUP_ROOT/databases/postgres"

docker exec test-postgres pg_dumpall -U postgres 2>/dev/null | gzip > "$BACKUP_ROOT/databases/postgres/all_databases.sql.gz"
if [ $? -eq 0 ]; then
    size=$(du -sh "$BACKUP_ROOT/databases/postgres/all_databases.sql.gz" | cut -f1)
    echo "  ✅ PostgreSQL备份完成 ($size)"
else
    echo "  ❌ PostgreSQL备份失败"
fi

# 3. 备份MongoDB
echo ""
echo "=== 3. 备份MongoDB ==="
mkdir -p "$BACKUP_ROOT/databases/mongodb"

docker exec test-mongodb mongodump --archive 2>/dev/null | gzip > "$BACKUP_ROOT/databases/mongodb/dump.archive.gz"
if [ $? -eq 0 ]; then
    size=$(du -sh "$BACKUP_ROOT/databases/mongodb/dump.archive.gz" | cut -f1)
    echo "  ✅ MongoDB备份完成 ($size)"
else
    echo "  ❌ MongoDB备份失败"
fi

# 4. 备份Redis
echo ""
echo "=== 4. 备份Redis ==="
mkdir -p "$BACKUP_ROOT/databases/redis"

docker exec test-redis redis-cli -a test_redis_password --rdb /tmp/dump.rdb SAVE 2>/dev/null
sleep 2
docker cp test-redis:/data/dump.rdb "$BACKUP_ROOT/databases/redis/dump.rdb" 2>/dev/null
if [ $? -eq 0 ]; then
    gzip "$BACKUP_ROOT/databases/redis/dump.rdb"
    size=$(du -sh "$BACKUP_ROOT/databases/redis/dump.rdb.gz" | cut -f1)
    echo "  ✅ Redis备份完成 ($size)"
else
    echo "  ❌ Redis备份失败"
fi

# 5. 备份Neo4j (DAO/区块链数据)
echo ""
echo "=== 5. 备份Neo4j (重要!) ==="
mkdir -p "$BACKUP_ROOT/databases/neo4j"

# 在线备份（推荐）
docker exec test-neo4j neo4j-admin database dump neo4j --to-path=/tmp 2>/dev/null
docker cp test-neo4j:/tmp/neo4j.dump "$BACKUP_ROOT/databases/neo4j/" 2>/dev/null

if [ $? -eq 0 ]; then
    gzip "$BACKUP_ROOT/databases/neo4j/neo4j.dump"
    size=$(du -sh "$BACKUP_ROOT/databases/neo4j/neo4j.dump.gz" | cut -f1)
    echo "  ✅ Neo4j备份完成 ($size)"
else
    echo "  ⚠️  Neo4j在线备份失败，尝试离线备份..."
    
    # 停止Neo4j容器进行离线备份
    docker stop test-neo4j
    docker cp test-neo4j:/data "$BACKUP_ROOT/databases/neo4j/" 2>/dev/null
    if [ $? -eq 0 ]; then
        tar -czf "$BACKUP_ROOT/databases/neo4j/data.tar.gz" -C "$BACKUP_ROOT/databases/neo4j/" data/
        rm -rf "$BACKUP_ROOT/databases/neo4j/data"
        docker start test-neo4j
        size=$(du -sh "$BACKUP_ROOT/databases/neo4j/data.tar.gz" | cut -f1)
        echo "  ✅ Neo4j离线备份完成 ($size)"
    else
        docker start test-neo4j
        echo "  ❌ Neo4j备份完全失败"
    fi
fi

# 6. 备份Elasticsearch索引（可选）
echo ""
echo "=== 6. 备份Elasticsearch (可选) ==="
mkdir -p "$BACKUP_ROOT/databases/elasticsearch"
echo "  ⚠️  Elasticsearch数据量大，本脚本跳过"
echo "  提示: 可使用Elasticsearch快照API进行备份"

# 7. 备份Weaviate（可选）
echo ""
echo "=== 7. 备份Weaviate (可选) ==="
mkdir -p "$BACKUP_ROOT/databases/weaviate"
echo "  ⚠️  Weaviate数据，本脚本跳过"

# 8. 备份服务配置和代码
echo ""
echo "=== 8. 备份服务配置和代码 ==="

# 配置文件
if [ -d "/opt/services/configs" ]; then
    cp -r /opt/services/configs/* "$BACKUP_ROOT/configs/" 2>/dev/null
    echo "  ✅ configs目录备份完成"
else
    echo "  ⚠️  configs目录不存在"
fi

cp /opt/services/docker-compose.yml "$BACKUP_ROOT/configs/" 2>/dev/null && echo "  ✅ docker-compose.yml备份完成"

# 脚本
if [ -d "/opt/services/scripts" ]; then
    cp -r /opt/services/scripts/* "$BACKUP_ROOT/scripts/" 2>/dev/null
    echo "  ✅ scripts目录备份完成"
else
    echo "  ⚠️  scripts目录不存在"
fi

# 管理脚本
if [ -f "$HOME/manage_tencent_services.sh" ]; then
    cp "$HOME/manage_tencent_services.sh" "$BACKUP_ROOT/scripts/" 2>/dev/null
    echo "  ✅ 管理脚本备份完成"
fi

# 服务代码（仅关键文件，排除大文件）
echo ""
echo "=== 9. 备份服务代码 ==="

# Zervigo
if [ -d "/opt/services/zervigo" ]; then
    mkdir -p "$BACKUP_ROOT/services/zervigo"
    tar --exclude='logs/*' --exclude='*.log' \
        -czf "$BACKUP_ROOT/services/zervigo.tar.gz" \
        -C /opt/services zervigo/
    size=$(du -sh "$BACKUP_ROOT/services/zervigo.tar.gz" | cut -f1)
    echo "  ✅ Zervigo备份完成 ($size)"
fi

# AI Service (排除venv和临时文件)
if [ -d "/opt/services/ai-service-1/current" ]; then
    mkdir -p "$BACKUP_ROOT/services/ai-service-1"
    tar --exclude='venv' --exclude='__pycache__' --exclude='*.pyc' \
        --exclude='temp/*' --exclude='uploads/*' --exclude='*.log' \
        -czf "$BACKUP_ROOT/services/ai-service-1.tar.gz" \
        -C /opt/services/ai-service-1 current/
    size=$(du -sh "$BACKUP_ROOT/services/ai-service-1.tar.gz" | cut -f1)
    echo "  ✅ AI Service备份完成 ($size)"
fi

# 10. 备份Docker配置
echo ""
echo "=== 10. 备份Docker配置 ==="
mkdir -p "$BACKUP_ROOT/docker"

for container in test-mysql test-postgres test-redis test-mongodb test-neo4j test-elasticsearch test-weaviate; do
    docker inspect $container > "$BACKUP_ROOT/docker/${container}_config.json" 2>/dev/null
done

echo "  ✅ Docker配置备份完成"

# 11. 生成备份清单
echo ""
echo "=== 11. 生成备份清单 ==="

cat > "$BACKUP_ROOT/BACKUP_MANIFEST.txt" << MANIFEST
腾讯云服务器完整备份清单
========================================
备份时间: $TIMESTAMP
服务器: $(hostname)
服务器IP: 101.33.251.158
备份路径: $BACKUP_ROOT

系统信息:
  - 内核版本: $(uname -r)
  - Ubuntu版本: $(lsb_release -d | cut -f2)
  - 磁盘使用: $(df -h / | awk 'NR==2{print $3 "/" $2 " (" $5 ")"}')
  - 内存使用: $(free -h | awk 'NR==2{print $3 "/" $2}')

数据库备份:
MANIFEST

find "$BACKUP_ROOT/databases" -type f -exec ls -lh {} \; | awk '{printf "  %-60s %10s\n", $9, $5}' >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"

echo "" >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"
echo "服务代码备份:" >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"
find "$BACKUP_ROOT/services" -type f -exec ls -lh {} \; | awk '{printf "  %-60s %10s\n", $9, $5}' >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"

echo "" >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"
echo "总大小:" >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"
du -sh "$BACKUP_ROOT" | awk '{printf "  %s\n", $1}' >> "$BACKUP_ROOT/BACKUP_MANIFEST.txt"

# 显示清单
cat "$BACKUP_ROOT/BACKUP_MANIFEST.txt"

# 12. 创建压缩包
echo ""
echo "=== 12. 创建压缩包 ==="

cd "$BACKUP_DIR"
tar -czf "backup_${TIMESTAMP}.tar.gz" "backup_${TIMESTAMP}/" 2>/dev/null

if [ $? -eq 0 ]; then
    size=$(du -sh "backup_${TIMESTAMP}.tar.gz" | cut -f1)
    echo "  ✅ 压缩完成: backup_${TIMESTAMP}.tar.gz ($size)"
    
    # 删除临时目录
    rm -rf "$BACKUP_ROOT"
    
    echo ""
    echo "=========================================="
    echo "✅ 备份完成！"
    echo "=========================================="
    echo ""
    echo "备份文件: $BACKUP_DIR/backup_${TIMESTAMP}.tar.gz"
    echo "文件大小: $size"
    echo ""
    echo "下载到本地的命令:"
    echo "  scp -i ~/Downloads/basic.pem ubuntu@101.33.251.158:$BACKUP_DIR/backup_${TIMESTAMP}.tar.gz ./"
    echo ""
    echo "或使用rsync (支持断点续传):"
    echo "  rsync -avz --progress -e 'ssh -i ~/Downloads/basic.pem' ubuntu@101.33.251.158:$BACKUP_DIR/backup_${TIMESTAMP}.tar.gz ./"
    echo ""
    
    # 列出所有备份
    echo "当前所有备份:"
    ls -lh "$BACKUP_DIR"/backup_*.tar.gz 2>/dev/null || echo "  仅有此备份"
    
else
    echo "  ❌ 压缩失败"
    exit 1
fi

echo ""
echo "=========================================="
echo "备份脚本执行完毕"
echo "=========================================="

