package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ServiceHandshake 服务握手信息
type ServiceHandshake struct {
	ServiceID       string
	ServiceName     string
	ServiceSecret   string
	ServiceType     string
	CentralBrainURL string
	AuthServiceURL  string
	Timeout         time.Duration
}

// HandshakeResult 握手结果
type HandshakeResult struct {
	Success      bool   `json:"success"`
	ServiceToken string `json:"service_token,omitempty"`
	Error        string `json:"error,omitempty"`
}

// Handshake 执行服务握手
func Handshake(config *ServiceHandshake) (*HandshakeResult, error) {
	// 1. 向Auth Service注册并获取Service Token
	fmt.Printf("DEBUG: Service Handshake - 开始握手: ServiceID=%s\n", config.ServiceID)

	// 构建请求
	reqBody := map[string]string{
		"service_id":     config.ServiceID,
		"service_secret": config.ServiceSecret,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求到Auth Service
	url := fmt.Sprintf("%s/api/v1/auth/service/login", config.AuthServiceURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("请求auth-service失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	fmt.Printf("DEBUG: Service Handshake - Auth Service响应: %s\n", string(body))

	// 解析响应
	var authResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ServiceToken string `json:"service_token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 响应: %s", err, string(body))
	}

	// 检查响应状态
	if authResp.Code != 0 {
		return &HandshakeResult{
			Success: false,
			Error:   authResp.Message,
		}, nil
	}

	fmt.Printf("DEBUG: Service Handshake - 握手成功，获取Service Token: %s\n",
		func() string {
			if len(authResp.Data.ServiceToken) > 50 {
				return authResp.Data.ServiceToken[:50] + "..."
			}
			return authResp.Data.ServiceToken
		}())

	// 2. 注册到Central Brain（可选）
	if config.CentralBrainURL != "" {
		registerURL := strings.TrimRight(config.CentralBrainURL, "/") + "/internal/v1/services/register"
		serviceType := config.ServiceType
		if serviceType == "" {
			serviceType = "core"
		}

		payload := map[string]string{
			"service_id":   config.ServiceID,
			"service_name": config.ServiceName,
			"service_type": serviceType,
		}

		body, _ := json.Marshal(payload)
		req, err := http.NewRequest(http.MethodPost, registerURL, bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("WARN: 构建向Central Brain注册请求失败: %v\n", err)
		} else {
			req.Header.Set("Content-Type", "application/json")
			if authResp.Data.ServiceToken != "" {
				req.Header.Set("Authorization", "Bearer "+authResp.Data.ServiceToken)
			}

			timeout := config.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}

			client := &http.Client{Timeout: timeout}
			if resp, err := client.Do(req); err != nil {
				fmt.Printf("WARN: 注册服务到Central Brain失败: %v\n", err)
			} else {
				defer resp.Body.Close()
				if resp.StatusCode >= 300 {
					bodyBytes, _ := io.ReadAll(resp.Body)
					fmt.Printf("WARN: Central Brain注册返回非成功状态 %d: %s\n", resp.StatusCode, string(bodyBytes))
				} else {
					fmt.Printf("DEBUG: Service Handshake - 已向Central Brain注册服务 %s\n", config.ServiceID)
				}
			}
		}
	}

	return &HandshakeResult{
		Success:      true,
		ServiceToken: authResp.Data.ServiceToken,
	}, nil
}

// ValidateServiceToken 验证Service Token
func ValidateServiceToken(token, authServiceURL string) (bool, error) {
	// 构建验证请求
	reqBody := map[string]string{
		"service_token": token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	// 发送验证请求
	url := fmt.Sprintf("%s/api/v1/auth/service/validate", authServiceURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// 解析响应
	var validateResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &validateResp); err != nil {
		return false, err
	}

	return validateResp.Code == 0 && validateResp.Data.Valid, nil
}
