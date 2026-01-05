package permission

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PermissionClient Permission Service客户端
type PermissionClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPermissionClient 创建Permission Service客户端
func NewPermissionClient(baseURL string) *PermissionClient {
	return &PermissionClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID     string `json:"user_id"`
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	Permission string `json:"permission,omitempty"`
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Allowed bool   `json:"allowed"`
		Reason  string `json:"reason,omitempty"`
	} `json:"data"`
}

// PermissionResponse Permission Service响应格式
type PermissionResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UserPermissions 用户权限信息
type UserPermissions struct {
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// GetUserRoles 获取用户角色列表
func (c *PermissionClient) GetUserRoles(userID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s/roles", c.baseURL, userID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Permission Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Permission Service返回错误状态码: %d", resp.StatusCode)
	}

	var permResp PermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&permResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if permResp.Code != 0 {
		return nil, fmt.Errorf("Permission Service返回错误: %s", permResp.Message)
	}

	// 转换数据
	rolesJSON, _ := json.Marshal(permResp.Data)
	var roles []map[string]interface{}
	if err := json.Unmarshal(rolesJSON, &roles); err != nil {
		return nil, fmt.Errorf("解析用户角色失败: %v", err)
	}

	var roleNames []string
	for _, role := range roles {
		if roleName, ok := role["roleName"].(string); ok {
			roleNames = append(roleNames, roleName)
		}
	}

	return roleNames, nil
}

// GetUserPermissions 获取用户权限列表（通过角色）
func (c *PermissionClient) GetUserPermissions(userID string) (*UserPermissions, error) {
	// 先获取用户角色
	roles, err := c.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	// 获取每个角色的权限
	var allPermissions []string
	permissionSet := make(map[string]bool)

	for _, roleName := range roles {
		// 获取角色ID（需要先调用获取角色详情）
		rolePerms, err := c.getRolePermissionsByRoleName(roleName)
		if err == nil {
			for _, perm := range rolePerms {
				if !permissionSet[perm] {
					permissionSet[perm] = true
					allPermissions = append(allPermissions, perm)
				}
			}
		}
	}

	return &UserPermissions{
		UserID:      userID,
		Roles:       roles,
		Permissions: allPermissions,
	}, nil
}

// getRolePermissionsByRoleName 通过角色名获取权限（内部方法）
func (c *PermissionClient) getRolePermissionsByRoleName(roleName string) ([]string, error) {
	// 先获取所有角色，找到对应的角色ID
	url := fmt.Sprintf("%s/api/v1/roles", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rolesResp PermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&rolesResp); err != nil {
		return nil, err
	}

	if rolesResp.Code != 0 {
		return nil, fmt.Errorf("获取角色列表失败: %s", rolesResp.Message)
	}

	// 查找角色ID
	rolesJSON, _ := json.Marshal(rolesResp.Data)
	var roles []map[string]interface{}
	json.Unmarshal(rolesJSON, &roles)

	var roleID string
	for _, role := range roles {
		if name, ok := role["roleName"].(string); ok && name == roleName {
			if id, ok := role["id"].(float64); ok {
				roleID = fmt.Sprintf("%.0f", id)
				break
			}
		}
	}

	if roleID == "" {
		return nil, fmt.Errorf("角色不存在: %s", roleName)
	}

	// 获取角色权限
	return c.GetRolePermissions(roleID)
}

// GetRolePermissions 获取角色权限列表
func (c *PermissionClient) GetRolePermissions(roleID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/roles/%s/permissions", c.baseURL, roleID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Permission Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Permission Service返回错误状态码: %d", resp.StatusCode)
	}

	var permResp PermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&permResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if permResp.Code != 0 {
		return nil, fmt.Errorf("Permission Service返回错误: %s", permResp.Message)
	}

	// 转换数据
	permsJSON, _ := json.Marshal(permResp.Data)
	var permissions []map[string]interface{}
	if err := json.Unmarshal(permsJSON, &permissions); err != nil {
		return nil, fmt.Errorf("解析角色权限失败: %v", err)
	}

	var permissionCodes []string
	for _, perm := range permissions {
		if code, ok := perm["permissionCode"].(string); ok {
			permissionCodes = append(permissionCodes, code)
		}
	}

	return permissionCodes, nil
}

// GetAllRoles 获取所有角色列表（公开）
func (c *PermissionClient) GetAllRoles() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/roles", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Permission Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Permission Service返回错误状态码: %d", resp.StatusCode)
	}

	var permResp PermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&permResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if permResp.Code != 0 {
		return nil, fmt.Errorf("Permission Service返回错误: %s", permResp.Message)
	}

	// 转换数据
	rolesJSON, _ := json.Marshal(permResp.Data)
	var roles []map[string]interface{}
	if err := json.Unmarshal(rolesJSON, &roles); err != nil {
		return nil, fmt.Errorf("解析角色列表失败: %v", err)
	}

	return roles, nil
}

// GetAllPermissions 获取所有权限列表（公开）
func (c *PermissionClient) GetAllPermissions() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/permissions", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Permission Service失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Permission Service返回错误状态码: %d", resp.StatusCode)
	}

	var permResp PermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&permResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if permResp.Code != 0 {
		return nil, fmt.Errorf("Permission Service返回错误: %s", permResp.Message)
	}

	// 转换数据
	permsJSON, _ := json.Marshal(permResp.Data)
	var permissions []map[string]interface{}
	if err := json.Unmarshal(permsJSON, &permissions); err != nil {
		return nil, fmt.Errorf("解析权限列表失败: %v", err)
	}

	return permissions, nil
}
