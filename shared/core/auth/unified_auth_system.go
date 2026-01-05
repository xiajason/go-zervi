package auth

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UnifiedAuthSystem 统一认证系统
type UnifiedAuthSystem struct {
	db         *sql.DB
	jwtSecret  string
	roleConfig *RoleConfig
	dbType     string // "postgresql" 或 "mysql"
}

// detectDatabaseType 检测数据库类型
func detectDatabaseType(db *sql.DB) string {
	var version string
	// 尝试执行 MySQL 的 VERSION() 查询
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		log.Printf("WARN: 无法检测数据库版本: %v", err)
		return "mysql" // 默认尝试 MySQL
	}

	log.Printf("DEBUG: 数据库版本字符串: %s", version)

	// 检查版本字符串中的关键词
	if version != "" {
		lowerVersion := fmt.Sprintf("%s", version)
		if len(lowerVersion) > 10 {
			log.Printf("DEBUG: 版本字符串前10个字符: %s", lowerVersion[:10])
		}
		// MySQL 包含 "mysql" 或版本号开头
		if len(version) > 4 && (version[0] >= '0' && version[0] <= '9') {
			log.Printf("INFO: 检测到 MySQL (版本号开头: %s)", string(version[0]))
			return "mysql"
		}
		// PostgreSQL 包含 "PostgreSQL"
		if len(version) > 10 && (version[0:10] == "PostgreSQL") {
			log.Printf("INFO: 检测到 PostgreSQL")
			return "postgresql"
		}
	}

	// 如果无法确定，尝试通过 driver 名称判断
	driver := db.Driver()
	log.Printf("DEBUG: 数据库驱动: %T", driver)

	// 默认返回 mysql
	log.Printf("WARN: 无法确定数据库类型，默认使用 mysql")
	return "mysql"
}

// RoleConfig 角色配置
type RoleConfig struct {
	Roles map[string]*RoleInfo `json:"roles"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

type permissionDefinition struct {
	name        string
	resource    string
	action      string
	description string
	service     string
}

// UserInfo 用户信息（统一结构）
type UserInfo struct {
	ID                 int        `json:"id" db:"id"`
	Username           string     `json:"username" db:"username"`
	Email              string     `json:"email" db:"email"`
	Phone              *string    `json:"phone" db:"phone"`
	PasswordHash       string     `json:"-" db:"password_hash"`
	Role               string     `json:"role" db:"role"` // 通过查询获取
	Status             string     `json:"status" db:"status"`
	EmailVerified      bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified      bool       `json:"phone_verified" db:"phone_verified"`
	SubscriptionStatus string     `json:"subscription_status" db:"subscription_status"`
	SubscriptionType   *string    `json:"subscription_type" db:"subscription_type"`
	SubscriptionExpiry *time.Time `json:"subscription_expiry" db:"subscription_expires_at"`
	LastLogin          *time.Time `json:"last_login" db:"last_login_at"`
	CreatedAt          *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID      int      `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthResult 认证结果
type AuthResult struct {
	Success     bool      `json:"success"`
	Token       string    `json:"token,omitempty"`
	User        *UserInfo `json:"user,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	Error       string    `json:"error,omitempty"`
	ErrorCode   string    `json:"error_code,omitempty"`
}

// NewUnifiedAuthSystem 创建统一认证系统
func NewUnifiedAuthSystem(db *sql.DB, jwtSecret string) *UnifiedAuthSystem {
	roleConfig := &RoleConfig{
		Roles: map[string]*RoleInfo{
			"guest": {
				Name:        "guest",
				Level:       1,
				Permissions: []string{"read:public"},
				Description: "访客用户",
			},
			"user": {
				Name:        "user",
				Level:       2,
				Permissions: []string{"read:public", "read:own", "write:own"},
				Description: "普通用户",
			},
			"admin": {
				Name:        "admin",
				Level:       3,
				Permissions: []string{"read:public", "read:own", "write:own", "read:all", "write:all", "delete:own"},
				Description: "管理员",
			},
			"super_admin": {
				Name:        "super_admin",
				Level:       4,
				Permissions: []string{"*"},
				Description: "超级管理员",
			},
		},
	}

	uas := &UnifiedAuthSystem{
		db:         db,
		jwtSecret:  jwtSecret,
		roleConfig: roleConfig,
		dbType:     detectDatabaseType(db),
	}
	log.Printf("INFO: UnifiedAuthSystem 检测到数据库类型: %s", uas.dbType)
	return uas
}

func (uas *UnifiedAuthSystem) isPostgres() bool {
	return strings.EqualFold(uas.dbType, "postgresql")
}

func (uas *UnifiedAuthSystem) placeholder(idx int) string {
	if uas.isPostgres() {
		return fmt.Sprintf("$%d", idx)
	}
	return "?"
}

func (uas *UnifiedAuthSystem) makePlaceholders(count int) string {
	parts := make([]string, count)
	for i := 0; i < count; i++ {
		parts[i] = uas.placeholder(i + 1)
	}
	return strings.Join(parts, ", ")
}

func (uas *UnifiedAuthSystem) tableExists(tableName string) (bool, error) {
	var query string
	var args []interface{}

	if uas.isPostgres() {
		query = `SELECT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public' AND table_name = $1
        )`
		args = []interface{}{tableName}
	} else {
		query = `SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`
		args = []interface{}{tableName}
	}

	var exists bool
	if err := uas.db.QueryRow(query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// InitializeDatabase 初始化数据库表结构
func (uas *UnifiedAuthSystem) InitializeDatabase() error {
	tables := []string{
		"zervigo_auth_users",
		"zervigo_auth_roles",
		"zervigo_auth_permissions",
		"zervigo_auth_role_permissions",
		"zervigo_auth_user_roles",
		"zervigo_auth_tokens",
		"zervigo_auth_login_logs",
		"zervigo_service_credentials",
	}

	for _, table := range tables {
		exists, err := uas.tableExists(table)
		if err != nil {
			return fmt.Errorf("检查表 %s 是否存在失败: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("表 %s 不存在，请先运行数据库初始化脚本", table)
		}
	}

	if err := uas.initializePermissions(); err != nil {
		return fmt.Errorf("初始化权限数据失败: %w", err)
	}

	if err := uas.createDefaultSuperAdmin(); err != nil {
		return fmt.Errorf("创建默认超级管理员失败: %w", err)
	}

	return nil
}

// initializePermissions 初始化权限数据
func (uas *UnifiedAuthSystem) initializePermissions() error {
	permissions := []permissionDefinition{
		{"read:public", "public", "read", "读取公开内容", "auth"},
		{"read:own", "own", "read", "读取自己的内容", "auth"},
		{"write:own", "own", "write", "修改自己的内容", "auth"},
		{"read:all", "all", "read", "读取所有内容", "auth"},
		{"write:all", "all", "write", "修改所有内容", "auth"},
		{"delete:own", "own", "delete", "删除自己的内容", "auth"},
		{"delete:all", "all", "delete", "删除所有内容", "auth"},
		{"admin:users", "users", "admin", "用户管理", "auth"},
		{"admin:system", "system", "admin", "系统管理", "auth"},
	}

	tx, err := uas.db.Begin()
	if err != nil {
		return fmt.Errorf("开启权限初始化事务失败: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	roleIDs := make(map[string]int64)
	for roleName, roleInfo := range uas.roleConfig.Roles {
		roleID, upsertErr := uas.upsertRoleTx(tx, roleName, roleInfo)
		if upsertErr != nil {
			err = fmt.Errorf("初始化角色 %s 失败: %w", roleName, upsertErr)
			return err
		}
		roleIDs[roleName] = roleID
	}

	permissionIDs := make(map[string]int64)
	for _, perm := range permissions {
		permID, upsertErr := uas.upsertPermissionTx(tx, perm)
		if upsertErr != nil {
			err = fmt.Errorf("初始化权限 %s 失败: %w", perm.name, upsertErr)
			return err
		}
		permissionIDs[perm.name] = permID
	}

	for roleName, roleInfo := range uas.roleConfig.Roles {
		roleID := roleIDs[roleName]
		if roleID == 0 {
			continue
		}

		if len(roleInfo.Permissions) == 1 && roleInfo.Permissions[0] == "*" {
			for _, permID := range permissionIDs {
				if ensureErr := uas.ensureRolePermissionTx(tx, roleID, permID); ensureErr != nil {
					err = fmt.Errorf("为角色 %s 分配权限失败: %w", roleName, ensureErr)
					return err
				}
			}
			continue
		}

		for _, permName := range roleInfo.Permissions {
			permID, ok := permissionIDs[permName]
			if !ok {
				log.Printf("WARN: 未找到权限 %s，跳过分配给角色 %s", permName, roleName)
				continue
			}
			if ensureErr := uas.ensureRolePermissionTx(tx, roleID, permID); ensureErr != nil {
				err = fmt.Errorf("为角色 %s 分配权限 %s 失败: %w", roleName, permName, ensureErr)
				return err
			}
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("提交权限初始化事务失败: %w", commitErr)
	}

	return nil
}

func (uas *UnifiedAuthSystem) upsertRoleTx(tx *sql.Tx, roleName string, info *RoleInfo) (int64, error) {
	if info == nil {
		info = &RoleInfo{Name: roleName}
	}

	if uas.isPostgres() {
		query := `
			INSERT INTO zervigo_auth_roles (role_name, role_description, role_level, status, updated_at)
			VALUES ($1, $2, $3, 'active', NOW())
			ON CONFLICT (role_name)
			DO UPDATE SET role_description = EXCLUDED.role_description,
			              role_level = EXCLUDED.role_level,
			              status = 'active',
			              updated_at = NOW()
			RETURNING id
		`
		var roleID int64
		if err := tx.QueryRow(query, roleName, info.Description, info.Level).Scan(&roleID); err != nil {
			return 0, err
		}
		return roleID, nil
	}

	query := `
		INSERT INTO zervigo_auth_roles (role_name, role_description, role_level, status, updated_at)
		VALUES (?, ?, ?, 'active', NOW())
		ON DUPLICATE KEY UPDATE role_description = VALUES(role_description),
		                        role_level = VALUES(role_level),
		                        status = 'active',
		                        updated_at = VALUES(updated_at)
	`
	result, err := tx.Exec(query, roleName, info.Description, info.Level)
	if err != nil {
		return 0, err
	}
	if roleID, err := result.LastInsertId(); err == nil && roleID > 0 {
		return roleID, nil
	}
	return uas.fetchRoleIDTx(tx, roleName)
}

func (uas *UnifiedAuthSystem) fetchRoleIDTx(tx *sql.Tx, roleName string) (int64, error) {
	query := fmt.Sprintf("SELECT id FROM zervigo_auth_roles WHERE role_name = %s", uas.placeholder(1))
	var roleID int64
	if err := tx.QueryRow(query, roleName).Scan(&roleID); err != nil {
		return 0, err
	}
	return roleID, nil
}

func (uas *UnifiedAuthSystem) upsertPermissionTx(tx *sql.Tx, perm permissionDefinition) (int64, error) {
	if uas.isPostgres() {
		query := `
			INSERT INTO zervigo_auth_permissions (permission_name, permission_code, permission_description, service_name, resource_type, action, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW())
			ON CONFLICT (permission_name)
			DO UPDATE SET permission_description = EXCLUDED.permission_description,
			              service_name = EXCLUDED.service_name,
			              resource_type = EXCLUDED.resource_type,
			              action = EXCLUDED.action,
			              status = TRUE
			RETURNING id
		`
		var permID int64
		if err := tx.QueryRow(query, perm.name, perm.name, perm.description, perm.service, perm.resource, perm.action).Scan(&permID); err != nil {
			return 0, err
		}
		return permID, nil
	}

	query := `
		INSERT INTO zervigo_auth_permissions (permission_name, permission_code, permission_description, service_name, resource_type, action, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, TRUE, NOW())
		ON DUPLICATE KEY UPDATE permission_description = VALUES(permission_description),
		                        service_name = VALUES(service_name),
		                        resource_type = VALUES(resource_type),
		                        action = VALUES(action),
		                        status = TRUE
	`
	result, err := tx.Exec(query, perm.name, perm.name, perm.description, perm.service, perm.resource, perm.action)
	if err != nil {
		return 0, err
	}
	if permID, err := result.LastInsertId(); err == nil && permID > 0 {
		return permID, nil
	}
	return uas.fetchPermissionIDTx(tx, perm.name)
}

func (uas *UnifiedAuthSystem) fetchPermissionIDTx(tx *sql.Tx, permName string) (int64, error) {
	query := fmt.Sprintf("SELECT id FROM zervigo_auth_permissions WHERE permission_name = %s", uas.placeholder(1))
	var permID int64
	if err := tx.QueryRow(query, permName).Scan(&permID); err != nil {
		return 0, err
	}
	return permID, nil
}

func (uas *UnifiedAuthSystem) fetchUserIDTx(tx *sql.Tx, username string) (int64, error) {
	query := fmt.Sprintf("SELECT id FROM zervigo_auth_users WHERE username = %s", uas.placeholder(1))
	var userID int64
	if err := tx.QueryRow(query, username).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func (uas *UnifiedAuthSystem) scanUser(row *sql.Row) (*UserInfo, error) {
	var (
		id                 int64
		phone              sql.NullString
		subscriptionType   sql.NullString
		subscriptionExpiry sql.NullTime
		lastLogin          sql.NullTime
		createdAt          sql.NullTime
		updatedAt          sql.NullTime
		role               sql.NullString
	)

	user := &UserInfo{}
	if err := row.Scan(
		&id,
		&user.Username,
		&user.Email,
		&phone,
		&user.PasswordHash,
		&user.Status,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.SubscriptionStatus,
		&subscriptionType,
		&subscriptionExpiry,
		&lastLogin,
		&createdAt,
		&updatedAt,
		&role,
	); err != nil {
		return nil, err
	}

	user.ID = int(id)
	if phone.Valid {
		user.Phone = &phone.String
	}
	if subscriptionType.Valid {
		user.SubscriptionType = &subscriptionType.String
	}
	if subscriptionExpiry.Valid {
		t := subscriptionExpiry.Time
		user.SubscriptionExpiry = &t
	}
	if lastLogin.Valid {
		t := lastLogin.Time
		user.LastLogin = &t
	}
	if createdAt.Valid {
		t := createdAt.Time
		user.CreatedAt = &t
	}
	if updatedAt.Valid {
		t := updatedAt.Time
		user.UpdatedAt = &t
	}
	if role.Valid {
		user.Role = role.String
	} else {
		user.Role = "user"
	}

	return user, nil
}

func (uas *UnifiedAuthSystem) ensureRolePermissionTx(tx *sql.Tx, roleID, permID int64) error {
	if uas.isPostgres() {
		query := `
			INSERT INTO zervigo_auth_role_permissions (role_id, permission_id)
			VALUES ($1, $2)
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`
		_, err := tx.Exec(query, roleID, permID)
		return err
	}

	query := `INSERT IGNORE INTO zervigo_auth_role_permissions (role_id, permission_id) VALUES (?, ?)`
	_, err := tx.Exec(query, roleID, permID)
	return err
}

// createDefaultSuperAdmin 创建默认超级管理员
func (uas *UnifiedAuthSystem) createDefaultSuperAdmin() error {
	tx, err := uas.db.Begin()
	if err != nil {
		return fmt.Errorf("开启超级管理员初始化事务失败: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	checkQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM zervigo_auth_users u
		JOIN zervigo_auth_user_roles ur ON u.id = ur.user_id
		JOIN zervigo_auth_roles r ON ur.role_id = r.id
		WHERE r.role_name = %s
	`, uas.placeholder(1))
	var count int
	if err = tx.QueryRow(checkQuery, "super_admin").Scan(&count); err != nil {
		return fmt.Errorf("检查超级管理员是否存在失败: %w", err)
	}
	if count > 0 {
		err = tx.Commit()
		return err
	}

	password := "admin123"
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return fmt.Errorf("生成超级管理员密码哈希失败: %w", hashErr)
	}

	var userID int64
	username := "admin"
	email := "admin@jobfirst.com"

	if uas.isPostgres() {
		insertQuery := `
			INSERT INTO zervigo_auth_users (username, email, password_hash, status, email_verified, phone_verified, subscription_status, subscription_type, created_at, updated_at)
			VALUES ($1, $2, $3, 'active', TRUE, FALSE, 'free', 'basic', NOW(), NOW())
			ON CONFLICT (username) DO NOTHING
			RETURNING id
		`
		err = tx.QueryRow(insertQuery, username, email, string(hashedPassword)).Scan(&userID)
		if err == sql.ErrNoRows {
			err = nil
			userID, err = uas.fetchUserIDTx(tx, username)
		}
	} else {
		insertQuery := `
			INSERT INTO zervigo_auth_users (username, email, password_hash, status, email_verified, phone_verified, subscription_status, subscription_type, created_at, updated_at)
			VALUES (?, ?, ?, 'active', 1, 0, 'free', 'basic', NOW(), NOW())
			ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)
		`
		var result sql.Result
		result, err = tx.Exec(insertQuery, username, email, string(hashedPassword))
		if err == nil {
			if userID, err = result.LastInsertId(); err == nil && userID > 0 {
				// inserted new user
			} else {
				userID, err = uas.fetchUserIDTx(tx, username)
			}
		}
	}
	if err != nil {
		return fmt.Errorf("插入超级管理员失败: %w", err)
	}

	roleID, err := uas.fetchRoleIDTx(tx, "super_admin")
	if err != nil {
		return fmt.Errorf("获取super_admin角色ID失败: %w", err)
	}

	if uas.isPostgres() {
		query := `
			INSERT INTO zervigo_auth_user_roles (user_id, role_id, status, assigned_at)
			VALUES ($1, $2, 'active', NOW())
			ON CONFLICT (user_id, role_id) DO NOTHING
		`
		if _, err = tx.Exec(query, userID, roleID); err != nil {
			return fmt.Errorf("为超级管理员分配角色失败: %w", err)
		}
	} else {
		query := `INSERT IGNORE INTO zervigo_auth_user_roles (user_id, role_id, status, assigned_at) VALUES (?, ?, 'active', NOW())`
		if _, err = tx.Exec(query, userID, roleID); err != nil {
			return fmt.Errorf("为超级管理员分配角色失败: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交超级管理员初始化事务失败: %w", err)
	}

	log.Printf("默认超级管理员已创建: username=%s, password=%s", username, password)
	return nil
}

// Authenticate 用户认证
func (uas *UnifiedAuthSystem) Authenticate(username, password string) (*AuthResult, error) {
	// 查询用户信息
	user, err := uas.getUserByUsername(username)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "用户不存在",
			ErrorCode: "USER_NOT_FOUND",
		}, nil
	}

	// 检查用户状态
	if user.Status != "active" {
		return &AuthResult{
			Success:   false,
			Error:     "用户账户已被禁用",
			ErrorCode: "USER_DISABLED",
		}, nil
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "密码错误",
			ErrorCode: "INVALID_PASSWORD",
		}, nil
	}

	// 获取用户权限
	permissions, err := uas.getUserPermissions(user.Role)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "获取用户权限失败",
			ErrorCode: "PERMISSION_ERROR",
		}, nil
	}

	// 生成JWT token
	token, err := uas.generateJWT(user, permissions)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "生成token失败",
			ErrorCode: "TOKEN_ERROR",
		}, nil
	}

	// 更新最后登录时间
	uas.updateLastLogin(user.ID)

	// 记录访问日志
	uas.logAccess(user.ID, "login", "auth", "success", "", "")

	return &AuthResult{
		Success:     true,
		Token:       token,
		User:        user,
		Permissions: permissions,
	}, nil
}

// ValidateJWT 验证JWT token
func (uas *UnifiedAuthSystem) ValidateJWT(tokenString string) (*AuthResult, error) {
	// 解析JWT token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(uas.jwtSecret), nil
	})

	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "无效的token",
			ErrorCode: "INVALID_TOKEN",
		}, nil
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return &AuthResult{
			Success:   false,
			Error:     "token验证失败",
			ErrorCode: "TOKEN_VALIDATION_FAILED",
		}, nil
	}

	// 检查token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return &AuthResult{
			Success:   false,
			Error:     "token已过期",
			ErrorCode: "TOKEN_EXPIRED",
		}, nil
	}

	// 获取用户信息
	user, err := uas.getUserByID(claims.UserID)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "用户不存在",
			ErrorCode: "USER_NOT_FOUND",
		}, nil
	}

	// 检查用户状态
	if user.Status != "active" {
		return &AuthResult{
			Success:   false,
			Error:     "用户账户已被禁用",
			ErrorCode: "USER_DISABLED",
		}, nil
	}

	return &AuthResult{
		Success:     true,
		User:        user,
		Permissions: claims.Permissions,
	}, nil
}

// CheckPermission 检查用户权限
func (uas *UnifiedAuthSystem) CheckPermission(userID int, permission string) (bool, error) {
	// 获取用户信息
	user, err := uas.getUserByID(userID)
	if err != nil {
		return false, err
	}

	// 超级管理员拥有所有权限
	if user.Role == "super_admin" {
		return true, nil
	}

	// 检查角色权限
	permissions, err := uas.getUserPermissions(user.Role)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permission || perm == "*" {
			return true, nil
		}
	}

	return false, nil
}

// getUserByUsername 根据用户名获取用户信息
func (uas *UnifiedAuthSystem) getUserByUsername(username string) (*UserInfo, error) {
	query := fmt.Sprintf(`
        SELECT u.id, u.username, u.email, u.phone, u.password_hash, u.status,
               u.email_verified, u.phone_verified, u.subscription_status,
               u.subscription_type, u.subscription_expires_at, u.last_login_at,
               u.created_at, u.updated_at, COALESCE(r.role_name, 'user')
        FROM zervigo_auth_users u
        LEFT JOIN zervigo_auth_user_roles ur ON u.id = ur.user_id AND (ur.status IS NULL OR ur.status = 'active')
        LEFT JOIN zervigo_auth_roles r ON ur.role_id = r.id
        WHERE u.username = %s
        ORDER BY COALESCE(r.role_level, 0) DESC
        LIMIT 1
    `, uas.placeholder(1))

	return uas.scanUser(uas.db.QueryRow(query, username))
}

// getUserByID 根据ID获取用户信息
func (uas *UnifiedAuthSystem) getUserByID(userID int) (*UserInfo, error) {
	query := fmt.Sprintf(`
        SELECT u.id, u.username, u.email, u.phone, u.password_hash, u.status,
               u.email_verified, u.phone_verified, u.subscription_status,
               u.subscription_type, u.subscription_expires_at, u.last_login_at,
               u.created_at, u.updated_at, COALESCE(r.role_name, 'user')
        FROM zervigo_auth_users u
        LEFT JOIN zervigo_auth_user_roles ur ON u.id = ur.user_id AND (ur.status IS NULL OR ur.status = 'active')
        LEFT JOIN zervigo_auth_roles r ON ur.role_id = r.id
        WHERE u.id = %s
        ORDER BY COALESCE(r.role_level, 0) DESC
        LIMIT 1
    `, uas.placeholder(1))

	return uas.scanUser(uas.db.QueryRow(query, userID))
}

// getUserPermissions 获取用户权限
func (uas *UnifiedAuthSystem) getUserPermissions(role string) ([]string, error) {
	query := fmt.Sprintf(`
        SELECT p.permission_code
        FROM zervigo_auth_roles r
        JOIN zervigo_auth_role_permissions rp ON r.id = rp.role_id
        JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
        WHERE r.role_name = %s
        ORDER BY p.permission_code
    `, uas.placeholder(1))

	rows, err := uas.db.Query(query, role)
	if err != nil {
		return nil, fmt.Errorf("查询角色权限失败: %w", err)
	}
	defer rows.Close()

	permissions := make([]string, 0)
	for rows.Next() {
		var permission string
		if scanErr := rows.Scan(&permission); scanErr != nil {
			log.Printf("WARN: 读取角色权限失败: %v", scanErr)
			continue
		}
		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历角色权限失败: %w", err)
	}

	if len(permissions) == 0 {
		if roleInfo, exists := uas.roleConfig.Roles[role]; exists {
			if len(roleInfo.Permissions) == 1 && roleInfo.Permissions[0] == "*" {
				return uas.getAllPermissionCodes()
			}
			return append([]string{}, roleInfo.Permissions...), nil
		}
	}

	return permissions, nil
}

func (uas *UnifiedAuthSystem) getAllPermissionCodes() ([]string, error) {
	rows, err := uas.db.Query("SELECT permission_code FROM zervigo_auth_permissions ORDER BY permission_code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if scanErr := rows.Scan(&code); scanErr != nil {
			return nil, scanErr
		}
		permissions = append(permissions, code)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

// generateJWT 生成JWT token
func (uas *UnifiedAuthSystem) generateJWT(user *UserInfo, permissions []string) (string, error) {
	roleInfo, exists := uas.roleConfig.Roles[user.Role]
	if !exists {
		roleInfo = &RoleInfo{Level: 1}
	}

	claims := &JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Level:       roleInfo.Level,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7天，适配测试需要
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "jobfirst-auth",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uas.jwtSecret))
}

// updateLastLogin 更新最后登录时间
func (uas *UnifiedAuthSystem) updateLastLogin(userID int) {
	query := fmt.Sprintf("UPDATE zervigo_auth_users SET last_login_at = NOW(), updated_at = NOW() WHERE id = %s", uas.placeholder(1))
	if _, err := uas.db.Exec(query, userID); err != nil {
		log.Printf("WARN: 更新用户 %d 最后登录时间失败: %v", userID, err)
	}
}

// logAccess 记录访问日志
func (uas *UnifiedAuthSystem) logAccess(userID int, action, resource, result, ipAddress, userAgent string) {
	method := action
	if resource != "" {
		method = fmt.Sprintf("%s:%s", action, resource)
	}

	success := strings.EqualFold(result, "success")
	failureReason := ""
	if !success {
		failureReason = result
	}

	fallbackUsername := fmt.Sprintf("user:%d", userID)
	if userID == 0 {
		fallbackUsername = "anonymous"
	}

	var err error
	if uas.isPostgres() {
		query := `
            INSERT INTO zervigo_auth_login_logs
            (user_id, username, login_method, success, failure_reason, client_ip, user_agent, login_at)
            VALUES ($1,
                    COALESCE((SELECT username FROM zervigo_auth_users WHERE id = $1), $2),
                    $3, $4, $5, $6, $7, NOW())
        `
		_, err = uas.db.Exec(query, userID, fallbackUsername, method, success, failureReason, ipAddress, userAgent)
	} else {
		query := `
            INSERT INTO zervigo_auth_login_logs
            (user_id, username, login_method, success, failure_reason, client_ip, user_agent, login_at)
            VALUES (?, IFNULL((SELECT username FROM zervigo_auth_users WHERE id = ? LIMIT 1), ?), ?, ?, ?, ?, NOW())
        `
		_, err = uas.db.Exec(query, userID, userID, fallbackUsername, method, success, failureReason, ipAddress, userAgent)
	}

	if err != nil {
		log.Printf("WARN: 记录访问日志失败: %v", err)
	}
}

// Close 关闭数据库连接
func (uas *UnifiedAuthSystem) Close() error {
	if uas.db != nil {
		return uas.db.Close()
	}
	return nil
}
