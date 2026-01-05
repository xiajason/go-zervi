package context

import (
	"context"
	"errors"
)

// tenantIDKey 用于在context中存储租户ID的key类型
type tenantIDKey struct{}

var (
	// ErrNoTenantPermission 表示没有租户权限
	ErrNoTenantPermission = errors.New("no tenant permission")
	// ErrTenantNotFound 表示租户未找到
	ErrTenantNotFound = errors.New("tenant not found")
)

// GetTenantID 从context获取租户ID
// 如果context中没有租户ID，返回ErrTenantNotFound错误
func GetTenantID(ctx context.Context) (int64, error) {
	tenantID, ok := ctx.Value(tenantIDKey{}).(int64)
	if !ok || tenantID == 0 {
		return 0, ErrTenantNotFound
	}
	return tenantID, nil
}

// SetTenantID 设置租户ID到context
// 返回一个新的context，包含租户ID
func SetTenantID(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

// MustGetTenantID 获取租户ID（如果不存在则panic）
// 仅在确定context中一定有租户ID时使用
func MustGetTenantID(ctx context.Context) int64 {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		panic(err)
	}
	return tenantID
}

// WithTenantID 创建带租户ID的context（便捷方法）
func WithTenantID(ctx context.Context, tenantID int64) context.Context {
	return SetTenantID(ctx, tenantID)
}

