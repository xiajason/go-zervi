package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/szjason72/zervigo/shared/core/response"
)

// AuthClient 认证服务客户端
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAuthClient 创建认证服务客户端
func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ValidateTokenRequest 验证Token请求
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse 验证Token响应
type ValidateTokenResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *AuthResult `json:"data"`
}

// ValidateToken 验证JWT Token（调用auth-service）
func (c *AuthClient) ValidateToken(token string) (*AuthResult, error) {
	// 构建请求
	reqBody := ValidateTokenRequest{
		Token: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求到auth-service
	url := fmt.Sprintf("%s/api/v1/auth/validate", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("请求auth-service失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var validateResp ValidateTokenResponse
	if err := json.Unmarshal(body, &validateResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if validateResp.Code != response.CodeSuccess {
		return &AuthResult{
			Success:   false,
			Error:     validateResp.Message,
			ErrorCode: fmt.Sprintf("ERR_%d", validateResp.Code),
		}, nil
	}

	// 返回验证结果
	if validateResp.Data == nil {
		return &AuthResult{
			Success:   false,
			Error:     "验证结果为空",
			ErrorCode: "EMPTY_RESULT",
		}, nil
	}

	return validateResp.Data, nil
}

// GetUserRequest 获取用户信息请求
type GetUserRequest struct {
	UserID int `json:"user_id"`
}

// GetUser 获取用户信息（调用auth-service）
func (c *AuthClient) GetUser(userID int) (*UserInfo, error) {
	// 发送GET请求到auth-service
	url := fmt.Sprintf("%s/api/v1/auth/user?user_id=%d", c.baseURL, userID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求auth-service失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 调试：打印响应内容
	fmt.Printf("DEBUG: AuthClient.GetUser - 响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: AuthClient.GetUser - 响应内容: %s\n", string(body))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var user UserInfo
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 响应内容: %s", err, string(body))
	}

	return &user, nil
}

// CheckPermission 检查用户权限（调用auth-service）
func (c *AuthClient) CheckPermission(userID int, permission string) (bool, error) {
	// 发送GET请求到auth-service
	url := fmt.Sprintf("%s/api/v1/auth/permission?user_id=%d&permission=%s", c.baseURL, userID, permission)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("请求auth-service失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var permResp struct {
		UserID        int    `json:"user_id"`
		Permission    string `json:"permission"`
		HasPermission bool   `json:"has_permission"`
	}

	if err := json.Unmarshal(body, &permResp); err != nil {
		return false, fmt.Errorf("解析响应失败: %w", err)
	}

	return permResp.HasPermission, nil
}
