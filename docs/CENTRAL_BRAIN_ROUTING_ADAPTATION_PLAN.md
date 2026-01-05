# Central Brain路由适配与前端协同方案

## 📋 方案概述

**目标**: 
1. 在Central Brain中注册所有基础设施服务路由
2. 支持小程序端"多对一"路由适配（一个页面调用多个服务，统一入口）
3. 支持Web端"多对多"路由适配（多个页面调用多个服务）
4. 提供灵活的路径映射和聚合接口支持

**设计原则**:
- ✅ 统一网关入口：所有前端请求都通过Central Brain
- ✅ 路径透明：前端无需知道后端服务端口
- ✅ 聚合支持：支持一个请求聚合多个服务数据
- ✅ 灵活映射：支持路径映射和重写

---

## 🎯 路由架构设计

### 1. 统一路由注册方案

#### 1.1 服务分类

**业务服务**（已注册）:
- `/api/v1/auth` → Auth Service (8207)
- `/api/v1/user` → User Service (8082)
- `/api/v1/job` → Job Service (8084)
- `/api/v1/resume` → Resume Service (8085)
- `/api/v1/company` → Company Service (8083)

**高级服务**（已注册）:
- `/api/v1/ai` → AI Service (8100)
- `/api/v1/blockchain` → Blockchain Service (8208)

**基础设施服务**（需要注册）:
- `/api/v1/notification` → Notification Service (8605)
- `/api/v1/banner` → Banner Service (8612)
- `/api/v1/template` → Template Service (8611)
- `/api/v1/statistics` → Statistics Service (8606)

**未来服务**（预留）:
- `/api/v1/resource` → Resource Service (待实现)
- `/api/v1/dict` → Dict Service (待实现)

---

### 2. 多对一路由适配方案（小程序端）

#### 2.1 场景分析

**小程序首页场景**:
```typescript
// 首页需要聚合多个服务的数据
Page({
  onLoad() {
    // 需要调用多个服务
    Promise.all([
      ApiService.getBanners(),           // /api/v1/banner/banners
      ApiService.getNotifications(),     // /api/v1/notification/notifications
      ApiService.getUserInfo(),          // /api/v1/user/info
      ApiService.getJobRecommendations() // /api/v1/job/recommendations
    ]).then(([banners, notifications, userInfo, jobs]) => {
      // 更新页面数据
    })
  }
})
```

**问题**: 
- 小程序需要发起4个HTTP请求
- 每个请求都需要经过Central Brain代理
- 网络延迟叠加

**解决方案**: ⭐ **聚合接口**

---

#### 2.2 聚合接口设计

**方案A: 统一聚合接口**（推荐）

```go
// Central Brain提供聚合接口
GET /api/v1/aggregate/home
// 返回首页所需的所有数据
{
  "banners": [...],
  "notifications": [...],
  "userInfo": {...},
  "jobRecommendations": [...]
}
```

**方案B: 小程序专用路径映射**

```go
// 映射小程序常用路径
GET /api/v1/personal/home/banners      → /api/v1/banner/banners
GET /api/v1/personal/home/notifications → /api/v1/notification/notifications
GET /api/v1/personal/mine/info         → /api/v1/user/info
```

**推荐**: ✅ **方案A + 方案B结合**
- 提供聚合接口减少请求次数
- 提供路径映射兼容实际项目API路径

---

### 3. 多对多路由适配方案（Web端）

#### 3.1 场景分析

**Web端不同页面场景**:
```typescript
// 职位列表页
GET /api/v1/job/list
GET /api/v1/job/filters

// 职位详情页
GET /api/v1/job/:id
GET /api/v1/company/:id
GET /api/v1/resume/recommendations/:jobId

// 用户中心页
GET /api/v1/user/info
GET /api/v1/resume/list
GET /api/v1/notification/notifications
```

**解决方案**: ✅ **保持现有路由设计**
- Web端每个页面独立调用所需服务
- Central Brain负责路由代理和负载均衡
- 无需特殊处理

---

## 🔧 实施方案

### Phase 1: 扩展配置结构

#### 1.1 更新Config结构

```go
// shared/core/shared/config.go
type Config struct {
    // ... 现有配置 ...
    
    // 基础设施服务端口
    NotificationServicePort int
    BannerServicePort       int
    TemplateServicePort     int
    StatisticsServicePort   int
    
    // 路由配置
    Routing struct {
        EnableAggregation bool                    // 是否启用聚合接口
        PathMappings      map[string]string      // 路径映射（小程序兼容）
        AggregationRoutes map[string][]string    // 聚合路由配置
    }
}
```

#### 1.2 更新配置加载

```go
func GetDefaultConfig() *Config {
    config := &Config{
        // ... 现有配置 ...
        
        NotificationServicePort: getEnvInt("NOTIFICATION_SERVICE_PORT", 8605),
        BannerServicePort:       getEnvInt("BANNER_SERVICE_PORT", 8612),
        TemplateServicePort:     getEnvInt("TEMPLATE_SERVICE_PORT", 8611),
        StatisticsServicePort:   getEnvInt("STATISTICS_SERVICE_PORT", 8606),
    }
    
    // 路由配置
    config.Routing.EnableAggregation = getEnvBool("ROUTING_ENABLE_AGGREGATION", true)
    
    // 路径映射（小程序兼容）
    config.Routing.PathMappings = map[string]string{
        "/api/v1/personal/home/banners":       "/api/v1/banner/banners",
        "/api/v1/personal/home/notifications": "/api/v1/notification/notifications",
        "/api/v1/personal/mine/info":           "/api/v1/user/info",
        "/api/v1/personal/resume/templates":    "/api/v1/template/public/templates",
    }
    
    // 聚合路由配置
    config.Routing.AggregationRoutes = map[string][]string{
        "/api/v1/aggregate/home": {
            "/api/v1/banner/banners",
            "/api/v1/notification/notifications?pageNum=1&pageSize=10",
            "/api/v1/user/info",
        },
        "/api/v1/aggregate/user-center": {
            "/api/v1/user/info",
            "/api/v1/resume/list/summary",
            "/api/v1/notification/notifications?pageNum=1&pageSize=10",
        },
    }
    
    return config
}
```

---

### Phase 2: 扩展Central Brain路由注册

#### 2.1 注册基础设施服务

```go
// shared/central-brain/centralbrain.go

func (cb *CentralBrain) registerServiceProxies() {
    serviceHost := cb.config.ServiceDiscovery.ServiceHost
    
    services := map[string]ServiceProxy{
        // ... 现有业务服务 ...
        
        // 基础设施服务
        "notification": {
            ServiceName: "notification-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", serviceHost, cb.config.NotificationServicePort),
            PathPrefix:  "/api/v1/notification",
        },
        "banner": {
            ServiceName: "banner-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", serviceHost, cb.config.BannerServicePort),
            PathPrefix:  "/api/v1/banner",
        },
        "template": {
            ServiceName: "template-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", serviceHost, cb.config.TemplateServicePort),
            PathPrefix:  "/api/v1/template",
        },
        "statistics": {
            ServiceName: "statistics-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", serviceHost, cb.config.StatisticsServicePort),
            PathPrefix:  "/api/v1/statistics",
        },
    }
    
    // 注册服务代理
    for serviceKey, service := range services {
        cb.registerServiceProxy(serviceKey, service)
    }
    
    // 注册路径映射（小程序兼容）
    cb.registerPathMappings()
    
    // 注册聚合接口（如果启用）
    if cb.config.Routing.EnableAggregation {
        cb.registerAggregationRoutes()
    }
}
```

---

#### 2.2 路径映射实现

```go
// registerPathMappings 注册路径映射（小程序兼容）
func (cb *CentralBrain) registerPathMappings() {
    for sourcePath, targetPath := range cb.config.Routing.PathMappings {
        // 解析目标路径，确定服务
        targetService := cb.resolveServiceFromPath(targetPath)
        if targetService == nil {
            continue
        }
        
        // 注册映射路由
        cb.router.Any(sourcePath, func(c *gin.Context) {
            // 重写路径
            c.Request.URL.Path = targetPath
            // 代理到目标服务
            cb.proxyRequest(c, *targetService)
        })
        
        fmt.Printf("✅ 注册路径映射: %s -> %s\n", sourcePath, targetPath)
    }
}

// resolveServiceFromPath 从路径解析服务
func (cb *CentralBrain) resolveServiceFromPath(path string) *ServiceProxy {
    // 匹配PathPrefix
    if strings.HasPrefix(path, "/api/v1/banner") {
        return &ServiceProxy{
            ServiceName: "banner-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", cb.config.ServiceDiscovery.ServiceHost, cb.config.BannerServicePort),
            PathPrefix:  "/api/v1/banner",
        }
    }
    if strings.HasPrefix(path, "/api/v1/notification") {
        return &ServiceProxy{
            ServiceName: "notification-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", cb.config.ServiceDiscovery.ServiceHost, cb.config.NotificationServicePort),
            PathPrefix:  "/api/v1/notification",
        }
    }
    // ... 其他服务映射
    return nil
}
```

---

#### 2.3 聚合接口实现

```go
// registerAggregationRoutes 注册聚合接口
func (cb *CentralBrain) registerAggregationRoutes() {
    for aggregatePath, targetPaths := range cb.config.Routing.AggregationRoutes {
        cb.router.GET(aggregatePath, func(c *gin.Context) {
            cb.handleAggregation(c, targetPaths)
        })
        
        fmt.Printf("✅ 注册聚合接口: %s (聚合 %d 个服务)\n", aggregatePath, len(targetPaths))
    }
}

// handleAggregation 处理聚合请求
func (cb *CentralBrain) handleAggregation(c *gin.Context, targetPaths []string) {
    // 使用goroutine并发调用多个服务
    type result struct {
        path     string
        data     interface{}
        err      error
        statusCode int
    }
    
    results := make(chan result, len(targetPaths))
    
    // 并发请求
    for _, targetPath := range targetPaths {
        go func(path string) {
            data, statusCode, err := cb.fetchServiceData(path, c.Request)
            results <- result{
                path:       path,
                data:       data,
                err:        err,
                statusCode: statusCode,
            }
        }(targetPath)
    }
    
    // 收集结果
    aggregatedData := make(map[string]interface{})
    hasError := false
    
    for i := 0; i < len(targetPaths); i++ {
        res := <-results
        if res.err != nil || res.statusCode != http.StatusOK {
            hasError = true
            aggregatedData[res.path] = map[string]interface{}{
                "error": res.err.Error(),
            }
        } else {
            // 提取data字段（如果存在）
            if dataMap, ok := res.data.(map[string]interface{}); ok {
                if data, exists := dataMap["data"]; exists {
                    aggregatedData[res.path] = data
                } else {
                    aggregatedData[res.path] = res.data
                }
            } else {
                aggregatedData[res.path] = res.data
            }
        }
    }
    
    // 返回聚合结果
    statusCode := http.StatusOK
    if hasError {
        statusCode = http.StatusPartialContent // 206
    }
    
    c.JSON(statusCode, gin.H{
        "code": 0,
        "message": "聚合数据获取成功",
        "data": aggregatedData,
        "timestamp": time.Now().Unix(),
    })
}

// fetchServiceData 从服务获取数据
func (cb *CentralBrain) fetchServiceData(path string, originalReq *http.Request) (interface{}, int, error) {
    // 解析目标服务
    targetService := cb.resolveServiceFromPath(path)
    if targetService == nil {
        return nil, http.StatusBadRequest, fmt.Errorf("无法解析服务路径: %s", path)
    }
    
    // 构建目标URL
    targetPath := strings.TrimPrefix(path, targetService.PathPrefix)
    if targetPath == "" {
        targetPath = "/"
    }
    
    // 解析查询参数
    if idx := strings.Index(path, "?"); idx != -1 {
        targetPath = path[idx:]
    }
    
    targetURL := targetService.BaseURL + targetPath
    
    // 创建请求
    req, err := http.NewRequest("GET", targetURL, nil)
    if err != nil {
        return nil, http.StatusInternalServerError, err
    }
    
    // 复制认证头
    userToken := cb.extractUserToken(originalReq)
    if userToken != "" {
        req.Header.Set("Authorization", "Bearer "+userToken)
    }
    
    // 添加服务token
    serviceToken := cb.getServiceToken()
    if serviceToken != "" {
        req.Header.Set("X-Service-Token", serviceToken)
        req.Header.Set("X-Service-ID", "central-brain")
    }
    
    // 发送请求
    resp, err := cb.httpClient.Do(req)
    if err != nil {
        return nil, http.StatusInternalServerError, err
    }
    defer resp.Body.Close()
    
    // 读取响应
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, resp.StatusCode, err
    }
    
    // 解析JSON
    var data interface{}
    if err := json.Unmarshal(respBody, &data); err != nil {
        return nil, resp.StatusCode, err
    }
    
    return data, resp.StatusCode, nil
}
```

---

### Phase 3: 前端适配

#### 3.1 小程序端API服务增强

```typescript
// frontend/src/services/api.ts

// 聚合接口
export class ApiService {
    // ... 现有方法 ...
    
    // 首页聚合数据（小程序优化）
    static async getHomeData() {
        return request({
            url: '/api/v1/aggregate/home',
            method: 'GET'
        })
    }
    
    // 用户中心聚合数据
    static async getUserCenterData() {
        return request({
            url: '/api/v1/aggregate/user-center',
            method: 'GET'
        })
    }
    
    // 兼容实际项目路径（小程序）
    static async getBanners() {
        return request({
            url: '/api/v1/personal/home/banners',
            method: 'GET'
        })
    }
    
    static async getNotifications(params: any) {
        return request({
            url: '/api/v1/personal/home/notifications',
            method: 'GET',
            data: params
        })
    }
    
    static async getPersonalInfo() {
        return request({
            url: '/api/v1/personal/mine/info',
            method: 'GET'
        })
    }
    
    static async getResumeTemplates() {
        return request({
            url: '/api/v1/personal/resume/templates',
            method: 'GET'
        })
    }
}
```

---

#### 3.2 小程序首页使用示例

```typescript
// frontend/src/pages/index/index.tsx

import { ApiService } from '@/services/api'

Page({
  data: {
    banners: [],
    notifications: [],
    userInfo: null,
    jobs: []
  },
  
  async onLoad() {
    // 方案1: 使用聚合接口（推荐，减少请求次数）
    try {
      const homeData = await ApiService.getHomeData()
      this.setData({
        banners: homeData.data['/api/v1/banner/banners'] || [],
        notifications: homeData.data['/api/v1/notification/notifications'] || [],
        userInfo: homeData.data['/api/v1/user/info'] || null
      })
    } catch (error) {
      console.error('获取首页数据失败:', error)
    }
    
    // 方案2: 使用兼容路径（如果聚合接口不可用）
    // Promise.all([
    //   ApiService.getBanners(),
    //   ApiService.getNotifications({ pageNum: 1, pageSize: 10 }),
    //   ApiService.getPersonalInfo()
    // ]).then(([banners, notifications, userInfo]) => {
    //   this.setData({ banners, notifications, userInfo })
    // })
  }
})
```

---

#### 3.3 Web端保持现有方式

```typescript
// Web端可以继续使用独立API调用
// 因为Web端可以并发处理多个请求，不需要聚合

// 职位列表页
const [jobs, filters] = await Promise.all([
  ApiService.getJobList({ page: 1 }),
  ApiService.getJobFilters()
])

// 职位详情页
const [jobDetail, companyInfo, recommendations] = await Promise.all([
  ApiService.getJobDetail(jobId),
  ApiService.getCompanyDetail(companyId),
  ApiService.getResumeRecommendations(jobId)
])
```

---

## 📊 路由映射表

### 小程序兼容路径映射

| 小程序路径 | 实际服务路径 | 服务 |
|-----------|------------|------|
| `/api/v1/personal/home/banners` | `/api/v1/banner/banners` | Banner Service |
| `/api/v1/personal/home/notifications` | `/api/v1/notification/notifications` | Notification Service |
| `/api/v1/personal/mine/info` | `/api/v1/user/info` | User Service |
| `/api/v1/personal/mine/points` | `/api/v1/user/points` | User Service |
| `/api/v1/personal/resume/templates` | `/api/v1/template/public/templates` | Template Service |
| `/api/v1/personal/resume/list/summary` | `/api/v1/resume/list/summary` | Resume Service |

### 聚合接口

| 聚合路径 | 聚合的服务 | 用途 |
|---------|-----------|------|
| `/api/v1/aggregate/home` | banner, notification, user | 小程序首页 |
| `/api/v1/aggregate/user-center` | user, resume, notification | 用户中心页 |

---

## 🔧 配置文件更新

### configs/local.env

```bash
# 基础设施服务端口配置
NOTIFICATION_SERVICE_PORT=8605
BANNER_SERVICE_PORT=8612
TEMPLATE_SERVICE_PORT=8611
STATISTICS_SERVICE_PORT=8606

# 路由配置
ROUTING_ENABLE_AGGREGATION=true
```

---

## 📋 实施计划

### Step 1: 配置扩展（1小时）

1. ✅ 更新`shared/core/shared/config.go`，添加基础设施服务端口配置
2. ✅ 更新`configs/local.env`，添加端口配置
3. ✅ 测试配置加载

### Step 2: 路由注册（2小时）

1. ✅ 更新`shared/central-brain/centralbrain.go`，注册基础设施服务路由
2. ✅ 测试路由代理功能
3. ✅ 验证服务间通信

### Step 3: 路径映射（2小时）

1. ✅ 实现路径映射功能
2. ✅ 配置小程序兼容路径映射
3. ✅ 测试路径映射

### Step 4: 聚合接口（3小时）

1. ✅ 实现聚合接口功能
2. ✅ 配置聚合路由
3. ✅ 测试聚合接口性能和错误处理

### Step 5: 前端适配（2小时）

1. ✅ 更新前端API服务，添加聚合接口和兼容路径
2. ✅ 更新小程序首页，使用聚合接口
3. ✅ 测试前端功能

### Step 6: 测试验证（2小时）

1. ✅ 端到端测试（小程序首页）
2. ✅ 端到端测试（Web端）
3. ✅ 性能测试（聚合接口 vs 独立请求）
4. ✅ 错误处理测试

**总计**: 约12小时（1.5个工作日）

---

## ✅ 验收标准

### 功能验收

- [ ] 所有基础设施服务可以通过Central Brain访问
- [ ] 小程序兼容路径映射正常工作
- [ ] 聚合接口正常工作
- [ ] 错误处理正确（部分服务失败时返回206状态码）
- [ ] 认证token正确传递

### 性能验收

- [ ] 聚合接口性能优于独立请求（减少请求次数）
- [ ] 路由代理延迟 < 10ms
- [ ] 并发处理能力正常

### 兼容性验收

- [ ] 小程序端可以正常使用聚合接口和兼容路径
- [ ] Web端可以正常使用独立API调用
- [ ] 向后兼容（现有API调用不受影响）

---

## 🎯 方案优势

### 1. 统一入口 ✅
- 所有前端请求都通过Central Brain
- 前端无需知道后端服务端口

### 2. 小程序优化 ✅
- 聚合接口减少请求次数
- 路径映射兼容实际项目API

### 3. Web端灵活 ✅
- 保持现有独立API调用方式
- 支持并发请求，性能不受影响

### 4. 可扩展性 ✅
- 易于添加新的服务路由
- 易于添加新的聚合接口
- 易于添加新的路径映射

### 5. 向后兼容 ✅
- 不影响现有API调用
- 渐进式迁移

---

**方案生成时间**: 2025-01-29  
**预计实施时间**: 1.5个工作日  
**优先级**: 🔥🔥🔥 **最高**

