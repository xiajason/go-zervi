# 基于PostgreSQL RLS的权限管理方案

## 问题分析

当前core中的数据库管理器存在以下问题：
1. **缺乏权限感知**：没有基于用户角色的权限控制机制
2. **数据库特性未充分利用**：PostgreSQL的RLS、角色系统等高级特性未被使用
3. **MySQL兼容性不足**：没有为MySQL提供等效的权限控制方案
4. **缺乏动态权限**：无法根据用户角色动态调整数据访问权限

## 解决方案：增强数据库管理器

### 1. 数据库类型检测和特性支持

```go
// DatabaseType 数据库类型
type DatabaseType string

const (
    DatabaseTypePostgreSQL DatabaseType = "postgresql"
    DatabaseTypeMySQL      DatabaseType = "mysql"
    DatabaseTypeRedis      DatabaseType = "redis"
    DatabaseTypeNeo4j      DatabaseType = "neo4j"
)

// DatabaseCapabilities 数据库能力
type DatabaseCapabilities struct {
    SupportsRLS           bool // 行级安全
    SupportsRoles         bool // 角色系统
    SupportsViews         bool // 视图
    SupportsFunctions     bool // 存储过程/函数
    SupportsTriggers      bool // 触发器
    SupportsJSON          bool // JSON支持
    SupportsArray         bool // 数组支持
    SupportsVector        bool // 向量搜索
}

// GetCapabilities 获取数据库能力
func (dm *Manager) GetCapabilities(dbType DatabaseType) DatabaseCapabilities {
    switch dbType {
    case DatabaseTypePostgreSQL:
        return DatabaseCapabilities{
            SupportsRLS:       true,
            SupportsRoles:     true,
            SupportsViews:     true,
            SupportsFunctions: true,
            SupportsTriggers:  true,
            SupportsJSON:      true,
            SupportsArray:     true,
            SupportsVector:    true,
        }
    case DatabaseTypeMySQL:
        return DatabaseCapabilities{
            SupportsRLS:       false,
            SupportsRoles:     true,  // MySQL 8.0+
            SupportsViews:     true,
            SupportsFunctions: true,
            SupportsTriggers:  true,
            SupportsJSON:      true,
            SupportsArray:     false,
            SupportsVector:    false,
        }
    default:
        return DatabaseCapabilities{}
    }
}
```

### 2. 权限感知的数据库管理器

```go
// PermissionAwareManager 权限感知的数据库管理器
type PermissionAwareManager struct {
    *Manager
    currentUser *UserContext
    dbType      DatabaseType
    capabilities DatabaseCapabilities
}

// UserContext 用户上下文
type UserContext struct {
    UserID   uint
    Username string
    Roles    []string
    Permissions []string
    DatabaseRole string // 数据库角色名
}

// NewPermissionAwareManager 创建权限感知的数据库管理器
func (dm *Manager) NewPermissionAwareManager(user *UserContext) *PermissionAwareManager {
    // 检测主要数据库类型
    var dbType DatabaseType
    if dm.PostgreSQL != nil {
        dbType = DatabaseTypePostgreSQL
    } else if dm.MySQL != nil {
        dbType = DatabaseTypeMySQL
    }
    
    return &PermissionAwareManager{
        Manager:     dm,
        currentUser: user,
        dbType:      dbType,
        capabilities: dm.GetCapabilities(dbType),
    }
}

// GetPermissionAwareDB 获取权限感知的数据库连接
func (pam *PermissionAwareManager) GetPermissionAwareDB() (*gorm.DB, error) {
    switch pam.dbType {
    case DatabaseTypePostgreSQL:
        return pam.getPostgreSQLWithRLS()
    case DatabaseTypeMySQL:
        return pam.getMySQLWithPermissions()
    default:
        return nil, fmt.Errorf("不支持的数据库类型: %s", pam.dbType)
    }
}

// getPostgreSQLWithRLS 获取带RLS的PostgreSQL连接
func (pam *PermissionAwareManager) getPostgreSQLWithRLS() (*gorm.DB, error) {
    if pam.Manager.PostgreSQL == nil {
        return nil, fmt.Errorf("PostgreSQL未初始化")
    }
    
    // 设置数据库角色
    if err := pam.setPostgreSQLRole(); err != nil {
        return nil, fmt.Errorf("设置PostgreSQL角色失败: %w", err)
    }
    
    return pam.Manager.PostgreSQL.GetDB(), nil
}

// getMySQLWithPermissions 获取带权限控制的MySQL连接
func (pam *PermissionAwareManager) getMySQLWithPermissions() (*gorm.DB, error) {
    if pam.Manager.MySQL == nil {
        return nil, fmt.Errorf("MySQL未初始化")
    }
    
    // MySQL通过应用层权限控制
    db := pam.Manager.MySQL.GetDB()
    
    // 设置用户上下文到GORM的Session中
    session := db.Session(&gorm.Session{
        Context: context.WithValue(context.Background(), "user_context", pam.currentUser),
    })
    
    return session, nil
}

// setPostgreSQLRole 设置PostgreSQL角色
func (pam *PermissionAwareManager) setPostgreSQLRole() error {
    if pam.currentUser.DatabaseRole == "" {
        pam.currentUser.DatabaseRole = "zervigo_user"
    }
    
    // 执行SET ROLE命令
    sql := fmt.Sprintf("SET ROLE %s", pam.currentUser.DatabaseRole)
    return pam.Manager.PostgreSQL.Exec(sql)
}
```

### 3. PostgreSQL RLS策略管理

```go
// RLSPolicy RLS策略
type RLSPolicy struct {
    TableName    string
    PolicyName   string
    PolicyType   string // "SELECT", "INSERT", "UPDATE", "DELETE", "ALL"
    PolicyExpression string
    IsEnabled    bool
}

// CreateRLSPolicy 创建RLS策略
func (pam *PermissionAwareManager) CreateRLSPolicy(policy RLSPolicy) error {
    if !pam.capabilities.SupportsRLS {
        return fmt.Errorf("当前数据库不支持RLS")
    }
    
    sql := fmt.Sprintf(`
        CREATE POLICY %s ON %s
        FOR %s
        TO %s
        USING (%s)
    `, policy.PolicyName, policy.TableName, policy.PolicyType, 
       pam.currentUser.DatabaseRole, policy.PolicyExpression)
    
    return pam.Manager.PostgreSQL.Exec(sql)
}

// EnableRLS 启用表的RLS
func (pam *PermissionAwareManager) EnableRLS(tableName string) error {
    if !pam.capabilities.SupportsRLS {
        return fmt.Errorf("当前数据库不支持RLS")
    }
    
    sql := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", tableName)
    return pam.Manager.PostgreSQL.Exec(sql)
}

// CreateRoleBasedView 创建基于角色的视图
func (pam *PermissionAwareManager) CreateRoleBasedView(viewName, baseTable, whereClause string) error {
    if !pam.capabilities.SupportsViews {
        return fmt.Errorf("当前数据库不支持视图")
    }
    
    sql := fmt.Sprintf(`
        CREATE OR REPLACE VIEW %s AS
        SELECT * FROM %s
        WHERE %s
    `, viewName, baseTable, whereClause)
    
    return pam.Manager.PostgreSQL.Exec(sql)
}
```

### 4. MySQL权限控制方案

```go
// MySQLPermissionFilter MySQL权限过滤器
type MySQLPermissionFilter struct {
    UserContext *UserContext
    TableName   string
    WhereClause string
}

// ApplyPermissionFilter 应用权限过滤
func (pam *PermissionAwareManager) ApplyPermissionFilter(query *gorm.DB, tableName string) *gorm.DB {
    if pam.dbType != DatabaseTypeMySQL {
        return query
    }
    
    // 根据用户角色和权限添加WHERE条件
    whereClause := pam.buildPermissionWhereClause(tableName)
    if whereClause != "" {
        query = query.Where(whereClause)
    }
    
    return query
}

// buildPermissionWhereClause 构建权限WHERE子句
func (pam *PermissionAwareManager) buildPermissionWhereClause(tableName string) string {
    // 根据表名和用户权限构建WHERE条件
    switch tableName {
    case "resumes":
        if pam.hasPermission("resume:view_all") {
            return "" // 可以查看所有简历
        } else if pam.hasPermission("resume:view_own") {
            return "user_id = ?" // 只能查看自己的简历
        } else if pam.hasPermission("resume:view_enterprise") {
            return "enterprise_id IN (?)" // 只能查看企业相关简历
        }
    case "approvals":
        if pam.hasPermission("approve:view_all") {
            return ""
        } else if pam.hasPermission("approve:view_assigned") {
            return "assigned_to = ?"
        }
    }
    
    return "1=0" // 默认无权限
}

// hasPermission 检查是否有特定权限
func (pam *PermissionAwareManager) hasPermission(permission string) bool {
    for _, p := range pam.currentUser.Permissions {
        if p == permission {
            return true
        }
    }
    return false
}
```

### 5. 统一的权限感知查询接口

```go
// PermissionAwareQuery 权限感知查询
type PermissionAwareQuery struct {
    db     *gorm.DB
    manager *PermissionAwareManager
}

// Find 权限感知的查找
func (paq *PermissionAwareQuery) Find(dest interface{}, tableName string, conds ...interface{}) error {
    query := paq.db.Table(tableName)
    
    // 应用权限过滤
    query = paq.manager.ApplyPermissionFilter(query, tableName)
    
    // 添加其他条件
    if len(conds) > 0 {
        query = query.Where(conds[0], conds[1:]...)
    }
    
    return query.Find(dest).Error
}

// First 权限感知的查找第一条
func (paq *PermissionAwareQuery) First(dest interface{}, tableName string, conds ...interface{}) error {
    query := paq.db.Table(tableName)
    
    // 应用权限过滤
    query = paq.manager.ApplyPermissionFilter(query, tableName)
    
    // 添加其他条件
    if len(conds) > 0 {
        query = query.Where(conds[0], conds[1:]...)
    }
    
    return query.First(dest).Error
}

// Create 权限感知的创建
func (paq *PermissionAwareQuery) Create(value interface{}, tableName string) error {
    // 检查创建权限
    if !paq.manager.hasCreatePermission(tableName) {
        return fmt.Errorf("无权限创建 %s", tableName)
    }
    
    return paq.db.Table(tableName).Create(value).Error
}

// Update 权限感知的更新
func (paq *PermissionAwareQuery) Update(value interface{}, tableName string, conds ...interface{}) error {
    // 检查更新权限
    if !paq.manager.hasUpdatePermission(tableName) {
        return fmt.Errorf("无权限更新 %s", tableName)
    }
    
    query := paq.db.Table(tableName)
    
    // 应用权限过滤
    query = paq.manager.ApplyPermissionFilter(query, tableName)
    
    // 添加其他条件
    if len(conds) > 0 {
        query = query.Where(conds[0], conds[1:]...)
    }
    
    return query.Updates(value).Error
}

// Delete 权限感知的删除
func (paq *PermissionAwareQuery) Delete(value interface{}, tableName string, conds ...interface{}) error {
    // 检查删除权限
    if !paq.manager.hasDeletePermission(tableName) {
        return fmt.Errorf("无权限删除 %s", tableName)
    }
    
    query := paq.db.Table(tableName)
    
    // 应用权限过滤
    query = paq.manager.ApplyPermissionFilter(query, tableName)
    
    // 添加其他条件
    if len(conds) > 0 {
        query = query.Where(conds[0], conds[1:]...)
    }
    
    return query.Delete(value).Error
}
```

### 6. 使用示例

```go
// 在服务中使用权限感知的数据库管理器
func (s *ResumeService) GetResumeList(userID uint) ([]Resume, error) {
    // 获取用户上下文
    userCtx := s.getUserContext(userID)
    
    // 创建权限感知的数据库管理器
    pam := s.dbManager.NewPermissionAwareManager(userCtx)
    
    // 获取权限感知的数据库连接
    db, err := pam.GetPermissionAwareDB()
    if err != nil {
        return nil, err
    }
    
    // 创建权限感知查询
    paq := &PermissionAwareQuery{
        db:     db,
        manager: pam,
    }
    
    // 执行权限感知查询
    var resumes []Resume
    err = paq.Find(&resumes, "resumes")
    if err != nil {
        return nil, err
    }
    
    return resumes, nil
}
```

## 优势总结

1. **数据库特性充分利用**：PostgreSQL使用RLS，MySQL使用应用层权限控制
2. **统一接口**：无论底层是PostgreSQL还是MySQL，都提供相同的权限感知接口
3. **性能优化**：PostgreSQL在数据库层面过滤，MySQL在应用层面优化
4. **安全性高**：PostgreSQL的RLS提供数据库级别的安全保障
5. **向后兼容**：保持对现有代码的兼容性
6. **灵活扩展**：可以轻松添加新的权限控制策略

这个方案既发挥了PostgreSQL的优势，又保持了对MySQL的兼容性，您觉得如何？
