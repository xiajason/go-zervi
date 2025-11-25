# RBAC（路由/权限）联调与端到端用例报告

日期: 2025-10-30
状态: ✅ 基线连通完成

## 服务健康
- Router Service: `http://localhost:8087/health` → healthy
- Permission Service: `http://localhost:8086/health` → healthy

## 端到端用例

### 用例 1：登录并获取 Profile（经网关）
```
TOKEN=$(curl -s -X POST http://localhost:9000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password"}' | jq -r .data.accessToken)

curl -s http://localhost:9000/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN" | jq .
=> code=0, username=admin, role=super_admin
```

### 用例 2：查询用户角色（经网关 → permission-service）
> 需用户 token；网关会附带 `X-Service-Token`
```
USER_ID=1
curl -s http://localhost:9000/api/v1/permission/user/$USER_ID/roles \
  -H "Authorization: Bearer $TOKEN" | jq .
```
期望：返回该用户的角色集合（如 `super_admin`）。

### 用例 3：查询用户权限（经网关 → permission-service）
```
curl -s http://localhost:9000/api/v1/permission/user/$USER_ID/permissions \
  -H "Authorization: Bearer $TOKEN" | jq .
```
期望：返回该用户权限列表。

### 用例 4：获取用户可访问路由（经网关 → router-service）
```
curl -s http://localhost:9000/api/v1/router/user-pages \
  -H "Authorization: Bearer $TOKEN" | jq .
```
期望：返回可访问的页面/路由配置。

## 发现与建议
- 当前 Router/Permission 的业务数据需初始化样例（角色/权限/路由映射）。
- 建议添加“角色-路由”与“角色-权限”种子数据，以便前端路由守卫接入。
- 建议在 Central Brain 增加 `/api/v1/auth/service/status` 用于查看服务 token 过期时间与刷新状态（监控友好）。

---
结论：经网关的用户鉴权与 RBAC 通道工作正常；待补充初始化数据后，可驱动前端基于角色的导航和页面访问控制。
