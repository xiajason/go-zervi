# 实际项目API设计参考与Zervigo业务层对比分析报告

## 📋 报告概述

**分析日期**: 2025-01-29  
**参考项目**: `/Users/szjason72/resume-center/miniprogram-4`  
**分析目的**: 基于实际运行的小程序项目，发现真实API设计模式和前端需求，补充Zervigo业务层设计  
**分析范围**: API路径设计、数据模型、功能模块、文件上传、资源管理

---

## 🎯 关键发现总结

### ✅ 发现的宝贵设计模式

1. **API路径层次化设计** ✅
   - `/personal/*` - 个人用户相关
   - `/resource/*` - 资源管理（文件、字典）
   - 清晰的业务模块划分

2. **文件上传和资源管理分离** ✅
   - 独立的资源服务（Resource Service）
   - 使用resourceId管理文件
   - 支持批量获取资源URL

3. **简历数据结构设计** ✅
   - 完整的数据结构（basicInfo, jobIntention, workExperiences等）
   - 支持多个求职意向
   - 支持简历权限和黑名单管理

4. **首页数据聚合** ✅
   - 横幅（banners）接口
   - 通知（notifications）接口
   - 分页支持

---

## 📊 实际项目API完整清单

### 1. 认证相关API (`/personal/authentication/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/personal/authentication/login` | POST | 微信登录（openIdCode, phoneCode） | ⚠️ **需要补充** |
| `/personal/authentication/check` | POST | Token有效性检查 | ⚠️ **需要补充** |
| `/personal/authentication/getUserPhone` | GET | 获取用户手机号 | ⚠️ **需要补充** |
| `/personal/authentication/getUserIdKey` | POST | 获取用户ID密钥（实名认证） | ⚠️ **需要补充** |
| `/personal/authentication/certification` | POST | 实名认证 | ⚠️ **需要补充** |
| `/personal/authentication/logout` | POST | 登出 | ✅ 已有 |
| `/personal/authentication/getMyUserIdKey` | GET | 获取我的ID密钥 | ⚠️ **需要补充** |
| `/personal/authentication/cancellation` | POST | 注销账号 | ⚠️ **需要补充** |

**关键发现**:
- ✅ 微信小程序登录流程（openIdCode + phoneCode）
- ✅ Token检查机制
- ✅ 实名认证流程
- ✅ 账号注销功能

---

### 2. 简历相关API (`/personal/resume/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/personal/resume/list/summary` | GET | 获取简历列表（摘要） | ⚠️ **需要补充** |
| `/personal/resume/create` | POST | 创建简历 | ✅ 需实现 |
| `/personal/resume/publish/:resumeId` | POST | 发布简历 | ⚠️ **需要补充** |
| `/personal/resume/publish/:resumeId` | DELETE | 删除简历 | ✅ 需实现 |
| `/personal/resume/detail/:resumeId` | GET | 获取简历详情 | ✅ 需实现 |
| `/personal/resume/update/:resumeId` | PUT | 更新简历 | ✅ 需实现 |
| `/personal/resume/templates` | GET | 获取简历模板列表 | ⚠️ **需要补充** |
| `/personal/resume/upload/:resumeId` | POST | 上传简历文件 | ⚠️ **需要补充** |
| `/personal/resume/permission/:resumeId` | GET | 获取简历权限设置 | ⚠️ **需要补充** |
| `/personal/resume/permission/:resumeId` | PUT | 更新简历权限设置 | ⚠️ **需要补充** |
| `/personal/resume/blacklist/:resumeId` | GET | 获取简历黑名单 | ⚠️ **需要补充** |
| `/personal/resume/blacklist/:resumeId` | PUT | 设置简历黑名单 | ⚠️ **需要补充** |
| `/personal/resume/preview/:resumeId` | GET | 简历预览（HTML/图片） | ⚠️ **需要补充** |

**关键发现**:
- ✅ **简历摘要列表**（summary）- 前端首页需要简历列表，但不需要完整详情
- ✅ **简历发布/取消发布** - 前端需要发布功能
- ✅ **简历权限管理** - 前端需要设置谁可以查看简历
- ✅ **简历黑名单** - 前端需要屏蔽某些公司或用户
- ✅ **简历预览** - 前端需要生成预览图片或HTML

**简历数据结构**:
```typescript
{
  resumeId: string
  basicInfo: {
    name: string
    phone: string
    email: string
    gender: string
    birthday: string
    city: string
    photoResourceId: string
    // ...
  }
  jobIntention: {
    status: string  // 求职状态
    details: Array<{
      industry: string
      position: string
      city: string
      nature: string  // 工作性质
      salary: string  // 期望薪资
    }>
  }
  workExperiences: Array<{
    startDate: string
    endDate: string
    companyName: string
    position: string
    nature: string
    description: string
    salary: string
  }>
  educationExperiences: Array<{
    schoolName: string
    major: string
    degree: string
    startDate: string
    endDate: string
  }>
  projectExperiences: Array<{
    projectName: string
    startDate: string
    endDate: string
    description: string
    technologies: string[]
  }>
  trainingExperiences: Array<{...}>
  honors: Array<{...}>
  attachments: Array<{
    resourceId: string
    fileName: string
  }>
  selfEvaluation: string
  templateId: string
  blockChainHash: string
  publishTime: string
}
```

---

### 3. 用户/个人中心API (`/personal/mine/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/personal/mine/info` | GET | 获取用户信息 | ⚠️ **需要补充** |
| `/personal/mine/points` | GET | 获取积分 | ⚠️ **需要补充** |
| `/personal/mine/points/bill` | GET | 获取积分账单 | ⚠️ **需要补充** |
| `/personal/mine/approve/history` | GET | 获取审批历史 | ⚠️ **需要补充** |
| `/personal/mine/view/history` | GET | 获取查看历史 | ⚠️ **需要补充** |
| `/personal/mine/certification` | GET | 获取认证状态 | ⚠️ **需要补充** |
| `/personal/mine/avatar` | PUT | 更新头像（resourceId） | ⚠️ **需要补充** |

**关键发现**:
- ✅ **积分系统** - 前端需要积分和积分账单
- ✅ **查看历史** - 前端需要记录简历查看历史
- ✅ **审批历史** - 前端需要审批记录
- ✅ **头像更新** - 使用resourceId模式

---

### 4. 首页API (`/personal/home/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/personal/home/banners` | GET | 获取首页横幅 | ⚠️ **需要补充** |
| `/personal/home/notifications` | GET | 获取通知列表（分页） | ⚠️ **需要补充** |

**关键发现**:
- ✅ **首页横幅** - 前端首页需要banner轮播
- ✅ **通知列表** - 前端首页需要通知消息（分页）

**数据结构**:
```typescript
// Banner
{
  resourceId: string  // 图片资源ID
  // ...其他字段
}

// Notification
{
  records: Array<{
    // 通知内容
  }>
  total: number
  pageNum: number
  pageSize: number
}
```

---

### 5. 资源管理API (`/resource/*`) ⭐ **极其重要**

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/resource/upload` | POST | 文件上传（支持fileName参数） | ❌ **缺失** |
| `/resource/urls` | POST | 批量获取资源URL（resourceIds数组） | ❌ **缺失** |
| `/resource/url/:resourceId` | GET | 获取单个资源URL | ❌ **缺失** |

**关键发现**: ⭐⭐⭐⭐⭐
- ✅ **独立的资源服务** - 文件上传和URL管理分离
- ✅ **resourceId模式** - 所有文件都用resourceId引用
- ✅ **批量获取URL** - 前端需要批量获取多个资源URL（性能优化）
- ✅ **fileName参数** - 上传时指定文件名

**使用场景**:
```typescript
// 1. 上传头像
const resourceId = await upload('/resource/upload?fileName=avatar.jpg', { filePath, name: 'file' })

// 2. 更新用户头像
await updateAvatar({ resourceId })

// 3. 批量获取资源URL（简历列表）
const urls = await getResources([resourceId1, resourceId2, resourceId3])

// 4. 获取单个资源URL
const url = await getResource(resourceId)
```

---

### 6. 字典数据API (`/resource/dict/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/resource/dict/type/list` | GET | 获取字典类型列表 | ❌ **缺失** |
| `/resource/dict/data` | GET | 获取字典数据（dictTypeId） | ❌ **缺失** |
| `/resource/dict/search/school` | GET | 搜索学校（keyword） | ❌ **缺失** |

**关键发现**:
- ✅ **字典系统** - 前端需要字典数据（行业、职位、城市、工作性质等）
- ✅ **学校搜索** - 前端需要学校自动补全

**使用场景**:
```typescript
// 1. 获取字典类型
const types = await getDictTypes()

// 2. 获取字典数据（如"求职状态"）
const statusDict = await getDictData('求职状态')

// 3. 搜索学校
const schools = await searchSchool('清华大学')
```

---

### 7. 审批相关API (`/personal/approve/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/personal/approve/list` | GET | 获取审批列表（分页） | ⚠️ **需要补充** |
| `/personal/approve/handle/:approveId` | POST | 处理审批（res: 0/1） | ⚠️ **需要补充** |

**关键发现**:
- ✅ **审批流程** - 前端需要审批中心功能
- ✅ **审批处理** - 前端需要审批/拒绝操作

---

### 8. 聊天/知识库API (`/personal/chat/*`)

| 接口路径 | 方法 | 功能 | Zervigo状态 |
|---------|------|------|------------|
| `/personal/chat/usual` | GET | 获取常用问题列表 | ⚠️ **需要补充** |
| `/personal/chat/chat` | GET | 聊天（query参数） | ⚠️ **需要补充** |

**关键发现**:
- ✅ **常用问题** - 前端需要预设问题
- ✅ **简单聊天** - GET请求的聊天接口（可能是FAQ）

---

## 🎯 关键设计模式总结

### 1. API路径设计模式 ✅

**实际项目采用**:
```
/personal/{module}/{action}/{id?}
```

**示例**:
- `/personal/resume/list/summary`
- `/personal/resume/detail/:resumeId`
- `/personal/mine/info`
- `/personal/home/banners`

**Zervigo当前设计**:
```
/api/v1/{service}/{action}/{id?}
```

**建议**: ✅ **保持Zervigo设计，但可以参考实际项目的action命名**

---

### 2. 文件上传和资源管理 ⭐⭐⭐⭐⭐

**实际项目设计**:
```
独立Resource Service:
- POST /resource/upload?fileName=xxx
- POST /resource/urls (批量获取)
- GET /resource/url/:resourceId
```

**Zervigo当前状态**: ❌ **完全缺失**

**建议**: ⚠️ **必须实现**
1. 创建独立的Resource Service或文件上传服务
2. 实现resourceId管理模式
3. 支持批量获取URL（性能优化）

---

### 3. 简历数据结构设计 ✅

**实际项目数据结构**:
```typescript
{
  basicInfo: {...}
  jobIntention: {
    status: string
    details: Array<{...}>  // 支持多个求职意向
  }
  workExperiences: Array<{...}>
  educationExperiences: Array<{...}>
  projectExperiences: Array<{...}>
  trainingExperiences: Array<{...}>
  honors: Array<{...}>
  attachments: Array<{...}>
  selfEvaluation: string
  templateId: string
  blockChainHash: string
}
```

**Zervigo数据库表结构**:
- ✅ `brew_jobfirst_v3_resumes` - 主表
- ✅ `brew_jobfirst_v3_work_experiences` - 工作经历
- ✅ `brew_jobfirst_v3_educations` - 教育经历
- ✅ `brew_jobfirst_v3_projects` - 项目经历
- ✅ `brew_jobfirst_v3_certifications` - 证书
- ⚠️ **缺少**: trainingExperiences（培训经历）
- ⚠️ **缺少**: honors（荣誉奖项）
- ⚠️ **缺少**: attachments（附件）

**匹配度**: ⚠️ **80%**

---

### 4. 首页数据聚合 ✅

**实际项目设计**:
```
GET /personal/home/banners
GET /personal/home/notifications?pageNum=1&pageSize=10
```

**Zervigo当前状态**: ❌ **完全缺失**

**建议**: ⚠️ **必须实现**
1. 创建首页数据聚合接口
2. 横幅管理（需要banner表）
3. 通知列表（需要notification表）

---

### 5. 字典数据系统 ⭐⭐⭐⭐⭐

**实际项目设计**:
```
GET /resource/dict/type/list
GET /resource/dict/data?dictTypeId=xxx
GET /resource/dict/search/school?keyword=xxx
```

**Zervigo当前状态**: ❌ **完全缺失**

**关键发现**:
- ✅ 前端需要大量字典数据（行业、职位、城市、工作性质、求职状态等）
- ✅ 前端需要学校搜索功能
- ✅ 前端需要职位类别搜索

**建议**: ⚠️ **必须实现**
1. 创建字典服务或字典表
2. 实现字典数据查询接口
3. 实现学校搜索接口

---

### 6. 简历权限和隐私管理 ⭐⭐⭐⭐

**实际项目设计**:
```
GET /personal/resume/permission/:resumeId
PUT /personal/resume/permission/:resumeId
GET /personal/resume/blacklist/:resumeId
PUT /personal/resume/blacklist/:resumeId
```

**Zervigo当前状态**: ❌ **完全缺失**

**关键发现**:
- ✅ 前端需要简历权限设置（谁可以查看）
- ✅ 前端需要黑名单功能（屏蔽某些公司）

**建议**: ⚠️ **必须实现**
1. 简历权限管理表
2. 简历黑名单表
3. 相关API接口

---

### 7. 简历预览功能 ⭐⭐⭐⭐

**实际项目设计**:
```
GET /personal/resume/preview/:resumeId?width=375&height=667
返回: HTML字符串或图片URL
```

**关键发现**:
- ✅ 前端需要简历预览功能
- ✅ 支持不同尺寸的预览

**建议**: ⚠️ **必须实现**
1. 简历模板渲染服务
2. HTML生成或图片生成
3. 预览接口

---

### 8. 简历发布/取消发布 ⭐⭐⭐⭐

**实际项目设计**:
```
POST /personal/resume/publish/:resumeId
DELETE /personal/resume/publish/:resumeId
```

**关键发现**:
- ✅ 前端需要简历发布功能
- ✅ 前端需要取消发布功能（删除）

**建议**: ⚠️ **必须实现**
1. 简历状态管理（draft, published, unpublished）
2. 发布接口
3. 取消发布接口

---

## 📊 Zervigo缺失的关键功能清单

### 🔥 高优先级（必须实现）

| 功能 | 实际项目API | Zervigo状态 | 优先级 |
|------|------------|------------|--------|
| **文件上传服务** | `/resource/upload` | ❌ 缺失 | 🔥🔥🔥 |
| **资源URL管理** | `/resource/urls`, `/resource/url/:resourceId` | ❌ 缺失 | 🔥🔥🔥 |
| **字典数据系统** | `/resource/dict/*` | ❌ 缺失 | 🔥🔥🔥 |
| **简历摘要列表** | `/personal/resume/list/summary` | ❌ 缺失 | 🔥🔥🔥 |
| **简历发布/取消发布** | `/personal/resume/publish/:resumeId` | ❌ 缺失 | 🔥🔥🔥 |
| **简历预览** | `/personal/resume/preview/:resumeId` | ❌ 缺失 | 🔥🔥🔥 |
| **简历权限管理** | `/personal/resume/permission/:resumeId` | ❌ 缺失 | 🔥🔥🔥 |
| **简历黑名单** | `/personal/resume/blacklist/:resumeId` | ❌ 缺失 | 🔥🔥🔥 |
| **首页横幅** | `/personal/home/banners` | ❌ 缺失 | 🔥🔥🔥 |
| **首页通知** | `/personal/home/notifications` | ❌ 缺失 | 🔥🔥🔥 |
| **用户信息** | `/personal/mine/info` | ❌ 缺失 | 🔥🔥🔥 |
| **用户头像更新** | `/personal/mine/avatar` | ❌ 缺失 | 🔥🔥🔥 |

### ⚠️ 中优先级（建议实现）

| 功能 | 实际项目API | Zervigo状态 | 优先级 |
|------|------------|------------|--------|
| **积分系统** | `/personal/mine/points`, `/personal/mine/points/bill` | ❌ 缺失 | 🔥🔥 |
| **查看历史** | `/personal/mine/view/history` | ❌ 缺失 | 🔥🔥 |
| **审批中心** | `/personal/approve/list`, `/personal/approve/handle/:approveId` | ❌ 缺失 | 🔥🔥 |
| **微信登录** | `/personal/authentication/login` | ❌ 缺失 | 🔥🔥 |
| **实名认证** | `/personal/authentication/certification` | ❌ 缺失 | 🔥🔥 |
| **Token检查** | `/personal/authentication/check` | ❌ 缺失 | 🔥🔥 |

---

## 🎯 实际项目数据模型参考

### 简历数据结构（完整版）

```typescript
interface Resume {
  resumeId: string
  
  // 基本信息
  basicInfo: {
    name: string
    phone: string
    email: string
    gender: string  // "男" | "女"
    birthday: string  // "YYYY-MM-DD"
    city: string
    photoResourceId: string  // 头像资源ID
    huKou: string  // 户口
    currentIdentity: string  // 当前身份
    workTime: string  // 工作年限
  }
  
  // 求职意向（支持多个）
  jobIntention: {
    status: string  // 求职状态 "在职-考虑机会" | "离职-立即到岗" 等
    details: Array<{
      industry: string  // 行业
      position: string  // 职位
      city: string  // 城市
      nature: string  // 工作性质 "全职" | "兼职" | "实习"
      salary: string  // 期望薪资
    }>
  }
  
  // 工作经历
  workExperiences: Array<{
    startDate: string  // "YYYY-MM"
    endDate: string  // "YYYY-MM"
    companyName: string
    position: string
    positionShort: string  // 职位简称
    nature: string  // 工作性质
    description: string  // 工作描述
    salary: string  // 薪资
  }>
  
  // 教育经历
  educationExperiences: Array<{
    schoolName: string
    major: string
    degree: string  // 学历
    startDate: string
    endDate: string
  }>
  
  // 项目经历
  projectExperiences: Array<{
    projectName: string
    startDate: string
    endDate: string
    description: string
    technologies: string[]
  }>
  
  // 培训经历
  trainingExperiences: Array<{
    // ...
  }>
  
  // 荣誉奖项
  honors: Array<{
    // ...
  }>
  
  // 附件
  attachments: Array<{
    resourceId: string
    fileName: string
  }>
  
  // 自我评价
  selfEvaluation: string
  
  // 模板ID
  templateId: string
  
  // 区块链哈希
  blockChainHash: string
  
  // 发布时间
  publishTime: string
  
  // 更新时间
  updateTime: string
}
```

---

## 📋 对比分析结果

### Zervigo前端需求 vs 实际项目API

| 需求类别 | Zervigo前端需求 | 实际项目API | 匹配度 |
|---------|----------------|------------|--------|
| **认证** | 登录、注册、登出 | 微信登录、Token检查、登出 | ⚠️ **60%** |
| **用户信息** | 用户信息查询、更新 | 用户信息查询、头像更新 | ⚠️ **70%** |
| **简历管理** | 列表、创建、编辑、删除 | 摘要列表、创建、更新、删除、发布、预览、权限 | ⚠️ **50%** |
| **职位管理** | 列表、详情、搜索、收藏、申请 | 无（可能是另一个服务） | ⚠️ **未知** |
| **文件上传** | 头像、简历文件 | 资源上传、URL管理 | ❌ **0%** |
| **首页数据** | 首页展示 | 横幅、通知 | ❌ **0%** |
| **字典数据** | 行业、职位、城市等 | 字典系统、学校搜索 | ❌ **0%** |

**总体匹配度**: ⚠️ **40%**

---

## 🎯 关键建议和行动计划

### 第一阶段：核心功能补充（必须立即实现）

#### 1. 资源服务（Resource Service）⭐⭐⭐⭐⭐

**必须实现**:
```go
// 文件上传
POST /api/v1/resource/upload?fileName=xxx
- 支持文件上传
- 返回resourceId

// 批量获取URL
POST /api/v1/resource/urls
Body: { resourceIds: []string }
- 返回资源URL映射

// 获取单个URL
GET /api/v1/resource/url/:resourceId
- 返回资源URL
```

**数据库表**:
- `resources` - 资源表（resourceId, fileName, filePath, fileSize, fileType, created_at）

**实施优先级**: 🔥🔥🔥 **最高**

---

#### 2. 字典服务（Dict Service）⭐⭐⭐⭐⭐

**必须实现**:
```go
// 获取字典类型列表
GET /api/v1/dict/types

// 获取字典数据
GET /api/v1/dict/data?type=求职状态
- 返回字典数据列表

// 搜索学校
GET /api/v1/dict/schools/search?keyword=清华
- 返回学校列表
```

**数据库表**:
- `dict_types` - 字典类型表
- `dict_data` - 字典数据表
- `schools` - 学校表（可选）

**实施优先级**: 🔥🔥🔥 **最高**

---

#### 3. 简历服务补充API ⭐⭐⭐⭐

**必须实现**:
```go
// 简历摘要列表
GET /api/v1/resume/list/summary
- 返回简历摘要列表（不包含完整详情）

// 简历发布
POST /api/v1/resume/:resumeId/publish

// 简历取消发布
DELETE /api/v1/resume/:resumeId/publish

// 简历预览
GET /api/v1/resume/:resumeId/preview?width=375&height=667
- 返回HTML或图片URL

// 简历权限管理
GET /api/v1/resume/:resumeId/permission
PUT /api/v1/resume/:resumeId/permission

// 简历黑名单
GET /api/v1/resume/:resumeId/blacklist
PUT /api/v1/resume/:resumeId/blacklist
```

**数据库表补充**:
- `resume_permissions` - 简历权限表
- `resume_blacklist` - 简历黑名单表

**实施优先级**: 🔥🔥🔥 **最高**

---

#### 4. 首页服务（Home Service）⭐⭐⭐⭐

**必须实现**:
```go
// 获取横幅
GET /api/v1/home/banners
- 返回横幅列表（resourceId）

// 获取通知
GET /api/v1/home/notifications?pageNum=1&pageSize=10
- 返回通知列表（分页）
```

**数据库表**:
- `banners` - 横幅表
- `notifications` - 通知表

**实施优先级**: 🔥🔥 **高**

---

#### 5. 用户服务补充API ⭐⭐⭐⭐

**必须实现**:
```go
// 获取用户信息
GET /api/v1/user/info
- 返回用户信息（包含resourceId）

// 更新用户信息
PUT /api/v1/user/info

// 更新用户头像
PUT /api/v1/user/avatar
Body: { resourceId: string }

// Token检查
POST /api/v1/auth/check
- 检查Token有效性
```

**实施优先级**: 🔥🔥🔥 **最高**

---

### 第二阶段：增强功能实现

#### 1. 积分系统 ⭐⭐⭐

**实施优先级**: 🔥🔥 **中**

#### 2. 查看历史 ⭐⭐⭐

**实施优先级**: 🔥🔥 **中**

#### 3. 审批中心 ⭐⭐⭐

**实施优先级**: 🔥🔥 **中**

---

## 🔍 数据库表结构补充建议

### 需要新增的表

1. **resources** - 资源表
```sql
CREATE TABLE resources (
  resource_id VARCHAR(100) PRIMARY KEY,
  file_name VARCHAR(255),
  file_path VARCHAR(500),
  file_size BIGINT,
  file_type VARCHAR(50),
  uploader_id BIGINT,
  created_at TIMESTAMP
);
```

2. **dict_types** - 字典类型表
```sql
CREATE TABLE dict_types (
  dict_type_id BIGINT PRIMARY KEY,
  dict_type_name VARCHAR(100) UNIQUE,
  dict_type_code VARCHAR(100) UNIQUE,
  description TEXT,
  status VARCHAR(20)
);
```

3. **dict_data** - 字典数据表
```sql
CREATE TABLE dict_data (
  dict_data_id BIGINT PRIMARY KEY,
  dict_type_id BIGINT,
  dict_label VARCHAR(100),
  dict_value VARCHAR(100),
  sort_order INT,
  status VARCHAR(20)
);
```

4. **banners** - 横幅表
```sql
CREATE TABLE banners (
  banner_id BIGINT PRIMARY KEY,
  resource_id VARCHAR(100),
  title VARCHAR(200),
  link_url VARCHAR(500),
  sort_order INT,
  status VARCHAR(20),
  created_at TIMESTAMP
);
```

5. **notifications** - 通知表
```sql
CREATE TABLE notifications (
  notification_id BIGINT PRIMARY KEY,
  user_id BIGINT,
  title VARCHAR(200),
  content TEXT,
  type VARCHAR(50),
  is_read BOOLEAN,
  created_at TIMESTAMP
);
```

6. **resume_permissions** - 简历权限表
```sql
CREATE TABLE resume_permissions (
  permission_id BIGINT PRIMARY KEY,
  resume_id BIGINT,
  permission_type VARCHAR(50),  -- public, private, company_only
  allowed_companies JSONB,  -- 允许查看的公司列表
  created_at TIMESTAMP
);
```

7. **resume_blacklist** - 简历黑名单表
```sql
CREATE TABLE resume_blacklist (
  blacklist_id BIGINT PRIMARY KEY,
  resume_id BIGINT,
  company_ids BIGINT[],  -- 屏蔽的公司ID列表
  user_ids BIGINT[],  -- 屏蔽的用户ID列表
  created_at TIMESTAMP
);
```

---

## 📊 实际项目API路径映射到Zervigo

### 建议的Zervigo API路径设计

| 实际项目路径 | Zervigo建议路径 | 服务 |
|------------|---------------|------|
| `/personal/resume/list/summary` | `/api/v1/resume/list/summary` | Resume Service |
| `/personal/resume/create` | `/api/v1/resume` (POST) | Resume Service |
| `/personal/resume/update/:resumeId` | `/api/v1/resume/:resumeId` (PUT) | Resume Service |
| `/personal/resume/publish/:resumeId` | `/api/v1/resume/:resumeId/publish` (POST) | Resume Service |
| `/personal/resume/preview/:resumeId` | `/api/v1/resume/:resumeId/preview` | Resume Service |
| `/personal/mine/info` | `/api/v1/user/info` | User Service |
| `/personal/mine/avatar` | `/api/v1/user/avatar` (PUT) | User Service |
| `/personal/home/banners` | `/api/v1/home/banners` | Home Service或Banner Service |
| `/personal/home/notifications` | `/api/v1/home/notifications` | Notification Service |
| `/resource/upload` | `/api/v1/resource/upload` | Resource Service |
| `/resource/urls` | `/api/v1/resource/urls` (POST) | Resource Service |
| `/resource/dict/data` | `/api/v1/dict/data` | Dict Service |

---

## ✅ 总结和关键发现

### 🎯 核心发现

1. **资源管理是核心基础设施** ⭐⭐⭐⭐⭐
   - 所有文件（头像、简历、附件）都用resourceId管理
   - 独立的资源服务，支持批量获取URL
   - **Zervigo完全缺失此功能**

2. **字典系统是必需的基础功能** ⭐⭐⭐⭐⭐
   - 前端需要大量字典数据（行业、职位、城市、工作性质等）
   - 学校搜索功能
   - **Zervigo完全缺失此功能**

3. **简历功能比想象中复杂** ⭐⭐⭐⭐
   - 简历摘要列表（性能优化）
   - 简历发布/取消发布
   - 简历权限管理
   - 简历黑名单
   - 简历预览（HTML/图片生成）
   - **Zervigo只实现了基础CRUD**

4. **首页数据聚合是必需的** ⭐⭐⭐⭐
   - 横幅管理
   - 通知列表
   - **Zervigo完全缺失此功能**

5. **用户中心功能丰富** ⭐⭐⭐⭐
   - 用户信息管理
   - 头像更新（resourceId模式）
   - 积分系统
   - 查看历史
   - **Zervigo只实现了基础功能**

---

### 📋 必须立即补充的功能

**优先级排序**:

1. 🔥🔥🔥 **资源服务（Resource Service）**
   - 文件上传
   - 资源URL管理
   - 批量获取URL

2. 🔥🔥🔥 **字典服务（Dict Service）**
   - 字典类型和数据
   - 学校搜索

3. 🔥🔥🔥 **简历服务补充**
   - 摘要列表
   - 发布/取消发布
   - 权限管理
   - 黑名单管理
   - 预览功能

4. 🔥🔥🔥 **用户服务补充**
   - 用户信息查询和更新
   - 头像更新
   - Token检查

5. 🔥🔥 **首页服务**
   - 横幅
   - 通知列表

---

### 🎯 实施建议

**不能只关注数据库表结构，还需要关注**:

1. ✅ **API接口设计** - 参考实际项目的API路径和参数设计
2. ✅ **文件上传和资源管理** - 独立的资源服务是必需的
3. ✅ **字典数据系统** - 前端需要大量字典数据
4. ✅ **首页数据聚合** - 前端首页需要横幅和通知
5. ✅ **简历功能增强** - 发布、权限、预览等功能

**建议调整第二阶段实施计划**:

1. **先实现基础设施**:
   - Resource Service（文件上传）
   - Dict Service（字典数据）

2. **再实现核心业务**:
   - 用户服务完整API
   - 简历服务完整API（包含发布、权限、预览）
   - 职位服务完整API

3. **最后实现增强功能**:
   - 首页服务
   - 积分系统
   - 审批中心

---

**报告生成时间**: 2025-01-29  
**参考项目**: resume-center/miniprogram-4  
**关键发现**: 实际项目比我们想象的更复杂，需要补充大量基础设施功能  
**建议**: 立即开始资源服务和字典服务的实现

