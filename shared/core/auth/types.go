package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// User 用户基础信息（与现有users表兼容）
type User struct {
	ID                    uint           `json:"id" gorm:"primaryKey;column:id"`
	Username              string         `json:"username" gorm:"column:username;size:100;uniqueIndex"`
	Email                 string         `json:"email" gorm:"column:email;size:255;uniqueIndex"`
	Phone                 *string        `json:"phone" gorm:"column:phone"`
	PasswordHash          string         `json:"-" gorm:"column:password_hash"`
	Status                string         `json:"status" gorm:"column:status"`
	EmailVerified         bool           `json:"email_verified" gorm:"column:email_verified"`
	PhoneVerified         bool           `json:"phone_verified" gorm:"column:phone_verified"`
	SubscriptionStatus    string         `json:"subscription_status" gorm:"column:subscription_status"`
	SubscriptionType      *string        `json:"subscription_type" gorm:"column:subscription_type"`
	SubscriptionExpiresAt *time.Time     `json:"subscription_expires_at" gorm:"column:subscription_expires_at"`
	CreatedAt             time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt             time.Time      `json:"updated_at" gorm:"column:updated_at"`
	LastLoginAt           *time.Time     `json:"last_login_at" gorm:"column:last_login_at"`
	DeletedAt             gorm.DeletedAt `json:"-" gorm:"column:deleted_at"`
	Role                  string         `json:"role" gorm:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "zervigo_auth_users"
}

// DevTeamUser 开发团队成员
type DevTeamUser struct {
	ID                        uint       `json:"id" gorm:"primaryKey"`
	UserID                    uint       `json:"user_id" gorm:"uniqueIndex"`
	TeamRole                  string     `json:"team_role" gorm:"type:enum('super_admin','system_admin','dev_lead','frontend_dev','backend_dev','qa_engineer','guest');default:guest"`
	SSHPublicKey              string     `json:"ssh_public_key" gorm:"type:text"`
	ServerAccessLevel         string     `json:"server_access_level" gorm:"type:enum('full','limited','readonly','none');default:limited"`
	CodeAccessModules         string     `json:"code_access_modules" gorm:"type:json"`
	DatabaseAccess            string     `json:"database_access" gorm:"type:json"`
	ServiceRestartPermissions string     `json:"service_restart_permissions" gorm:"type:json"`
	LastLoginAt               *time.Time `json:"last_login_at"`
	Status                    string     `json:"status" gorm:"type:enum('active','inactive','suspended');default:active"`
	CreatedAt                 time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt                 *time.Time `json:"deleted_at" gorm:"index"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// DevOperationLog 开发操作日志
type DevOperationLog struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id"`
	OperationType    string    `json:"operation_type" gorm:"type:varchar(100)"`
	OperationTarget  string    `json:"operation_target" gorm:"type:varchar(255)"`
	OperationDetails string    `json:"operation_details" gorm:"type:json"`
	IPAddress        string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent        string    `json:"user_agent" gorm:"type:text"`
	Status           string    `json:"status" gorm:"type:enum('success','failed','blocked');default:success"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// JWT Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success   bool        `json:"success"`
	Token     string      `json:"token"`
	User      User        `json:"user"`
	DevTeam   DevTeamUser `json:"dev_team,omitempty"`
	ExpiresAt string      `json:"expires_at"`
	Message   string      `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    User   `json:"user,omitempty"`
}

// Role 角色定义
type Role struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

// Permission 权限定义
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Scope    string `json:"scope"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret        string        `json:"jwt_secret"`
	TokenExpiry      time.Duration `json:"token_expiry"`
	RefreshExpiry    time.Duration `json:"refresh_expiry"`
	PasswordMin      int           `json:"password_min_length"`
	MaxLoginAttempts int           `json:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration"`
}

// OAuth2Client OAuth2 客户端
type OAuth2Client struct {
	ID           string    `json:"id" gorm:"primaryKey;column:id;type:varchar(255)"`
	Secret       string    `json:"-" gorm:"column:client_secret;type:varchar(255);not null"`
	Name         string    `json:"name" gorm:"column:name;type:varchar(255);not null"`
	RedirectURIs []string  `json:"redirect_uris" gorm:"column:redirect_uris;type:jsonb;default:'[]'"`
	Scopes       []string  `json:"scopes" gorm:"column:scopes;type:jsonb;default:'[]'"`
	Status       string    `json:"status" gorm:"column:status;type:varchar(50);default:'active'"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (OAuth2Client) TableName() string {
	return "oauth2_clients"
}

// AuthorizationCode 授权码
type AuthorizationCode struct {
	Code                string    `json:"code" gorm:"primaryKey;column:code;type:varchar(255)"`
	ClientID            string    `json:"client_id" gorm:"column:client_id;type:varchar(255);not null;index"`
	UserID              uint      `json:"user_id" gorm:"column:user_id;not null;index"`
	RedirectURI         string    `json:"redirect_uri" gorm:"column:redirect_uri;type:varchar(512);not null"`
	Scope               string    `json:"scope" gorm:"column:scope;type:varchar(255)"`
	CodeChallenge       string    `json:"code_challenge" gorm:"column:code_challenge;type:varchar(255)"`
	CodeChallengeMethod string    `json:"code_challenge_method" gorm:"column:code_challenge_method;type:varchar(10)"`
	ExpiresAt           time.Time `json:"expires_at" gorm:"column:expires_at;not null;index"`
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (AuthorizationCode) TableName() string {
	return "oauth2_authorization_codes"
}

// RefreshToken 刷新令牌
type RefreshToken struct {
	Token     string     `json:"token" gorm:"primaryKey;column:token;type:varchar(255)"`
	ClientID  string     `json:"client_id" gorm:"column:client_id;type:varchar(255);not null;index"`
	UserID    uint       `json:"user_id" gorm:"column:user_id;not null;index"`
	Scope     string     `json:"scope" gorm:"column:scope;type:varchar(255)"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"column:expires_at;not null;index"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	RevokedAt *time.Time `json:"revoked_at" gorm:"column:revoked_at;index"`
}

// TableName 指定表名
func (RefreshToken) TableName() string {
	return "oauth2_refresh_tokens"
}

// AccessToken 访问令牌（可选，如果使用数据库存储）
type AccessToken struct {
	Token     string     `json:"token" gorm:"primaryKey;column:token;type:varchar(255)"`
	ClientID  string     `json:"client_id" gorm:"column:client_id;type:varchar(255);not null;index"`
	UserID    uint       `json:"user_id" gorm:"column:user_id;not null;index"`
	Scope     string     `json:"scope" gorm:"column:scope;type:varchar(255)"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"column:expires_at;not null;index"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	RevokedAt *time.Time `json:"revoked_at" gorm:"column:revoked_at;index"`
}

// TableName 指定表名
func (AccessToken) TableName() string {
	return "oauth2_access_tokens"
}

// OAuth2AuthorizationRequest 授权请求
type OAuth2AuthorizationRequest struct {
	ClientID     string `json:"client_id" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	ResponseType string `json:"response_type" binding:"required"` // "code" for authorization code flow
	Scope        string `json:"scope"`
	State        string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`        // PKCE
	CodeChallengeMethod string `json:"code_challenge_method"` // "plain" or "S256"
}

// OAuth2TokenRequest Token 请求
type OAuth2TokenRequest struct {
	GrantType    string `json:"grant_type" binding:"required"` // "authorization_code" or "refresh_token"
	Code         string `json:"code"`                          // for authorization_code grant
	RedirectURI  string `json:"redirect_uri"`                  // for authorization_code grant
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
	RefreshToken string `json:"refresh_token"`                 // for refresh_token grant
	CodeVerifier string `json:"code_verifier"`                  // PKCE
}

// OAuth2TokenResponse Token 响应
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"` // "Bearer"
	ExpiresIn    int64  `json:"expires_in"` // seconds
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OAuth2UserInfoResponse 用户信息响应
type OAuth2UserInfoResponse struct {
	Sub           string `json:"sub"`           // Subject (user ID)
	Name          string `json:"name"`          // Username
	Email         string `json:"email"`         // Email
	EmailVerified bool   `json:"email_verified"` // Email verified
	Picture       string `json:"picture,omitempty"`
}
