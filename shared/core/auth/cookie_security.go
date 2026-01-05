package auth

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

// CookieSecurityConfig Cookie安全配置
type CookieSecurityConfig struct {
	// Cookie名称
	Name string
	// 过期时间（秒）
	MaxAge int
	// HttpOnly: 防止XSS攻击
	HttpOnly bool
	// Secure: HTTPS only（生产环境应为true）
	Secure bool
	// SameSite: CSRF保护
	SameSite http.SameSite
	// Path: Cookie路径
	Path string
	// Domain: Cookie域名（可选）
	Domain string
}

// DefaultCookieConfig 获取默认Cookie配置
func DefaultCookieConfig() *CookieSecurityConfig {
	// 从环境变量读取配置
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	secure := isProduction // 生产环境使用secure=True

	maxAge := 7 * 24 * 60 * 60 // 默认7天
	if maxAgeEnv := os.Getenv("COOKIE_MAX_AGE"); maxAgeEnv != "" {
		if parsed, err := strconv.Atoi(maxAgeEnv); err == nil {
			maxAge = parsed
		}
	}

	return &CookieSecurityConfig{
		Name:     "access_token",
		MaxAge:   maxAge,
		HttpOnly: true,                    // 防止XSS攻击
		Secure:   secure,                   // 生产环境使用HTTPS
		SameSite: http.SameSiteLaxMode,    // CSRF保护
		Path:     "/",
		Domain:   "",                       // 不设置域名，使用当前域名
	}
}

// SetSecureCookie 设置安全Cookie
func SetSecureCookie(w http.ResponseWriter, config *CookieSecurityConfig, value string) {
	if config == nil {
		config = DefaultCookieConfig()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.Name,
		Value:    value,
		MaxAge:   config.MaxAge,
		HttpOnly: config.HttpOnly,
		Secure:   config.Secure,
		SameSite: config.SameSite,
		Path:     config.Path,
		Domain:   config.Domain,
		Expires:  time.Now().Add(time.Duration(config.MaxAge) * time.Second),
	})
}

// DeleteSecureCookie 删除安全Cookie
func DeleteSecureCookie(w http.ResponseWriter, config *CookieSecurityConfig) {
	if config == nil {
		config = DefaultCookieConfig()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.Name,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: config.HttpOnly,
		Secure:   config.Secure,
		SameSite: config.SameSite,
		Path:     config.Path,
		Domain:   config.Domain,
		Expires:  time.Unix(0, 0),
	})
}




