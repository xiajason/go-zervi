package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ServiceCredentials 服务凭证
type ServiceCredentials struct {
	ID            int       `json:"id" db:"id"`
	ServiceID     string    `json:"service_id" db:"service_id"`
	ServiceName   string    `json:"service_name" db:"service_name"`
	ServiceType   string    `json:"service_type" db:"service_type"`
	ServiceSecret string    `json:"-" db:"service_secret"`
	Description   string    `json:"description" db:"description"`
	AllowedAPIs   []string  `json:"allowed_apis" db:"allowed_apis"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// ServiceTokenClaims 服务Token Claims
type ServiceTokenClaims struct {
	ServiceID   string   `json:"service_id"`
	ServiceName string   `json:"service_name"`
	ServiceType string   `json:"service_type"`
	AllowedAPIs []string `json:"allowed_apis"`
	jwt.RegisteredClaims
}

// ServiceAuthResult 服务认证结果
type ServiceAuthResult struct {
	Success   bool                `json:"success"`
	Service   *ServiceCredentials `json:"service,omitempty"`
	Token     string              `json:"token,omitempty"`
	ExpiresIn int64               `json:"expires_in,omitempty"`
	Error     string              `json:"error,omitempty"`
	ErrorCode string              `json:"error_code,omitempty"`
}

// ServiceAuthService 服务认证服务
type ServiceAuthService struct {
	db               *sql.DB
	serviceJWTSecret string // zervigo-mvp-secret-key-2025
}

// NewServiceAuthService 创建服务认证服务
func NewServiceAuthService(db *sql.DB, serviceJWTSecret string) *ServiceAuthService {
	return &ServiceAuthService{
		db:               db,
		serviceJWTSecret: serviceJWTSecret,
	}
}

// AuthenticateService 服务认证
func (sas *ServiceAuthService) AuthenticateService(serviceID, serviceSecret string) (*ServiceAuthResult, error) {
	// 查询服务凭证
	service, err := sas.getServiceCredentials(serviceID)
	if err != nil {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "服务不存在",
			ErrorCode: "SERVICE_NOT_FOUND",
		}, nil
	}

	// 检查服务状态
	if service.Status != "active" {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "服务已被禁用",
			ErrorCode: "SERVICE_DISABLED",
		}, nil
	}

	// 验证服务密钥
	if err := bcrypt.CompareHashAndPassword([]byte(service.ServiceSecret), []byte(serviceSecret)); err != nil {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "服务密钥错误",
			ErrorCode: "INVALID_SECRET",
		}, nil
	}

	// 更新最后使用时间
	sas.updateLastUsed(serviceID)

	// 生成服务token
	token, expiresIn, err := sas.generateServiceToken(service)
	if err != nil {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "生成token失败",
			ErrorCode: "TOKEN_ERROR",
		}, nil
	}

	return &ServiceAuthResult{
		Success:   true,
		Service:   service,
		Token:     token,
		ExpiresIn: expiresIn,
	}, nil
}

// ValidateServiceToken 验证服务token
func (sas *ServiceAuthService) ValidateServiceToken(tokenString string) (*ServiceAuthResult, error) {
	// 解析JWT token
	token, err := jwt.ParseWithClaims(tokenString, &ServiceTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(sas.serviceJWTSecret), nil
	})

	if err != nil {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "无效的token",
			ErrorCode: "INVALID_TOKEN",
		}, nil
	}

	claims, ok := token.Claims.(*ServiceTokenClaims)
	if !ok || !token.Valid {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "token验证失败",
			ErrorCode: "TOKEN_VALIDATION_FAILED",
		}, nil
	}

	// 检查token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "token已过期",
			ErrorCode: "TOKEN_EXPIRED",
		}, nil
	}

	// 查询服务信息
	service, err := sas.getServiceCredentials(claims.ServiceID)
	if err != nil {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "服务不存在",
			ErrorCode: "SERVICE_NOT_FOUND",
		}, nil
	}

	// 检查服务状态
	if service.Status != "active" {
		return &ServiceAuthResult{
			Success:   false,
			Error:     "服务已被禁用",
			ErrorCode: "SERVICE_DISABLED",
		}, nil
	}

	// 更新最后使用时间
	sas.updateLastUsed(claims.ServiceID)

	return &ServiceAuthResult{
		Success: true,
		Service: service,
	}, nil
}

// getServiceCredentials 获取服务凭证
func (sas *ServiceAuthService) getServiceCredentials(serviceID string) (*ServiceCredentials, error) {
	query := `
		SELECT id, service_id, service_name, service_type, service_secret, 
		       COALESCE(description, ''), allowed_apis, status, created_at, updated_at
		FROM zervigo_service_credentials
		WHERE service_id = $1
	`

	var service ServiceCredentials
	var allowedAPIs pq.StringArray
	err := sas.db.QueryRow(query, serviceID).Scan(
		&service.ID,
		&service.ServiceID,
		&service.ServiceName,
		&service.ServiceType,
		&service.ServiceSecret,
		&service.Description,
		&allowedAPIs,
		&service.Status,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// 处理allowed_apis数组
	if len(allowedAPIs) > 0 {
		service.AllowedAPIs = []string(allowedAPIs)
	} else {
		service.AllowedAPIs = []string{"*"} // 默认允许所有API
	}

	return &service, nil
}

// generateServiceToken 生成服务token
func (sas *ServiceAuthService) generateServiceToken(service *ServiceCredentials) (string, int64, error) {
	// Token有效期：24小时
	expiresIn := int64(24 * 60 * 60)
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := &ServiceTokenClaims{
		ServiceID:   service.ServiceID,
		ServiceName: service.ServiceName,
		ServiceType: service.ServiceType,
		AllowedAPIs: service.AllowedAPIs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zervigo-service-auth",
			Subject:   service.ServiceID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(sas.serviceJWTSecret))
	if err != nil {
		return "", 0, err
	}

	// 记录token（可选）
	go sas.recordServiceToken(service.ServiceID, tokenString, expiresAt)

	return tokenString, expiresIn, nil
}

// updateLastUsed 更新最后使用时间
func (sas *ServiceAuthService) updateLastUsed(serviceID string) {
	query := `UPDATE zervigo_service_credentials SET last_used_at = CURRENT_TIMESTAMP WHERE service_id = $1`
	sas.db.Exec(query, serviceID)
}

// recordServiceToken 记录服务token（用于审计）
func (sas *ServiceAuthService) recordServiceToken(serviceID, tokenString string, expiresAt time.Time) {
	// 计算token的hash值（不存储明文）
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	query := `
		INSERT INTO zervigo_service_tokens (service_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	sas.db.Exec(query, serviceID, tokenHash, expiresAt)
}

// CheckServicePermission 检查服务是否有权限调用特定API
func (sas *ServiceAuthService) CheckServicePermission(serviceID, apiPath string) (bool, error) {
	service, err := sas.getServiceCredentials(serviceID)
	if err != nil {
		return false, err
	}

	// 如果allowed_apis包含"*"，表示允许所有API
	for _, allowedAPI := range service.AllowedAPIs {
		if allowedAPI == "*" {
			return true, nil
		}
		if allowedAPI == apiPath {
			return true, nil
		}
	}

	return false, nil
}
