# 即插即用组件热加载机制设计

## 🎯 核心理念

**降级机制 = 微服务热加载** ✅

**关键理解**:
- ✅ Router Service和Permission Service可以**动态切换**降级模式和正常模式
- ✅ **无需重启Central Brain** - 自动检测并切换
- ✅ 数据就绪后**自动恢复** - 无缝切换到正常模式
- ✅ 支持**联调联试** - 开发过程中数据逐步完善，服务自动适应

---

## 🔄 热加载机制设计

### 工作流程

```
┌─────────────────────────────────────────────────┐
│  1. 初始状态（降级模式）                          │
│  Router Service: 降级模式（默认路由）             │
│  Central Brain: 检测到降级，使用默认路由         │
└─────────────────────────────────────────────────┘
                    ↓
         [数据逐步导入/配置逐步完善]
                    ↓
┌─────────────────────────────────────────────────┐
│  2. 自动检测（定期检查）                         │
│  Router Service: 检查数据库是否有数据           │
│  Central Brain: 定期检查Router Service状态      │
└─────────────────────────────────────────────────┘
                    ↓
         [检测到数据就绪]
                    ↓
┌─────────────────────────────────────────────────┐
│  3. 自动切换（无需重启）                         │
│  Router Service: 切换到正常模式                 │
│  Central Brain: 检测到正常模式，使用动态路由     │
│  ✅ 全程无需重启Central Brain！                  │
└─────────────────────────────────────────────────┘
```

---

## 💻 实现方案

### 1. Router Service热切换实现

#### 1.1 状态管理

```go
// services/infrastructure/router/main.go

type ServiceMode int

const (
    ModeUnknown ServiceMode = iota
    ModeDegraded            // 降级模式
    ModeNormal             // 正常模式
)

type RouterService struct {
    mode          ServiceMode
    db            *sql.DB
    defaultRoutes []RouteConfig
    mutex         sync.RWMutex
}

// 启动时初始化
func NewRouterService(db *sql.DB) *RouterService {
    rs := &RouterService{
        db:            db,
        defaultRoutes: getDefaultRoutes(),
    }
    
    // 初始检查，确定模式
    rs.checkAndSwitchMode()
    
    return rs
}

// checkAndSwitchMode 检查并切换模式（热切换）
func (rs *RouterService) checkAndSwitchMode() {
    rs.mutex.Lock()
    defer rs.mutex.Unlock()
    
    // 检查数据库是否有数据
    hasData := rs.checkRouteData()
    
    oldMode := rs.mode
    if hasData {
        rs.mode = ModeNormal
        if oldMode == ModeDegraded {
            log.Printf("✅ 检测到路由数据已就绪，自动切换到正常模式（热切换，无需重启）")
        }
    } else {
        rs.mode = ModeDegraded
        if oldMode == ModeNormal {
            log.Printf("⚠️ 路由数据不可用，切换到降级模式（热切换，无需重启）")
        }
    }
}
```

---

#### 1.2 后台自动检测

```go
// services/infrastructure/router/main.go

func (rs *RouterService) StartAutoRecovery() {
    ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
    
    go func() {
        for range ticker.C {
            rs.checkAndSwitchMode()
        }
    }()
    
    log.Printf("🔄 Router Service自动恢复机制已启动（每30秒检查一次）")
}

// 路由查询时动态选择
func (rs *RouterService) GetRoutes() []RouteConfig {
    rs.mutex.RLock()
    mode := rs.mode
    rs.mutex.RUnlock()
    
    if mode == ModeNormal {
        // 正常模式：查询数据库
        return rs.getRoutesFromDB()
    } else {
        // 降级模式：返回默认路由
        return rs.defaultRoutes
    }
}
```

---

#### 1.3 健康检查反映当前状态

```go
// services/infrastructure/router/main.go

func (rs *RouterService) GetHealth() gin.H {
    rs.mutex.RLock()
    mode := rs.mode
    rs.mutex.RUnlock()
    
    health := gin.H{
        "service": "router-service",
        "status":  "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    if mode == ModeDegraded {
        health["mode"] = "degraded"
        health["reason"] = "no_database_data"
        health["routes_count"] = len(rs.defaultRoutes)
        health["message"] = "服务运行在降级模式（使用默认路由），无需重启即可自动恢复"
    } else {
        health["mode"] = "normal"
        routes := rs.getRoutesFromDB()
        health["routes_count"] = len(routes)
        health["message"] = "服务运行在正常模式（动态路由）"
    }
    
    return health
}
```

---

### 2. Central Brain热检测实现

#### 2.1 Router Client支持热切换

```go
// shared/central-brain/router/client.go

type RouterClient struct {
    baseURL       string
    httpClient    *http.Client
    degradedMode  bool
    defaultRoutes []RouteConfig
    lastCheck     time.Time
    checkInterval time.Duration
    mutex         sync.RWMutex
}

func NewRouterClient(baseURL string) *RouterClient {
    client := &RouterClient{
        baseURL:       baseURL,
        httpClient:    &http.Client{Timeout: 3 * time.Second},
        degradedMode:  false,
        defaultRoutes: getDefaultRoutes(),
        lastCheck:     time.Now(),
        checkInterval: 30 * time.Second,
    }
    
    // 初始检查
    client.checkServiceMode()
    
    // 启动后台检查
    go client.startPeriodicCheck()
    
    return client
}

// checkServiceMode 检查Router Service模式（热检测）
func (rc *RouterClient) checkServiceMode() {
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    
    // 检查健康状态
    resp, err := rc.httpClient.Get(rc.baseURL + "/health")
    if err != nil {
        // 服务不可用，使用降级模式
        if !rc.degradedMode {
            log.Printf("⚠️ Router Service不可用，切换到降级模式（热切换，无需重启）")
        }
        rc.degradedMode = true
        rc.lastCheck = time.Now()
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        rc.degradedMode = true
        rc.lastCheck = time.Now()
        return
    }
    
    // 解析健康状态
    var health struct {
        Status string `json:"status"`
        Mode   string `json:"mode"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        rc.degradedMode = true
        rc.lastCheck = time.Now()
        return
    }
    
    // 判断模式
    oldMode := rc.degradedMode
    rc.degradedMode = (health.Mode == "degraded" || health.Status == "degraded")
    
    // 模式切换日志
    if oldMode != rc.degradedMode {
        if rc.degradedMode {
            log.Printf("⚠️ Router Service切换到降级模式（热切换，无需重启）")
        } else {
            log.Printf("✅ Router Service切换到正常模式（热切换，无需重启）")
        }
    }
    
    rc.lastCheck = time.Now()
}

// startPeriodicCheck 定期检查（热加载核心）
func (rc *RouterClient) startPeriodicCheck() {
    ticker := time.NewTicker(rc.checkInterval)
    
    for range ticker.C {
        rc.checkServiceMode()
    }
}

// GetAllRoutes 获取路由（自动选择模式）
func (rc *RouterClient) GetAllRoutes() ([]RouteConfig, error) {
    rc.mutex.RLock()
    degradedMode := rc.degradedMode
    rc.mutex.RUnlock()
    
    // 如果距离上次检查超过阈值，立即检查一次
    if time.Since(rc.lastCheck) > rc.checkInterval {
        rc.checkServiceMode()
        rc.mutex.RLock()
        degradedMode = rc.degradedMode
        rc.mutex.RUnlock()
    }
    
    if degradedMode {
        log.Printf("ℹ️ 使用默认路由配置（降级模式）")
        return rc.defaultRoutes, nil
    }
    
    // 正常模式：调用Router Service
    resp, err := rc.httpClient.Get(rc.baseURL + "/api/v1/router/routes")
    if err != nil {
        // 如果调用失败，临时切换到降级模式
        log.Printf("⚠️ 调用Router Service失败，临时使用默认路由: %v", err)
        rc.mutex.Lock()
        rc.degradedMode = true
        rc.mutex.Unlock()
        return rc.defaultRoutes, nil
    }
    defer resp.Body.Close()
    
    // 解析响应...
    var routerResp RouterResponse
    if err := json.NewDecoder(resp.Body).Decode(&routerResp); err != nil {
        return rc.defaultRoutes, nil
    }
    
    // 转换数据
    routesJSON, _ := json.Marshal(routerResp.Data)
    var routes []RouteConfig
    if err := json.Unmarshal(routesJSON, &routes); err != nil {
        return rc.defaultRoutes, nil
    }
    
    return routes, nil
}
```

---

### 3. 健康检查端点（支持手动触发）

```go
// services/infrastructure/router/main.go

// 添加手动触发检查的端点
r.POST("/api/v1/router/reload", func(c *gin.Context) {
    routerService.checkAndSwitchMode()
    
    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "message": "路由配置已重新加载",
        "mode": routerService.GetMode(),
    })
})

// 获取当前状态
r.GET("/api/v1/router/status", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "data": gin.H{
            "mode": routerService.GetMode(),
            "routes_count": len(routerService.GetRoutes()),
            "auto_recovery": true,
            "check_interval": 30,
        },
    })
})
```

---

## 📊 热加载效果演示

### 场景1: 开发环境联调

```
时间轴:
00:00 - Router Service启动（数据库无数据）
       → 自动进入降级模式
       → Central Brain检测到降级，使用默认路由

00:30 - 开发人员导入路由配置数据
       → Router Service后台检测到数据就绪
       → 自动切换到正常模式（无需重启）
       → Central Brain检测到模式切换（无需重启）
       → 自动开始使用动态路由

01:00 - 开发人员测试路由功能
       → 使用动态路由配置
       → 功能正常
```

**关键点**: ✅ 全程无需重启Central Brain！

---

### 场景2: 测试环境数据恢复

```
时间轴:
00:00 - Router Service正常运行（正常模式）
       → 使用动态路由

00:30 - 数据库故障（数据丢失）
       → Router Service检测到数据不可用
       → 自动切换到降级模式（无需重启）
       → Central Brain检测到模式切换（无需重启）
       → 自动使用默认路由

01:00 - 数据库恢复（数据恢复）
       → Router Service检测到数据就绪
       → 自动切换到正常模式（无需重启）
       → Central Brain检测到模式切换（无需重启）
       → 自动恢复使用动态路由
```

**关键点**: ✅ 自动恢复，无需人工干预！

---

## 🔧 配置参数

### 环境变量配置

```bash
# Router Service热加载配置
ROUTER_SERVICE_AUTO_RECOVERY=true        # 是否启用自动恢复
ROUTER_SERVICE_CHECK_INTERVAL=30         # 检查间隔（秒）
ROUTER_SERVICE_DEGRADED_MODE=true        # 是否允许降级模式

# Central Brain热检测配置
CENTRAL_BRAIN_ROUTER_CHECK_INTERVAL=30   # Router Service检查间隔（秒）
CENTRAL_BRAIN_AUTO_SWITCH=true           # 是否自动切换模式
```

---

## ✅ 优势总结

### 1. 无需重启 ✅

- ✅ Router Service模式切换：无需重启
- ✅ Central Brain模式切换：无需重启
- ✅ 数据就绪后自动恢复：无需重启

---

### 2. 开发友好 ✅

- ✅ 开发环境可以逐步导入数据
- ✅ 数据导入后立即生效（无需重启）
- ✅ 联调联试更方便

---

### 3. 运维友好 ✅

- ✅ 数据库故障时自动降级
- ✅ 数据库恢复后自动恢复
- ✅ 无需人工干预

---

### 4. 测试友好 ✅

- ✅ 测试环境可以模拟各种场景
- ✅ 数据注入后立即生效
- ✅ 无需重启服务

---

## 🎯 核心价值

### 理解确认

**您的理解完全正确！** ✅

**降级机制 = 微服务热加载**

- ✅ **无需重启Central Brain** - 自动检测并切换
- ✅ **数据就绪后自动恢复** - 无缝切换到正常模式
- ✅ **支持联调联试** - 开发过程中数据逐步完善，服务自动适应
- ✅ **自动适应数据状态** - 数据库故障时降级，恢复后自动恢复

---

## 📋 实施优先级

| 优先级 | 任务 | 预计时间 | 热加载效果 |
|--------|------|----------|-----------|
| 🔥 最高 | Router Service热切换实现 | 2小时 | ✅ 无需重启 |
| 🔥 高 | Central Brain热检测实现 | 1.5小时 | ✅ 无需重启 |
| ⭐ 中 | Permission Service热切换 | 2小时 | ✅ 无需重启 |
| ⭐ 低 | 手动触发检查端点 | 30分钟 | ✅ 立即生效 |

---

**报告生成时间**: 2025-01-29  
**核心理念**: **降级机制 = 微服务热加载 = 无需重启 = 自动适应**

