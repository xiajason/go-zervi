package auth

import (
	"sync"
	"time"
)

var (
	// tokenInvalidationTime 全局Token失效时间戳
	// 所有在此时间之前签发的Token都将被视为无效
	tokenInvalidationTime int64
	tokenInvalidationMux  sync.RWMutex
)

// init 初始化Token失效时间
func init() {
	// 初始化时设置为当前时间，使所有旧Token失效
	tokenInvalidationTime = time.Now().Unix()
}

// InvalidateAllTokens 使所有Token失效
// 调用此函数后，所有在此时间之前签发的Token都将被视为无效
func InvalidateAllTokens() {
	tokenInvalidationMux.Lock()
	defer tokenInvalidationMux.Unlock()
	
	tokenInvalidationTime = time.Now().Unix()
}

// GetTokenInvalidationTime 获取Token失效时间戳
func GetTokenInvalidationTime() int64 {
	tokenInvalidationMux.RLock()
	defer tokenInvalidationMux.RUnlock()
	
	return tokenInvalidationTime
}

// IsTokenValid 检查Token是否有效
// 如果Token的签发时间（iat）早于全局失效时间，则Token无效
func IsTokenValid(tokenIssuedAt int64) bool {
	tokenInvalidationMux.RLock()
	defer tokenInvalidationMux.RUnlock()
	
	// 如果Token签发时间早于失效时间，则Token无效
	return tokenIssuedAt >= tokenInvalidationTime
}

// SetTokenInvalidationTime 设置Token失效时间（用于测试或特殊场景）
func SetTokenInvalidationTime(timestamp int64) {
	tokenInvalidationMux.Lock()
	defer tokenInvalidationMux.Unlock()
	
	tokenInvalidationTime = timestamp
}




