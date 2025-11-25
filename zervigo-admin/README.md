# 🧠 Zervigo Admin - 动态管理后台

基于 Element Plus + Vue 3 + TypeScript 构建的现代化管理后台系统。

借鉴 VueCMF 的优秀设计模式，实现了完整的动态菜单和二级页面系统。

## ✨ 特性

- 🚀 **动态菜单系统** - 从后端 API 加载菜单配置
- 🎯 **智能路由注册** - 自动注册动态路由
- 🔍 **智能组件查找** - 借鉴 VueCMF 的最佳实践
- 💡 **优雅降级策略** - 找不到组件时使用通用模板
- 🧭 **面包屑导航** - 自动追踪路径
- 🎨 **图标动态渲染** - 支持 Element Plus 图标
- 📱 **响应式设计** - 适配各种屏幕尺寸
- 🔧 **TypeScript** - 完整的类型支持

## 📦 技术栈

- **框架**: Vue 3
- **路由**: Vue Router 4
- **UI**: Element Plus
- **语言**: TypeScript
- **构建**: Vite
- **状态**: Vue Composition API

## 🚀 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

### 3. 访问应用

打开浏览器访问前端地址（以终端输出为准）：
- 通常是: `http://localhost:3000`
- 或查看终端中 Vite 输出的实际地址

**默认账号**:
- 用户名: `admin`
- 密码: `admin123`

## 📁 项目结构

```
zervigo-admin/
├── src/
│   ├── api/                    # API 接口
│   │   ├── auth.ts            # 认证 API
│   │   ├── menu.ts            # 菜单 API (新增)
│   │   ├── request.ts         # 请求封装
│   │   └── system.ts          # 系统管理 API
│   ├── services/              # 业务服务
│   │   └── MenuService.ts     # 菜单服务 (新增)
│   ├── views/
│   │   ├── Layout.vue         # 布局组件 (升级)
│   │   ├── Home.vue           # 首页
│   │   ├── Login.vue          # 登录页
│   │   ├── template/          # 模板系统 (新增)
│   │   │   ├── Common.vue     # 通用模板
│   │   │   ├── ListTemplate.vue # 列表模板
│   │   │   └── system/        # 系统模块模板
│   │   └── system/            # 系统管理页面
│   │       ├── Users.vue      # 用户管理
│   │       ├── Roles.vue      # 角色管理
│   │       └── Permissions.vue # 权限管理
│   ├── router/
│   │   └── index.ts           # 路由配置 (升级)
│   └── main.ts
├── QUICK_START.md             # 快速启动指南
├── DYNAMIC_MENU_GUIDE.md      # 动态菜单详细指南
├── VUECMF_LESSONS_LEARNED.md  # VueCMF 学习总结
└── README.md                  # 本文件
```

## 🎯 核心功能

### 1. 动态菜单系统

菜单配置存储在后端数据库，前端自动加载并渲染：

```typescript
// 菜单自动加载
await menuService.loadMenu()

// 菜单刷新
await menuService.refresh()
```

### 2. 智能路由注册

根据菜单配置自动注册动态路由：

```typescript
// 自动查找对应的组件
const componentKey = Object.keys(modules).find(key => 
  key.includes(componentPath) || 
  key.endsWith(`/${componentPath}.vue`)
)

// 找不到时使用通用模板
component: componentKey 
  ? modules[componentKey] 
  : () => import('@/views/template/Common.vue')
```

### 3. 模板系统

提供可复用的页面模板：

- **Common.vue** - 通用降级模板
- **ListTemplate.vue** - 列表页模板
- **自定义模板** - 按需创建

## 📚 文档

- [快速启动指南](./QUICK_START.md) - 快速上手
- [动态菜单完整指南](./DYNAMIC_MENU_GUIDE.md) - 详细使用说明
- [VueCMF 学习总结](./VUECMF_LESSONS_LEARNED.md) - 技术经验分享

## 🔧 从 VueCMF 学到的经验

本项目借鉴了 VueCMF 的优秀设计：

1. ✅ **智能组件查找** - 避免硬编码路径
2. ✅ **优雅降级策略** - 提升用户体验
3. ✅ **调试友好设计** - 开发环境详细日志
4. ✅ **模块化架构** - 易于维护和扩展

详见: [VUECMF_LESSONS_LEARNED.md](./VUECMF_LESSONS_LEARNED.md)

## 🎨 核心改进

### 路由加载（借鉴 VueCMF）

**VueCMF 原始问题**:
```typescript
// ❌ 路径拼接错误
component: modules[import.meta.env.BASE_URL + 'src/views/' + path + '.vue']
```

**Zervigo 正确实现**:
```typescript
// ✅ 智能查找 + 降级策略
const componentKey = Object.keys(modules).find(key => 
  key.includes(componentPath) || 
  key.endsWith(`/${componentPath}.vue`)
)
component: componentKey 
  ? modules[componentKey] 
  : () => import('@/views/template/Common.vue')
```

## 🔍 调试技巧

### 查看菜单数据

```javascript
// 在浏览器控制台执行
console.log(menuService.menuTree.value)
console.log(menuService.menuList.value)
```

### 查看可用模板

开发环境下，路由加载时会输出：

```
[路由加载] 可用的模板文件: [...]
[路由注册] 用户管理 -> /system/users ✅
[路由注册] 职位管理 -> /business/jobs ⚠️ 使用通用模板
```

## 🐛 常见问题

### Q: 菜单没有显示？

检查：
1. 后端服务是否启动（端口 9000）
2. 菜单 API 是否正常
3. 浏览器控制台是否有错误

### Q: 点击菜单显示通用模板？

这是正常的降级行为！说明：
- 路由已正确注册
- 对应的专用组件不存在
- 系统使用通用模板提供友好提示

如需创建专用模板，参考控制台提示。

## 📊 API 接口

### 后端接口

```typescript
// 获取菜单列表
GET /api/v1/router/menu/list

// 获取导航菜单（树形）
GET /api/v1/router/menu/nav

// 获取用户权限菜单
GET /api/v1/router/user-pages
```

## 🎯 开发指南

### 添加新菜单

1. 在后端数据库中配置菜单
2. （可选）创建对应的 Vue 组件
3. 刷新页面即可看到新菜单

### 创建专用模板

```bash
# 1. 创建组件文件
# 例如：src/views/template/business/Jobs.vue

# 2. 实现组件
<template>
  <div>职位管理页面</div>
</template>

# 3. 刷新页面测试
```

## 🌟 特别感谢

- [VueCMF](https://github.com/vuecmf/vuecmf-web) - 提供了宝贵的设计参考
- [Element Plus](https://element-plus.org/) - 优秀的 UI 组件库
- [Vue 3](https://vuejs.org/) - 渐进式 JavaScript 框架

## 📄 License

MIT

## 👥 贡献

欢迎提交 Issue 和 Pull Request！

---

**更新日期**: 2024-11-05  
**版本**: v2.0.0 - 动态菜单完整版  
**技术亮点**: 借鉴 VueCMF 优秀设计，实现智能路由系统  
