package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型（包含租户ID和多租户支持）
// 所有业务模型都应该嵌入此模型
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  int64          `json:"tenant_id" gorm:"column:tenant_id;index;not null;default:1"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`
}

// BeforeCreate GORM Hook: 创建前自动设置tenant_id
// 如果tenant_id为0，从context获取并设置
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.TenantID == 0 {
		// 从context获取tenant_id
		if ctx := tx.Statement.Context; ctx != nil {
			if tenantID, ok := ctx.Value("tenant_id").(int64); ok && tenantID > 0 {
				m.TenantID = tenantID
			} else {
				// 如果context中没有，尝试从gin context获取
				// 注意：这需要在中间件中设置
				m.TenantID = 1 // 默认租户ID
			}
		} else {
			m.TenantID = 1 // 默认租户ID
		}
	}
	return nil
}

// ScopeTenant GORM Scope: 自动过滤租户
// 使用方法: db.Scopes(ScopeTenant(tenantID)).Find(&jobs)
func ScopeTenant(tenantID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// ScopeTenantFromContext GORM Scope: 从context获取租户ID并过滤
// 使用方法: db.Scopes(ScopeTenantFromContext(ctx)).Find(&jobs)
func ScopeTenantFromContext(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID, ok := ctx.Value("tenant_id").(int64); ok && tenantID > 0 {
			return db.Where("tenant_id = ?", tenantID)
		}
		// 如果没有tenant_id，返回原查询（可能需要在应用层处理）
		return db
	}
}

// SetTenantID 设置租户ID（用于手动设置）
func (m *BaseModel) SetTenantID(tenantID int64) {
	m.TenantID = tenantID
}

// GetTenantID 获取租户ID
func (m *BaseModel) GetTenantID() int64 {
	return m.TenantID
}

