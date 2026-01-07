# 🔧 重启 Auth Service 以应用修复

## ✅ 已完成的修复

修改了 `/shared/core/auth/unified_auth_api.go`，使登录接口同时支持：

### 1. 标准格式（curl 测试使用）
```json
{
  "username": "admin",
  "password": "admin123"
}
```

### 2. VueCMF 格式（前端使用）
```json
{
  "data": {
    "login_name": "admin",
    "password": "admin123"
  }
}
```

## 🚀 重启步骤

### 方法 1: 使用启动脚本重启

```bash
cd /Users/szjason72/gozervi/imartdevos/scripts

# 停止所有核心服务
./stop-core-services.sh 2>/dev/null || kill $(cat /tmp/auth-service.pid) 2>/dev/null

# 重新启动核心服务
./start-core-services.sh
```

### 方法 2: 手动重启 Auth Service

```bash
# 1. 停止旧进程
kill $(cat /tmp/auth-service.pid) 2>/dev/null
# 或者
pkill -f "auth.*8207"

# 2. 启动 Auth Service
cd /Users/szjason72/gozervi/zervigo.demo/shared/core
nohup go run ./cmd/unified-auth-service --port 8207 > /tmp/auth-service.log 2>&1 &
echo $! > /tmp/auth-service.pid

# 3. 验证启动
sleep 2
curl http://localhost:8207/health
```

### 方法 3: 如果 Central Brain 也在运行，一起重启

```bash
# 停止所有服务
pkill -f "central-brain"
pkill -f "auth.*8207"

# 重启 Auth Service
cd /Users/szjason72/gozervi/zervigo.demo/shared/core
nohup go run ./cmd/unified-auth-service --port 8207 > /tmp/auth-service.log 2>&1 &

# 重启 Central Brain
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
go run . &
```

## ✅ 验证修复

重启后，测试两种格式都能正常登录：

### 测试 1: 标准格式
```bash
curl -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq .
```

### 测试 2: VueCMF 格式
```bash
curl -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"data":{"login_name":"admin","password":"admin123"}}' \
  | jq .
```

两个都应该返回：
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "accessToken": "...",
    "userId": 1,
    "userName": "admin"
  }
}
```

## 🎯 前端测试

重启服务后：

1. **刷新浏览器页面**（Ctrl/Cmd + R）
2. **输入登录信息**：
   - 账号：`admin`
   - 密码：`admin123`
3. **点击登录按钮**

**应该成功登录！** ✅

