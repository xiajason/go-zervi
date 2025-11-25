# 📋 Zervigo API 响应格式说明

## 🎯 后端响应格式

### Zervigo 实际格式

根据 `shared/core/response/api_response.go` 分析，后端使用 `response.Success()` 返回：

```json
{
  "code": 0,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@zervigo.com",
      "role": "super_admin"
    }
  },
  "timestamp": 1730972400
}
```

**字段说明**:
- `code`: 0 表示成功，其他值表示错误
- `message`: 响应消息
- `data`: 业务数据
- `timestamp`: Unix 时间戳（毫秒）

### VueCMF 格式（兼容）

```json
{
  "code": 0,
  "data": { ... },
  "message": "success"
}
```

## 🔄 前端响应处理

### request.ts 响应拦截器

```typescript
request.interceptors.response.use(
  (response) => {
    const res = response.data
    
    // 兼容多种格式
    const isSuccess = 
      res.code === 0 ||               // VueCMF
      res.code === 200 ||             // HTTP Code
      res.status === 'success' ||     // Zervigo
      res.success === true            // 其他
    
    if (isSuccess) {
      return res.data || res   // 返回 data 部分
    } else {
      // 错误处理
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message))
    }
  }
)
```

### Login.vue 中的使用

```typescript
const res = await login(loginForm)

// res 已经是 response.data.data 了
// 所以直接访问 res.token，而不是 res.data.token
if (res && res.token) {
  localStorage.setItem('token', res.token)
  localStorage.setItem('userInfo', JSON.stringify(res.user))
  router.push('/home')
}
```

## 📊 数据流转过程

```
后端返回:
{
  status: "success",
  data: { token: "xxx", user: {...} },
  message: "Login successful"
}
           ↓
axios 接收: response.data
{
  status: "success",
  data: { token: "xxx", user: {...} },
  message: "Login successful"
}
           ↓
request.ts 拦截器处理:
const res = response.data  // 第一层解包
return res.data             // 第二层解包
           ↓
Login.vue 接收: res
{
  token: "xxx",
  user: {...}
}
           ↓
使用: res.token ✅ (正确)
     res.data.token ❌ (错误 - 会是 undefined)
```

## 🔍 调试技巧

### 1. 查看原始响应

在浏览器控制台的 Network 标签中：
1. 找到 `login` 请求
2. 查看 Response 标签
3. 查看实际返回的 JSON 格式

### 2. 使用控制台日志

Login.vue 中已添加调试日志：
```typescript
console.log('登录响应:', res)  // 查看拦截器处理后的数据
```

如果响应格式错误，会输出：
```javascript
console.error('登录响应格式错误:', res)
```

### 3. 使用 curl 测试后端

```bash
curl -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq .
```

查看实际的响应格式。

## ✅ 已修复的问题

1. **响应拦截器** - 兼容多种后端格式
2. **Login.vue** - 正确访问 `res.token`（不是 `res.data.token`）
3. **错误处理** - 添加详细的调试日志

## 🎯 预期响应格式

### 登录成功

```json
{
  "code": 0,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@zervigo.com",
      "role": "super_admin"
    }
  },
  "timestamp": 1730972400
}
```

### 登录失败

```json
{
  "code": 1001,
  "message": "Invalid credentials",
  "data": null,
  "timestamp": 1730972400
}
```

---

**更新日期**: 2024-11-06  
**版本**: v2.1.2  
**修复**: 响应格式处理逻辑  

