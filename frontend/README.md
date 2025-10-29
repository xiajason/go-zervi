# Zervigo MVP 前端项目

基于 Taro 3.x 的多端统一开发框架，支持编译到微信小程序、H5、React Native 等多个平台。

## 🚀 项目特性

- ✅ **多端支持**: 一套代码，多端运行（微信小程序、H5、React Native等）
- ✅ **TypeScript**: 完整的类型支持，提升开发效率
- ✅ **组件化**: 基于 React 的组件化开发
- ✅ **状态管理**: 支持 Redux/Zustand 等状态管理方案
- ✅ **路由管理**: 内置路由管理，支持页面跳转和参数传递
- ✅ **API集成**: 与后端微服务API完美集成
- ✅ **原型图集成**: 内置93张原型图设计参考

## 📁 项目结构

```
frontend/
├── config/                    # 构建配置
│   ├── index.js              # 默认配置
│   ├── dev.js                # 开发环境配置
│   └── prod.js               # 生产环境配置
├── src/                      # 源码目录
│   ├── pages/                # 页面文件
│   │   ├── index/            # 首页
│   │   ├── login/            # 登录页
│   │   ├── register/         # 注册页
│   │   ├── profile/          # 个人中心
│   │   ├── resume/           # 简历管理
│   │   ├── job/              # 职位列表
│   │   ├── company/          # 企业信息
│   │   ├── chat/             # AI聊天
│   │   └── search/           # 搜索页面
│   ├── components/           # 公共组件
│   ├── assets/               # 静态资源
│   │   └── prototypes/       # 原型图文件
│   ├── services/             # API服务
│   ├── utils/                # 工具函数
│   ├── store/                # 状态管理
│   ├── types/                # 类型定义
│   ├── app.tsx               # 应用入口
│   ├── app.config.ts         # 应用配置
│   └── app.scss              # 全局样式
├── dist/                     # 构建输出目录
├── package.json              # 项目依赖
├── project.config.json       # 小程序项目配置
├── babel.config.js           # Babel配置
├── tsconfig.json             # TypeScript配置
└── README.md                 # 项目说明
```

## 🛠️ 开发环境

### 环境要求

- Node.js >= 16.0.0
- npm >= 8.0.0 或 yarn >= 1.22.0
- Taro CLI >= 3.6.0

### 安装依赖

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install
# 或
yarn install
```

### 开发命令

```bash
# 微信小程序开发
npm run dev:weapp

# H5开发
npm run dev:h5

# 支付宝小程序开发
npm run dev:alipay

# 百度小程序开发
npm run dev:swan

# 字节跳动小程序开发
npm run dev:tt

# QQ小程序开发
npm run dev:qq

# 京东小程序开发
npm run dev:jd

# React Native开发
npm run dev:rn

# 快应用开发
npm run dev:quickapp
```

### 构建命令

```bash
# 构建微信小程序
npm run build:weapp

# 构建H5
npm run build:h5

# 构建支付宝小程序
npm run build:alipay

# 构建百度小程序
npm run build:swan

# 构建字节跳动小程序
npm run build:tt

# 构建QQ小程序
npm run build:qq

# 构建京东小程序
npm run build:jd

# 构建React Native
npm run build:rn

# 构建快应用
npm run build:quickapp
```

## 📱 多端适配

### 平台差异处理

Taro 提供了多端适配方案，可以通过以下方式处理平台差异：

```typescript
// 平台判断
import Taro from '@tarojs/taro'

if (process.env.TARO_ENV === 'weapp') {
  // 微信小程序特有逻辑
} else if (process.env.TARO_ENV === 'h5') {
  // H5特有逻辑
} else if (process.env.TARO_ENV === 'rn') {
  // React Native特有逻辑
}
```

### 样式适配

```scss
// 使用条件编译
/* #ifdef H5 */
.h5-specific {
  display: block;
}
/* #endif */

/* #ifdef MP */
.miniprogram-specific {
  display: block;
}
/* #endif */
```

## 🔧 配置说明

### 环境变量

项目支持多环境配置：

- `.env` - 默认环境变量
- `.env.local` - 本地环境变量
- `.env.development` - 开发环境变量
- `.env.production` - 生产环境变量

### API配置

```typescript
// config/index.js
const config = {
  defineConstants: {
    API_BASE_URL: '"http://localhost:9000"',  // 开发环境
    WS_BASE_URL: '"ws://localhost:9000"'
  }
}
```

### 小程序配置

```json
// project.config.json
{
  "appid": "wx1234567890abcdef",
  "projectname": "zervigo-mvp-frontend",
  "setting": {
    "urlCheck": true,
    "es6": true,
    "postcss": true,
    "minified": true
  }
}
```

## 📊 API集成

### 服务层

项目内置了完整的API服务层，支持：

- 认证服务 (登录、注册、登出)
- 用户服务 (用户信息管理)
- 职位服务 (职位搜索、详情)
- 简历服务 (简历管理、分析)
- AI服务 (智能匹配、聊天)
- 企业服务 (企业信息管理)
- 区块链服务 (数据验证、统计)

### 使用示例

```typescript
import { ApiService } from '@/services/api'

// 用户登录
const loginResult = await ApiService.login('username', 'password')

// 获取职位列表
const jobList = await ApiService.getJobList({ page: 1, pageSize: 10 })

// AI聊天
const chatResponse = await ApiService.aiChat('你好，AI助手')
```

## 🎨 原型图集成

项目集成了93张原型图设计参考：

- **总览模式**: 整体产品架构展示
- **标注模式**: 详细功能标注说明
- **演示模式**: 产品演示和用户体验

### 查看原型图

```bash
# 在浏览器中打开原型图
open src/assets/prototypes/总览模式.html
open src/assets/prototypes/标注模式.html
open src/assets/prototypes/演示模式.html
```

## 🧪 测试

```bash
# 运行测试
npm test

# 运行测试并生成覆盖率报告
npm run test:coverage

# 代码检查
npm run lint

# 代码检查并自动修复
npm run lint:fix
```

## 📦 部署

### 小程序部署

1. 构建小程序代码：
```bash
npm run build:weapp
```

2. 使用微信开发者工具打开 `dist` 目录

3. 上传代码到微信小程序后台

### H5部署

1. 构建H5代码：
```bash
npm run build:h5
```

2. 将 `dist` 目录部署到Web服务器

### React Native部署

1. 构建RN代码：
```bash
npm run build:rn
```

2. 按照React Native部署流程进行部署

## 🔗 相关链接

- [Taro官方文档](https://docs.taro.zone/)
- [React官方文档](https://react.dev/)
- [TypeScript官方文档](https://www.typescriptlang.org/)
- [微信小程序开发文档](https://developers.weixin.qq.com/miniprogram/dev/)

## 📄 许可证

MIT License

## 👥 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 📞 联系方式

- 项目维护者: Zervigo Team
- 邮箱: team@zervigo.com
- 项目地址: https://github.com/zervigo/mvpdemo
