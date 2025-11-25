# 📮 Postman Collection 使用指南

> 为前端团队提供完整的API测试和调试工具

## 🎯 目的

这份指南帮助前端开发人员快速配置和使用Postman测试Zervigo MVP的API接口。

---

## 📦 安装和配置

### 1. 安装Postman

- **下载地址**: https://www.postman.com/downloads/
- **推荐版本**: 最新稳定版本

### 2. 导入环境配置

创建Postman Environment，命名为 `Zervigo Dev`:

```json
{
  "name": "Zervigo Dev",
  "values": [
    {
      "key": "base_url",
      "value": "http://localhost:9000",
      "type": "default",
      "enabled": true
    },
    {
      "key": "token",
      "value": "",
      "type": "secret",
      "enabled": true
    },
    {
      "key": "user_id",
      "value": "",
      "type": "default",
      "enabled": true
    },
    {
      "key": "job_id",
      "value": "",
      "type": "default",
      "enabled": true
    },
    {
      "key": "resume_id",
      "value": "",
      "type": "default",
      "enabled": true
    },
    {
      "key": "company_id",
      "value": "",
      "type": "default",
      "enabled": true
    }
  ],
  "_postman_variable_scope": "environment"
}
```

### 3. 全局请求配置

在Postman的 **Collection Settings** → **Pre-request Script** 中添加：

```javascript
// 自动添加Authorization Token（如果已设置）
if (pm.environment.get("token")) {
    pm.request.headers.add({
        key: 'Authorization',
        value: 'Bearer ' + pm.environment.get("token")
    });
}

// 添加时间戳
pm.environment.set("timestamp", Date.now());
```

在 **Tests** 中添加：

```javascript
// 自动保存Token
if (pm.response.code === 200) {
    try {
        const jsonData = pm.response.json();
        if (jsonData.token) {
            pm.environment.set("token", jsonData.token);
            console.log("Token已自动保存");
        }
        if (jsonData.user_id) {
            pm.environment.set("user_id", jsonData.user_id);
        }
    } catch (e) {
        // 忽略JSON解析错误
    }
}

// 自动保存ID
if (pm.response.code === 200) {
    try {
        const jsonData = pm.response.json();
        if (jsonData.job_id) {
            pm.environment.set("job_id", jsonData.job_id);
        }
        if (jsonData.resume_id) {
            pm.environment.set("resume_id", jsonData.resume_id);
        }
        if (jsonData.company_id) {
            pm.environment.set("company_id", jsonData.company_id);
        }
    } catch (e) {
        // 忽略JSON解析错误
    }
}
```

---

## 🔐 认证服务 (Auth Service)

### 1. 用户注册

```http
POST {{base_url}}/api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "Test123456!",
  "email": "test@example.com",
  "phone": "13800138000"
}
```

**预期响应**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser"
  }
}
```

### 2. 用户登录

```http
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "Test123456!"
}
```

**预期响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": 1,
    "username": "testuser",
    "role": "user",
    "expire_time": 1735689600
  }
}
```

**Tests脚本**:
```javascript
pm.test("登录成功", function () {
    pm.response.to.have.status(200);
    const jsonData = pm.response.json();
    pm.expect(jsonData.data.token).to.exist;
});

pm.test("自动保存Token", function () {
    const jsonData = pm.response.json();
    if (jsonData.data.token) {
        pm.environment.set("token", jsonData.data.token);
    }
});
```

### 3. 刷新Token

```http
POST {{base_url}}/api/v1/auth/refresh
Content-Type: application/json

{
  "token": "{{token}}"
}
```

### 4. 获取用户信息

```http
GET {{base_url}}/api/v1/auth/user/info
Authorization: Bearer {{token}}
```

### 5. 获取权限列表

```http
GET {{base_url}}/api/v1/auth/user/permissions
Authorization: Bearer {{token}}
```

### 6. 获取角色列表

```http
GET {{base_url}}/api/v1/auth/user/roles
Authorization: Bearer {{token}}
```

### 7. 用户登出

```http
POST {{base_url}}/api/v1/auth/logout
Authorization: Bearer {{token}}
```

---

## 👤 用户服务 (User Service)

### 1. 获取用户信息

```http
GET {{base_url}}/api/v1/user/info
Authorization: Bearer {{token}}
```

### 2. 更新用户信息

```http
PUT {{base_url}}/api/v1/user/info
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "real_name": "张三",
  "avatar": "https://example.com/avatar.jpg",
  "gender": 1,
  "birthday": "1990-01-01",
  "location": "深圳",
  "bio": "这是一段个人简介"
}
```

### 3. 用户列表

```http
GET {{base_url}}/api/v1/user/list?page=1&page_size=10
Authorization: Bearer {{token}}
```

### 4. 用户统计

```http
GET {{base_url}}/api/v1/user/stats
Authorization: Bearer {{token}}
```

### 5. 修改密码

```http
PUT {{base_url}}/api/v1/user/password
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "old_password": "Old123456!",
  "new_password": "New123456!"
}
```

### 6. 上传头像

```http
POST {{base_url}}/api/v1/user/avatar
Authorization: Bearer {{token}}
Content-Type: multipart/form-data

file: [选择图片文件]
```

---

## 💼 职位服务 (Job Service)

### 1. 创建职位

```http
POST {{base_url}}/api/v1/job
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "company_id": 1,
  "job_title": "高级后端工程师",
  "job_description": "负责系统后端开发...",
  "job_requirements": "3年以上Go语言开发经验...",
  "job_type": "全职",
  "work_location": "深圳南山区",
  "salary_min": 25000,
  "salary_max": 40000,
  "experience": "3-5年",
  "education": "本科及以上",
  "skills": ["Go", "微服务", "分布式系统"],
  "benefits": ["五险一金", "带薪年假"]
}
```

**Tests脚本**:
```javascript
pm.test("创建职位成功", function () {
    pm.response.to.have.status(200);
});

// 自动保存job_id供后续请求使用
const jsonData = pm.response.json();
if (jsonData.data && jsonData.data.job_id) {
    pm.environment.set("job_id", jsonData.data.job_id);
}
```

### 2. 获取职位详情

```http
GET {{base_url}}/api/v1/job/{{job_id}}
Authorization: Bearer {{token}}
```

### 3. 更新职位

```http
PUT {{base_url}}/api/v1/job/{{job_id}}
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "job_title": "高级后端工程师（更新）",
  "salary_max": 45000
}
```

### 4. 删除职位

```http
DELETE {{base_url}}/api/v1/job/{{job_id}}
Authorization: Bearer {{token}}
```

### 5. 搜索职位

```http
POST {{base_url}}/api/v1/job/search
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "keyword": "后端工程师",
  "location": "深圳",
  "job_type": "全职",
  "salary_min": 20000,
  "salary_max": 50000,
  "experience": "3-5年",
  "education": "本科及以上",
  "skills": ["Go", "Python"],
  "page": 1,
  "page_size": 20
}
```

### 6. 职位列表

```http
GET {{base_url}}/api/v1/job/list?page=1&page_size=10&status=1
Authorization: Bearer {{token}}
```

### 7. 获取企业职位列表

```http
GET {{base_url}}/api/v1/job/company/{{company_id}}?page=1&page_size=10
Authorization: Bearer {{token}}
```

### 8. 职位推荐

```http
POST {{base_url}}/api/v1/job/recommend
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "user_id": {{user_id}},
  "limit": 10
}
```

---

## 📄 简历服务 (Resume Service)

### 1. 创建简历

```http
POST {{base_url}}/api/v1/resume
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "resume_name": "我的简历",
  "personal_info": {
    "real_name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "gender": 1,
    "birthday": "1995-01-01",
    "location": "深圳",
    "bio": "5年后端开发经验",
    "avatar": "https://example.com/avatar.jpg"
  },
  "work_experience": [
    {
      "company_name": "XX科技公司",
      "position": "高级后端工程师",
      "start_date": "2019-06",
      "end_date": "2024-12",
      "description": "负责核心业务系统开发",
      "achievements": ["完成高并发系统优化", "设计微服务架构"]
    }
  ],
  "education": [
    {
      "school_name": "XX大学",
      "major": "计算机科学",
      "degree": "本科",
      "start_date": "2015-09",
      "end_date": "2019-06",
      "description": "主修课程包括数据结构、算法..."
    }
  ],
  "skills": [
    {
      "skill_name": "Go",
      "skill_level": "精通",
      "category": "编程语言"
    }
  ]
}
```

### 2. 获取简历详情

```http
GET {{base_url}}/api/v1/resume/{{resume_id}}
Authorization: Bearer {{token}}
```

### 3. 更新简历

```http
PUT {{base_url}}/api/v1/resume/{{resume_id}}
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "resume_name": "我的简历（更新版）"
}
```

### 4. 删除简历

```http
DELETE {{base_url}}/api/v1/resume/{{resume_id}}
Authorization: Bearer {{token}}
```

### 5. 获取用户的简历列表

```http
GET {{base_url}}/api/v1/resume/user/{{user_id}}
Authorization: Bearer {{token}}
```

### 6. 简历分析

```http
POST {{base_url}}/api/v1/resume/analyze
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "resume_id": {{resume_id}},
  "job_id": {{job_id}}
}
```

### 7. 简历匹配

```http
POST {{base_url}}/api/v1/resume/match
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "resume_id": {{resume_id}},
  "job_id": {{job_id}}
}
```

### 8. 上传简历文件

```http
POST {{base_url}}/api/v1/resume/upload
Authorization: Bearer {{token}}
Content-Type: multipart/form-data

file: [选择简历文件（PDF/DOC/DOCX）]
```

### 9. 解析简历文件

```http
POST {{base_url}}/api/v1/resume/parse
Authorization: Bearer {{token}}
Content-Type: multipart/form-data

file: [选择简历文件]
```

---

## 🏢 企业服务 (Company Service)

### 1. 创建企业

```http
POST {{base_url}}/api/v1/company
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "company_name": "XX科技有限公司",
  "company_logo": "https://example.com/logo.jpg",
  "company_description": "专注于互联网产品开发",
  "industry": "互联网/IT",
  "company_size": "50-200人",
  "website": "https://example.com",
  "address": "深圳市南山区科技园",
  "city": "深圳",
  "province": "广东",
  "country": "中国",
  "contact_person": "李经理",
  "contact_phone": "0755-12345678",
  "contact_email": "hr@example.com"
}
```

### 2. 获取企业详情

```http
GET {{base_url}}/api/v1/company/{{company_id}}
Authorization: Bearer {{token}}
```

### 3. 更新企业信息

```http
PUT {{base_url}}/api/v1/company/{{company_id}}
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "company_description": "更新后的企业简介"
}
```

### 4. 删除企业

```http
DELETE {{base_url}}/api/v1/company/{{company_id}}
Authorization: Bearer {{token}}
```

### 5. 搜索企业

```http
POST {{base_url}}/api/v1/company/search
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "keyword": "科技",
  "industry": "互联网/IT",
  "city": "深圳",
  "page": 1,
  "page_size": 20
}
```

### 6. 企业列表

```http
GET {{base_url}}/api/v1/company/list?page=1&page_size=10
Authorization: Bearer {{token}}
```

### 7. 企业认证

```http
POST {{base_url}}/api/v1/company/verify
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "company_id": {{company_id}},
  "business_license": "123456789012345",
  "tax_number": "91110000000000000",
  "legal_person": "张三",
  "legal_person_id": "110101199001011234",
  "verification_documents": [
    "https://example.com/doc1.jpg",
    "https://example.com/doc2.jpg"
  ]
}
```

### 8. 企业统计

```http
GET {{base_url}}/api/v1/company/stats
Authorization: Bearer {{token}}
```

### 9. 上传企业Logo

```http
POST {{base_url}}/api/v1/company/logo
Authorization: Bearer {{token}}
Content-Type: multipart/form-data

file: [选择Logo图片文件]
```

---

## 🤖 AI服务 (AI Service)

### 1. 职位匹配

```http
POST {{base_url}}/api/v1/ai/match
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "resume_id": {{resume_id}},
  "job_id": {{job_id}},
  "match_type": "resume_job"
}
```

### 2. 简历分析

```http
POST {{base_url}}/api/v1/ai/resume/analyze
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "resume_id": {{resume_id}},
  "analysis_type": "comprehensive"
}
```

### 3. AI聊天

```http
POST {{base_url}}/api/v1/ai/chat
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "message": "这个职位的技能要求是什么？",
  "context": "职位信息",
  "user_id": {{user_id}},
  "chat_type": "job"
}
```

### 4. 智能推荐

```http
POST {{base_url}}/api/v1/ai/recommend
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "user_id": {{user_id}},
  "recommend_type": "jobs",
  "limit": 10
}
```

### 5. 获取匹配历史

```http
GET {{base_url}}/api/v1/ai/match/history/{{user_id}}
Authorization: Bearer {{token}}
```

### 6. 获取分析历史

```http
GET {{base_url}}/api/v1/ai/analysis/history/{{user_id}}
Authorization: Bearer {{token}}
```

### 7. 获取聊天历史

```http
GET {{base_url}}/api/v1/ai/chat/history/{{session_id}}
Authorization: Bearer {{token}}
```

### 8. AI服务健康检查

```http
GET {{base_url}}/api/v1/ai/health
Authorization: Bearer {{token}}
```

---

## 🔗 区块链服务 (Blockchain Service)

### 1. 健康检查

```http
GET {{base_url}}/api/v1/blockchain/health
Authorization: Bearer {{token}}
```

### 2. 记录数据哈希

```http
POST {{base_url}}/api/v1/blockchain/record
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "data_type": "resume_update",
  "data_id": {{resume_id}},
  "data_hash": "0xabc123..."
}
```

---

## 🔍 常用调试场景

### 场景1: 首次登录获取Token

```
1. POST /api/v1/auth/login
   → 获取Token并保存到环境变量
   → 后续请求自动使用该Token
```

### 场景2: 创建完整的职位发布流程

```
1. POST /api/v1/company
   → 获取company_id
   
2. POST /api/v1/job
   → 使用company_id创建职位
   → 获取job_id
   
3. GET /api/v1/job/{{job_id}}
   → 验证职位创建成功
   
4. POST /api/v1/job/search
   → 测试职位搜索功能
```

### 场景3: 简历匹配流程

```
1. POST /api/v1/resume
   → 创建简历
   → 获取resume_id
   
2. POST /api/v1/resume/match
   → 测试简历与职位匹配
   
3. POST /api/v1/ai/match
   → AI智能匹配分析
   
4. GET /api/v1/ai/match/history/{{user_id}}
   → 查看匹配历史
```

### 场景4: 测试分页功能

```
1. GET /api/v1/job/list?page=1&page_size=10
2. GET /api/v1/job/list?page=2&page_size=10
3. 验证total字段正确性
```

### 场景5: 测试权限控制

```
1. 使用普通用户Token
   → POST /api/v1/job → 应该403 Forbidden
   
2. 使用企业用户Token
   → POST /api/v1/job → 应该200 OK
```

---

## 📊 响应断言示例

### 通用断言

```javascript
// 状态码断言
pm.test("状态码为200", function () {
    pm.response.to.have.status(200);
});

// 响应时间断言
pm.test("响应时间小于500ms", function () {
    pm.expect(pm.response.responseTime).to.be.below(500);
});

// JSON结构断言
pm.test("响应为JSON格式", function () {
    pm.response.to.be.json;
});

// 业务码断言
pm.test("业务码为200", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.code).to.eql(200);
});
```

### 数据断言

```javascript
// 字段存在性断言
pm.test("返回数据包含必要字段", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.data).to.have.property('id');
    pm.expect(jsonData.data).to.have.property('name');
});

// 字段类型断言
pm.test("ID为数字类型", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.data.id).to.be.a('number');
});

// 数组长度断言
pm.test("列表不为空", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.data.list).to.be.an('array');
    pm.expect(jsonData.data.list.length).to.be.above(0);
});
```

---

## 🛠️ 常见问题排查

### 问题1: Token过期

**现象**: 401 Unauthorized

**解决方案**:
1. 重新登录获取新Token
2. 检查Token格式是否正确: `Bearer {token}`
3. 检查环境变量是否更新

### 问题2: CORS跨域错误

**现象**: Access-Control-Allow-Origin error

**解决方案**:
1. 检查后端CORS配置
2. 确认前端Origin在白名单中
3. 检查OPTIONS预检请求

### 问题3: 500 Internal Server Error

**现象**: 服务器内部错误

**解决方案**:
1. 查看后端日志: `tail -f logs/*.log`
2. 检查数据库连接
3. 检查请求参数格式
4. 检查必要服务是否启动

### 问题4: 分页参数无效

**现象**: 分页结果不符合预期

**解决方案**:
1. 检查page和page_size参数类型（应为数字）
2. 验证total字段是否正确
3. 检查数据库查询逻辑

---

## 📝 Postman Collection导出

### 导出方法

1. 在Postman中完成所有请求配置
2. 点击Collection → Export
3. 选择Collection v2.1格式
4. 保存为 `Zervigo_MVP_API.postman_collection.json`

### 分享给团队

```bash
# 1. 上传到项目仓库
git add docs/Zervigo_MVP_API.postman_collection.json
git commit -m "Add Postman API collection"
git push

# 2. 或上传到Postman Cloud
# - 点击Collection → Share
# - 选择"Share via link"
# - 复制链接发送给团队
```

---

## 📚 扩展阅读

- [API接口文档](BACKEND_FRONTEND_HANDOVER_CHECKLIST.md)
- [数据库设计](MICROSERVICE_DATABASE_DESIGN.md)
- [项目README](../../README.md)
- [部署文档](DEPLOYMENT.md)

---

**最后更新**: 2025-01-XX  
**维护团队**: Zervigo 后端团队

