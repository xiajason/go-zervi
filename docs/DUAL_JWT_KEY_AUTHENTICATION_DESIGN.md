# 双JWT密钥认证架构设计方案

## 🎯 设计方案概述

您提出的**双JWT密钥认证架构**是一个非常清晰和合理的设计：

### 密钥用途划分

1. **`zervigo-mvp-secret-key-2025`** - 微服务集群内部认证凭证
   - **用途**: 服务间相互认证
   - **场景**: Central Brain ↔ Auth Service, Permission Service ↔ Router Service 等
   - **特点**: 内部专用，不对外暴露

2. **`jobfirst-unified-auth-secret-key-2024`** - 外部用户API访问认证凭证
   - **用途**: 用户访问API的认证
   - **场景**: 前端应用 → API Gateway → 业务服务
   - **特点**: 对外暴露，用户可见

## ✅ 方案可行性分析

### 优势

1. **职责清晰**
   - 内部认证和外部认证完全分离
   - 降低密钥泄露风险（即使外部密钥泄露，也无法伪造服务）

2. **安全性高**
   - 服务间使用独立的密钥，攻击者无法通过获取用户token伪装服务
   - 用户token无法用于服务间调用

3. **易于管理**
   - 可以独立更新密钥而不影响另一方
   - 便于密钥轮换和权限控制

4. **符合微服务最佳实践**
   - 服务间认证和用户认证分离
   - 支持细粒度的权限控制

### 潜在问题

1. **需要维护两套JWT逻辑**
   - 需要区分服务token和用户token
   - 需要两套Claims结构

2. **代码复杂度增加**
   - 需要判断请求来源（服务间还是用户请求）
   - 需要两套验证逻辑

## 🏗️ 架构设计

### 服务间认证流程

```
┌─────────────────┐
│  Central Brain  │
│  (API Gateway)  │
└────────┬────────┘
         │ 1. 验证用户token (jobfirst-2024)
         │ 2. 获取用户信息
         ▼
┌─────────────────┐
│  Auth Service   │ด้วย zervigo-2025 token
└────────┬────────┘
         │ 3. 验证服务token (zervigo-2025)
         │ 4. 返回用户权限信息
         ▼
┌─────────────────┐
│Permission Service│ด้วย zervigo-2025 token
└─────────────────┘
```

### 用户认证流程

```
┌─────────┐
│ 前端应用 │
└────┬────┘
     │ 1. 用户登录请求
     ▼
┌─────────────────┐
│  Central Brain  │
│  (API Gateway)  │
└────┬────────────┘
     │ 2. 转发到Auth Service
     ▼
┌─────────────────┐
│  Auth Service   │
│  (生成jobfirst-2024 token)
└────┬────────────┘
     │ 3. 返回用户token (jobfirst-2024)
     ▼
┌─────────┐
│ 前端应用 │
│ (保存token)
└────┬────┘
     │ 4. 后续请求携带token
     ▼
┌─────────────────┐
│  Central Brain  │
│  验证用户token  │
│ (jobfirst-2024)
└────┬────────────┘
     │ 5. 转发到业务服务（添加服务token）
     ▼
┌─────────────────┐
│   业务服务      │
│ (验证服务token) │
│ (zervigo-2025)
└─────────────────┘
```

## 🔧 实现细节

### 1. JWT Claims结构设计

#### 服务Token Claims (zervigo marvel-2025)
```go
type ServiceTokenClaims struct {
    ServiceID      string   `json:"service_id"`       // 服务ID
    ServiceName    string   `json:"service_name"`     // 服务名称
    ServiceType    string   `json:"service_type"`     // core/infrastructure/business
    AllowedAPIs    []string `json:"allowed_apis"`     // 允许调用的API列表
    jwt.RegisteredClaims
}
```

#### 用户Token Claims (jobfirst-2024)
```go
type UserTokenClaims struct {
    UserID      uint     `json:"user_id"`
    Username    string   `json:"username"`
    Email       string   `json:"email"`
    Role        string   `json:"role"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}
```

### 2. Auth Service职责扩展

Auth Service需要提供：

#### 服务认证API
```go
// POST /api/v1/auth/service/login
// 服务启动时调用，获取服务token
{
    "service_id": "auth-service",
    "service_secret": "xxx"
}

// 返回服务token (使用zervigo-2025密钥)
{
    "service_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
}
```

#### 服务Token验证API
```go
// POST /api/v1/auth/service/validate
// 验证服务token
{
    "service_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

// 返回验证结果
{
    "valid": true,
    "service_id": "central-brain",
    "service_name": "Central Brain",
    "allowed_apis": ["*"]
}
```

#### 用户认证API (保持现有)
```go
安全
// POST /api/v1/auth/login
// 用户登录，生成用户token (使用jobfirst-2024密钥)
```

### 3. Central Brain实现

```go
// Central Brain需要：
// 1. 验证用户token (jobfirst-2024)
// 2. 携带服务token (zervigo-2025) 调用其他服务

func (cb *CentralBrain) proxyRequest(c *gin.Context, service ServiceProxy) {
    // 1. 验证用户token（如果是用户请求）
    userToken := extractUserToken(c)
    if userToken != "" {
        userInfo, err := cb.validateUserToken(userToken) // jobfirst-2024
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid user token"})
            return
        }
        c.Set("user", userInfo)
    }
    
    // 2. 获取服务token (zervigo-2025)
    serviceToken := cb.getServiceToken()
    
    // 3. 转发请求，添加服务token
    req.Header.Set("X-Service-Token", serviceToken)
    req.Header.Set("X-Service-ID", "central-brain")
    
    // 4. 如果是用户请求，也保留用户token
    if userToken != "" {
        req.Header.Set("Authorization", "Bearer "+userToken)
    }
}
```

### 4. 其他服务实现

```go
// 业务服务需要验证两种token：
// 1. 服务token (zervigo-2025) - 验证请求来源
// 2. 用户token (jobfirst-2024) - 验证用户身份

func RequireServiceAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 验证服务token
        serviceToken := c.GetHeader("X-Service-Token")
        if serviceToken == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing service token"})
            return
        }
        
        // 验证服务token (zervigo-2025)
        serviceInfo, err := validateServiceToken(serviceToken)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid service token"})
            return
        }
        
        c.Set("service", serviceInfo)
        c.Next()
    }
}
```

## 📋 实施步骤

### Phase 1: 准备阶段
1. ✅ 确认密钥用途划分（已完成）
2. 在数据库中创建服务凭证表
3. 扩展Auth Service支持服务认证

### Phase 2: 实现服务认证
1. 实现服务token生成和验证
2. endure Auth Service支持服务登录和验证API
3. 更新服务Claims结构

### Phase 3: 更新服务实现
1. 겹 Central Brain添加服务token转发
2. 更新各服务支持服务token验证
3. 保持用户token验证（jobfirst-2024）不变

### Phase 4: 测试验证
1. 测试服务间认证
2. 测试用户认证不受影响
3. 测试混合场景（用户通过Gateway访问）

## ❓ 需要确认的问题

1. **服务凭证管理**
   - 服务ID和Secret如何生成和存储？
   - 是否需要在数据库中创建 `zervigo_service_credentials` 表？

2. **服务token生命周期**
   - 服务token有效期多久？（建议：24小时或更长）
   - 是否需要refresh机制？

3. **服务发现集成**
   - 服务token获取是否集成到Consul注册流程中？
   - 服务启动时自动获取token？

4. **向后兼容**
   - 现有服务是否需要逐步迁移？
   - 还是统一一次性更新？

5. **错误处理**
   - 服务token失效时的处理策略？
   - 是否需要自动重新获取？

## ✅ 结论

**这个设计方案完全可行！** 这是一个经典的内外分离认证架构，具有以下优势：

- ✅ 职责清晰：内部认证和外部认证分离
- ✅ 安全性高：密钥隔离，降低泄露风险
- ✅ 易于管理：可以独立更新和维护
- ✅ 符合最佳实践：符合微服务安全架构模式

**建议立即开始实施！**
