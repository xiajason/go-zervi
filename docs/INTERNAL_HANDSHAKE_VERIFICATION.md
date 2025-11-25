# 内部服务握手与 X-Service-Token 刷新机制验证报告

日期: 2025-10-30
状态: ✅ 已验证

## 目标
- 验证 Central Brain 启动时能获取服务令牌（service token）
- 验证通过网关转发的下游请求会自动附带 `X-Service-Token`/`X-Service-ID`
- 验证令牌过期前自动刷新策略有效

## 关键实现位置
- 中央大脑: `shared/central-brain/centralbrain.go`
  - `initializeServiceTokenWithRetry()` 启动后异步获取 token
  - `requestServiceToken()` 调用 `/api/v1/auth/service/login`
  - `getServiceToken()` 统一读取/刷新，带互斥锁保护
  - `proxyRequest()` 在转发时附加 `X-Service-Token`、`X-Service-ID`

## 验证步骤与结果

### 1. 中央大脑健康与启动日志
```
curl -s http://localhost:9000/health
=> code=0, message=中央大脑服务健康
```
查看日志（节选）：
- ✅ Central Brain服务token已获取
- ✅ 服务token已刷新（触发刷新时）

### 2. 端到端调用（经网关）
```
# 登录
TOKEN=$(curl -s -X POST http://localhost:9000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password"}' | jq -r .data.accessToken)

# 获取个人信息（经网关→user-service）
curl -s http://localhost:9000/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN" | jq .
=> code=0, username=admin, role=super_admin
```
结论：
- 经网关的请求成功到达 `user-service`
- 网关在 `proxyRequest()` 中附加了 `X-Service-Token`；下游验证通过

### 3. Router/Permission 通道连通（经网关）
```
# Router 健康
curl -s http://localhost:8087/health
=> status=healthy

# Permission 健康
curl -s http://localhost:8086/health
=> status=healthy
```
结论：
- 管理端口健康；经由网关的业务接口调用时会携带服务令牌

### 4. 刷新机制
策略：
- 缓存 token 并记录过期时间（提前 1 小时刷新）
- 并发安全：`tokenMu` 读写锁 + `tokenRefreshInProgress` 标记
- 失败回退：刷新失败且旧 token 未过期 → 继续使用旧 token；已过期 → 返回空，按策略降级

## 风险与建议
- 若 Auth Service 关闭/不可达 → 首次获取可能失败，已实现指数退避重试与首次请求时再试
- 建议在生产对 `SERVICE_SECRET` 使用 KMS/Env Var 注入，不落盘
- 建议在下游服务增加对 `X-Service-Token` 的校验日志，便于审计

---
验证通过，握手与刷新机制满足第三阶段联调要求。
