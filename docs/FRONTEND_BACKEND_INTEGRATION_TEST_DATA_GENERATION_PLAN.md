# 前后端联调测试数据生成与验证方案

## 📋 问题分析

### 当前挑战

1. **测试数据不足**
   - 数据库只有少量基础测试数据
   - 无法支持完整的业务流测试
   - 前后端联调困难

2. **业务流测试需要**
   - 用户注册 → 登录 → 创建简历 → 搜索职位 → 申请职位 → AI匹配
   - 企业注册 → 认证 → 发布职位 → 查看简历 → 审核申请
   - 需要关联数据支持完整流程

3. **前端开发需求**
   - 需要真实的数据展示效果
   - 需要覆盖各种边界情况
   - 需要测试权限控制

---

## 🎯 解决方案设计

### 方案一：渐进式测试数据生成（推荐）

**特点**：逐步生成测试数据，支持不同阶段的开发需求

**优势**：
- ✅ 数据量可控，不会过多影响性能
- ✅ 可以按模块生成，逐步完善
- ✅ 支持快速重置和重新生成

**实施步骤**：
1. **基础数据层**：用户、角色、权限
2. **业务数据层**：简历、职位、企业
3. **关联数据层**：申请、匹配、聊天记录
4. **完整业务流**：端到端的业务场景

---

### 方案二：Mock数据服务

**特点**：独立的数据Mock服务，前端可以独立开发

**优势**：
- ✅ 前端和后端可以并行开发
- ✅ 不依赖真实数据库
- ✅ 可以快速模拟各种场景

**实施步骤**：
1. 搭建Mock服务（使用JSON Server或自定义）
2. 定义API数据格式
3. 前端接入Mock服务
4. 联调时切换到真实API

---

### 方案三：数据库种子脚本

**特点**：完整的SQL种子脚本，一键生成测试数据

**优势**：
- ✅ 数据一致性高
- ✅ 可以版本控制
- ✅ 支持快速重置

**实施步骤**：
1. 编写SQL种子脚本
2. 生成关联数据
3. 定期更新和维护

---

## 🚀 推荐方案：混合方案

### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    测试数据生成层                        │
├─────────────────────────────────────────────────────────┤
│  1. SQL种子脚本 (基础数据)                                │
│  2. Python数据生成器 (业务数据)                            │
│  3. Mock服务 (前端独立开发)                                │
│  4. API测试脚本 (业务流验证)                                │
└─────────────────────────────────────────────────────────┘
```

### 实施优先级

**Phase 1: MVP阶段（立即实施）**
- ✅ SQL种子脚本：基础用户、角色、权限
- ✅ Python生成器：少量简历、职位、企业数据
- ✅ 业务流验证：核心流程（注册→登录→创建简历→搜索职位）

**Phase 2: 功能完善阶段**
- ✅ 扩展数据生成器：更多业务数据
- ✅ Mock服务：前端独立开发
- ✅ 自动化测试：API端到端测试

**Phase 3: 优化阶段**
- ✅ 性能测试数据：大数据量
- ✅ 压力测试场景：并发测试
- ✅ 边界情况数据：异常场景

---

## 📊 测试数据生成器设计

### 1. 完整业务流数据生成器

**功能**：
- 生成用户（求职者、HR、企业管理员）
- 生成企业（已认证、未认证）
- 生成职位（不同类型、不同状态）
- 生成简历（不同格式、不同状态）
- 生成关联数据（申请、匹配、聊天）

**数据结构**：
```python
# 用户数据
users = [
    {
        "username": "job_seeker_001",
        "email": "jobseeker001@example.com",
        "role": "job_seeker",
        "profile": {...},
        "resumes": [...],
        "applications": [...]
    },
    # ...
]

# 企业数据
companies = [
    {
        "company_name": "科技公司A",
        "status": "verified",
        "jobs": [...],
        "applications": [...]
    },
    # ...
]
```

---

### 2. 数据关联性保证

**关键关联**：
- ✅ 用户 → 简历 → 职位申请 → 职位
- ✅ 企业 → 职位 → 职位申请 → 用户
- ✅ AI匹配记录 → 简历 + 职位
- ✅ 聊天记录 → 用户 + 会话

**生成策略**：
```python
# 1. 先生成基础数据（用户、企业）
users = generate_users(count=50)
companies = generate_companies(count=20)

# 2. 基于基础数据生成业务数据
for user in users:
    if user.role == 'job_seeker':
        resumes = generate_resumes(user, count=random.randint(1, 3))
        for resume in resumes:
            applications = generate_applications(resume, jobs)

for company in companies:
    jobs = generate_jobs(company, count=random.randint(5, 15))
```

---

### 3. 业务场景数据

**场景1：求职者完整流程**
```
1. 注册账户 → users表
2. 完善个人信息 → users表更新
3. 创建简历 → resumes表
4. 搜索职位 → jobs表查询
5. 申请职位 → job_applications表
6. 查看AI匹配 → ai_matches表
7. 查看聊天记录 → ai_chats表
```

**场景2：企业完整流程**
```
1. 注册企业账户 → users表 + companies表
2. 企业认证 → company_verifications表
3. 发布职位 → jobs表
4. 查看简历 → resumes表查询
5. 审核申请 → job_applications表更新
6. 查看统计 → statistics表
```

---

## 🛠️ 实施工具

### 1. SQL种子脚本生成器

**文件**：`databases/postgres/init/10-seed-test-data.sql`

**功能**：
- 生成基础用户（10个不同类型）
- 生成基础企业（5个）
- 生成基础职位（20个）
- 生成基础简历（15个）
- 生成关联数据（申请、匹配等）

**使用方式**：
```bash
# 执行种子脚本
psql -U postgres -d zervigo_unified -f databases/postgres/init/10-seed-test-data.sql
```

---

### 2. Python数据生成器（增强版）

**文件**：`scripts/comprehensive_test_data_generator.py`

**功能**：
- 生成大量测试数据（可配置数量）
- 支持业务流数据生成
- 支持数据重置和清理
- 生成测试报告

**使用方式**：
```bash
# 生成测试数据
python scripts/comprehensive_test_data_generator.py --users 100 --companies 20 --jobs 200 --resumes 150

# 重置测试数据
python scripts/comprehensive_test_data_generator.py --reset

# 生成特定业务场景数据
python scripts/comprehensive_test_data_generator.py --scenario job_seeker_flow
```

---

### 3. Mock服务

**文件**：`scripts/mock-api-server.py`

**功能**：
- 提供Mock API服务
- 返回预设的测试数据
- 支持自定义响应
- 支持延迟模拟

**使用方式**：
```bash
# 启动Mock服务
python scripts/mock-api-server.py --port 9999

# 前端连接Mock服务
# 修改 frontend/src/config/api.ts
# API_BASE_URL = 'http://localhost:9999/api'
```

---

### 4. API测试脚本

**文件**：`scripts/test-business-flow.sh`

**功能**：
- 测试完整业务流
- 验证API响应
- 生成测试报告

**使用方式**：
```bash
# 执行业务流测试
bash scripts/test-business-flow.sh

# 测试特定场景
bash scripts/test-business-flow.sh --scenario job_seeker_flow
```

---

## 📝 测试数据规格

### 基础数据规模

| 数据类型 | 数量 | 说明 |
|---------|------|------|
| 用户 | 50-100 | 包含求职者、HR、企业管理员 |
| 企业 | 20-30 | 包含已认证、未认证 |
| 职位 | 100-200 | 不同类型、不同状态 |
| 简历 | 80-150 | 不同格式、不同状态 |
| 申请记录 | 50-100 | 不同状态 |
| AI匹配记录 | 30-50 | 不同匹配度 |
| 聊天记录 | 100-200 | 不同类型 |

### 业务流数据

**求职者流程**：
- 10个求职者账户
- 每个账户1-3份简历
- 每个账户5-10个申请记录
- 每个账户10-20条AI匹配记录

**企业流程**：
- 5个企业账户
- 每个企业10-20个职位
- 每个企业20-50个申请记录
- 每个企业10-20个统计记录

---

## 🔄 前后端联调流程

### 阶段1：数据准备

```bash
# 1. 初始化数据库
psql -U postgres -d zervigo_unified -f databases/postgres/init/01-init-schema.sql
psql -U postgres -d zervigo_unified -f databases/postgres/init/02-zervigo-microservices-schema.sql

# 2. 生成测试数据
python scripts/comprehensive_test_data_generator.py --users 50 --companies 20 --jobs 100 --resumes 80

# 3. 验证数据
psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM users;"
psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM companies;"
psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM jobs;"
```

---

### 阶段2：后端启动

```bash
# 1. 启动基础设施
bash scripts/start-consul.sh
bash scripts/start-central-brain.sh

# 2. 启动业务服务
bash scripts/start-phase2-services.sh

# 3. 验证服务健康
bash scripts/comprehensive_health_check.sh
```

---

### 阶段3：前端接入

```bash
# 1. 前端开发环境
cd frontend
npm install
npm run dev:h5  # 或 npm run dev:weapp

# 2. 配置API地址
# frontend/src/config/api.ts
export const API_BASE_URL = 'http://localhost:9000/api/v1'

# 3. 测试登录
# 使用测试账号登录
# username: test_user_001
# password: test123456
```

---

### 阶段4：业务流验证

```bash
# 1. 测试用户注册
curl -X POST http://localhost:9000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test_user","email":"test@example.com","password":"test123456"}'

# 2. 测试登录
curl -X POST http://localhost:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_user","password":"test123456"}'

# 3. 测试创建简历
curl -X POST http://localhost:9000/api/v1/resume/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"resume_name":"我的简历","personal_info":{...}}'

# 4. 测试搜索职位
curl -X GET "http://localhost:9000/api/v1/job/search?keyword=软件工程师" \
  -H "Authorization: Bearer <token>"
```

---

## 🎨 Mock数据服务设计

### Mock服务架构

```
┌─────────────────────────────────────────────────────────┐
│                    Mock API Server                       │
├─────────────────────────────────────────────────────────┤
│  GET  /api/v1/users          → 返回用户列表                │
│  GET  /api/v1/users/:id      → 返回用户详情                │
│  POST /api/v1/auth/login     → 返回Mock Token             │
│  GET  /api/v1/jobs           → 返回职位列表                │
│  GET  /api/v1/resumes        → 返回简历列表                │
│  ...更多API                                                  │
└─────────────────────────────────────────────────────────┘
```

### Mock数据格式

**用户数据**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [
      {
        "user_id": 1,
        "username": "test_user_001",
        "email": "test001@example.com",
        "role": "job_seeker",
        "profile": {
          "real_name": "测试用户001",
          "avatar": "https://example.com/avatar1.jpg",
          "location": "深圳"
        }
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

**职位数据**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "jobs": [
      {
        "job_id": 1,
        "company_id": 1,
        "company_name": "科技公司A",
        "job_title": "高级软件工程师",
        "job_description": "负责后端服务开发...",
        "salary_min": 20000,
        "salary_max": 35000,
        "location": "深圳",
        "work_type": "full-time"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 📋 测试数据生成脚本实现

### 1. 完整业务流数据生成器

**实现思路**：
```python
#!/usr/bin/env python3
# scripts/comprehensive_test_data_generator.py

import asyncio
import asyncpg
import json
import random
from datetime import datetime, timedelta
from faker import Faker

fake = Faker('zh_CN')

class TestDataGenerator:
    def __init__(self, db_config):
        self.db_config = db_config
        self.conn = None
        
    async def connect(self):
        self.conn = await asyncpg.connect(**self.db_config)
        
    async def generate_users(self, count=50):
        """生成用户数据"""
        users = []
        for i in range(count):
            role = random.choice(['job_seeker', 'hr', 'company_admin'])
            user = {
                'username': f'test_user_{i+1:03d}',
                'email': f'test{i+1:03d}@example.com',
                'phone': f'138{random.randint(10000000, 99999999)}',
                'password_hash': '$2a$10$...',  # bcrypt hash of 'test123456'
                'role': role,
                'real_name': fake.name(),
                'location': fake.city(),
                'status': 'active'
            }
            users.append(user)
        return users
    
    async def generate_companies(self, count=20):
        """生成企业数据"""
        companies = []
        for i in range(count):
            company = {
                'company_name': f'{fake.company()}科技有限公司',
                'industry': random.choice(['互联网', '金融', '制造业', '服务业']),
                'company_size': random.choice(['1-50', '51-200', '201-500', '500+']),
                'city': fake.city(),
                'status': random.choice(['active', 'pending']),
                'verification_status': random.choice(['verified', 'pending', 'rejected'])
            }
            companies.append(company)
        return companies
    
    async def generate_jobs(self, companies, count=100):
        """生成职位数据"""
        jobs = []
        job_titles = ['软件工程师', '产品经理', 'UI设计师', '数据分析师', '测试工程师']
        for i in range(count):
            company = random.choice(companies)
            job = {
                'company_id': company['company_id'],
                'job_title': random.choice(job_titles),
                'job_description': fake.text(max_nb_chars=500),
                'job_requirements': fake.text(max_nb_chars=300),
                'salary_min': random.randint(10000, 20000),
                'salary_max': random.randint(25000, 50000),
                'location': fake.city(),
                'status': random.choice(['active', 'closed', 'draft'])
            }
            jobs.append(job)
        return jobs
    
    async def generate_resumes(self, users, count=80):
        """生成简历数据"""
        resumes = []
        for i in range(count):
            user = random.choice(users)
            resume = {
                'user_id': user['user_id'],
                'resume_name': f'{user["real_name"]}的简历',
                'personal_info': json.dumps({
                    'name': user['real_name'],
                    'email': user['email'],
                    'phone': user['phone'],
                    'location': user['location']
                }),
                'work_experience': json.dumps([
                    {
                        'company': fake.company(),
                        'position': fake.job(),
                        'duration': f'{random.randint(1, 5)}年',
                        'responsibilities': fake.text(max_nb_chars=200)
                    }
                ]),
                'education': json.dumps([
                    {
                        'school': fake.company() + '大学',
                        'degree': random.choice(['本科', '硕士', '博士']),
                        'major': fake.job(),
                        'graduation_year': str(random.randint(2015, 2023))
                    }
                ]),
                'skills': json.dumps([fake.job() for _ in range(random.randint(3, 8))]),
                'status': random.choice(['active', 'draft'])
            }
            resumes.append(resume)
        return resumes
    
    async def insert_data(self, table, data):
        """插入数据到数据库"""
        # 实现数据库插入逻辑
        pass
    
    async def generate_full_business_flow(self):
        """生成完整业务流数据"""
        # 1. 生成基础数据
        users = await self.generate_users(50)
        companies = await self.generate_companies(20)
        
        # 2. 插入基础数据
        await self.insert_data('users', users)
        await self.insert_data('companies', companies)
        
        # 3. 生成业务数据
        jobs = await self.generate_jobs(companies, 100)
        resumes = await self.generate_resumes(users, 80)
        
        # 4. 插入业务数据
        await self.insert_data('jobs', jobs)
        await self.insert_data('resumes', resumes)
        
        # 5. 生成关联数据
        applications = await self.generate_applications(resumes, jobs, 50)
        matches = await self.generate_ai_matches(resumes, jobs, 30)
        
        # 6. 插入关联数据
        await self.insert_data('job_applications', applications)
        await self.insert_data('ai_matches', matches)
        
        print("✅ 完整业务流数据生成完成！")
```

---

## 🔍 成功案例参考

### 案例1：分层数据生成

**经验**：
- ✅ 基础数据 → 业务数据 → 关联数据，分层生成
- ✅ 保证数据关联性，避免孤立数据
- ✅ 支持增量生成，可以逐步完善

---

### 案例2：Mock服务 + 真实数据混合

**经验**：
- ✅ 前端开发阶段使用Mock服务
- ✅ 联调阶段切换到真实API
- ✅ Mock数据格式与真实API保持一致

---

### 案例3：自动化测试数据生成

**经验**：
- ✅ 每次测试前自动生成测试数据
- ✅ 测试后自动清理
- ✅ 保证测试环境的一致性

---

## 📊 实施计划

### Phase 1: 基础数据生成（1-2天）

**任务**：
- ✅ 编写SQL种子脚本
- ✅ 生成基础用户、角色、权限数据
- ✅ 生成基础企业数据

**交付物**：
- `databases/postgres/init/10-seed-test-data.sql`

---

### Phase 2: 业务数据生成器（2-3天）

**任务**：
- ✅ 编写Python数据生成器
- ✅ 实现业务数据生成逻辑
- ✅ 实现数据关联逻辑

**交付物**：
- `scripts/comprehensive_test_data_generator.py`

---

### Phase 3: Mock服务（1-2天）

**任务**：
- ✅ 搭建Mock API服务
- ✅ 定义Mock数据格式
- ✅ 前端接入Mock服务

**交付物**：
- `scripts/mock-api-server.py`
- `frontend/src/mocks/` (Mock数据文件)

---

### Phase 4: 业务流验证（1天）

**任务**：
- ✅ 编写业务流测试脚本
- ✅ 验证完整业务流程
- ✅ 生成测试报告

**交付物**：
- `scripts/test-business-flow.sh`
- `docs/BUSINESS_FLOW_TEST_REPORT.md`

---

## ✅ 总结

### 核心方案

1. **SQL种子脚本**：基础数据，快速生成
2. **Python生成器**：业务数据，灵活可控
3. **Mock服务**：前端独立开发，并行工作
4. **自动化测试**：业务流验证，持续集成

### 关键优势

- ✅ **数据量可控**：可以按需生成数据
- ✅ **关联性保证**：数据之间有正确的关联关系
- ✅ **快速重置**：可以快速清理和重新生成
- ✅ **前后端并行**：Mock服务支持前端独立开发

### 下一步行动

1. **立即实施**：编写SQL种子脚本
2. **本周完成**：Python数据生成器
3. **下周完成**：Mock服务和前端接入
4. **持续优化**：根据实际需求调整数据生成策略

---

**报告生成时间**: 2025-01-29  
**关键建议**: 采用混合方案，SQL种子脚本 + Python生成器 + Mock服务，支持不同阶段的开发需求

