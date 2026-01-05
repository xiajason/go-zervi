# 🎯 Zervigo 微服务数据库架构设计

## 📊 **设计理念**

### ✅ **设计时间**: 2025-10-29 09:30
### 🎯 **设计目标**: 基于前辈成果，设计适合微服务架构的数据库结构

---

## 🏗️ **微服务数据库架构原则**

### **1. 服务边界清晰**
- 每个微服务拥有自己的核心表
- 跨服务数据通过API调用，避免直接数据库访问
- 共享数据通过事件驱动或API网关处理

### **2. 数据一致性**
- 强一致性：用户认证、权限管理
- 最终一致性：业务数据、统计数据
- 事件驱动：跨服务数据同步

### **3. 性能优化**
- 读写分离：读多写少的表使用从库
- 缓存策略：热点数据使用Redis缓存
- 索引优化：基于查询模式设计索引

---

## 🔐 **auth-service-go 数据库设计**

### **核心职责**
- 用户认证和授权
- JWT Token管理
- 权限和角色管理
- 登录审计

### **数据库表设计**

#### **1. 用户认证表 (zervigo_auth_users)**
```sql
CREATE TABLE zervigo_auth_users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    
    -- 用户状态
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, suspended, deleted
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    
    -- 订阅信息 (借鉴前辈设计)
    subscription_status VARCHAR(20) DEFAULT 'free', -- free, premium, enterprise
    subscription_type VARCHAR(20) DEFAULT 'basic', -- basic, pro, enterprise
    subscription_expires_at TIMESTAMP,
    subscription_features JSONB DEFAULT '{}',
    
    -- 版本访问权限 (Zervigo特色)
    accessible_versions TEXT[] DEFAULT ARRAY['basic'], -- basic, pro, future
    version_quota JSONB DEFAULT '{"basic": 1000, "pro": 5000, "future": 10000}',
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 索引
CREATE INDEX idx_auth_users_username ON zervigo_auth_users(username);
CREATE INDEX idx_auth_users_email ON zervigo_auth_users(email);
CREATE INDEX idx_auth_users_status ON zervigo_auth_users(status);
CREATE INDEX idx_auth_users_subscription ON zervigo_auth_users(subscription_status);
```

#### **2. 角色表 (zervigo_auth_roles)**
```sql
CREATE TABLE zervigo_auth_roles (
    id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    role_description TEXT,
    role_level INTEGER DEFAULT 1, -- 1-10, 数字越大权限越高
    
    -- 版本权限
    version_access TEXT[] DEFAULT ARRAY['basic'],
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 默认角色
INSERT INTO zervigo_auth_roles (role_name, role_description, role_level, version_access) VALUES
('super_admin', '超级管理员', 10, ARRAY['basic', 'pro', 'future']),
('admin', '系统管理员', 8, ARRAY['basic', 'pro']),
('hr_manager', 'HR经理', 6, ARRAY['basic', 'pro']),
('hr_user', 'HR用户', 4, ARRAY['basic']),
('job_seeker', '求职者', 2, ARRAY['basic']),
('company_admin', '企业管理员', 5, ARRAY['basic', 'pro']),
('company_user', '企业用户', 3, ARRAY['basic']);
```

#### **3. 权限表 (zervigo_auth_permissions)**
```sql
CREATE TABLE zervigo_auth_permissions (
    id BIGSERIAL PRIMARY KEY,
    permission_name VARCHAR(100) UNIQUE NOT NULL,
    permission_code VARCHAR(100) UNIQUE NOT NULL,
    permission_description TEXT,
    
    -- 服务权限
    service_name VARCHAR(50) NOT NULL, -- user-service, job-service, etc.
    resource_type VARCHAR(50) NOT NULL, -- user, job, company, resume
    action VARCHAR(20) NOT NULL, -- create, read, update, delete, list
    
    -- 版本权限
    version_access TEXT[] DEFAULT ARRAY['basic'],
    
    -- 状态
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 默认权限
INSERT INTO zervigo_auth_permissions (permission_name, permission_code, service_name, resource_type, action, version_access) VALUES
('用户管理-创建', 'user:create', 'user-service', 'user', 'create', ARRAY['basic', 'pro', 'future']),
('用户管理-读取', 'user:read', 'user-service', 'user', 'read', ARRAY['basic', 'pro', 'future']),
('用户管理-更新', 'user:update', 'user-service', 'user', 'update', ARRAY['basic', 'pro', 'future']),
('用户管理-删除', 'user:delete', 'user-service', 'user', 'delete', ARRAY['pro', 'future']),
('职位管理-创建', 'job:create', 'job-service', 'job', 'create', ARRAY['basic', 'pro', 'future']),
('职位管理-读取', 'job:read', 'job-service', 'job', 'read', ARRAY['basic', 'pro', 'future']),
('职位管理-更新', 'job:update', 'job-service', 'job', 'update', ARRAY['basic', 'pro', 'future']),
('职位管理-删除', 'job:delete', 'job-service', 'job', 'delete', ARRAY['pro', 'future']);
```

#### **4. 用户角色关联表 (zervigo_auth_user_roles)**
```sql
CREATE TABLE zervigo_auth_user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    role_id BIGINT REFERENCES zervigo_auth_roles(id) ON DELETE CASCADE,
    
    -- 分配信息
    assigned_by BIGINT REFERENCES zervigo_auth_users(id),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP, -- 角色过期时间
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    
    UNIQUE(user_id, role_id)
);
```

#### **5. 角色权限关联表 (zervigo_auth_role_permissions)**
```sql
CREATE TABLE zervigo_auth_role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES zervigo_auth_roles(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES zervigo_auth_permissions(id) ON DELETE CASCADE,
    
    -- 分配信息
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(role_id, permission_id)
);
```

#### **6. JWT Token管理表 (zervigo_auth_tokens)**
```sql
CREATE TABLE zervigo_auth_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES zervigo_auth_users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(20) DEFAULT 'access', -- access, refresh
    
    -- Token信息
    expires_at TIMESTAMP NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 客户端信息
    client_ip INET,
    user_agent TEXT,
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active', -- active, revoked, expired
    revoked_at TIMESTAMP,
    revoked_reason VARCHAR(100)
);

-- 索引
CREATE INDEX idx_auth_tokens_user_id ON zervigo_auth_tokens(user_id);
CREATE INDEX idx_auth_tokens_expires_at ON zervigo_auth_tokens(expires_at);
CREATE INDEX idx_auth_tokens_status ON zervigo_auth_tokens(status);
```

#### **7. 登录审计表 (zervigo_auth_login_logs)**
```sql
CREATE TABLE zervigo_auth_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES zervigo_auth_users(id) ON DELETE SET NULL,
    username VARCHAR(50),
    
    -- 登录信息
    login_method VARCHAR(20) NOT NULL, -- password, sms, email, oauth
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    
    -- 客户端信息
    client_ip INET,
    user_agent TEXT,
    device_info JSONB,
    
    -- 位置信息
    country VARCHAR(50),
    city VARCHAR(50),
    
    -- 时间
    login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_auth_login_logs_user_id ON zervigo_auth_login_logs(user_id);
CREATE INDEX idx_auth_login_logs_login_at ON zervigo_auth_login_logs(login_at);
CREATE INDEX idx_auth_login_logs_success ON zervigo_auth_login_logs(success);
```

---

## 👤 **user-service 数据库设计**

### **核心职责**
- 用户基本信息管理
- 用户档案和偏好设置
- 用户行为统计
- 用户关系管理

### **数据库表设计**

#### **1. 用户档案表 (zervigo_user_profiles)**
```sql
CREATE TABLE zervigo_user_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL, -- 关联auth-service的用户ID
    
    -- 基本信息
    real_name VARCHAR(50),
    nickname VARCHAR(50),
    avatar_url VARCHAR(500),
    gender INTEGER DEFAULT 0, -- 0:未知, 1:男, 2:女
    birthday DATE,
    
    -- 联系信息
    phone VARCHAR(20),
    email VARCHAR(100),
    wechat VARCHAR(50),
    qq VARCHAR(20),
    
    -- 地址信息
    country VARCHAR(50) DEFAULT '中国',
    province VARCHAR(50),
    city VARCHAR(50),
    district VARCHAR(50),
    address TEXT,
    
    -- 职业信息
    current_position VARCHAR(100),
    current_company VARCHAR(100),
    work_experience INTEGER DEFAULT 0, -- 工作年限
    education_level VARCHAR(20), -- 学历
    
    -- 偏好设置
    job_preferences JSONB DEFAULT '{}', -- 职位偏好
    location_preferences TEXT[], -- 地点偏好
    salary_expectation JSONB DEFAULT '{}', -- 薪资期望
    
    -- 状态
    profile_completeness INTEGER DEFAULT 0, -- 档案完整度 0-100
    is_public BOOLEAN DEFAULT FALSE, -- 是否公开档案
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_user_profiles_user_id ON zervigo_user_profiles(user_id);
CREATE INDEX idx_user_profiles_city ON zervigo_user_profiles(city);
CREATE INDEX idx_user_profiles_position ON zervigo_user_profiles(current_position);
```

#### **2. 用户技能表 (zervigo_user_skills)**
```sql
CREATE TABLE zervigo_user_skills (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    skill_name VARCHAR(100) NOT NULL,
    skill_category VARCHAR(50), -- 技术, 管理, 语言, 其他
    proficiency_level INTEGER DEFAULT 1, -- 1-5熟练度
    years_of_experience INTEGER DEFAULT 0,
    
    -- 验证信息
    verified BOOLEAN DEFAULT FALSE,
    verified_by VARCHAR(50), -- 验证方式
    verified_at TIMESTAMP,
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_user_skills_user_id ON zervigo_user_skills(user_id);
CREATE INDEX idx_user_skills_category ON zervigo_user_skills(skill_category);
```

#### **3. 用户教育经历表 (zervigo_user_education)**
```sql
CREATE TABLE zervigo_user_education (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- 学校信息
    school_name VARCHAR(200) NOT NULL,
    school_type VARCHAR(50), -- 985, 211, 普通本科, 专科, 其他
    major VARCHAR(100),
    degree VARCHAR(50), -- 博士, 硕士, 本科, 专科, 高中
    
    -- 时间信息
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE, -- 是否在读
    
    -- 成绩信息
    gpa DECIMAL(3,2),
    ranking VARCHAR(50), -- 排名信息
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_user_education_user_id ON zervigo_user_education(user_id);
CREATE INDEX idx_user_education_school ON zervigo_user_education(school_name);
```

#### **4. 用户工作经历表 (zervigo_user_experience)**
```sql
CREATE TABLE zervigo_user_experience (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- 公司信息
    company_name VARCHAR(200) NOT NULL,
    company_industry VARCHAR(100),
    company_size VARCHAR(50), -- 公司规模
    
    -- 职位信息
    position VARCHAR(100) NOT NULL,
    department VARCHAR(100),
    job_level VARCHAR(50), -- 级别
    
    -- 时间信息
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE, -- 是否当前工作
    
    -- 工作内容
    job_description TEXT,
    achievements TEXT, -- 工作成就
    skills_used TEXT[], -- 使用的技能
    
    -- 薪资信息
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency VARCHAR(10) DEFAULT 'CNY',
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_user_experience_user_id ON zervigo_user_experience(user_id);
CREATE INDEX idx_user_experience_company ON zervigo_user_experience(company_name);
CREATE INDEX idx_user_experience_position ON zervigo_user_experience(position);
```

#### **5. 用户行为统计表 (zervigo_user_statistics)**
```sql
CREATE TABLE zervigo_user_statistics (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    
    -- 活跃度统计
    login_count INTEGER DEFAULT 0,
    last_login_at TIMESTAMP,
    active_days INTEGER DEFAULT 0, -- 活跃天数
    
    -- 求职统计
    job_view_count INTEGER DEFAULT 0,
    job_apply_count INTEGER DEFAULT 0,
    resume_view_count INTEGER DEFAULT 0,
    resume_share_count INTEGER DEFAULT 0,
    
    -- AI使用统计
    ai_chat_count INTEGER DEFAULT 0,
    ai_analysis_count INTEGER DEFAULT 0,
    ai_quota_used INTEGER DEFAULT 0,
    
    -- 社交统计
    follow_count INTEGER DEFAULT 0, -- 关注数
    follower_count INTEGER DEFAULT 0, -- 粉丝数
    connection_count INTEGER DEFAULT 0, -- 连接数
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_user_statistics_user_id ON zervigo_user_statistics(user_id);
```

---

## 💼 **job-service 数据库设计**

### **核心职责**
- 职位信息管理
- 职位发布和更新
- 职位搜索和筛选
- 职位申请管理

### **数据库表设计**

#### **1. 职位表 (zervigo_jobs)**
```sql
CREATE TABLE zervigo_jobs (
    id BIGSERIAL PRIMARY KEY,
    
    -- 基本信息
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    requirements TEXT,
    responsibilities TEXT, -- 工作职责
    
    -- 公司信息
    company_id BIGINT NOT NULL, -- 关联company-service
    company_name VARCHAR(200), -- 冗余字段，提高查询性能
    company_logo VARCHAR(500),
    
    -- 职位分类
    job_category VARCHAR(100), -- 技术, 销售, 市场, 运营, 其他
    job_subcategory VARCHAR(100), -- 前端, 后端, 全栈, 等
    job_level VARCHAR(50), -- 初级, 中级, 高级, 专家
    
    -- 工作信息
    work_type VARCHAR(20) DEFAULT 'full-time', -- full-time, part-time, contract, intern
    work_location VARCHAR(200),
    work_address TEXT,
    remote_allowed BOOLEAN DEFAULT FALSE,
    
    -- 薪资信息
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency VARCHAR(10) DEFAULT 'CNY',
    salary_period VARCHAR(20) DEFAULT 'monthly', -- monthly, yearly, hourly
    salary_negotiable BOOLEAN DEFAULT TRUE,
    
    -- 要求信息
    experience_required VARCHAR(50), -- 经验要求
    education_required VARCHAR(50), -- 学历要求
    skills_required TEXT[], -- 技能要求
    languages_required TEXT[], -- 语言要求
    
    -- 福利信息
    benefits TEXT[], -- 福利待遇
    perks TEXT[], -- 额外福利
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'draft', -- draft, published, paused, closed, expired
    is_featured BOOLEAN DEFAULT FALSE, -- 是否推荐
    is_urgent BOOLEAN DEFAULT FALSE, -- 是否紧急
    
    -- 统计信息
    view_count INTEGER DEFAULT 0,
    apply_count INTEGER DEFAULT 0,
    favorite_count INTEGER DEFAULT 0,
    
    -- 时间信息
    publish_at TIMESTAMP,
    expire_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT NOT NULL -- 创建者ID
);

-- 索引
CREATE INDEX idx_jobs_company_id ON zervigo_jobs(company_id);
CREATE INDEX idx_jobs_title ON zervigo_jobs(title);
CREATE INDEX idx_jobs_location ON zervigo_jobs(work_location);
CREATE INDEX idx_jobs_category ON zervigo_jobs(job_category);
CREATE INDEX idx_jobs_status ON zervigo_jobs(status);
CREATE INDEX idx_jobs_salary ON zervigo_jobs(salary_min, salary_max);
CREATE INDEX idx_jobs_publish_at ON zervigo_jobs(publish_at);
CREATE INDEX idx_jobs_featured ON zervigo_jobs(is_featured);
```

#### **2. 职位申请表 (zervigo_job_applications)**
```sql
CREATE TABLE zervigo_job_applications (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT REFERENCES zervigo_jobs(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL, -- 关联auth-service的用户ID
    
    -- 申请信息
    resume_id BIGINT, -- 关联resume-service的简历ID
    cover_letter TEXT, -- 求职信
    application_source VARCHAR(50) DEFAULT 'web', -- web, mobile, api
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'pending', -- pending, reviewing, interviewed, offered, rejected, withdrawn
    application_stage VARCHAR(50), -- 申请阶段
    
    -- 处理信息
    reviewed_by BIGINT, -- 审核人ID
    reviewed_at TIMESTAMP,
    review_notes TEXT, -- 审核备注
    
    -- 面试信息
    interview_scheduled_at TIMESTAMP,
    interview_location VARCHAR(200),
    interview_notes TEXT,
    
    -- 结果信息
    offer_salary INTEGER,
    offer_start_date DATE,
    offer_expires_at TIMESTAMP,
    
    -- 时间戳
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(job_id, user_id) -- 一个用户只能申请一次同一个职位
);

-- 索引
CREATE INDEX idx_job_applications_job_id ON zervigo_job_applications(job_id);
CREATE INDEX idx_job_applications_user_id ON zervigo_job_applications(user_id);
CREATE INDEX idx_job_applications_status ON zervigo_job_applications(status);
CREATE INDEX idx_job_applications_applied_at ON zervigo_job_applications(applied_at);
```

#### **3. 职位收藏表 (zervigo_job_favorites)**
```sql
CREATE TABLE zervigo_job_favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    job_id BIGINT REFERENCES zervigo_jobs(id) ON DELETE CASCADE,
    
    -- 收藏信息
    favorite_type VARCHAR(20) DEFAULT 'favorite', -- favorite, bookmark, interested
    notes TEXT, -- 收藏备注
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, job_id)
);

-- 索引
CREATE INDEX idx_job_favorites_user_id ON zervigo_job_favorites(user_id);
CREATE INDEX idx_job_favorites_job_id ON zervigo_job_favorites(job_id);
```

#### **4. 职位搜索历史表 (zervigo_job_search_history)**
```sql
CREATE TABLE zervigo_job_search_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- 搜索条件
    search_keywords TEXT,
    search_filters JSONB, -- 搜索筛选条件
    search_location VARCHAR(200),
    
    -- 搜索结果
    result_count INTEGER DEFAULT 0,
    clicked_job_ids BIGINT[], -- 点击的职位ID列表
    
    -- 时间戳
    searched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_job_search_history_user_id ON zervigo_job_search_history(user_id);
CREATE INDEX idx_job_search_history_searched_at ON zervigo_job_search_history(searched_at);
```

---

## 🔗 **跨服务数据同步策略**

### **1. 事件驱动同步**
```yaml
用户注册事件:
  - auth-service: 创建用户认证信息
  - user-service: 创建用户档案
  - 通过消息队列异步同步

职位发布事件:
  - job-service: 创建职位信息
  - company-service: 更新公司职位统计
  - ai-service: 更新职位向量
```

### **2. API调用同步**
```yaml
实时数据获取:
  - job-service 通过API获取公司信息
  - user-service 通过API获取用户认证状态
  - resume-service 通过API获取用户档案
```

### **3. 缓存策略**
```yaml
Redis缓存:
  - 用户基本信息缓存 (TTL: 1小时)
  - 职位列表缓存 (TTL: 30分钟)
  - 公司信息缓存 (TTL: 2小时)
  - 权限信息缓存 (TTL: 10分钟)
```

---

## 🎯 **总结**

### **设计亮点**
1. **借鉴前辈成果**: 基于成熟的业务模型设计
2. **微服务适配**: 每个服务有清晰的边界和职责
3. **Zervigo特色**: 版本管理、订阅系统、AI集成
4. **性能优化**: 合理的索引设计和缓存策略
5. **扩展性**: 支持未来功能扩展

### **下一步行动**
1. **创建数据库迁移脚本**
2. **实现服务间API调用**
3. **配置Redis缓存策略**
4. **实现事件驱动同步**

**🎯 这样我们就能在前辈成果的基础上，构建出适合微服务架构的Zervigo系统！**
