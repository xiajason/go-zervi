# 🧠 AI增强中央大脑 - 使用指南

## ✅ 已完成的功能

### 1. Consul 监听器
- ✅ 自动监听前端服务注册
- ✅ 检测路由元数据变化
- ✅ 长轮询机制（5分钟超时）

### 2. AI 路由分析器
- ✅ 智能推断图标
- ✅ 智能推断表名
- ✅ 智能推荐父菜单
- ✅ 智能推荐排序位置
- ✅ 自动生成API映射
- ✅ 自动推断权限

### 3. 路由注册器
- ✅ 批量路由注册
- ✅ 数据库事务保护
- ✅ 重复检测
- ✅ API映射自动创建

## 📋 快速开始

### 第一步：启动Consul

```bash
# 如果还没有安装Consul
brew install consul  # macOS
# 或
apt-get install consul  # Ubuntu

# 启动Consul（开发模式）
consul agent -dev
```

### 第二步：配置中央大脑

编辑配置文件或设置环境变量：

```yaml
# config.yaml
service_discovery:
  enabled: true
  consul_url: "http://localhost:8500"
  service_host: "localhost"
```

或使用环境变量：

```bash
export SERVICE_DISCOVERY_ENABLED=true
export CONSUL_URL=http://localhost:8500
```

### 第三步：启动中央大脑

```bash
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
go run .
```

你应该看到：

```
✅ Consul监听器初始化成功
🔍 启动Consul监听器...
👀 开始监听前端服务: vuecmf-frontend
🧠 AI增强中央大脑：Consul自动路由注册已启动
```

## 🎯 使用方式

### 方式一：前端页面自动注册（推荐）

#### 1. 创建前端页面并添加元信息

```vue
<!-- src/views/template/MyAwesomePage.vue -->
<!-- @route-meta
  title: "我的超棒页面"
  icon: "Star"
  permission: "awesome:view"
  parent: "/system"
  description: "这是一个很酷的页面"
-->
<template>
  <div class="awesome-page">
    <h2>{{ title }}</h2>
    <p>这里是我的超棒页面内容</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const title = ref('我的超棒页面')
</script>
```

#### 2. 前端构建时自动扫描

将来需要添加 Vite 插件（route-scanner），现在我们使用手动方式模拟：

```json
// dist/route-manifest.json (手动创建用于测试)
{
  "version": "1.0.0",
  "generated_at": "2025-11-06T16:30:00Z",
  "routes": [
    {
      "path": "/system/myawesomepage",
      "component": "template/MyAwesomePage",
      "title": "我的超棒页面",
      "icon": "Star",
      "permission": "awesome:view",
      "parent": "/system",
      "description": "这是一个很酷的页面"
    }
  ]
}
```

#### 3. 注册前端服务到Consul

```bash
# 使用Consul API注册服务
curl -X PUT http://localhost:8500/v1/agent/service/register \
  -d '{
    "ID": "vuecmf-frontend-8081",
    "Name": "vuecmf-frontend",
    "Tags": ["frontend", "vuecmf"],
    "Port": 8081,
    "Address": "localhost",
    "Meta": {
      "routes": "[{\"path\":\"/system/myawesomepage\",\"component\":\"template/MyAwesomePage\",\"title\":\"我的超棒页面\",\"icon\":\"Star\"}]"
    },
    "Check": {
      "HTTP": "http://localhost:8081/health",
      "Interval": "10s"
    }
  }'
```

#### 4. 观察中央大脑日志

你应该看到：

```
🔔 检测到前端服务变化 (index: 123 -> 456)
📋 处理 1 个前端服务实例
🎯 发现 1 个路由定义
🚀 开始注册 1 个路由...
🧠 AI分析路由: /system/myawesomepage (我的超棒页面)
📌 找到指定的父菜单: /system (ID: 2)
🔗 生成 4 个API映射
✅ 分析完成: 父菜单ID=2, 排序=40, API数=4
➕ 注册新路由: /system/myawesomepage (我的超棒页面)
✅ 菜单已创建 (ID: 9)
✅ API映射已创建: myawesomepage.list -> /api/v1/myawesomepage/index
✅ API映射已创建: myawesomepage.detail -> /api/v1/myawesomepage/detail
✅ API映射已创建: myawesomepage.save -> /api/v1/myawesomepage/save
✅ API映射已创建: myawesomepage.delete -> /api/v1/myawesomepage/delete
🎉 路由注册成功: /system/myawesomepage
📊 注册结果: 成功=1, 跳过=0, 失败=0
📢 通知前端更新路由...
```

#### 5. 验证数据库

```sql
-- 查看新注册的菜单
SELECT * FROM menu WHERE path = '/system/myawesomepage';

-- 查看API映射
SELECT * FROM api_map WHERE table_name = 'myawesomepage';
```

### 方式二：手动触发注册（测试用）

如果Consul还没有完全配置好，可以使用手动方式测试：

```go
// test_route_registration.go
package main

import (
    "fmt"
    "log"
)

func testRouteRegistration() {
    // 创建测试路由
    testRoutes := []RouteMetadata{
        {
            Path:        "/system/test1",
            Component:   "template/Test1",
            Title:       "测试页面1",
            Icon:        "Document",
            Permission:  "test1:view",
            Parent:      "/system",
        },
        {
            Path:        "/test2",
            Component:   "template/Test2",
            Title:       "测试页面2",
            Icon:        "Setting",
        },
    }
    
    // 直接调用注册器
    registrar := NewRouteRegistrar(centralBrain)
    registrar.RegisterRoutes(testRoutes)
}
```

## 🔍 AI分析能力展示

### 示例1：系统管理页面

**输入：**
```json
{
  "path": "/system/audit",
  "component": "template/System/Audit",
  "title": "审计日志"
}
```

**AI推断：**
- ✅ 图标：`Document`（默认）
- ✅ 表名：`audit`
- ✅ 父菜单：`/system`（系统管理）
- ✅ 排序：`40`（在用户、角色、权限之后）
- ✅ API映射：
  - `audit.list` → `/api/v1/audit/index`
  - `audit.detail` → `/api/v1/audit/detail`
  - `audit.save` → `/api/v1/audit/save`
  - `audit.delete` → `/api/v1/audit/delete`
- ✅ 权限：`audit:view`, `audit:create`, `audit:update`, `audit:delete`

### 示例2：Dashboard页面

**输入：**
```json
{
  "path": "/analytics/dashboard",
  "component": "template/Analytics/Dashboard",
  "title": "数据分析",
  "icon": "DataAnalysis"
}
```

**AI推断：**
- ✅ 图标：`DataAnalysis`（使用指定的）
- ✅ 表名：`dashboard`
- ✅ 父菜单：`0`（根级别，因为找不到`/analytics`）
- ✅ 排序：`50`
- ✅ API映射：
  - `dashboard.list` → `/api/v1/dashboard/index`
  - `dashboard.stats` → `/api/v1/dashboard/stats`（Dashboard特有）
- ✅ 权限：`dashboard:view`, `dashboard:create`, `dashboard:update`, `dashboard:delete`

## 📊 监控和调试

### 查看Consul服务

```bash
# 查看所有服务
curl http://localhost:8500/v1/catalog/services

# 查看特定服务
curl http://localhost:8500/v1/catalog/service/vuecmf-frontend

# 查看服务元数据
curl http://localhost:8500/v1/catalog/service/vuecmf-frontend | jq '.[0].ServiceMeta'
```

### 查看中央大脑状态

```bash
# 健康检查
curl http://localhost:9000/health

# 查看性能指标
curl http://localhost:9000/api/v1/metrics

# 查看菜单列表（验证注册结果）
curl http://localhost:9000/api/v1/menu/nav | jq '.data.nav_menu'
```

### 数据库查询

```sql
-- 查看所有菜单
SELECT id, title, path, component_tpl, pid, sort_num 
FROM menu 
ORDER BY pid, sort_num;

-- 查看API映射
SELECT table_name, action_type, api_path 
FROM api_map 
ORDER BY table_name, action_type;

-- 查看最近注册的路由
SELECT * FROM menu 
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;
```

## 🚀 下一步

### 短期（本周）

1. ✅ Consul监听器 - 已完成
2. ✅ AI路由分析器 - 已完成
3. ✅ 路由注册器 - 已完成
4. 🔄 前端Vite插件开发
5. 🔄 前端Consul注册脚本

### 中期（下周）

6. 🔄 SSE事件推送（实时通知前端）
7. 🔄 前端路由自动更新
8. 🔄 完整E2E测试

### 长期（未来）

9. 📝 权限自动分配
10. 📝 后端API自动生成
11. 📝 测试代码自动生成

## 💡 最佳实践

### 1. 路由命名规范

```
推荐: /module/feature
示例: /system/users, /analytics/reports

不推荐: /users_list, /userManagement
```

### 2. 元信息完整性

```vue
<!-- 完整的元信息 -->
<!-- @route-meta
  title: "清晰的标题"
  icon: "合适的图标"
  permission: "resource:action"
  parent: "/明确的父路径"
  description: "详细的描述"
-->
```

### 3. 组件命名

```
推荐: template/Module/Feature.vue
示例: template/System/Users.vue

生成路径: /system/users
生成表名: users
```

## ❓ 常见问题

### Q1: Consul监听器没有反应？

检查：
1. Consul是否在运行？`consul members`
2. 配置是否正确？检查`SERVICE_DISCOVERY_ENABLED`
3. 前端服务是否注册？`curl localhost:8500/v1/catalog/service/vuecmf-frontend`

### Q2: 路由注册失败？

检查：
1. 数据库连接是否正常？
2. `menu`和`api_map`表是否存在？
3. 查看中央大脑日志中的错误信息

### Q3: 前端看不到新菜单？

原因：
- 前端需要重新加载菜单
- 目前需要手动刷新页面
- 未来会实现SSE实时推送

临时方案：
```javascript
// 浏览器控制台
localStorage.clear()
location.reload()
```

## 📚 相关文档

- [增强AI中央大脑-动态路由自动注册方案.md](./增强AI中央大脑-动态路由自动注册方案.md) - 完整架构设计
- [中央大脑介入时序分析.md](../zervigo-admin/中央大脑介入时序分析.md) - 时序分析
- [Consul官方文档](https://www.consul.io/docs)

---

**注意**：这是AI增强中央大脑的第一阶段实现。完整的自动化流程（包括前端Vite插件和实时更新）将在后续版本中完成。

当前版本适合用于：
- ✅ 手动测试路由注册
- ✅ 验证AI分析能力
- ✅ 演示自动化潜力
- ✅ 开发环境使用

