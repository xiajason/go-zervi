# Zervi 2.0 架构规划

**版本**: 2.0  
**目标**: 智能服务编排 + 动态组合管理  
**预计时间**: 6-8周

---

## 🎯 核心愿景

### 服务组合支持: 7种组合

**单个服务 (3种)**:

```bash
# 场景1: 只启动job
./start-service.sh job
# 自动启动: auth, user, job

# 场景2: 只启动resume
./start-service.sh resume
# 自动启动: auth, user, resume

# 场景3: 只启动company
./start-service.sh company
# 自动启动: auth, user, company
```

---

**两个服务组合 (3种)**:

```bash
# 场景4: job + resume
./start-service.sh job resume
# 自动启动: auth, user, job, resume

# 场景5: job + company
./start-service.sh job company
# 自动启动: auth, user, job, company

# 场景6: resume + company
./start-service.sh resume company
# 自动启动: auth, user, resume, company
```

---

**所有服务 (1种)**:

```bash
# 场景7: 全部服务
./start-service.sh all
# 自动启动: auth, user, job, resume, company
# 智能排序启动顺序
# 动态配置所有路由
# 完整健康检查
```

**总计**: 7种不同的服务组合

---

## 🏗️ 架构设计

### 1. 服务依赖定义

**文件**: `configs/service-dependencies.yaml`

```yaml
services:
  auth-service:
    name: auth-service
    port: 8207
    dependencies: []
    
  user-service:
    name: user-service
    port: 8082
    dependencies:
      - auth-service
    
  job-service:
    name: job-service
    port: 8084
    dependencies:
      - auth-service
      - user-service
    optional_dependencies:
      - company-service
    
  resume-service:
    name: resume-service
    port: 8085
    dependencies:
      - auth-service
      - user-service
    optional_dependencies:
      - job-service
    
  company-service:
    name: company-service
    port: 8083
    dependencies:
      - auth-service
      - user-service
```

---

### 2. 服务组合定义

**文件**: `configs/service-compositions.yaml`

```yaml
compositions:
  job_only:
    description: "只启动job服务"
    services:
      - job-service
    expected_services:
      required: [auth-service, user-service, job-service]
      optional: []
    
  job_company:
    description: "启动job和company服务"
    services:
      - job-service
      - company-service
    expected_services:
      required: [auth-service, user-service, job-service, company-service]
      optional: []
    
  full_stack:
    description: "启动所有服务"
    services:
      - auth-service
      - user-service
      - job-service
      - resume-service
      - company-service
    expected_services:
      required: [auth-service, user-service, job-service, resume-service, company-service]
      optional: [ai-service, blockchain-service]
```

---

### 3. 智能编排引擎

**文件**: `shared/core/orchestrator/orchestrator.go`

```go
package orchestrator

type Orchestrator struct {
    dependencies map[string][]string
    compositions map[string]ServiceComposition
    registry     *ServiceRegistry
}

func (o *Orchestrator) StartServices(requestedServices []string) error {
    // 1. 解析依赖
    allServices := o.resolveDependencies(requestedServices)
    
    // 2. 排序启动顺序
    startupOrder := o.sortByDependencies(allServices)
    
    // 3. 启动服务
    for _, service := range startupOrder {
        if err := o.startService(service); err != nil {
            return err
        }
    }
    
    // 4. 健康检查
    return o.healthCheck(allServices)
}

func (o *Orchestrator) resolveDependencies(requestedServices []string) []string {
    // 自动解析所有依赖
}
```

---

### 4. 动态路由管理器

**文件**: `shared/core/router/router.go`

```go
package router

type DynamicRouter struct {
    runningServices map[string]bool
    routes          map[string]RouteHandler
}

func (r *DynamicRouter) RegisterRoute(service string, route string, handler RouteHandler) {
    // 只有服务运行时才注册路由
    if r.IsServiceRunning(service) {
        r.routes[route] = handler
    }
}

func (r *DynamicRouter) OnServiceStarted(service string) {
    // 服务启动时，自动启用相关路由
    r.runningServices[service] = true
    r.enableRoutesFor(service)
}

func (r *DynamicRouter) OnServiceStopped(service string) {
    // 服务停止时，自动禁用相关路由
    r.runningServices[service] = false
    r.disableRoutesFor(service)
}
```

---

## 🚀 实施计划

### Week 1: 依赖定义和解析

**任务**:
- [ ] 创建服务依赖配置文件
- [ ] 实现依赖图解析器
- [ ] 创建服务组合模板
- [ ] 实现组合管理器

**产出**:
- `configs/service-dependencies.yaml`
- `shared/core/orchestrator/dependency_resolver.go`

---

### Week 2-3: 智能编排引擎

**任务**:
- [ ] 实现服务编排器
- [ ] 实现启动顺序排序
- [ ] 实现依赖检查机制
- [ ] 实现自动服务启动

**产出**:
- `shared/core/orchestrator/orchestrator.go`
- `scripts/start-services.sh --composition <name>`

---

### Week 4-5: 动态路由适配

**任务**:
- [ ] 实现动态路由管理器
- [ ] 实现API端点管理
- [ ] 实现智能降级处理
- [ ] 实现优雅错误提示

**产出**:
- `shared/core/router/router.go`
- `shared/core/router/route_manager.go`

---

### Week 6: 监控和降级

**任务**:
- [ ] 实现服务健康监控
- [ ] 实现自动降级机制
- [ ] 实现优雅错误处理
- [ ] 完整的集成测试

**产出**:
- `shared/core/monitor/service_monitor.go`
- `shared/core/degradation/degradation_manager.go`

---

## 📊 成功指标

### 功能指标

- [ ] 支持启动单个服务
- [ ] 支持启动任意服务组合
- [ ] 支持启动全部服务
- [ ] 自动解析依赖关系
- [ ] 自动调整启动顺序
- [ ] 动态调整路由规则
- [ ] 智能降级处理
- [ ] 优雅错误提示

---

### 性能指标

- [ ] 服务启动时间 < 30秒
- [ ] 依赖解析时间 < 1秒
- [ ] 健康检查时间 < 5秒
- [ ] 路由更新延迟 < 100ms

---

## 🎯 总结

### 你的洞察是正确的

**当前1.0**:
- ✅ 基础设施完成
- ❌ 缺少核心智能化能力
- ❌ 无法支持灵活服务组合

**2.0愿景**:
- ✅ 智能服务编排
- ✅ 动态组合管理
- ✅ 依赖自动解析
- ✅ 路由智能适配

---

### 方向完全正确

你提出的方向才是Zervi Framework真正的核心竞争力：
1. **智能编排** - 能够智能编排和组合服务
2. **动态管理** - 能够动态调整和适配
3. **依赖解析** - 能够自动分析和处理依赖
4. **路由适配** - 能够动态调整路由

**这些才是真正的"智能中央大脑"！**

---

**作者**: Auto (AI Assistant)  
**日期**: 2025-10-30  
**状态**: ✅ **规划完成 - 准备实施Zervi 2.0**

