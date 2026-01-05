package main

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Tenant 租户模型
type Tenant struct {
	ID          int64          `json:"id" gorm:"primaryKey;column:id"`
	Name        string         `json:"name" gorm:"column:name;size:100;not null"`
	Code        string         `json:"code" gorm:"column:code;size:50;uniqueIndex;not null"`
	Description string         `json:"description" gorm:"column:description;type:text"`
	Status      string         `json:"status" gorm:"column:status;size:20;default:active;index"`
	Settings    TenantSettings `json:"settings" gorm:"column:settings;type:jsonb;default:'{}'"`
	Quota       TenantQuota    `json:"quota" gorm:"column:quota;type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"column:deleted_at;index"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "zervigo_tenants"
}

// TenantSettings 租户配置（JSON格式）
type TenantSettings struct {
	Theme          string            `json:"theme,omitempty"`
	Language       string            `json:"language,omitempty"`
	Timezone       string            `json:"timezone,omitempty"`
	CustomSettings map[string]interface{} `json:"custom_settings,omitempty"`
}

// Value 实现driver.Valuer接口
func (s TenantSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan 实现sql.Scanner接口
func (s *TenantSettings) Scan(value interface{}) error {
	if value == nil {
		*s = TenantSettings{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// TenantQuota 租户配额（JSON格式）
type TenantQuota struct {
	MaxUsers       int `json:"max_users,omitempty"`
	MaxStorage     int `json:"max_storage,omitempty"` // MB
	MaxAPICalls    int `json:"max_api_calls,omitempty"`
	MaxAICalls     int `json:"max_ai_calls,omitempty"`
	CustomQuota    map[string]interface{} `json:"custom_quota,omitempty"`
}

// Value 实现driver.Valuer接口
func (q TenantQuota) Value() (driver.Value, error) {
	return json.Marshal(q)
}

// Scan 实现sql.Scanner接口
func (q *TenantQuota) Scan(value interface{}) error {
	if value == nil {
		*q = TenantQuota{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, q)
}

// UserTenant 用户-租户关联模型
type UserTenant struct {
	ID        int64         `json:"id" gorm:"primaryKey;column:id"`
	UserID    int64         `json:"user_id" gorm:"column:user_id;not null;index"`
	TenantID  int64         `json:"tenant_id" gorm:"column:tenant_id;not null;index"`
	Role      string        `json:"role" gorm:"column:role;size:50;default:member;index"`
	Status    string        `json:"status" gorm:"column:status;size:20;default:active;index"`
	JoinedAt  time.Time     `json:"joined_at" gorm:"column:joined_at;autoCreateTime"`
	CreatedAt time.Time     `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (UserTenant) TableName() string {
	return "zervigo_user_tenants"
}

// TenantStatus 租户状态常量
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusDeleted   = "deleted"
)

// TenantRole 租户角色常量
const (
	TenantRoleOwner  = "owner"
	TenantRoleAdmin  = "admin"
	TenantRoleMember = "member"
	TenantRoleGuest  = "guest"
)

// UserTenantStatus 用户-租户状态常量
const (
	UserTenantStatusActive   = "active"
	UserTenantStatusInactive = "inactive"
)

