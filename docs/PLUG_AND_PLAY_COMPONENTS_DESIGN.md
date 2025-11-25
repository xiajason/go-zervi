# Router Service和Permission Service即插即用架构设计

## 🎯 设计目标

**核心理念**: Router Service和Permission Service设计为**即插即用型组件（Plug-and-Play Components）**，支持内部降级机制（Degraded Mode）。

**关键特性**:
1. ✅ **P0基础设施**: Central Brain必须可用（无降级）
2. ✅ **即插即用**: Router Service和Permission Service可选择性启用
3. ✅ **自动降级**: 数据库无数据或服务不可用时，自动启用降级模式
4. ✅ **优雅降级**: 降级模式下提供基础功能，完整功能待数据就绪后自动启用

---

## 🏗️ 架构设计

### 服务分层

```
┌─────────────────────────────────────────────────┐
│         P0基础设施层（必需，无降级）              │
│  ┌──────────────────────────────────────────┐   │
│  │  Central Brain (9000)                   │   │
│  │  - API网关                               │   │
│  │  - 服务代理                               │   │
│  │  - 健康检查                               │   │
│  │  - 熔断器                                 │   │
│  └──────────────────────────────────────────┘   │
│                                                 │
│  ┌──────────────────────────────────────────┐   │
│  │  Auth Service (8207)                    │   │
│  │  - 统一认证                               │   │
│  │  - JWT管理                                │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│    P1即插即用层（可选，支持降级）                │
│  ┌──────────────────┐  ┌──────────────────┐    │
│  │ Router Service   │  │Permission Service│    │
│  │   (8087)         │  │   (8086)         │    │
│  │                  │  │                  │    │
│  │ 即插即用组件      │  │ 即插即用组件      │    │
│  │                  │  │                  │    │
│  │ 降级模式:        │  │ 降级模式:        │    │
│  │ - 返回默认路由   │  │ - 开放所有权限   │    │
│  │ - 基本路由功能   │  │ - 基础权限检查   │    │
│  └──────────────────┘  └──────────────────┘    │
└─────────────────────────────────────────────────┘
```

---

## 🔧 降级机制设计

### 降级触发条件

#### Router Service降级条件

1. **服务不可用**
   - Router Service未启动
   - Router Service健康检查失败
   - 网络连接超时

2. **数据库无数据**
   - `route_config`表为空
   - `frontend_page_config`表为空
   - 数据库连接失败

3. **配置禁用**
   - `ROUTER_SERVICE_ENABLED=false`

---

#### Permission Service降级条件

1. **服务不可用**
   - Permission Service未启动
   - Permission Service健康检查失败
   - 网络连接超时

2. **数据库无数据 hoje**
   - 权限表为空
   - 角色表为空
   - 数据库连接失败

3. **配置禁用**
   - `PERMISSION_SERVICE_ENABLED=false`

---

### 降级行为设计

#### Router Service降级模式

**正常模式**:
- ✅ 查询数据库获取路由配置
- ✅ 根据用户角色过滤路由
- ✅ 提供动态路由功能

**降级模式**:
- 🔄 返回默认路由配置（内置）
- 🔄 所有用户获得相同的基础路由
- 🔄 记录降级日志，便于监控
- ✅ 服务仍然可用，提供基础功能

**降级策略**:
```go
// 默认路由配置（内置）
defaultRoutes := []RouteConfig{
    {Path: "/", Name: "首页", AccessLevel: "public"},
    {Path: "/login", Name: "登录", AccessCommission: "public"},
    {Path: "/register", Name: "注册", AccessLevel: "public"},
    // 更多默认路由...
}
```

---

#### Permission Service降级模式

**正常模式**:
- ✅ 查询数据库获取权限配置
- ✅ 根据用户角色验证权限
- ✅ 提供细粒度权限控制

**降级模式**:
- 🔄 使用默认权限策略（开放策略）
- 🔄 使其有认证用户获得基础权限
- 🔄 记录降级日志，便于监控
- ✅ 服务仍然可用，提供基础功能

**降级策略**:
```go
// 默认权限策略（开放策略）
defaultPermissionPolicy := &PermissionPolicy{
    AuthenticatedUsers: []string{"read", "write"}, // 基础权限
    PublicAccess: []string{"read"},                 // 公开访问
    AdminOnly: []string{"delete", "admin"},         // 管理员权限
}
```

---

## 💻 实现方案

### 1. Router Service降级实现

#### 1.1 服务启动时自检

```go
// services/infrastructure/router/main.go

func main() {
    // 初始化JobFirst核心包
    core, err := jobfirst.NewCore("")
    if err != nil {
        log.Printf("⚠️ JobFirst核心包初始化失败: %v，启用降级模式", err)
        startInDegradedMode()
        return
    }
    defer core.Close()

    // 数据库连接检查（可选）
    sqlDB, err := core.Database.GetPostgreSQL().GetSQLDB()
    if err != nil {
        log.Printf("⚠️ 数据库连接失败: %v，启用降级模式", err)
        startInDegradedMode()
        return
    }

    // 检查数据库是否有路由数据
    hasData := checkRouteData(sqlDB)
    if !hasData {
        log.Printf("ℹ️ 数据库无路由配置数据，启用降级模式（使用默认路由）")
        // 继续启动，但使用降级模式
    }

    // 正常启动
    startWithNormalMode(sqlDB, hasData)
}

// checkRouteData 检查数据库是否有路由数据
func checkRouteData(db *sql.DB) bool {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM route_config").Scan(&count)
    if err != nil {
        return false
    }
    return count > 0
}

// startInDegradedMode 降级模式启动
func startInDegradedMode() {
    r := gin.Default()
    
    // 设置标准路由
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "service": "router-service",
            "status":  "degraded",  // 降级状态
            "mode":    "degraded",
            "message": "服务运行在降级模式（无数据库连接）",
        })
    })

    // 提供默认路由配置
    r.GET("/api/v1/router/routes", func(c *gin.Context) {
        routes := getDefaultRoutes()
        c.JSON(http.StatusOK, gin.H{
            "code": 0,
            "data": routes,
            "mode": "degraded",
        })
    })

    // 启动服务
    r.Run(":8087")
}

// getDefaultRoutes 获取默认路由配置
func getDefaultRoutes() []RouteConfig {
    return []RouteConfig{
        {Path: "/", Name: "首页", AccessLevel: "public"},
        {Path: "/login", Name: "登录", AccessLevel: "public"},
        {Path: "/register", Name: "注册", AccessLevel: "public"},
        {Path: "/profile", Name: "个人中心", AccessLevel: "authenticated"},
        {Path: "/resume", Name: "简历管理", AccessLevel: "authenticated"},
        {Path: "/job", Name: "职位搜索", AccessLevel: "public"},
        {Path: "/company", Name: "企业中心", AccessLevel: "authenticated"},
    }
}
```

---

#### 1.2 Central Brain对Router Service的调用（支持降级）

```go
// shared/central-brain/router/client.go

type RouterClient struct {
    baseURL     string
    httpClient  *http.Client
    degradedMode bool  // 降级模式标志
    defaultRoutes []RouteConfig  // 默认路由配置
}

func NewRouterClient(baseURL string) *RouterClient {
    client := &RouterClient{
        baseURL: baseURL,
        httpClient: &http.Client{Timeout: 3 * time.Second},
        degradedMode: false,
        defaultRoutes: getDefaultRoutes(),
    }

    // 检查服务可用性
    client.checkServiceAvailability()

    return client
}

func (rc *RouterClient) checkServiceAvailability() {
    resp, err := rc.httpClient.Get(rc.baseURL + "/health")
    if err != nil {
        log.Printf("⚠️ Router Service不可用，启用降级模式: %v", err)
        rc.degradedMode = true
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("⚠️ Router Service健康检查失败，启用降级模式")
        rc.degradedMode = true
        return
    }

    // 检查是否已经是降级模式
    var health struct {
        Status string `json:"status"`
        Mode   string `json:"mode"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&health); err == nil {
        if health.Mode == "degraded" || health.Status == "degraded" {
            log.Printf("ℹ️ Router Service运行在降级模式")
            rc.degradedMode = true
        }
   ʣ
}

func (rc *RouterClient) GetAllRoutes() ([]RouteConfig, error) {
    if rc.degradedMode {
        log.Printf("ℹ️ 使用默认路由配置（降级模式）")
        return rc.defaultRoutes, nil
    }

    // 正常模式：调用Router Service
    resp, err := rc.httpClient.Get(rc.baseURL + "/api/v1/router/routes")
    if err != nil {
        log.Printf("⚠️ 调用Router Service失败，使用默认路由: %v", err)
        rc.degradedMode = true
        return rc.defaultRoutes, nil
    }
    defer resp.Body.Close()

    // 解析响应...
}
```

---

### 2. Permission Service降级实现

#### 2.1 服务启动时自检

```go
// services/infrastructure/permission/main.go

func main() {
    // 初始化JobFirst核心包
    core,922 err := jobfirst.NewCore("")
    if err != nil {
        log.Printf("⚠️ JobFirst核心包初始化失败: %v，启用降级模式", err)
        startInDegradedMode()
        return
    }
    defer core.Close()

    // 数据库连接检查（可选）
    sqlDB, err := core.Database.GetPostgreSQL().GetSQLDB()
    if err != nil {
        log.Printf("⚠️ 数据库连接失败: %v，启用降级模式", err)
        startInDegradedMode()
        return
    }

    // 检查数据库是否有权限数据
    hasData := checkPermissionData(sqlDB)
    if !hasData {
        log.Printf("ℹ️ 数据库无权限дей配置数据，启用降级模式（使用默认权限）")
        // 继续启动，但使用降级模式
    }

    // 正常启动
    startWithNormalMode(sqlDB, hasData)
}

// checkPermissionData 检查数据库是否有权限数据
func checkPermissionData(db *sql.DB) bool {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM permission").Scan(&count)
    if err != nil {
        return false
    }
    return count > 0
}

// startInDegradedMode 降级模式启动
func startInDegradedMode() {
    r := gin.Default()
    
    // 设置标准路由
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "service": "permission-service",
            "status":  "degraded",
            "mode":    "degraded",
            "message": "服务运行在降级模式（无数据库连接 aucune données）",
        })
    })

    // 提供默认权限策略
    r.GET("/api/v1/permission/check", func(c *gin.Context) {
        // 降级模式：开放所有权限
        c.JSON(http.StatusOK, gin.H{
            "code": 0,
            "data": gin.H{
                "allowed": true,
                "reason":  "degraded_mode",
            },
            "mode": "degraded",
        })
    })

    // 启动服务
    r.Run(":8086")
}

// 默认权限策略（开放策略）
func getDefaultPermissionPolicy() *PermissionPolicy {
    return &PermissionPolicy{
        AuthenticatedUsers: []string{"read", "write"},
        PublicAccess:       []string{"read"},
        AdminOnly:          []string{"delete", "admin"},
    }
}
```

---

## 📊 降级状态监控

### 健康检查响应格式

**正常模式**:
```json
{
  "service": "router-service",
  "status": "healthy",
  "mode": "normal",
  "routes_count": 15,
  "timestamp": "2025-01-29T10:00:00Z"
}
```

**降级模式**:
```json
{
  "service": "router-service",
  "status": "degraded",
  "mode": "degraded",
  "reason": "no_database_data",
  "default_routes_count": 7,
  "timestamp": "2025-01-29T10:00:00Z"
}
```

---

## 🔄 自动恢复机制

### 自动检测数据就绪

```go
// services/infrastructure/router/main.go

func startWithAutoRecovery(db *sql.DB) {
    // 启动时检查数据
    hasData := checkRouteData(db)
    
    if !hasData {
        // 降级模式启动
        startInDegradedModeWithRecovery(db)
        return
    }

    // 正常模式启动
    startWithNormalMode(db, true)
}

// startInDegradedModeWithRecovery 降级模式启动（带自动恢复）
func startInDegradedModeWithRecovery(db *sql.DB) {
    r := gin.Default()
    
    // 设置降级模式路由
    setupDegradedRoutes(r)

    // 后台检查数据就绪
    go func() {
        ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
        for range ticker.C {
            if checkRouteData(db) {
                log.Printf("✅ 检测到路由数据已就绪，自动切换到正常模式")
                // 重新加载路由配置
                reloadRoutes(db)
                ticker.Stop()
            }
        }
    }()

    r.Run(":8087")
}
```

---

## 📋 配置支持

### 环境变量配置

```bash
# Router Service配置
ROUTER_SERVICE_ENABLED=true              # 是否启用Router Service
ROUTER_SERVICE_DEGRADED_MODE=true        # 是否允许降级模式
ROUTER_SERVICE_AUTO_RECOVERY=true        # 是否自动恢复
ROUTER_SERVICE_CHECK_INTERVAL=30         # 数据检查间隔（秒）

# Permission Service配置
PERMISSION_SERVICE_ENABLED=true          # 是否启用Permission Service
PERMISSION_SERVICE_DEGRADED_MODE=true    # 是否允许降级模式
PERMISSION_SERVICE_AUTO_RECOVERY=true    # 是否自动恢复
PERMISSION_SERVICE_CHECK_INTERVAL=30     # 数据检查间隔（秒）
```

---

## ✅ 优势总结

### 1. 高可用性 ✅

- ✅ Central Brain始终可用（P0级别）
- ✅ Router Service和Permission Service降级但不中断
- ✅ 提供基础功能，不阻塞核心流程

---

### 2. 渐进式部署 ✅

- ✅ 可以先启动Central Brain（核心功能可用）
- ✅ 逐步添加Router Service和Permission Service
- ✅ 数据就绪后自动启用完整功能

---

### 3. 开发友好 ✅

- ✅ 开发环境无需完整数据即可启动
- ✅ 测试环境可以使用降级模式
- ✅ 生产环境数据就绪后自动启用Cookie功能

---

### 4. 运维友好 ✅

- ✅ 数据库故障时系统仍可用
- ✅ 清晰的降级状态监控
- ✅ 自动恢复机制

---

## 🚀 实施优先级

| 优先级 | 任务 | 预计时间 |
|--------|------|----------|
| 🔥 最高 | Router Service降级模式实现 | 2小时 |
| 🔥 高 | Permission Service降级模式实现 | 2小时 |
| ⭐ 中 | Central Brain降级处理 | 1小时 |
| ⭐ 低 | 自动恢复机制 | 1小时 |

---

**报告生成时间**: 2025-01-29  
**设计理念**: **即插即用、优雅降级、自动恢复**

