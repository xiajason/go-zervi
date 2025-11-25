package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/core/response"
)

// ZerviAuthAdapter Go-Zervi认证适配器
// 将jobfirst-core的认证中间件适配到Go-Zervi统一认证系统
type ZerviAuthAdapter struct {
	unifiedAuth *UnifiedAuthSystem
}

// NewZerviAuthAdapter 创建Go-Zervi认证适配器
func NewZerviAuthAdapter(db *sql.DB, jwtSecret string) *ZerviAuthAdapter {
	unifiedAuth := NewUnifiedAuthSystem(db, jwtSecret)
	return &ZerviAuthAdapter{
		unifiedAuth: unifiedAuth,
	}
}

// RequireAuth 需要登录的中间件（适配jobfirst-core接口）
func (adapter *ZerviAuthAdapter) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: Zervi认证中间件 - 开始处理请求: %s %s\n", c.Request.Method, c.Request.URL.Path)

		token := adapter.extractToken(c)
		fmt.Printf("DEBUG: Zervi认证中间件 - 提取到的token: %s\n", func() string {
			if token == "" {
				return "空token"
			}
			if len(token) > 50 {
				return token[:50] + "..."
			}
			return token
		}())

		if token == "" {
			fmt.Printf("DEBUG: Zervi认证中间件 - token为空，返回未登录\n")
			adapter.writeErrorResponse(c, response.Error(response.CodeUnauthorized, "未登录"))
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: Zervi认证中间件 - 开始验证token\n")
		result, err := adapter.unifiedAuth.ValidateJWT(token)
		if err != nil {
			fmt.Printf("DEBUG: Zervi认证中间件 - token验证失败: %v\n", err)
			adapter.writeErrorResponse(c, response.Error(response.CodeUnauthorized, "无效的token"))
			c.Abort()
			return
		}

		if !result.Success {
			fmt.Printf("DEBUG: Zervi认证中间件 - token验证失败: %s\n", result.Error)
			adapter.writeErrorResponse(c, response.Error(response.CodeUnauthorized, result.Error))
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: Zervi认证中间件 - token验证成功，用户ID: %d, 用户名: %s, 角色: %s\n",
			result.User.ID, result.User.Username, result.User.Role)

		// 设置用户信息到上下文（保持与jobfirst-core兼容）
		c.Set("user_id", result.User.ID)
		c.Set("username", result.User.Username)
		c.Set("role", result.User.Role)
		c.Set("email", result.User.Email)

		fmt.Printf("DEBUG: Zervi认证中间件 - 用户信息已设置到上下文，继续处理请求\n")
		c.Next()
	}
}

// RequireDevTeam 需要开发团队权限的中间件（适配jobfirst-core接口）
func (adapter *ZerviAuthAdapter) RequireDevTeam() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已登录
		adapter.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查开发团队权限
		role := c.GetString("role")
		devTeamRoles := []string{"super_admin", "system_admin", "dev_lead", "frontend_dev", "backend_dev", "qa_engineer"}

		hasPermission := false
		for _, devRole := range devTeamRoles {
			if role == devRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			adapter.writeErrorResponse(c, response.Error(response.CodeForbidden, "需要开发团队权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken 提取token（与jobfirst-core保持兼容）
func (adapter *ZerviAuthAdapter) extractToken(c *gin.Context) string {
	// 1. 从Authorization header获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		log.Printf("DEBUG: extractToken - Authorization header: '%s'", authHeader)
		parts := []string{}
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			parts = []string{authHeader[:7], authHeader[7:]}
		} else {
			parts = []string{authHeader}
		}
		log.Printf("DEBUG: extractToken - 分割后的parts: %v, 长度: %d", parts, len(parts))
		if len(parts) == 2 && parts[0] == "Bearer " {
			token := parts[1]
			log.Printf("DEBUG: extractToken - 找到Bearer token: %s", token)
			return token
		}
	} else {
		log.Printf("DEBUG: extractToken - Authorization header为空")
	}

	// 2. 从查询参数获取
	token := c.Query("token")
	if token != "" {
		log.Printf("DEBUG: extractToken - 查询参数token: '%s'", token)
		return token
	} else {
		log.Printf("DEBUG: extractToken - 查询参数token: ''")
	}

	// 3. 从Cookie获取
	cookie, err := c.Cookie("token")
	if err != nil {
		log.Printf("DEBUG: extractToken - Cookie token: '', 错误: %v", err)
	} else {
		log.Printf("DEBUG: extractToken - Cookie token: '%s'", cookie)
		return cookie
	}

	log.Printf("DEBUG: extractToken - 未找到任何token")
	return ""
}

// writeErrorResponse 写入错误响应（使用Go-Zervi标准格式）
func (adapter *ZerviAuthAdapter) writeErrorResponse(c *gin.Context, resp *response.ApiResponse) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, resp) // 使用200状态码，错误信息在响应体中
}

// writeSuccessResponse 写入成功响应（使用Go-Zervi标准格式）
func (adapter *ZerviAuthAdapter) writeSuccessResponse(c *gin.Context, resp *response.ApiResponse) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, resp)
}

// Login 登录方法（适配jobfirst-core接口）
func (adapter *ZerviAuthAdapter) Login(req ZerviLoginRequest, clientIP, userAgent string) (*ZerviLoginResponse, error) {
	result, err := adapter.unifiedAuth.Authenticate(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	// 构建响应（保持与jobfirst-core兼容）
	response := &ZerviLoginResponse{
		Token:       result.Token,
		User:        result.User,
		Permissions: result.Permissions,
	}

	return response, nil
}

// Register 注册方法（适配jobfirst-core接口）
func (adapter *ZerviAuthAdapter) Register(req ZerviRegisterRequest) (*ZerviRegisterResponse, error) {
	// 这里可以实现注册逻辑，或者返回错误表示不支持
	return nil, fmt.Errorf("注册功能暂未实现，请使用统一认证服务")
}

// 兼容性类型定义
type ZerviLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ZerviLoginResponse struct {
	Token       string    `json:"token"`
	User        *UserInfo `json:"user"`
	Permissions []string  `json:"permissions"`
}

type ZerviRegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ZerviRegisterResponse struct {
	User *UserInfo `json:"user"`
}
