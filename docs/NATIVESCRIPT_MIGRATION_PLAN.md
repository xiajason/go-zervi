# NativeScript-Vue 迁移计划

## 🎯 目标

- 用 NativeScript-Vue 重建移动端客户端，替换已归档的 Taro 多端方案。
- 复用现有微服务接口、中央大脑路由及权限体系，保持业务契约不变。
- 为未来 H5/小程序等端留出共享逻辑抽象层。

---

## 🏗️ 目录规划

```
apps/
└── nativescript/
    ├── app/                # 页面、组件、store
    ├── app.scss            # 全局样式
    ├── nativescript.config.ts
    ├── package.json
    └── tsconfig.json
packages/
├── shared/                 # TS 类型、API SDK、权限常量
└── ui/                     # 通用逻辑组件（可选）
```

原有 Taro 资源统一移至 `cleanup-backup/`，仅作历史参考。

---

## 🔑 实施步骤

1. **初始化工程**
   - 安装 CLI：`npm install -g @nativescript/cli`
   - `ns create nativescript --appPath apps/nativescript --vue`
   - 固定 Node 版本（推荐 22）并写入 `package.json`。

2. **环境配置**
   - 在 `nativescript.config.ts` 配置应用 ID、打包信息。
   - 使用 `.env.development/.env.production` 注入 `CENTRAL_BRAIN_BASE_URL` 等后端地址。
   - 在 `webpack.config.js` 中加载环境变量。

3. **API 封装**
   - 在 `packages/shared/api` 实现请求工具，使用 `@nativescript/core/http` 与 `ApplicationSettings` 管理 Token。
   - 统一处理 401（跳转登录）与错误提示。

4. **核心模块迁移**
   - 认证/Login → `app/views/auth`。
   - Profile、Job、Resume、Company 等业务页面 → 对应模块文件夹。
   - 使用 Pinia/状态管理同步权限、路由信息。

5. **动态路由与权限**
   - 启动时请求 `central-brain` 的 `/api/v1/router/user-pages`、`/api/v1/permission/...`。
   - 在路由钩子中校验角色/权限，未授权跳转统一页面。

6. **测试与交付**
   - `ns run ios|android --env.development` 本地调试。
   - `ns build` 生成产物，集成到 CI/CD 流程。
   - 补充真机截图、联调记录。

---

## ✅ 验收清单

- [ ] 登录/退出流程
- [ ] Profile 页面读取/编辑
- [ ] 职位/简历/公司模块列表与详情
- [ ] AI 服务调试入口
- [ ] 权限不足时的拦截与提示
- [ ] 打包脚本与版本管理更新

---

## 🔄 后续展望

- 在 `packages/shared` 抽象更多跨端逻辑，未来若重启小程序/H5 端，可直接复用。
- 评估是否需要生成原型到 NativeScript 页面模板的自动化脚本。
- 整合移动端监控、日志上报，保持与后端服务链路一致。

> 原有 Taro 文档与脚本已迁移至 `cleanup-backup/`。若需历史方案，请在该目录查看。

