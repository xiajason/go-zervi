package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantCodeExists    = errors.New("tenant code already exists")
	ErrUserNotInTenant     = errors.New("user not in tenant")
	ErrInvalidTenantStatus = errors.New("invalid tenant status")
	ErrInvalidTenantRole   = errors.New("invalid tenant role")
)

// TenantService 租户管理服务
type TenantService struct {
	db *gorm.DB
}

// NewTenantService 创建租户服务实例
func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{db: db}
}

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name        string         `json:"name" binding:"required"`
	Code        string         `json:"code" binding:"required"`
	Description string         `json:"description"`
	Settings    TenantSettings `json:"settings"`
	Quota       TenantQuota    `json:"quota"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Status      *string         `json:"status"`
	Settings    *TenantSettings `json:"settings"`
	Quota       *TenantQuota    `json:"quota"`
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	Items      []Tenant `json:"items"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req CreateTenantRequest, creatorUserID int64) (*Tenant, error) {
	// 检查code是否已存在
	var count int64
	if err := s.db.WithContext(ctx).Model(&Tenant{}).
		Where("code = ?", req.Code).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check tenant code: %w", err)
	}
	if count > 0 {
		return nil, ErrTenantCodeExists
	}

	// 创建租户
	tenant := Tenant{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      TenantStatusActive,
		Settings:    req.Settings,
		Quota:       req.Quota,
	}

	if err := s.db.WithContext(ctx).Create(&tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 将创建者添加为租户所有者
	userTenant := UserTenant{
		UserID:   creatorUserID,
		TenantID: tenant.ID,
		Role:     TenantRoleOwner,
		Status:   UserTenantStatusActive,
		JoinedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&userTenant).Error; err != nil {
		// 如果添加用户失败，删除刚创建的租户
		s.db.WithContext(ctx).Delete(&tenant)
		return nil, fmt.Errorf("failed to add creator to tenant: %w", err)
	}

	return &tenant, nil
}

// GetTenant 获取租户详情
func (s *TenantService) GetTenant(ctx context.Context, tenantID int64) (*Tenant, error) {
	var tenant Tenant
	if err := s.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return &tenant, nil
}

// GetTenantByCode 根据code获取租户
func (s *TenantService) GetTenantByCode(ctx context.Context, code string) (*Tenant, error) {
	var tenant Tenant
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by code: %w", err)
	}
	return &tenant, nil
}

// ListTenants 获取租户列表（分页）
func (s *TenantService) ListTenants(ctx context.Context, page, pageSize int, status string) (*TenantListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&Tenant{})

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count tenants: %w", err)
	}

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	// 查询数据
	var tenants []Tenant
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return &TenantListResponse{
		Items:      tenants,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID int64, req UpdateTenantRequest) (*Tenant, error) {
	var tenant Tenant
	if err := s.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Description != nil {
		tenant.Description = *req.Description
	}
	if req.Status != nil {
		// 验证状态
		validStatuses := []string{TenantStatusActive, TenantStatusSuspended, TenantStatusDeleted}
		isValid := false
		for _, vs := range validStatuses {
			if *req.Status == vs {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, ErrInvalidTenantStatus
		}
		tenant.Status = *req.Status
	}
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	if req.Quota != nil {
		tenant.Quota = *req.Quota
	}

	tenant.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return &tenant, nil
}

// DeleteTenant 删除租户（软删除）
func (s *TenantService) DeleteTenant(ctx context.Context, tenantID int64) error {
	var tenant Tenant
	if err := s.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// 软删除
	if err := s.db.WithContext(ctx).Delete(&tenant).Error; err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// AddUserToTenant 添加用户到租户
func (s *TenantService) AddUserToTenant(ctx context.Context, tenantID, userID int64, role string) error {
	// 验证角色
	validRoles := []string{TenantRoleOwner, TenantRoleAdmin, TenantRoleMember, TenantRoleGuest}
	isValidRole := false
	for _, vr := range validRoles {
		if role == vr {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		return ErrInvalidTenantRole
	}

	// 检查是否已存在
	var count int64
	if err := s.db.WithContext(ctx).Model(&UserTenant{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check user tenant: %w", err)
	}
	if count > 0 {
		return errors.New("user already in tenant")
	}

	userTenant := UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		Status:   UserTenantStatusActive,
		JoinedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&userTenant).Error; err != nil {
		return fmt.Errorf("failed to add user to tenant: %w", err)
	}

	return nil
}

// RemoveUserFromTenant 从租户移除用户
func (s *TenantService) RemoveUserFromTenant(ctx context.Context, tenantID, userID int64) error {
	result := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Delete(&UserTenant{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to remove user from tenant: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotInTenant
	}

	return nil
}

// GetUserTenants 获取用户所属的租户列表
func (s *TenantService) GetUserTenants(ctx context.Context, userID int64) ([]Tenant, error) {
	var tenants []Tenant
	if err := s.db.WithContext(ctx).
		Table("zervigo_tenants").
		Joins("INNER JOIN zervigo_user_tenants ON zervigo_tenants.id = zervigo_user_tenants.tenant_id").
		Where("zervigo_user_tenants.user_id = ? AND zervigo_user_tenants.status = ?", userID, UserTenantStatusActive).
		Where("zervigo_tenants.status = ?", TenantStatusActive).
		Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}
	return tenants, nil
}

// GetTenantUsers 获取租户的用户列表
func (s *TenantService) GetTenantUsers(ctx context.Context, tenantID int64) ([]UserTenant, error) {
	var userTenants []UserTenant
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID, UserTenantStatusActive).
		Find(&userTenants).Error; err != nil {
		return nil, fmt.Errorf("failed to get tenant users: %w", err)
	}
	return userTenants, nil
}

// SwitchTenant 切换租户（更新用户的当前租户ID）
func (s *TenantService) SwitchTenant(ctx context.Context, userID, tenantID int64) error {
	// 验证用户是否属于该租户
	var count int64
	if err := s.db.WithContext(ctx).Model(&UserTenant{}).
		Where("user_id = ? AND tenant_id = ? AND status = ?", userID, tenantID, UserTenantStatusActive).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check user tenant: %w", err)
	}
	if count == 0 {
		return ErrUserNotInTenant
	}

	// 验证租户状态
	var tenant Tenant
	if err := s.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant.Status != TenantStatusActive {
		return errors.New("tenant is not active")
	}

	// 更新用户的last_tenant_id（更新zervigo_auth_users表）
	// 先获取当前租户ID作为last_tenant_id
	var currentTenantID *int64
	var userTenant UserTenant
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, UserTenantStatusActive).
		Order("joined_at DESC").
		First(&userTenant).Error; err == nil {
		currentTenantID = &userTenant.TenantID
	}

	// 更新用户表的last_tenant_id
	// 注意：tenant_id字段在User模型中标记为gorm:"-"，表示不存储到数据库
	// 所以这里只更新last_tenant_id字段
	updateData := map[string]interface{}{
		"updated_at": time.Now(),
	}
	
	// 更新last_tenant_id字段
	if currentTenantID != nil {
		updateData["last_tenant_id"] = *currentTenantID
	} else {
		// 如果没有当前租户，将新租户ID设置为last_tenant_id
		updateData["last_tenant_id"] = tenantID
	}

	// 使用原生SQL更新zervigo_auth_users表
	if err := s.db.WithContext(ctx).
		Table("zervigo_auth_users").
		Where("id = ?", userID).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update user tenant: %w", err)
	}

	return nil
}

// CheckUserTenantPermission 检查用户在租户中的权限
func (s *TenantService) CheckUserTenantPermission(ctx context.Context, userID, tenantID int64) (*UserTenant, error) {
	var userTenant UserTenant
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND status = ?", userID, tenantID, UserTenantStatusActive).
		First(&userTenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotInTenant
		}
		return nil, fmt.Errorf("failed to check user tenant permission: %w", err)
	}
	return &userTenant, nil
}

