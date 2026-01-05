# 测试数据生成与三阶段实施计划协调方案

## 📋 协调原则

**核心理念**: 测试数据生成不是独立任务，而是**贯穿整个三阶段实施过程**的支撑性工作，与每个阶段的开发需求紧密结合。

---

## 🎯 与三阶段计划的对应关系

### Phase 1: 核心基础设施建设 (2-3周)

**测试数据需求**: ⭐ **基础数据**

**需要的测试数据**:
- ✅ 基础用户数据（测试认证功能）
- ✅ 角色和权限数据（测试权限系统）
- ✅ 服务凭证数据（测试服务间认证）

**实施时机**:
- **第1周**: 认证和网关集成时，需要测试用户数据
- **第2周**: 数据库集成时，需要基础数据验证
- **第3周**: 健康检查时，需要数据完整性验证

**协调方案**:
```
Phase 1 Week 1:
  - 启动认证服务
  - ✅ 生成基础测试用户（10个）
  - ✅ 生成角色和权限数据
  - 测试登录和Token生成

Phase 1 Week 2:
  - 数据库初始化
  - ✅ 执行SQL种子脚本（基础数据）
  - 验证数据完整性
  - 测试数据库连接

Phase 1 Week 3:
  - 版本管理测试
  - ✅ 数据备份和恢复测试
  - 健康检查验证数据
```

---

### Phase 2: 业务层构建 (3-4周)

**测试数据需求**: ⭐⭐⭐ **完整业务数据**

**需要的测试数据**:
- ✅ 扩展用户数据（50-100个，不同角色）
- ✅ 企业数据（20-30个）
- ✅ 职位数据（100-200个）
- ✅ 简历数据（80-150个）
- ✅ 关联数据（申请、匹配等）

**实施时机**:
- **第1周**: 用户和简历服务开发，需要测试数据
- **第2周**: 职位和公司服务开发，需要业务数据
- **第3-4周**: 业务逻辑集成，需要完整业务流数据

**协调方案**:
```
Phase 2 Week 1 (用户和简历服务):
  - 开发用户服务
  - ✅ 生成求职者用户（30个）
  - ✅ 生成简历数据（50个）
  - 测试用户CRUD和简历CRUD

Phase 2 Week 2 (职位和公司服务):
  - 开发公司服务
  - ✅ 生成企业用户（10个）
  - ✅ 生成企业数据（20个）
  - ✅ 生成职位数据（100个）
  - 测试公司认证和职位发布

Phase 2 Week 3-4 (业务集成):
  - 业务流程测试
  - ✅ 生成完整业务流数据
  - ✅ 生成关联数据（申请记录、匹配记录）
  - 测试端到端业务场景
```

---

### Phase 3: 微服务整体集成 (2-3周)

**测试数据需求**: ⭐⭐ **联调数据 + Mock数据**

**需要的测试数据**:
- ✅ 完整业务流数据（支持端到端测试）
- ✅ Mock数据服务（支持前端独立开发）
- ✅ 边界测试数据（异常场景）

**实施时机**:
- **第1周**: AI服务集成，需要匹配数据
- **第2周**: 支持服务集成，需要统计数据
- **第3周**: 前后端联调，需要完整数据和Mock服务

**协调方案**:
```
Phase 3 Week 1 (AI服务集成):
  - 集成AI服务
  - ✅ 生成AI匹配测试数据（简历+职位配对）
  - ✅ 生成聊天记录数据
  - 测试AI功能

Phase 3 Week 2 (支持服务集成):
  - 集成通知、横幅、模板服务
  - ✅ 生成通知数据
  - ✅ 生成横幅数据
  - ✅ 生成模板使用数据
  - 测试支持服务

Phase 3 Week 3 (前后端联调):
  - 前端开发
  - ✅ 搭建Mock API服务
  - ✅ 生成完整业务流数据
  - ✅ 前后端联调测试
  - ✅ 端到端业务场景验证
```

---

## 📊 实施时间表

| 阶段 | 时间 | 主要任务 | 测试数据任务 | 交付物 |
|------|------|----------|------------|--------|
| **Phase 1** | Week 1 | Campaign认证服务 | ✅ 生成基础用户数据（10个） | SQL种子脚本（基础数据） |
| | Week 2 | 数据库集成 | ✅ 执行种子脚本，验证数据 | 基础数据验证报告 |
| | Week 3 | 健康检查 | ✅ 数据完整性验证 | 数据健康检查报告 |
| **Phase 2** | Week 1 | 用户+简历服务 | ✅ 生成扩展用户和简历数据 | Python生成器（用户+简历） |
| | Week 2 | 职位+公司服务 | ✅ 生成企业和职位数据 | Python生成器（企业+职位） |
| | Week 3-4 | 业务集成 | ✅ 生成完整业务流数据 | 完整业务数据验证 |
| **Phase 3** | Week 1 | AI服务集成 | ✅ 生成AI匹配数据 | AI测试数据 |
| | Week 2 | 支持服务集成 | ✅ 生成通知、横幅数据 | 支持服务数据 |
| | Week 3 | 前后端联调 | ✅ Mock服务 + 完整数据 | Mock服务 + 联调验证 |

---

## 🛠️ 分阶段实施细节

### Phase 1: 测试数据准备（与基础设施同步）

#### 任务1: SQL种子脚本（Week 1）

**文件**: `databases/postgres/init/10-seed-basic-test-data.sql`

**内容**:
```sql
-- 1. 基础用户数据（10个）
INSERT INTO zervigo_auth_users (username, email, password_hash, status, email_verified) VALUES
('test_admin', 'admin@test.com', '$2a$10$...', 'active', true),
('test_user_001', 'user001@test.com', '$2a$10$...', 'active', true),
('test_user_002', 'user002@test.com', '$2a$10$...', 'active', true),
-- ... 更多测试用户
;

-- 2. 角色和权限关联
INSERT INTO zervigo_auth_user_roles (user_id, role_id) VALUES
(1, 1), -- admin role
(2, 3), -- job_seeker role
(3, 4), -- company_admin role
-- ...
;

-- 3. 服务凭证数据（测试服务间认证）
INSERT INTO service_credentials (service_id, service_secret) VALUES
('central-brain', 'test_secret_key'),
('user-service', 'test_secret_key'),
-- ...
;
```

**执行时机**:
- Phase 1 Week 1: 认证服务启动后立即执行
- 用于测试登录和Token生成

---

#### 任务2: 数据库验证（Week 2）

**脚本**: `scripts/verify-test-data.sh`

**功能**:
- 验证测试数据完整性
- 检查数据关联性
- 验证权限配置

**执行时机**:
- Phase 1 Week 2: 数据库集成测试时

---

### Phase 2: 业务数据生成（与业务开发同步）

#### 任务3: Python数据生成器 - 用户和简历（Week 1）

**文件**: `scripts/generate-user-resume-data.py`

**功能**:
- 生成30个求职者用户
- 生成50份简历
- 建立用户-简历关联

**执行时机**:
- Phase 2 Week 1: 用户和简历服务开发时
- 支持服务功能测试

**命令**:
```bash
python scripts/generate-user-resume-data.py --users 30 --resumes 50
```

---

#### 任务4: Python数据生成器 - 企业和职位（Week 2）

**文件**: `scripts/generate-company-job-data.py`

**功能**:
- 生成10个企业用户
- 生成20个企业
- 生成100个职位
- 建立企业-职位关联

**执行时机**:
- Phase 2 Week 2: 职位和公司服务开发时
- 支持服务功能测试

**命令**:
```bash
python scripts/generate-company-job-data.py --companies 20 --jobs 100
```

---

#### 任务5: 完整业务流数据生成（Week 3-4）

**文件**: `scripts/generate-full-business-flow-data.py`

**功能**:
- 生成职位申请记录（50-100个）
- 生成AI匹配记录（30-50个）
- 生成聊天记录（100-200个）
- 验证完整业务流

**执行时机**:
- Phase 2 Week 3-4: 业务逻辑集成测试时

**命令**:
```bash
python scripts/generate-full-business-flow-data.py --applications 50 --matches 30 --chats 100
```

---

### Phase 3: 联调数据准备（与集成测试同步）

#### 任务6: AI服务测试数据（Week 1）

**文件**: `scripts/generate-ai-test-data.py`

**功能**:
- 生成简历-职位匹配对
- 生成偏离度数据
- 生成推荐理由

**执行时机**:
- Phase 3 Week 1: AI服务集成测试时

---

#### 任务7: Mock服务搭建（Week 3）

**文件**: `scripts/mock-api-server.py`

**功能**:
- 提供Mock API服务
- 返回预设测试数据
- 支持前端独立开发

**执行时机**:
- Phase 3 Week 3: 前后端联调前

---

## 🔄 协调实施流程

### 总体流程

```
Phase 1 开始
  ↓
[Week 1] 认证服务开发
  → ✅ 同步生成基础测试用户数据
  → ✅ SQL种子脚本
  ↓
[Week 2] 数据库集成
  → ✅ 执行种子脚本
  → ✅ 验证数据完整性
  ↓
[Week 3] 健康检查
  → ✅ 数据健康验证
  ↓
Phase 2 开始
  ↓
[Week 1] 用户+简历服务开发
  → ✅ 同步生成用户和简历数据
  → ✅ Python生成器（用户+简历）
  ↓
[Week 2] 职位+公司服务开发
  → ✅ 同步生成企业和职位数据
  → ✅ Python生成器（企业+职位）
  ↓
[Week 3-4] 业务集成
  → ✅ 生成完整业务流数据
  → ✅ 业务场景验证
Calculator Phase 2 完成
  ↓
Phase 3 开始
  ↓
[Week 1] AI服务集成
  → ✅ 生成AI测试数据
  ↓
[Week 2] 支持服务集成
  → ✅ 生成通知、横幅数据
  ↓
[Week 3] 前后端联调
  → ✅ Mock服务搭建
  → ✅ 完整数据验证
  → ✅ 端到端测试
  ↓
项目完成
```

---

## 📋 具体实施步骤

### Step 1: Phase 1 Week 1 - 基础数据准备

**时间**: Phase 1 Week 1 开始前

**任务**:
1. ✅ 创建SQL种子脚本
2. ✅ 生成10个基础测试用户
3. ✅ 配置角色和权限
4. ✅ 准备服务凭证数据

**交付物**:
- `databases/postgres/init/10-seed-basic-test-data.sql`

**验证**:
```bash
# 1. 执行种子脚本
psql -U postgres -d zervigo_unified -f databases/postgres/init/10-seed-basic-test-data.sql

# 2. 验证数据
psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM zervigo_auth_users;"
# 应该返回 >= 10

# 3. 测试登录
curl -X POST http://localhost:8207/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_user_001","password":"test123456"}'
```

---

### Step 2: Phase 2 Week 1 - 用户和简历数据生成

**时间**: Phase 2 Week 1 开始

**任务**:
1. ✅ 增强Python数据生成器
2. ✅ 生成30个求职者用户
3. ✅ 生成50份简历
4. ✅ 建立用户-简历关联

**交付物**:
- `scripts/generate-user-resume-data.py`
- 扩展的用户和简历数据

**验证**:
```bash
# 1. 生成数据
python scripts/generate-user-resume-data.py --users 30 --resumes 50

# 2. 验证数据
psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM zervigo_auth_users WHERE role = 'job_seeker';"
# 应该返回 >= 30

psql -U postgres -d zervigo_unified -近距离 "SELECT COUNT(*) FROM zervigo_resumes;"
# 应该返回 >= 50
```

---

### Step 3: Phase 2 Week 2 - 企业和职位数据生成

**时间**: Phase 2 Week 2 开始

**任务**:
1. ✅ 创建企业数据生成器we
2. ✅ 生成20个企业
3. ✅ 生成100个职位
4. ✅ 建立企业-职位关联

**交付物**:
- `scripts/generate-company-job-data.py`
- 企业 and职位数据

**验证**:
```bash
# 1. 生成数据
python scripts/generate-company-job-data.py --companies 20 --jobs 100

# 2. 验证数据
psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM zervigo_companies;"
# 应该返回 >= 20

psql -U postgres -d zervigo_unified -c "SELECT COUNT(*) FROM zervigo_jobs;"
# 应该返回 >= 100
```

---

### Step 4: Phase 2 Week 3-4 - 完整业务流数据

**时间**: Phase 2 Week 3-4

**任务**:
1. ✅ 生成职位申请记录
2. ✅ 生成AI匹配记录
3. ✅ 生成聊天记录
4. ✅ 验证完整业务流

**交付物**:
- `scripts/generate-full-business-flow-data.py`
- 完整的业务流数据

**验证**:
```bash
# 1. 生成数据
python scripts/generate-full-business-flow-data.py --applications 50 --matches 30 --chats 100

# 2. 验证业务流
# 测试：用户登录 → 创建简历 → 搜索职位 → 申请职位 → 查看匹配
```

---

### Step 5: Phase 3 Week 1 - AI测试数据

**时间**: Phase 3 Week 1

**任务**:
1. ✅ 生成简历-职位匹配对
2. ✅ 生成匹配分数数据
3. ✅ 生成推荐理由

**交付物**:
- `scripts/generate-ai-test-data.py`
- AI测试数据

---

### Step 6: Phase 3 Week 3 - Mock服务

**时间**: Phase 3 Week 3

**任务**:
1. ✅ 搭建Mock API服务
2. ✅ 配置Mock数据
3. ✅ 前端接入Mock服务

**交付物**:
- `scripts/mock-api-server.py`
- `frontend/src/mocks/` (Mock数据文件)

---

## 🎯 关键协调点

### 1. 数据生成时机

**原则**: 数据生成**早于**服务开发需求，但**不早于**基础设施就绪

**具体**:
- Phase 1基础数据: 在认证服务启动后立即生成
- Phase 2业务数据: 在服务开发前1-2天生成
- Phase 3联调数据: 在集成测试前生成

---

### 2. 数据量控制

**原则**: 数据量根据**当前阶段需求**逐步增加

**具体**:
- Phase 1: 少量基础数据（10个用户）
- Phase 2: 中等规模业务数据（50-100个用户，100个职位）
- Phase 3: 完整业务流数据（支持端到端测试）

---

### 3. 数据关联性

**原则**: 数据生成必须保证**关联完整性**

**具体**:
- 用户 → 简历关联
- 企业 → 职位关联
- 用户 → 申请 → 职位关联
- 简历 → 匹配 → 职位关联

---

### 4. 数据重置机制

**原则**: 支持**快速重置**和**重新生成**

**具体**:
```bash
# 重置Phase 1基础数据
psql -U postgres -d zervigo_unified -f databases/postgres/init/10-seed-basic-test-data.sql --clean

# 重置Phase 2业务数据
python scripts/generate-user-resume-data.py --reset

# 重置完整数据
./scripts/reset-test-data.sh
```

---

## ✅ 验收标准

### Phase 1验收

- [ ] 基础用户数据生成成功（>=10个）
- [ ] 角色和权限配置正确
- [ ] 服务凭证数据配置正确
- [ ] 可以成功登录并获取Token

---

### Phase 2验收

- [ ] 用户数据 >= 50个（包含不同角色）
- [ ] 企业数据 >= 20个
- [ ] 职位数据 >= 100个
- [ ] 简历数据 >= 80个
- [ ] 关联数据完整性验证通过
- [ ] 业务流测试通过

---

### Phase 3验收

- [ ] AI测试数据生成成功
- [ ] Mock服务正常运行
- [ ] 前端可以接入Mock服务
- [ ] 完整业务流端到端测试通过
- [ ] 前后端联调测试通过

---

## 📊 实施检查清单

### Phase 1检查清单

- [ ] SQL种子脚本创建完成
- [ ] 基础用户数据生成（10个）
- [ ] 角色和权限配置完成
- [ ] 服务凭证数据配置完成
- [ ] 数据验证脚本通过
- [ ] 登录测试通过

---

### Phase 2检查清单

- [ ] Python用户数据生成器完成
- [ ] Python简历数据生成器完成
- [ ] Python企业数据生成器完成
- [ ] Python职位数据生成器完成
- [ ] 完整业务流数据生成器完成
- [ ] 数据关联性验证通过
- [ ] 业务场景测试通过

---

### Phase 3检查清单

- [ ] AI测试数据生成器完成
- [ ] Mock API服务搭建完成
- [ ] Mock数据配置完成
- [ ] 前端Mock接入完成
- [ ] 完整业务流测试通过
- [ ] 前后端联调测试通过

---

## 🚀 立即行动

### 第一步：创建Phase 1基础数据（今天）

**任务**:
1. ✅ 创建SQL种子脚本
2. ✅ 生成基础测试用户
3. ✅ 配置角色和权限

**预计时间**: 2-3小时

---

### 第二步：准备Phase 2数据生成器（本周）

**任务**:
1. ✅ 设计Python数据生成器架构
2. ✅ 实现用户和简历生成器
3. ✅ 实现企业和职位生成器

**预计时间**: 1-2天

---

### 第三步：准备Phase 3 Mock服务（下周）

**任务**:
1. ✅ 设计Mock服务架构
2. ✅ 实现Mock API服务
3. ✅ 配置Mock数据

**预计时间**: 1-2天

---

## ✅ 总结

### 核心原则

1. **同步进行**: 测试数据生成与开发进度同步，不超前不滞后
2. **按需生成**: 根据当前阶段需求生成对应规模的数据
3. **关联完整**: 确保数据之间的关联关系正确
4. **快速重置**: 支持快速清理和重新生成

### 协调优势

- ✅ **不冲突**: 测试数据生成作为支撑性工作，不影响原有计划
- ✅ **有节奏**: 与三阶段计划完美配合，每个阶段都有对应数据
- ✅ **可验证**: 每个阶段都有明确的数据验收标准
- ✅ **可持续**: 支持持续集成和持续测试

---

**报告生成时间**: 2025-01-29  
**关键结论**: 测试数据生成方案与三阶段实施计划完美协调，作为支撑性工作贯穿整个开发过程

