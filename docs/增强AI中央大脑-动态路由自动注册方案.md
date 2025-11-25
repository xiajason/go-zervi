# 🧠 增强AI中央大脑 - 动态路由自动注册方案

## 🎯 核心目标

实现**前端页面 → Consul发现 → 数据库注册 → AI中央大脑适配**的全自动化流程。

## 📋 当前问题分析

### 现状
```
❌ 手动创建前端页面 (Index.vue, Company/Index.vue...)
❌ 手动在数据库中配置菜单
❌ 前端硬编码路由路径
❌ 中央大脑被动响应
```

### 理想状态
```
✅ 开发者创建新页面 → 自动被发现
✅ 页面元信息 → 自动注册到 Consul
✅ Consul → 通知中央大脑
✅ 中央大脑 → 自动写入数据库
✅ 前端 → 自动获取最新路由
```

## 🏗️ 架构设计

### 完整流程图

```
┌─────────────────────────────────────────────────────────────┐
│                    Step 1: 开发者创建页面                     │
├─────────────────────────────────────────────────────────────┤
│  开发者:                                                      │
│    创建 src/views/template/MyNewPage.vue                     │
│    添加页面元信息：                                           │
│    <!-- @route-meta                                          │
│       title: "我的新页面"                                     │
│       icon: "Document"                                       │
│       permission: "mypage:view"                              │
│    -->                                                       │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│               Step 2: 前端构建时自动扫描                      │
├─────────────────────────────────────────────────────────────┤
│  Vite Plugin (route-scanner):                               │
│    扫描所有 .vue 文件                                        │
│    提取 @route-meta 元信息                                   │
│    生成 route-manifest.json                                  │
│    {                                                         │
│      "routes": [                                             │
│        {                                                     │
│          "path": "/mynewpage",                               │
│          "component": "template/MyNewPage",                  │
│          "meta": {...}                                       │
│        }                                                     │
│      ]                                                       │
│    }                                                         │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│            Step 3: 注册到 Consul (服务发现)                  │
├─────────────────────────────────────────────────────────────┤
│  Frontend Service Registration:                             │
│    服务名: "vuecmf-frontend"                                 │
│    端口: 8081                                                │
│    标签: ["frontend", "vuecmf"]                             │
│    元数据:                                                   │
│      - routes: route-manifest.json 的内容                   │
│      - version: "1.0.0"                                      │
│      - build_time: "2025-11-06T16:30:00"                    │
│                                                              │
│  Consul:                                                     │
│    POST /v1/agent/service/register                          │
│    {                                                         │
│      "ID": "vuecmf-frontend-8081",                          │
│      "Name": "vuecmf-frontend",                             │
│      "Tags": ["frontend"],                                  │
│      "Meta": {                                               │
│        "routes": "..."                                       │
│      }                                                       │
│    }                                                         │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│         Step 4: AI中央大脑监听 Consul 变化                   │
├─────────────────────────────────────────────────────────────┤
│  Central Brain - Consul Watcher:                            │
│    监听服务: "vuecmf-frontend"                              │
│    事件类型: ServiceRegistered, ServiceDeregistered         │
│                                                              │
│  触发器:                                                     │
│    新路由检测 → 调用 RouteAnalyzer                          │
│                                                              │
│  RouteAnalyzer (AI增强):                                    │
│    分析新路由的:                                             │
│      - 功能类型 (CRUD / Dashboard / Report)                 │
│      - 所需权限                                              │
│      - 依赖的后端API                                         │
│      - 推荐的菜单位置                                        │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│           Step 5: 自动生成并写入数据库配置                   │
├─────────────────────────────────────────────────────────────┤
│  Central Brain - Route Registrar:                           │
│                                                              │
│  1. 检查路由是否已存在                                       │
│     SELECT * FROM menu WHERE path = '/mynewpage'            │
│                                                              │
│  2. 如果不存在，生成菜单配置：                               │
│     INSERT INTO menu (                                       │
│       title, path, component_tpl,                            │
│       icon, permission, pid, sort_num                        │
│     ) VALUES (                                               │
│       '我的新页面',                                          │
│       '/mynewpage',                                          │
│       'template/MyNewPage',                                  │
│       'Document',                                            │
│       'mypage:view',                                         │
│       2,  -- 父菜单ID (AI推荐)                              │
│       10  -- 排序 (AI推荐)                                  │
│     )                                                        │
│                                                              │
│  3. 生成API映射配置：                                        │
│     INSERT INTO api_map (                                    │
│       table_name, action_type, api_path                      │
│     ) VALUES (                                               │
│       'mynewpage', 'list', '/api/v1/mynewpage/list'         │
│     )                                                        │
│                                                              │
│  4. 生成模型配置（如需要）：                                 │
│     INSERT INTO model_config (...)                          │
│     INSERT INTO model_field (...)                           │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│              Step 6: 通知前端更新路由                        │
├─────────────────────────────────────────────────────────────┤
│  Central Brain:                                              │
│    发送 WebSocket 消息或 SSE 事件                           │
│    事件: "routes-updated"                                    │
│    数据: { "new_routes": [...], "timestamp": ... }          │
│                                                              │
│  Frontend:                                                   │
│    监听 "routes-updated" 事件                                │
│    重新调用 loadMenu()                                       │
│    动态注册新路由                                            │
│    更新菜单显示                                              │
│                                                              │
│  用户体验:                                                   │
│    无需刷新页面，新菜单自动出现 ✨                          │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 实现方案

### 1. 前端 - Vite 插件（自动扫描）

```typescript
// vite-plugin-route-scanner.ts
import { Plugin } from 'vite'
import { parse } from '@vue/compiler-sfc'
import fs from 'fs'
import path from 'path'
import glob from 'glob'

interface RouteMetadata {
  path: string
  component: string
  title?: string
  icon?: string
  permission?: string
  description?: string
}

export function routeScannerPlugin(): Plugin {
  return {
    name: 'vite-plugin-route-scanner',
    
    buildStart() {
      console.log('🔍 扫描路由定义...')
      
      // 扫描所有 Vue 文件
      const vueFiles = glob.sync('src/views/template/**/*.vue')
      const routes: RouteMetadata[] = []
      
      for (const file of vueFiles) {
        const content = fs.readFileSync(file, 'utf-8')
        const route = this.extractRouteMetadata(file, content)
        if (route) {
          routes.push(route)
        }
      }
      
      // 生成 route-manifest.json
      const manifest = {
        version: '1.0.0',
        generated_at: new Date().toISOString(),
        routes: routes
      }
      
      fs.writeFileSync(
        'dist/route-manifest.json',
        JSON.stringify(manifest, null, 2)
      )
      
      console.log(`✅ 发现 ${routes.length} 个路由`)
    },
    
    extractRouteMetadata(filePath: string, content: string): RouteMetadata | null {
      // 提取文件中的 @route-meta 注释
      const metaRegex = /<!--\s*@route-meta\s*([\s\S]*?)\s*-->/
      const match = content.match(metaRegex)
      
      if (!match) return null
      
      // 解析 YAML 格式的元信息
      const metaContent = match[1]
      const meta: any = {}
      
      metaContent.split('\n').forEach(line => {
        const [key, ...valueParts] = line.trim().split(':')
        if (key && valueParts.length > 0) {
          meta[key.trim()] = valueParts.join(':').trim().replace(/["']/g, '')
        }
      })
      
      // 生成路由路径
      const relativePath = path.relative('src/views/template', filePath)
      const routePath = '/' + relativePath
        .replace(/\.vue$/, '')
        .replace(/\/Index$/, '')
        .toLowerCase()
      
      return {
        path: routePath,
        component: relativePath.replace(/\.vue$/, ''),
        ...meta
      }
    }
  }
}
```

### 2. 前端 - Consul 注册服务

```typescript
// consul-register.ts
import Consul from 'consul'
import fs from 'fs'

const consul = new Consul({
  host: 'localhost',
  port: '8500',
  promisify: true
})

async function registerFrontendService() {
  // 读取路由清单
  const manifest = JSON.parse(
    fs.readFileSync('dist/route-manifest.json', 'utf-8')
  )
  
  const serviceDefinition = {
    ID: 'vuecmf-frontend-8081',
    Name: 'vuecmf-frontend',
    Tags: ['frontend', 'vuecmf', 'web'],
    Port: 8081,
    Address: 'localhost',
    Meta: {
      routes: JSON.stringify(manifest.routes),
      version: manifest.version,
      generated_at: manifest.generated_at
    },
    Check: {
      HTTP: 'http://localhost:8081/health',
      Interval: '10s',
      Timeout: '5s'
    }
  }
  
  await consul.agent.service.register(serviceDefinition)
  console.log('✅ 前端服务已注册到 Consul')
}

// 在前端启动时调用
registerFrontendService()
```

### 3. 后端 - Consul 监听器

```go
// central-brain/consul_watcher.go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/hashicorp/consul/api"
)

type ConsulWatcher struct {
    client *api.Client
    brain  *CentralBrain
}

func NewConsulWatcher(consulURL string, brain *CentralBrain) *ConsulWatcher {
    config := api.DefaultConfig()
    config.Address = consulURL
    
    client, err := api.NewClient(config)
    if err != nil {
        log.Fatalf("创建 Consul 客户端失败: %v", err)
    }
    
    return &ConsulWatcher{
        client: client,
        brain:  brain,
    }
}

// WatchFrontendServices 监听前端服务变化
func (w *ConsulWatcher) WatchFrontendServices() {
    lastIndex := uint64(0)
    
    for {
        // 监听 vuecmf-frontend 服务
        queryOptions := &api.QueryOptions{
            WaitIndex: lastIndex,
            WaitTime:  5 * time.Minute,
        }
        
        services, meta, err := w.client.Catalog().Service(
            "vuecmf-frontend",
            "",
            queryOptions,
        )
        
        if err != nil {
            log.Printf("❌ Consul 查询失败: %v", err)
            time.Sleep(10 * time.Second)
            continue
        }
        
        // 检测变化
        if meta.LastIndex != lastIndex {
            log.Printf("🔔 检测到前端服务变化 (index: %d -> %d)", 
                lastIndex, meta.LastIndex)
            
            w.processServiceUpdate(services)
            lastIndex = meta.LastIndex
        }
    }
}

// processServiceUpdate 处理服务更新
func (w *ConsulWatcher) processServiceUpdate(services []*api.CatalogService) {
    for _, service := range services {
        // 提取路由信息
        routesJSON := service.ServiceMeta["routes"]
        if routesJSON == "" {
            continue
        }
        
        var routes []RouteMetadata
        if err := json.Unmarshal([]byte(routesJSON), &routes); err != nil {
            log.Printf("❌ 解析路由元数据失败: %v", err)
            continue
        }
        
        log.Printf("📋 发现 %d 个新路由", len(routes))
        
        // 调用 AI 中央大脑处理
        w.brain.RegisterRoutesAutomatically(routes)
    }
}
```

### 4. 后端 - AI 路由分析器

```go
// central-brain/route_analyzer.go
package main

import (
    "fmt"
    "strings"
)

type RouteMetadata struct {
    Path        string `json:"path"`
    Component   string `json:"component"`
    Title       string `json:"title"`
    Icon        string `json:"icon"`
    Permission  string `json:"permission"`
    Description string `json:"description"`
}

type RouteAnalyzer struct {
    brain *CentralBrain
}

// AnalyzeRoute AI分析路由
func (ra *RouteAnalyzer) AnalyzeRoute(route RouteMetadata) *MenuConfig {
    log.Printf("🧠 AI分析路由: %s", route.Path)
    
    config := &MenuConfig{
        Path:          route.Path,
        ComponentTpl:  route.Component,
        Title:         route.Title,
        Icon:          route.Icon,
        Permission:    route.Permission,
    }
    
    // AI 推断功能类型
    config.FunctionType = ra.inferFunctionType(route)
    
    // AI 推荐父菜单
    config.ParentID = ra.recommendParentMenu(route)
    
    // AI 推荐排序位置
    config.SortNum = ra.recommendSortOrder(route)
    
    // AI 生成 API 映射
    config.APIMapping = ra.generateAPIMapping(route)
    
    // AI 推断所需权限
    config.RequiredPermissions = ra.inferPermissions(route)
    
    return config
}

// inferFunctionType 推断功能类型
func (ra *RouteAnalyzer) inferFunctionType(route RouteMetadata) string {
    component := strings.ToLower(route.Component)
    
    if strings.Contains(component, "list") || 
       strings.Contains(component, "table") {
        return "list"  // 列表页
    }
    
    if strings.Contains(component, "form") || 
       strings.Contains(component, "edit") {
        return "form"  // 表单页
    }
    
    if strings.Contains(component, "detail") || 
       strings.Contains(component, "view") {
        return "detail"  // 详情页
    }
    
    if strings.Contains(component, "dashboard") || 
       strings.Contains(route.Path, "index") {
        return "dashboard"  // 仪表板
    }
    
    return "custom"  // 自定义页面
}

// recommendParentMenu AI推荐父菜单
func (ra *RouteAnalyzer) recommendParentMenu(route RouteMetadata) int {
    pathParts := strings.Split(route.Path, "/")
    
    if len(pathParts) > 2 {
        // 有层级结构，查找父菜单
        parentPath := "/" + pathParts[1]
        
        var parentID int
        err := ra.brain.db.QueryRow(`
            SELECT id FROM menu 
            WHERE path = $1
            LIMIT 1
        `, parentPath).Scan(&parentID)
        
        if err == nil {
            return parentID
        }
    }
    
    // 默认放在根级别
    return 0
}

// recommendSortOrder AI推荐排序
func (ra *RouteAnalyzer) recommendSortOrder(route RouteMetadata) int {
    // 查询同级菜单的最大排序号
    var maxSort int
    ra.brain.db.QueryRow(`
        SELECT COALESCE(MAX(sort_num), 0)
        FROM menu
        WHERE pid = $1
    `, route.ParentID).Scan(&maxSort)
    
    return maxSort + 10
}

// generateAPIMapping 生成API映射
func (ra *RouteAnalyzer) generateAPIMapping(route RouteMetadata) map[string]string {
    // 从路径推断表名
    pathParts := strings.Split(route.Path, "/")
    tableName := pathParts[len(pathParts)-1]
    
    mapping := map[string]string{
        "list":   fmt.Sprintf("/api/v1/%s/index", tableName),
        "detail": fmt.Sprintf("/api/v1/%s/detail", tableName),
        "save":   fmt.Sprintf("/api/v1/%s/save", tableName),
        "delete": fmt.Sprintf("/api/v1/%s/delete", tableName),
    }
    
    return mapping
}

// inferPermissions 推断所需权限
func (ra *RouteAnalyzer) inferPermissions(route RouteMetadata) []string {
    if route.Permission != "" {
        return []string{route.Permission}
    }
    
    // 从路径推断权限
    pathParts := strings.Split(route.Path, "/")
    resource := pathParts[len(pathParts)-1]
    
    return []string{
        fmt.Sprintf("%s:view", resource),
        fmt.Sprintf("%s:create", resource),
        fmt.Sprintf("%s:update", resource),
        fmt.Sprintf("%s:delete", resource),
    }
}
```

### 5. 后端 - 自动注册到数据库

```go
// central-brain/route_registrar.go
package main

import (
    "database/sql"
    "fmt"
    "log"
)

type RouteRegistrar struct {
    db       *sql.DB
    analyzer *RouteAnalyzer
}

// RegisterRoutesAutomatically 自动注册路由
func (cb *CentralBrain) RegisterRoutesAutomatically(routes []RouteMetadata) {
    registrar := &RouteRegistrar{
        db:       cb.db,
        analyzer: &RouteAnalyzer{brain: cb},
    }
    
    for _, route := range routes {
        if err := registrar.RegisterRoute(route); err != nil {
            log.Printf("❌ 注册路由失败 %s: %v", route.Path, err)
        } else {
            log.Printf("✅ 路由已注册: %s", route.Path)
        }
    }
    
    // 通知前端更新
    cb.notifyFrontendRoutesUpdated()
}

// RegisterRoute 注册单个路由
func (rr *RouteRegistrar) RegisterRoute(route RouteMetadata) error {
    // 1. 检查路由是否已存在
    var exists bool
    err := rr.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM menu WHERE path = $1
        )
    `, route.Path).Scan(&exists)
    
    if err != nil {
        return fmt.Errorf("检查路由失败: %v", err)
    }
    
    if exists {
        log.Printf("⚠️  路由已存在，跳过: %s", route.Path)
        return nil
    }
    
    // 2. AI 分析路由
    config := rr.analyzer.AnalyzeRoute(route)
    
    // 3. 插入菜单配置
    _, err = rr.db.Exec(`
        INSERT INTO menu (
            title, path, component_tpl, icon,
            pid, sort_num, status, created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, 10, NOW()
        )
    `, config.Title, config.Path, config.ComponentTpl,
       config.Icon, config.ParentID, config.SortNum)
    
    if err != nil {
        return fmt.Errorf("插入菜单失败: %v", err)
    }
    
    // 4. 插入 API 映射
    for actionType, apiPath := range config.APIMapping {
        _, err = rr.db.Exec(`
            INSERT INTO api_map (
                table_name, action_type, api_path, created_at
            ) VALUES (
                $1, $2, $3, NOW()
            )
        `, config.TableName, actionType, apiPath)
        
        if err != nil {
            log.Printf("⚠️  插入API映射失败: %v", err)
        }
    }
    
    log.Printf("🎉 路由注册成功: %s (父菜单ID: %d, 排序: %d)", 
        route.Path, config.ParentID, config.SortNum)
    
    return nil
}
```

### 6. 前端 - 实时路由更新

```typescript
// src/services/RouteUpdateService.ts
import { EventSource } from 'eventsource'

class RouteUpdateService {
  private eventSource: EventSource | null = null
  
  /**
   * 监听路由更新
   */
  startListening() {
    // 使用 Server-Sent Events (SSE)
    this.eventSource = new EventSource('http://localhost:9000/api/v1/events/routes')
    
    this.eventSource.addEventListener('routes-updated', (event) => {
      const data = JSON.parse(event.data)
      console.log('🔔 收到路由更新通知:', data)
      
      // 重新加载菜单
      this.reloadRoutes()
    })
    
    this.eventSource.onerror = (error) => {
      console.error('❌ SSE连接错误:', error)
    }
  }
  
  /**
   * 重新加载路由
   */
  async reloadRoutes() {
    const layoutService = new LayoutService()
    await layoutService.loadMenu()
    
    ElMessage.success('菜单已更新 ✨')
  }
  
  /**
   * 停止监听
   */
  stopListening() {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
  }
}

export default new RouteUpdateService()
```

## 📊 实施步骤

### 阶段 1：基础设施（本周）

1. ✅ 安装 Consul
2. ✅ 配置中央大脑连接 Consul
3. ✅ 实现 Consul Watcher

### 阶段 2：前端自动化（下周）

4. 🔄 开发 Vite 路由扫描插件
5. 🔄 实现 Consul 注册脚本
6. 🔄 添加路由元信息注释

### 阶段 3：AI分析（第3周）

7. 🔄 实现 RouteAnalyzer
8. 🔄 实现 RouteRegistrar
9. 🔄 测试自动注册流程

### 阶段 4：实时更新（第4周）

10. 🔄 实现 SSE 事件推送
11. 🔄 前端实时路由更新
12. 🔄 完整集成测试

## 🎯 预期效果

### Before（现在）
```bash
# 开发者工作流
1. 创建 MyNewPage.vue
2. 手动在数据库添加菜单记录
3. 手动配置 API 映射
4. 手动配置权限
5. 重启后端
6. 刷新前端

耗时：30-60分钟 😫
```

### After（实现后）
```bash
# 开发者工作流
1. 创建 MyNewPage.vue（添加 @route-meta 注释）
2. npm run build  # 自动完成所有注册
3. 等待 5-10 秒
4. ✨ 新菜单自动出现在前端

耗时：2-3分钟 🎉
```

## 💡 总结

这个方案实现了：

1. ✅ **完全自动化** - 开发者只需创建 Vue 文件
2. ✅ **服务发现** - 通过 Consul 实现
3. ✅ **AI 增强** - 智能推荐菜单位置、权限、API映射
4. ✅ **实时更新** - 前端无需刷新即可看到新菜单
5. ✅ **零侵入** - 不影响现有代码

这才是真正的**增强AI中央大脑**！🧠✨

