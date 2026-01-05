# Go-Zervi框架业务功能实现计划

## 📋 基于学习文档的业务实现策略

基于对 `/Users/szjason72/study/szbolent/zervi.test/docs` 的深入学习，我们将在Go-Zervi框架中实现完整的简历隐私保护、权限管理和积分系统。

## 🎯 核心业务功能实现计划

### 第一阶段：简历隐私保护系统 (1-2周)

#### 1. 简历权限管理API

**实现目标**: 基于学习文档实现完整的简历隐私保护机制

```yaml
核心API:
  - GET /api/v1/resume/permission/{resumeId}     # 获取简历权限配置
  - PUT /api/v1/resume/permission/{resumeId}     # 更新简历权限配置
  - GET /api/v1/resume/blacklist/{resumeId}      # 获取黑名单
  - PUT /api/v1/resume/blacklist/{resumeId}      # 更新黑名单

隐私级别控制:
  - PUBLIC: 公开简历，所有企业可查看
  - PRIVATE: 私密简历，仅指定企业可查看
  - FRIENDS: 好友可见，需要审批

权限控制:
  - allowDownload: 是否允许下载
  - requireApproval: 是否需要审批
  - allowedEnterprises: 允许的企业ID列表
  - deniedEnterprises: 禁止的企业ID列表
```

#### 2. 数据库表结构设计

**基于学习文档的完整表结构**:

```sql
-- 简历权限表
CREATE TABLE resume_permission (
    id BIGSERIAL PRIMARY KEY,
    resume_id VARCHAR(50) UNIQUE NOT NULL,
    privacy_level VARCHAR(20) DEFAULT 'PRIVATE',
    allow_download BOOLEAN DEFAULT false,
    require_approval BOOLEAN DEFAULT true,
    allowed_enterprises TEXT, -- JSON数组
    denied_enterprises TEXT, -- JSON数组
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 简历黑名单表
CREATE TABLE resume_blacklist (
    id BIGSERIAL PRIMARY KEY,
    resume_id VARCHAR(50) NOT NULL,
    enterprise_id VARCHAR(50) NOT NULL,
    enterprise_name VARCHAR(100) NOT NULL,
    reason VARCHAR(255),
    add_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 审批记录表
CREATE TABLE approve_record (
    approve_id VARCHAR(50) PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- 审批类型：简历查看、简历下载、简历收藏
    user_id BIGINT NOT NULL,
    enterprise_id VARCHAR(50) NOT NULL,
    enterprise_name VARCHAR(100) NOT NULL,
    resume_id VARCHAR(50) NOT NULL,
    resume_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT '待审批', -- 待审批、已通过、已拒绝
    cost INT DEFAULT 0, -- 积分消耗
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    handle_time TIMESTAMP
);

-- 积分账单表
CREATE TABLE points_bill (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL, -- 收入、支出
    amount INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    balance INT NOT NULL, -- 余额
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 第二阶段：权限管理系统 (1-2周)

#### 1. 角色权限管理API

```yaml
角色管理:
  - GET /api/v1/roles/                    # 获取角色列表
  - POST /api/v1/roles/                    # 创建角色
  - PUT /api/v1/roles/{id}                 # 更新角色
  - DELETE /api/v1/roles/{id}              # 删除角色

权限管理:
  - GET /api/v1/permissions/               # 获取权限列表
  - POST /api/v1/permissions/              # 创建权限
  - PUT /api/v1/permissions/{id}           # 更新权限
  - DELETE /api/v1/permissions/{id}        # 删除权限

用户角色分配:
  - GET /api/v1/users/{id}/roles           # 获取用户角色
  - POST /api/v1/users/{id}/roles          # 分配角色
  - DELETE /api/v1/users/{id}/roles/{roleId} # 移除角色
```

#### 2. 基于RBAC的权限控制

```go
// 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id")
        if !hasPermission(userID, permission) {
            c.JSON(http.StatusForbidden, gin.H{
                "code": 403,
                "message": "权限不足",
                "error_code": "INSUFFICIENT_PERMISSION",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

// 简历权限检查
func CheckResumePermission(resumeID string, userID uint, action string) bool {
    // 检查用户是否有权限对简历执行指定操作
    // 考虑隐私级别、审批状态、黑名单等因素
}
```

### 第三阶段：积分系统和审批流程 (1-2周)

#### 1. 积分管理API

```yaml
积分查询:
  - GET /api/v1/points/user/{userId}       # 获取用户积分
  - GET /api/v1/points/user/{userId}/balance # 获取积分余额
  - GET /api/v1/points/user/{userId}/bill   # 获取积分账单

积分操作:
  - POST /api/v1/points/award               # 奖励积分
  - POST /api/v1/points/deduct              # 扣除积分
  - POST /api/v1/points/freeze              # 冻结积分
  - POST /api/v1/points/unfreeze            # 解冻积分
```

#### 2. 审批流程API

```yaml
审批管理:
  - GET /api/v1/approve/list                # 获取审批列表
  - POST /api/v1/approve/handle/{approveId} # 处理审批
  - GET /api/v1/approve/history             # 获取审批历史

审批类型:
  - 简历查看: 企业查看简历需要用户审批
  - 简历下载: 企业下载简历需要用户审批
  - 简历收藏: 企业收藏简历需要用户审批
```

## 🛠️ 技术实现方案

### 1. 服务架构设计

```yaml
简历服务 (resume-service):
  端口: 8085
  功能:
    - 简历CRUD操作
    - 简历权限管理
    - 简历黑名单管理
    - 简历模板管理

权限服务 (permission-service):
  端口: 8086
  功能:
    - 角色管理
    - 权限管理
    - 用户角色分配
    - 权限检查

积分服务 (points-service):
  端口: 8087
  功能:
    - 积分管理
    - 积分账单
    - 积分奖励/扣除
    - 积分冻结/解冻

审批服务 (approve-service):
  端口: 8088
  功能:
    - 审批流程管理
    - 审批历史记录
    - 审批通知
    - 审批统计
```

### 2. 数据库设计

```yaml
核心表:
  - zervigo_auth_users: 用户表
  - zervigo_auth_roles: 角色表
  - zervigo_auth_permissions: 权限表
  - zervigo_auth_user_roles: 用户角色关联表
  - zervigo_auth_role_permissions: 角色权限关联表

业务表:
  - resume: 简历表
  - resume_permission: 简历权限表
  - resume_blacklist: 简历黑名单表
  - approve_record: 审批记录表
  - points_bill: 积分账单表
  - view_history: 查看历史表
```

### 3. API响应格式标准化

```go
// 标准响应格式
type ApiResponse struct {
    Code      int         `json:"code"`      // 0表示成功
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    ErrorCode string      `json:"error_code,omitempty"`
    Timestamp int64       `json:"timestamp"`
}

// 分页响应格式
type PageResponse struct {
    List     interface{} `json:"list"`      // 必须是list
    Total    int64       `json:"total"`
    PageNum  int         `json:"pageNum"`
    PageSize int         `json:"pageSize"`
    Pages    int         `json:"pages,omitempty"`
}
```

## 🎯 实现优先级

### P0 - 必须立即实现 (影响核心功能)

1. **简历权限管理API**
   - 简历权限配置
   - 黑名单管理
   - 隐私级别控制

2. **权限检查中间件**
   - 基于RBAC的权限验证
   - 简历访问权限控制
   - API访问权限控制

### P1 - 应该尽快实现 (提升用户体验)

1. **积分管理系统**
   - 积分查询和操作
   - 积分账单记录
   - 积分奖励机制

2. **审批流程系统**
   - 审批申请和处理
   - 审批历史记录
   - 审批通知机制

### P2 - 可以后期优化 (锦上添花)

1. **高级权限功能**
   - 动态权限分配
   - 权限继承
   - 权限审计

2. **积分系统扩展**
   - 积分等级系统
   - 积分兑换商城
   - 积分排行榜

## 📊 预期成果

### 第一阶段完成后

- ✅ 完整的简历隐私保护机制
- ✅ 基于RBAC的权限管理系统
- ✅ 标准化的API响应格式
- ✅ 完整的数据库表结构

### 第二阶段完成后

- ✅ 积分管理系统
- ✅ 审批流程系统
- ✅ 用户角色管理
- ✅ 权限分配机制

### 第三阶段完成后

- ✅ 完整的业务功能
- ✅ 端到端的测试验证
- ✅ 性能优化和监控
- ✅ 文档和部署指南

## 🚀 下一步行动

1. **立即开始**: 实现简历权限管理API
2. **并行开发**: 权限管理系统和积分系统
3. **集成测试**: 端到端功能验证
4. **性能优化**: 数据库优化和缓存机制

---

**总结**: 基于学习文档，我们完全可以在Go-Zervi框架中实现完整的简历隐私保护、权限管理和积分系统。这将是一个功能完整、架构清晰、安全可靠的微服务系统。
