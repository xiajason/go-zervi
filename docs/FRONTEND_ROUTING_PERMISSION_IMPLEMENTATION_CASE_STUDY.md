# 小程序与Web端动态路由权限管理实现案例学习报告

## 📋 报告概述

**学习时间**: 2025-01-29  
**学习项目**: 
1. 小程序端：`/Users/szjason72/resume-center/miniprogram-4`
2. Web端：`/Users/szjason72/resume-center/简历中心后台代码/resume-centre`

**学习目的**: 理解实际项目中路由权限管理的实现方式，为Zervigo项目提供参考

---

## 🎯 小程序端实现分析

### 1. 认证流程

**实现方式**:
```typescript
// app.ts - 应用启动
onLaunch() {
  wx.login({
    success: res => {
      // 获取微信code，发送到后端换取openId和token
    }
  })
}

// login.ts - 登录页面
getPhone(e: any) {
  userModel.login(e.detail.code).then((res: any) => {
    // 登录成功，保存token
    wx.setStorageSync('accessToken', res.accessToken)
    wx.isLogin = true
    wx.switchTab({ url: "/pages/home/home" })
  })
}
```

**关键发现**:
- ✅ 使用`wx.getStorageSync('accessToken')`存储token
- ✅ 使用全局变量`wx.isLogin`标记登录状态
- ✅ Token验证通过`userModel.check()`检查

---

### 2. HTTP请求封装

**实现方式**:
```typescript
// utils/http.ts
class HTTP {
  request(url: string, params: Record<string, any> = {}, data = {}, method: METHOD_TYPE = 'GET') {
    wx.request({
      url: config.api_base_url + url,
      method,
      data,
      header: {
        'content-type': 'application/json',
        accessToken: wx.getStorageSync('accessToken')  // ⭐ 关键：从本地存储读取token
      },
      success: (res: any) => {
        if(res.data.code === 0) {
          resolve(res.data.data)
        } else if (res.data.code === 100001 || res.data.code === 100002) {
          // Token失效处理
          wx.showToast({
            title: "登录已过期",
            icon:'none',
            duration:2000
          })
          reject(res.data.code)
        }
      }
    })
  }
}
```

**关键发现**:
- ✅ 统一的HTTP请求封装
- ✅ 自动添加accessToken到header
- ✅ Token失效时统一处理（code 100001/100002）
- ✅ 错误统一提示

---

### 3. Token验证机制

**实现方式**:
```typescript
// pages/home/home.ts
init() {
  if(wx.getStorageSync('accessToken')) {
    userModel.check().then((status: any) => {
      // status: 0=正常, 1=需要登录, 2=需要人脸识别
      this.setLoginStatus(true)
      this.checkStatus(status)
    }).catch(() => {
      this.setLoginStatus(false)
    })
  } else {
    this.setLoginStatus(false)
  }
}

checkStatus(status: number) {
  if(status === 1) {
    wx.reLaunch({ url: "/pages/login/login" })
    return
  }
  if(status === 2) {
    wx.redirectTo({ url: "/pages/face/face" })
    return
  }
  // 正常状态，加载数据
  this.getHomeInfo()
}
```

**关键发现**:
- ✅ 页面加载时检查token有效性
- ✅ 根据token状态（status）决定页面跳转
- ✅ 支持多级认证状态（已登录、未登录、需要人脸识别）

---

### 4. 路由使用方式

**实现方式**:
```typescript
// models/resume.ts - 直接硬编码API路径
class ResumeModel extends HTTP {
  getResumes() {
    return this.request('/personal/resume/list/summary', {}, {}, 'GET')
  }
  
  getResumeInfo(resumeId: string) {
    return this.request(`/personal/resume/detail/${resumeId}`, {}, {}, 'GET')
  }
}

// pages/home/home.ts - 直接调用API
getHomeInfo() {
  homeModel.getBanner().then((res: any) => {
    // 处理数据
  })
  
  homeModel.getNotifications({
    pageNum: 1,
    pageSize: 10
  }).then((res: any) => {
    // 处理数据
  })
}
```

**关键发现**:
- ⚠️ **API路径硬编码**在代码中
- ⚠️ **没有动态路由管理**
- ⚠️ **权限控制在后端API层面**，前端不感知
- ✅ 使用Model层封装API调用，代码组织清晰

---

### 5. 权限控制实现

**小程序端权限控制特点**:
- ❌ **前端不做权限控制**，直接调用API
- ✅ **权限控制在后端**，无权限时返回错误码
- ✅ **前端根据错误码处理**（如code 100001/100002表示需要登录）

**优缺点分析**:
- ✅ 优点：实现简单，前端代码清晰
- ❌ 缺点：前端无法提前知道哪些功能可用，用户体验可能不佳
- ❌ 缺点：权限变更需要修改代码，不够灵活

---

## 🌐 Web端实现分析（Java Spring Cloud Gateway）

### 1. API Gateway架构

**实现方式**:
```java
// AuthorizeFilter.java - 全局过滤器
@Component
public class AuthorizeFilter implements GlobalFilter, Ordered {
  private static final String AUTHORIZE_TOKEN = "accessToken";
  
  @Value("${whites}") private String[] whites;  // 白名单配置
  
  @Override
  public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
    ServerHttpRequest request = exchange.getRequest();
    
    // 1. 检查白名单
    if (whites != null) {
      for (String white : whites) {
        if (request.getURI().getPath().contains(white)) {
          return chain.filter(exchange);  // 白名单直接放行
        }
      }
    }
    
    // 2. 提取token
    HttpHeaders headers = request.getHeaders();
    String token = headers.getFirst(AUTHORIZE_TOKEN);
    
    if (StringUtils.isEmpty(token)) {
      return unauthorizedResponse(exchange, ErrorCode.GATEWAY_ACCESS_TOKEN_EMPTY);
    }
    
    // 3. 验证token
    try {
      TokenSubject tokenSubject = jwtComponent.parseToken(token);
      if (tokenSubject == null) {
        return unauthorizedResponse(exchange, ErrorCode.GATEWAY_ACCESS_TOKEN_INVALID);
      }
    } catch (Exception e) {
      return unauthorizedResponse(exchange, ErrorCode.GATEWAY_ACCESS_TOKEN_INVALID);
    }
    
    // 4. 验证通过，继续转发
    return chain.filter(exchange);
  }
}
```

**关键发现**:
- ✅ 使用全局过滤器统一处理认证
- ✅ 配置白名单，某些路径不需要认证
- ✅ Token验证失败时返回统一错误码
- ✅ 使用JWT解析token，获取用户信息

---

### 2. 路由配置

**实现方式**:
```yaml
# application.yml
spring:
  cloud:
    gateway:
      routes:
        - id: 资源接口
          uri: lb://resource
          predicates:
            - Path=/resource/**
        - id: 个人端后台
          uri: lb://personal
          predicates:
            - Path=/personal/**
        - id: 企业端后台
          uri: lb://enterprise
          predicates:
            - Path=/enterprise/**
```

**关键发现**:
- ✅ 使用配置文件定义路由规则
- ✅ 使用服务发现（lb://）进行负载均衡
- ✅ 路径匹配模式（Path=/personal/**）
- ⚠️ **路由配置是静态的**，不是从数据库读取

---

### 3. 白名单配置

**实现方式**:
```yaml
whites: >
  /v2/api-docs,
  /admin/version/,
  /admin/authentication/login,
  /personal/version/,
  /personal/authentication/login,
  /personal/home/banners,
  /enterprise/version/,
  /enterprise/authentication/login,
  /resource/version/,
  /resource/ocr/general,
  /resource/dict/data,
  /resource/urls,
  /open/version/,
  /open/api/statistics/resume
```

**关键发现**:
- ✅ 白名单配置清晰，公开API不需要认证
- ✅ 登录接口在白名单中
- ✅ 版本检查接口在白名单中
- ✅ 静态资源接口在白名单中

---

## 🔍 对比分析：实际项目 vs Zervigo需求

### 实际项目的实现方式

| 项目 | 路由管理 | 权限控制 | 前端实现 |
|------|---------|---------|---------|
| **小程序端** | 硬编码API路径 | 后端API层面 | 直接调用API，根据错误码处理 |
| **Web端（Java）** | 配置文件静态路由 | Gateway过滤器 | 直接调用API，Gateway统一认证 |

### Zervigo的需求（基于Router Service）

| 需求 | 实现方式 | 优势 |
|------|---------|------|
| **动态路由** | 从数据库读取路由配置 | ✅ 灵活，可动态调整 |
| **权限控制** | 根据用户角色和权限过滤路由 | ✅ 精细权限控制 |
| **前端路由获取** | 调用`/api/v1/router/user-routes`获取可访问路由 | ✅ 前端知道哪些功能可用 |

---

## 💡 关键发现和启示

### 1. 实际项目的实现方式（简单但不够灵活）

**小程序端**:
- ✅ 使用统一的HTTP请求封装
- ✅ Token自动添加到header
- ✅ 错误统一处理
- ❌ API路径硬编码
- ❌ 没有动态路由管理
- ❌ 权限控制在后端，前端不感知

**Web端（Java）**:
- ✅ Gateway统一认证
- ✅ 白名单配置清晰
- ✅ 使用服务发现和负载均衡
- ❌ 路由配置是静态的（配置文件）
- ❌ 没有基于角色的动态路由

---

### 2. Zervigo需要实现的（更灵活但更复杂）

**基于Router Service的动态路由系统**:
- ✅ 路由配置存储在数据库，可动态调整
- ✅ 根据用户角色和权限返回可访问路由
- ✅ 前端可以根据路由列表动态构建UI
- ⚠️ 实现复杂度更高
- ⚠️ 需要数据库支持

---

### 3. 最佳实践建议

#### 方案A：混合方案（推荐）

**小程序端**（参考实际项目）:
```typescript
// 简化的实现方式
// 1. 登录后保存token
wx.setStorageSync('accessToken', token)

// 2. 直接调用API（路径硬编码）
const banners = await ApiService.getBanners()

// 3. 根据错误码处理权限问题
if (error.code === 403) {
  // 无权限，隐藏功能
}
```

**Web端**（参考Java Gateway）:
```typescript
// 类似的实现方式
// 1. 统一的HTTP拦截器添加token
axios.interceptors.request.use(config => {
  config.headers.accessToken = localStorage.getItem('accessToken')
  return config
})

// 2. 直接调用API
const banners = await ApiService.getBanners()
```

**但是**，如果Zervigo需要更精细的权限控制：
- ✅ 使用Router Service的动态路由
- ✅ 前端调用`/api/v1/router/user-routes`获取路由列表
- ✅ 根据路由列表动态构建UI

---

#### 方案B：完全动态路由方案（Zervigo Router Service）

**小程序端**:
```typescript
// 1. 登录后获取路由列表
const routes = await RouteService.getUserRoutes()

// 2. 根据路由列表构建API调用
if (RouteService.hasRoute('home.banners')) {
  const banners = await ApiService.getBanners()
}
```

**Web端**:
```typescript
// 类似的实现
const routes = await RouteService.getUserRoutes()

// 根据路由列表动态显示菜单和功能
routes.forEach(route => {
  if (route.routeKey === 'admin.users') {
    // 显示用户管理菜单
  }
})
```

---

## 🎯 为Zervigo设计的综合方案

### 方案设计原则

1. **兼容性优先**
   - 保持现有API调用方式（向后兼容）
   - 同时支持动态路由（可选功能）

2. **渐进式迁移**
   - 第一阶段：简单的静态路由代理（类似Java Gateway）
   - 第二阶段：集成Router Service的动态路由

3. **灵活配置**
   - 支持白名单配置（公开API）
   - 支持动态路由配置（数据库）
   - 支持静态路由配置（配置文件）

---

### 实施建议

#### Phase 1: 基础路由代理（立即实施）

**实现方式**（参考Java Gateway）:
```go
// Central Brain实现类似Java Gateway的功能
// 1. 全局过滤器验证token
// 2. 白名单配置
// 3. 路径匹配路由转发

// 优点：实现简单，快速可用
// 缺点：路由配置是静态的
```

#### Phase 2: 集成Router Service（后续优化）

**实现方式**（参考Router Service）:
```go
// Central Brain集成Router Service
// 1. 从数据库读取路由配置
// 2. 根据用户角色和权限过滤路由
// 3. 提供路由列表API给前端
// 4. 动态路由代理

// 优点：灵活，支持精细权限控制
// 缺点：复杂度较高
```

---

## 📊 小程序端最佳实践总结

### 1. Token管理
```typescript
// ✅ 推荐方式
// 存储token
wx.setStorageSync('accessToken', token)

// 获取token
const token = wx.getStorageSync('accessToken')

// 清除token
wx.clearStorageSync()
```

### 2. HTTP请求封装
```typescript
// ✅ 推荐方式
class HTTP {
  request(url, params, data, method) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: config.api_base_url + url,
        method,
        data,
        header: {
          'content-type': 'application/json',
          accessToken: wx.getStorageSync('accessToken')  // 自动添加token
        },
        success: (res) => {
          if (res.data.code === 0) {
            resolve(res.data.data)
          } else if (res.data.code === 100001 || res.data.code === 100002) {
            // Token失效处理
            wx.showToast({ title: "登录已过期" })
            reject(res.data.code)
          } else {
            wx.showToast({ title: res.data.msg || res.data.error })
          }
        }
      })
    })
  }
}
```

### 3. Token验证
```typescript
// ✅ 推荐方式
// 页面加载时检查token
onLoad() {
  if (wx.getStorageSync('accessToken')) {
    userModel.check().then((status) => {
      if (status === 0) {
        // 正常，加载数据
        this.loadData()
      } else if (status === 1) {
        // 需要登录
        wx.reLaunch({ url: "/pages/login/login" })
      }
    })
  } else {
    // 未登录
    wx.reLaunch({ url: "/pages/login/login" })
  }
}
```

### 4. 错误处理
```typescript
// ✅ 推荐方式
// 统一错误处理
if (res.data.code === 100001 || res.data.code === 100002) {
  wx.showToast({
    title: "登录已过期",
    icon: 'none',
    duration: 2000
  })
  reject(res.data.code)
}
```

---

## 📊 Web端最佳实践总结（Java Gateway）

### 1. Gateway过滤器
```java
// ✅ 推荐方式
@Component
public class AuthorizeFilter implements GlobalFilter, Ordered {
  @Override
  public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
    // 1. 检查白名单
    // 2. 提取token
    // 3. 验证token
    // 4. 继续转发
  }
}
```

### 2. 路由配置
```yaml
# ✅ 推荐方式
spring:
  cloud:
    gateway:
      routes:
        - id: 个人端后台
          uri: lb://personal
          predicates:
            - Path=/personal/**
```

### 3. 白名单配置
```yaml
# ✅ 推荐方式
whites: >
  /personal/authentication/login,
  /personal/home/banners,
  /resource/version/
```

---

## 🎯 为Zervigo设计的综合方案

### 方案设计：混合路由管理

#### 1. 静态路由 + 动态路由结合

**静态路由**（类似Java Gateway）:
```go
// 配置文件定义的基础路由
routes := map[string]ServiceProxy{
  "auth": {
    ServiceName: "auth-service",
    BaseURL:     "http://localhost:8207",
    PathPrefix:  "/api/v1/auth",
  },
  // ...
}
```

**动态路由**（集成Router Service）:
```go
// 从数据库读取的动态路由
routes := cb.getAllRouteConfigs()  // 从数据库查询
accessibleRoutes := cb.getAccessibleRoutes(userRoles)  // 根据权限过滤
```

**优先级**:
1. 先检查动态路由（数据库）
2. 如果没有匹配，检查静态路由（配置文件）
3. 如果都没有，返回404

---

#### 2. 前端适配方案

**小程序端**（参考实际项目）:
```typescript
// 方案1：简单方式（推荐用于MVP）
// 直接调用API，根据错误码处理权限
const banners = await ApiService.getBanners()
if (error.code === 403) {
  // 无权限，隐藏功能
}

// 方案2：动态路由（用于后续优化）
// 先获取路由列表
const routes = await RouteService.getUserRoutes()
if (RouteService.hasRoute('home.banners')) {
  const banners = await ApiService.getBanners()
}
```

**Web端**（参考Java Gateway）:
```typescript
// 类似的实现方式
// 统一的HTTP拦截器
axios.interceptors.request.use(config => {
  config.headers.accessToken = localStorage.getItem('accessToken')
  return config
})

// 直接调用API
const banners = await ApiService.getBanners()
```

---

### 3. 权限控制流程

**推荐流程**:
```
1. 用户登录
   ↓
2. 获取Token + 用户信息
   ↓
3. （可选）调用 /api/v1/router/user-routes 获取可访问路由
   ↓
4. 前端根据路由列表动态构建UI（如果有动态路由）
   ↓
5. 用户操作触发API调用
   ↓
6. Central Brain验证Token和权限
   ↓
7. 转发到目标服务
```

---

## 📋 实施建议

### 立即实施（MVP阶段）

**参考实际项目的简单实现**:
1. ✅ 统一的HTTP请求封装（自动添加token）
2. ✅ Token验证和错误处理
3. ✅ 静态路由代理（类似Java Gateway）
4. ✅ 白名单配置

**优点**:
- ✅ 实现简单，快速可用
- ✅ 代码清晰，易于维护
- ✅ 符合实际项目的最佳实践

---

### 后续优化（增强阶段）

**集成Router Service的动态路由**:
1. ⚠️ Central Brain集成Router Service
2. ⚠️ 提供路由列表API给前端
3. ⚠️ 前端根据路由列表动态构建UI
4. ⚠️ 动态路由代理和权限验证

**优点**:
- ✅ 灵活，支持精细权限控制
- ✅ 可动态调整路由配置
- ✅ 前端可以提前知道哪些功能可用

---

## ✅ 总结

### 实际项目的成功经验

1. **小程序端**:
   - ✅ 统一的HTTP请求封装
   - ✅ Token自动管理
   - ✅ 错误统一处理
   - ⚠️ API路径硬编码（简单但不够灵活）

2. **Web端（Java Gateway）**:
   - ✅ Gateway统一认证
   - ✅ 白名单配置清晰
   - ✅ 使用服务发现和负载均衡
   - ⚠️ 路由配置是静态的（简单但不够灵活）

### Zervigo应该采用的方案

**MVP阶段**（参考实际项目）:
- ✅ 简单的静态路由代理
- ✅ 统一的Token验证
- ✅ 白名单配置

**后续优化**（集成Router Service）:
- ✅ 动态路由管理
- ✅ 基于角色的权限控制
- ✅ 前端路由列表API

**混合方案**:
- ✅ 静态路由 + 动态路由结合
- ✅ 先检查动态路由，再检查静态路由
- ✅ 向后兼容，渐进式迁移

---

**报告生成时间**: 2025-01-29  
**关键启示**: 实际项目使用简单的静态路由，但Zervigo需要更灵活的动态路由系统  
**建议**: 先实现简单的静态路由（MVP），再逐步集成动态路由系统

