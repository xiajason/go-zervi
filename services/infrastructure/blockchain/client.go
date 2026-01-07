package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 区块链服务客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建区块链服务客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RecordVersionStatusChange 记录版本状态变化
func (c *Client) RecordVersionStatusChange(ctx context.Context, req *VersionStatusChangeRequest) (*BlockchainResponse, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/version/status/record", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var blockchainResp BlockchainResponse
	if err := json.Unmarshal(body, &blockchainResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %v", err)
	}

	return &blockchainResp, nil
}

// RecordPermissionChange 记录权限变更
func (c *Client) RecordPermissionChange(ctx context.Context, req *PermissionChangeRequest) (*BlockchainResponse, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/permission/change/record", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var blockchainResp BlockchainResponse
	if err := json.Unmarshal(body, &blockchainResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %v", err)
	}

	return &blockchainResp, nil
}

// GetUserStatusHistory 获取用户状态历史
func (c *Client) GetUserStatusHistory(ctx context.Context, userID string) (*BlockchainResponse, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/version/status/history/%s", c.baseURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var blockchainResp BlockchainResponse
	if err := json.Unmarshal(body, &blockchainResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %v", err)
	}

	return &blockchainResp, nil
}

// GetUserPermissionHistory 获取用户权限历史
func (c *Client) GetUserPermissionHistory(ctx context.Context, userID string) (*BlockchainResponse, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/permission/change/history/%s", c.baseURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var blockchainResp BlockchainResponse
	if err := json.Unmarshal(body, &blockchainResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %v", err)
	}

	return &blockchainResp, nil
}

// ValidateDataConsistency 数据一致性校验
func (c *Client) ValidateDataConsistency(ctx context.Context) (*BlockchainResponse, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/consistency/validate", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var blockchainResp BlockchainResponse
	if err := json.Unmarshal(body, &blockchainResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %v", err)
	}

	return &blockchainResp, nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) (*BlockchainResponse, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d, %s", resp.StatusCode, string(body))
	}

	var blockchainResp BlockchainResponse
	if err := json.Unmarshal(body, &blockchainResp); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %v", err)
	}

	return &blockchainResp, nil
}
