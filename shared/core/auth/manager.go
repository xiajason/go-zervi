package auth

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthManager 认证管理器
type AuthManager struct {
	db     *gorm.DB
	config AuthConfig
}

// NewAuthManager 创建认证管理器
func NewAuthManager(db *gorm.DB, config AuthConfig) *AuthManager {
	return &AuthManager{
		db:     db,
		config: config,
	}
}

// Register 用户注册
func (am *AuthManager) Register(req RegisterRequest) (*RegisterResponse, error) {
	// 检查用户名和邮箱是否已存在
	var existingUser User
	if err := am.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名或邮箱已存在")
	}

	// 哈希密码
	passwordHash, err := am.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	phone := strings.TrimSpace(req.Phone)
	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}

	// 创建用户
	user := User{
		Username:           req.Username,
		Email:              req.Email,
		PasswordHash:       passwordHash,
		Status:             "active",
		EmailVerified:      false,
		PhoneVerified:      false,
		SubscriptionStatus: "free",
		Phone:              phonePtr,
	}

	if err := am.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	if err := am.assignRole(user.ID, "user"); err != nil {
		log.Printf("WARN: 为用户分配默认角色失败 user_id=%d: %v", user.ID, err)
	}

	fullName := strings.TrimSpace(strings.TrimSpace(req.FirstName + " " + req.LastName))
	if fullName != "" || phone != "" {
		if err := am.upsertUserProfile(user.ID, fullName, phone); err != nil {
			log.Printf("WARN: 创建用户档案失败 user_id=%d: %v", user.ID, err)
		}
	}

	role, err := am.fetchPrimaryRole(user.ID)
	if err != nil {
		role = "user"
	}
	user.Role = role

	return &RegisterResponse{
		Success: true,
		Message: "注册成功",
		User:    user,
	}, nil
}

// Login 用户登录
func (am *AuthManager) Login(req LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	// 查找用户
	var user User
	if err := am.db.Where("username = ? AND status = 'active'", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在或已被禁用")
	}

	// 验证密码
	if !am.validatePassword(req.Password, user.PasswordHash) {
		am.logLoginAttempt(user.ID, clientIP, userAgent, "failed", "密码错误")
		return nil, errors.New("密码错误")
	}

	role, err := am.fetchPrimaryRole(user.ID)
	if err != nil {
		log.Printf("WARN: 获取用户角色失败 user_id=%d: %v", user.ID, err)
		role = "user"
	}
	user.Role = role

	// 生成JWT token
	token, expiresAt, err := am.generateToken(user.ID, user.Username, role)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	am.db.Model(&user).Update("last_login_at", now)

	// 记录登录日志
	am.logLoginAttempt(user.ID, clientIP, userAgent, "success", "登录成功")

	response := &LoginResponse{
		Success:   true,
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Message:   "登录成功",
	}

	// 检查是否为开发团队成员
	var devTeam DevTeamUser
	if err := am.db.Where("user_id = ? AND status = 'active'", user.ID).First(&devTeam).Error; err == nil {
		response.DevTeam = devTeam
	}

	return response, nil
}

// SuperAdminLogin 超级管理员登录
func (am *AuthManager) SuperAdminLogin(req LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	// 查找用户
	var user User
	if err := am.db.Where("username = ? AND status = 'active'", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在或已被禁用")
	}

	// 验证密码
	if !am.validatePassword(req.Password, user.PasswordHash) {
		am.logLoginAttempt(user.ID, clientIP, userAgent, "failed", "密码错误")
		return nil, errors.New("密码错误")
	}

	// 检查是否为超级管理员
	var devTeam DevTeamUser
	if err := am.db.Where("user_id = ? AND team_role = 'super_admin' AND status = 'active'", user.ID).First(&devTeam).Error; err != nil {
		return nil, errors.New("您不是超级管理员")
	}

	role, err := am.fetchPrimaryRole(user.ID)
	if err != nil {
		role = "super_admin"
	}

	// 生成JWT token
	token, expiresAt, err := am.generateToken(user.ID, user.Username, role)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	am.db.Model(&user).Update("last_login_at", now)
	am.db.Model(&devTeam).Update("last_login_at", now)

	// 记录登录日志
	am.logLoginAttempt(user.ID, clientIP, userAgent, "success", "超级管理员登录成功")

	user.Role = role

	return &LoginResponse{
		Success:   true,
		Token:     token,
		User:      user,
		DevTeam:   devTeam,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Message:   "超级管理员登录成功",
	}, nil
}

// ValidateToken 验证JWT token
func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	if len(tokenString) > 50 {
		log.Printf("DEBUG: 开始验证JWT token: %s...", tokenString[:50])
	} else {
		log.Printf("DEBUG: 开始验证JWT token: %s", tokenString)
	}
	log.Printf("DEBUG: 使用JWT secret: %s", am.config.JWTSecret)

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		log.Printf("DEBUG: JWT解析 - 算法: %v", token.Header["alg"])
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("DEBUG: JWT解析失败 - 不支持的签名方法: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		log.Printf("DEBUG: JWT解析 - 使用secret进行验证")
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		log.Printf("DEBUG: JWT解析失败: %v", err)
		return nil, fmt.Errorf("token解析失败: %w", err)
	}

	if !token.Valid {
		log.Printf("DEBUG: JWT token无效")
		return nil, errors.New("无效的token")
	}

	// 检查token是否过期
	currentTime := time.Now().Unix()
	log.Printf("DEBUG: 当前时间: %d, Token过期时间: %d", currentTime, claims.Exp)
	if currentTime > claims.Exp {
		log.Printf("DEBUG: Token已过期")
		return nil, errors.New("token已过期")
	}

	return claims, nil
}

// GetUserByID 根据ID获取用户
func (am *AuthManager) GetUserByID(userID uint) (*User, error) {
	var user User
	if err := am.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}
	if role, err := am.fetchPrimaryRole(user.ID); err == nil {
		user.Role = role
	}
	return &user, nil
}

// GetDevTeamUser 获取开发团队成员信息
func (am *AuthManager) GetDevTeamUser(userID uint) (*DevTeamUser, error) {
	var devTeam DevTeamUser
	if err := am.db.Preload("User").Where("user_id = ? AND status = 'active'", userID).First(&devTeam).Error; err != nil {
		return nil, fmt.Errorf("不是开发团队成员: %w", err)
	}
	return &devTeam, nil
}

// CheckPermission 检查用户权限
func (am *AuthManager) CheckPermission(userID uint, requiredRole string) (bool, error) {
	var devTeam DevTeamUser
	if err := am.db.Where("user_id = ? AND status = 'active'", userID).First(&devTeam).Error; err != nil {
		return false, errors.New("用户不是开发团队成员")
	}

	// 超级管理员拥有所有权限
	if devTeam.TeamRole == "super_admin" {
		return true, nil
	}

	// 检查角色权限
	return devTeam.TeamRole == requiredRole, nil
}

// 辅助方法

// hashPassword 哈希密码
func (am *AuthManager) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// validatePassword 验证密码
func (am *AuthManager) validatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken 生成JWT token
func (am *AuthManager) generateToken(userID uint, username, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(am.config.TokenExpiry)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Exp:      expiresAt.Unix(),
		Iat:      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(am.config.JWTSecret))

	return tokenString, expiresAt, err
}

// logLoginAttempt 记录登录尝试
func (am *AuthManager) logLoginAttempt(userID uint, ipAddress, userAgent, status, message string) {
	log := DevOperationLog{
		UserID:           userID,
		OperationType:    "login_attempt",
		OperationTarget:  "auth",
		OperationDetails: fmt.Sprintf(`{"status": "%s", "message": "%s"}`, status, message),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Status:           status,
		CreatedAt:        time.Now(),
	}
	am.db.Create(&log)
}

func (am *AuthManager) upsertUserProfile(userID uint, fullName, phone string) error {
	if am.db == nil {
		return errors.New("database not initialized")
	}

	nameValue := interface{}(nil)
	if strings.TrimSpace(fullName) != "" {
		nameValue = strings.TrimSpace(fullName)
	}

	phoneValue := interface{}(nil)
	if strings.TrimSpace(phone) != "" {
		phoneValue = strings.TrimSpace(phone)
	}

	dialect := strings.ToLower(am.db.Dialector.Name())
	if dialect == "postgres" {
		query := `
            INSERT INTO zervigo_user_profiles (user_id, real_name, phone, created_at, updated_at)
            VALUES (?, ?, ?, NOW(), NOW())
            ON CONFLICT (user_id) DO UPDATE SET real_name = EXCLUDED.real_name, phone = EXCLUDED.phone, updated_at = NOW()
        `
		return am.db.Exec(query, userID, nameValue, phoneValue).Error
	}

	query := `
        INSERT INTO zervigo_user_profiles (user_id, real_name, phone, created_at, updated_at)
        VALUES (?, ?, ?, NOW(), NOW())
        ON DUPLICATE KEY UPDATE real_name = VALUES(real_name), phone = VALUES(phone), updated_at = VALUES(updated_at)
    `
	return am.db.Exec(query, userID, nameValue, phoneValue).Error
}

func (am *AuthManager) assignRole(userID uint, roleName string) error {
	if am.db == nil {
		return errors.New("database not initialized")
	}

	type roleRow struct {
		ID uint
	}

	var role roleRow
	if err := am.db.Raw("SELECT id FROM zervigo_auth_roles WHERE role_name = ? LIMIT 1", roleName).Scan(&role).Error; err != nil {
		return fmt.Errorf("查询角色失败: %w", err)
	}

	if role.ID == 0 {
		return fmt.Errorf("角色 %s 不存在", roleName)
	}

	dial := strings.ToLower(am.db.Dialector.Name())
	if dial == "postgres" {
		query := `
            INSERT INTO zervigo_auth_user_roles (user_id, role_id, status, assigned_at)
            VALUES (?, ?, 'active', NOW())
            ON CONFLICT (user_id, role_id) DO UPDATE SET status = 'active', assigned_at = NOW()
        `
		return am.db.Exec(query, userID, role.ID).Error
	}

	query := `
        INSERT INTO zervigo_auth_user_roles (user_id, role_id, status, assigned_at)
        VALUES (?, ?, 'active', NOW())
        ON DUPLICATE KEY UPDATE status = 'active', assigned_at = NOW()
    `
	return am.db.Exec(query, userID, role.ID).Error
}

func (am *AuthManager) fetchPrimaryRole(userID uint) (string, error) {
	if am.db == nil {
		return "user", errors.New("database not initialized")
	}

	var roleName string
	err := am.db.Raw(`
        SELECT r.role_name
        FROM zervigo_auth_user_roles ur
        JOIN zervigo_auth_roles r ON ur.role_id = r.id
        WHERE ur.user_id = ? AND (ur.status IS NULL OR ur.status = 'active')
        ORDER BY r.role_level DESC, ur.assigned_at DESC
        LIMIT 1
    `, userID).Scan(&roleName).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "user", nil
	}
	if err != nil {
		return "user", err
	}
	if roleName == "" {
		return "user", nil
	}
	return roleName, nil
}
