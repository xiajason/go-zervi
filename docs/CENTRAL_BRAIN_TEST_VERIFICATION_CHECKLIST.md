# Central Brain基础设施层测试验证准备检查清单

## 📋 准备状态检查

### ✅ 已完成的项目

1. **代码编译**
   - ✅ 所有中间件文件已创建
   - ✅ 编译通过，无错误
   - ✅ 依赖已安装（uuid, rate limit）

2. **功能实现**
   - ✅ 请求日志中间件（logging.go）
   - ✅ 性能指标中间件（metrics.go）
   - ✅ 限流中间件（ratelimit.go）
   - ✅ 熔断器中间件（circuitbreaker.go）
   - ✅ HTTP客户端连接池（pool.go）
   - ✅ 错误处理工具（error_handler.go）

3. **集成状态**
   - ✅ 中间件已集成到Central Brain
   - ✅ 管理API已注册
   - ✅ 服务代理已添加熔断器保护

---

## 🔍 启动前检查清单

### 必需条件

- [x] **Go环境**: Go 1.25+已安装
- [x] **配置文件**: configs/local.env存在
- [x] **端口可用**: 9000端口（或已停止旧进程）
- [x] **编译成功**: 无编译错误

### 可选条件

- [ ] **Auth Service**: 运行在8207端口（用于服务token）
- [ ] **数据库**: PostgreSQL运行（用于数据库检查）

---

## 🚀 启动方式

### 方式1: 使用测试脚本（推荐）

```bash
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo
bash scripts/test-central-brain-infrastructure.sh
```

**功能**:
- ✅ 自动检查环境
- ✅ 自动编译
- ✅ 后台启动服务
- ✅ 自动运行测试
- ✅ 显示日志和状态

---

### 方式2: 手动启动

```bash
# 1. 加载环境变量
cd /Users/szjason72/szbolent/Zervigo/zervigo.demo
set -a
source <(cat configs/local.env | grep "^[^#]" | grep -v "^$")
set +a

# 2. 编译
cd shared/central-brain
go build -o ../../bin/central-brain *.go

# 3. 启动
cd ../..
./bin/central-brain
```

---

## 🧪 测试验证项

### 基础功能测试

1. **健康检查**
   ```bash
   curl http://localhost:9000/health
   ```
   **预期**: 返回200状态码，包含service、status、version信息

2. **性能指标**
   ```bash
   curl http://localhost:9000/api/v1/metrics
   ```
   **预期**: 返回性能统计数据（total_requests、success_rate等）

3. **熔断器状态**
   ```bash
   curl http://localhost:9000/api/v1/circuit-breakers
   ```
   **预期**: 返回各服务的熔断器状态

---

### 功能验证测试

4. **TraceID生成**
   ```bash
   curl -H "X-Trace-ID: test-123" http://localhost:9000/health
   ```
   **预期**: 响应头包含X-Trace-ID

5. **结构化日志**
   ```bash
   tail -f /tmp/central-brain.log
   ```
   **预期**: 看到JSON格式的结构化日志

6. **限流测试**
   ```bash
   # 快速发送100个请求
   for i in {1..100}; do curl -s http://localhost:9000/health > /dev/null; done
   ```
   **预期**: 某些请求可能返回429（如果超过限制）

---

### 服务代理测试

7. **服务代理（需要Auth Service运行）**
   ```bash
   # 如果Auth Service运行在8207端口
   curl http://localhost:9000/api/v1/auth/health
   ```
   **预期**: 代理到Auth Service并返回响应

---

## ⚠️ 注意事项

### 1. Auth Service依赖

**情况**: Central Brain需要Auth Service来获取服务token

**处理**:
- ✅ **无需Auth Service**: Central Brain可以启动，但服务token获取会失败
- ✅ **日志会显示**: "⚠️ 获取服务token失败"，但不影响启动
- ✅ **功能影响**: 服务间认证可能失败，但不影响基础功能测试

---

### 2. 端口占用

**检查**:
```bash
lsof -ti:9000
```

**处理**:
```bash
# 停止现有进程
kill $(lsof -ti:9000)
```

---

### 3. 日志输出

**位置**: 
- 标准输出/标准错误
- 如果使用测试脚本：`/tmp/central-brain.log`

**查看**:
```bash
tail -f /tmp/central-brain.log
```

---

## ✅ 测试验证条件总结

### 当前状态

| 条件 | 状态 | 说明 |
|------|------|------|
| **代码编译** | ✅ 完成 | 无错误 |
| **依赖安装** | ✅ 完成 | uuid, rate limit已安装 |
| **配置文件** | ✅ 存在 | configs/local.env |
| **端口可用** | ⚠️ 检查 | 可能被占用 |
| **Auth Service** | ⚠️ 可选 | 用于服务token |
| **数据库** | ⚠️ 可选 | 用于数据库检查 |

---

### 结论

**✅ 具备启动和测试验证的条件**

**可以立即执行的测试**:
1. ✅ 启动服务
2. ✅ 健康检查
3. ✅ 性能指标查询
4. ✅ 熔断器状态查询
5. ✅ TraceID生成验证
6. ✅ 结构化日志验证
7. ⚠️ 限流测试（需要高频请求）
8. ⚠️ 服务代理测试（需要下游服务运行）

---

**建议**: 使用测试脚本 `scripts/test-central-brain-infrastructure.sh` 进行自动化测试验证

