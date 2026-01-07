package vuecmf

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// APIMappingService API 映射服务
type APIMappingService struct {
	db    *sql.DB
	cache *redis.Client
	ctx   context.Context
}

// NewAPIMappingService 创建 API 映射服务
func NewAPIMappingService(db *sql.DB, cache *redis.Client) *APIMappingService {
	return &APIMappingService{
		db:    db,
		cache: cache,
		ctx:   context.Background(),
	}
}

// GetAPIPath 获取 API 路径
func (s *APIMappingService) GetAPIPath(tableName, actionType string, appID int) (string, error) {
	if appID == 0 {
		appID = 1
	}

	// 1. 先查 Redis 缓存
	cacheKey := fmt.Sprintf("api_map:%s:%s:%d", tableName, actionType, appID)
	if s.cache != nil {
		if cached, err := s.cache.Get(s.ctx, cacheKey).Result(); err == nil {
			return cached, nil
		}
	}

	// 2. 查询数据库（使用简化的 vuecmf_api_map 表）
	var apiPath string
	query := `
		SELECT api_path
		FROM vuecmf_api_map
		WHERE table_name = $1 
		AND action_type = $2
	`

	err := s.db.QueryRow(query, tableName, actionType).Scan(&apiPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("API 映射不存在: %s.%s", tableName, actionType)
		}
		return "", fmt.Errorf("查询 API 映射失败: %v", err)
	}

	// 3. 缓存结果（1小时）
	if s.cache != nil {
		s.cache.Set(s.ctx, cacheKey, apiPath, 1*time.Hour)
	}

	return apiPath, nil
}

// GetAllMappings 获取所有 API 映射（用于调试）
func (s *APIMappingService) GetAllMappings() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			table_name,
			action_type,
			api_path,
			request_method,
			COALESCE(note, '') as note
		FROM vuecmf_api_map
		ORDER BY id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var tableName, actionType, apiPath, method, note string
		if err := rows.Scan(&tableName, &actionType, &apiPath, &method, &note); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"table_name":  tableName,
			"action_type": actionType,
			"api_path":    apiPath,
			"method":      method,
			"note":        note,
		})
	}

	return results, nil
}

// ClearCache 清除缓存
func (s *APIMappingService) ClearCache() error {
	if s.cache == nil {
		return nil
	}

	pattern := "api_map:*"
	iter := s.cache.Scan(s.ctx, 0, pattern, 0).Iterator()

	for iter.Next(s.ctx) {
		err := s.cache.Del(s.ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	return nil
}

