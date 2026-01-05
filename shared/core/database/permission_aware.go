package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

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
	SupportsRLS       bool // 行级安全
	SupportsRoles     bool // 角色系统
	SupportsViews     bool // 视图
	SupportsFunctions bool // 存储过程/函数
	SupportsTriggers  bool // 触发器
	SupportsJSON      bool // JSON支持
	SupportsArray     bool // 数组支持
	SupportsVector    bool // 向量搜索
}

// UserContext 用户上下文
type UserContext struct {
	UserID       uint
	Username     string
	Roles        []string
	Permissions  []string
	DatabaseRole string // 数据库角色名
}

// RLSPolicy RLS策略
type RLSPolicy struct {
	TableName        string
	PolicyName       string
	PolicyType       string // "SELECT", "INSERT", "UPDATE", "DELETE", "ALL"
	PolicyExpression string
	IsEnabled        bool
}

// PermissionAwareManager 权限感知的数据库管理器
type PermissionAwareManager struct {
	*Manager
	currentUser  *UserContext
	dbType       DatabaseType
	capabilities DatabaseCapabilities
}

// PermissionAwareQuery 权限感知查询
type PermissionAwareQuery struct {
	db      *gorm.DB
	manager *PermissionAwareManager
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
			SupportsRoles:     true, // MySQL 8.0+
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
		Manager:      dm,
		currentUser:  user,
		dbType:       dbType,
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

// hasCreatePermission 检查是否有创建权限
func (pam *PermissionAwareManager) hasCreatePermission(tableName string) bool {
	switch tableName {
	case "resumes":
		return pam.hasPermission("resume:create")
	case "approvals":
		return pam.hasPermission("approve:create")
	}
	return false
}

// hasUpdatePermission 检查是否有更新权限
func (pam *PermissionAwareManager) hasUpdatePermission(tableName string) bool {
	switch tableName {
	case "resumes":
		return pam.hasPermission("resume:update")
	case "approvals":
		return pam.hasPermission("approve:update")
	}
	return false
}

// hasDeletePermission 检查是否有删除权限
func (pam *PermissionAwareManager) hasDeletePermission(tableName string) bool {
	switch tableName {
	case "resumes":
		return pam.hasPermission("resume:delete")
	case "approvals":
		return pam.hasPermission("approve:delete")
	}
	return false
}

// NewPermissionAwareQuery 创建权限感知查询
func (pam *PermissionAwareManager) NewPermissionAwareQuery() (*PermissionAwareQuery, error) {
	db, err := pam.GetPermissionAwareDB()
	if err != nil {
		return nil, err
	}

	return &PermissionAwareQuery{
		db:      db,
		manager: pam,
	}, nil
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

// GetCurrentUser 获取当前用户上下文
func (paq *PermissionAwareQuery) GetCurrentUser() *UserContext {
	return paq.manager.currentUser
}

// GetDatabaseType 获取数据库类型
func (paq *PermissionAwareQuery) GetDatabaseType() DatabaseType {
	return paq.manager.dbType
}

// GetCapabilities 获取数据库能力
func (paq *PermissionAwareQuery) GetCapabilities() DatabaseCapabilities {
	return paq.manager.capabilities
}
