package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AIEnhancer AI增强器 - 为中央大脑提供智能能力
type AIEnhancer struct {
	mu sync.RWMutex

	// 用户行为追踪
	userSessions map[int]*UserSession // userID -> session

	// 路径访问统计（用于AI分析）
	pathStats map[string]*PathAIStats

	// 预测缓存
	predictions map[string]*Prediction

	// 智能缓存
	smartCache map[string]*CachedData
}

// UserSession 用户会话
type UserSession struct {
	UserID      int
	SessionID   string
	StartTime   time.Time
	LastActive  time.Time
	Actions     []UserAction
	CurrentPath string
}

// UserAction 用户操作
type UserAction struct {
	Timestamp  time.Time
	ActionType string // "api_call", "page_view"
	Path       string
	Duration   int64 // ms
	Success    bool
}

// PathAIStats AI需要的路径统计
type PathAIStats struct {
	Path            string
	TotalAccess     int64
	AvgDuration     int64
	LastAccessTime  time.Time
	AccessFrequency float64 // 访问频率（次/分钟）
	DataChangeRate  float64 // 数据变化率（0-1）
	UserDistribution map[int]int // 哪些用户访问了这个路径
}

// Prediction 预测结果
type Prediction struct {
	CurrentPath  string
	NextPath     string
	Probability  float64
	PreloadData  interface{}
	GeneratedAt  time.Time
}

// CachedData 智能缓存数据
type CachedData struct {
	Data       interface{}
	CachedAt   time.Time
	ExpiresAt  time.Time
	HitCount   int64
}

// NewAIEnhancer 创建AI增强器
func NewAIEnhancer() *AIEnhancer {
	return &AIEnhancer{
		userSessions: make(map[int]*UserSession),
		pathStats:    make(map[string]*PathAIStats),
		predictions:  make(map[string]*Prediction),
		smartCache:   make(map[string]*CachedData),
	}
}

// RecordAction 记录用户操作（供AI分析）
func (ai *AIEnhancer) RecordAction(userID int, actionType, path string, duration int64, success bool) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	// 获取或创建用户会话
	session, exists := ai.userSessions[userID]
	if !exists {
		session = &UserSession{
			UserID:    userID,
			SessionID: fmt.Sprintf("session_%d_%d", userID, time.Now().Unix()),
			StartTime: time.Now(),
			Actions:   []UserAction{},
		}
		ai.userSessions[userID] = session
	}

	// 记录操作
	action := UserAction{
		Timestamp:  time.Now(),
		ActionType: actionType,
		Path:       path,
		Duration:   duration,
		Success:    success,
	}
	session.Actions = append(session.Actions, action)
	session.LastActive = time.Now()
	session.CurrentPath = path

	// 更新路径统计
	ai.updatePathStats(path, duration)
}

// updatePathStats 更新路径统计
func (ai *AIEnhancer) updatePathStats(path string, duration int64) {
	stats, exists := ai.pathStats[path]
	if !exists {
		stats = &PathAIStats{
			Path:             path,
			UserDistribution: make(map[int]int),
		}
		ai.pathStats[path] = stats
	}

	stats.TotalAccess++
	stats.AvgDuration = (stats.AvgDuration*int64(stats.TotalAccess-1) + duration) / int64(stats.TotalAccess)
	stats.LastAccessTime = time.Now()

	// 计算访问频率（简化版）
	if stats.TotalAccess > 1 {
		timeSinceFirst := time.Since(stats.LastAccessTime).Minutes()
		if timeSinceFirst > 0 {
			stats.AccessFrequency = float64(stats.TotalAccess) / timeSinceFirst
		}
	}
}

// PredictNextAction AI预测下一步操作（简化版 - 基于规则）
func (ai *AIEnhancer) PredictNextAction(userID int, currentPath string) *Prediction {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	// 简单规则预测（后续可替换为AI模型）
	prediction := &Prediction{
		CurrentPath: currentPath,
		GeneratedAt: time.Now(),
	}

	// 规则1: 查看列表 → 很可能编辑
	if containsKeyword(currentPath, []string{"index", "list"}) {
		prediction.NextPath = replaceLast(currentPath, "index", "save")
		prediction.Probability = 0.7
		return prediction
	}

	// 规则2: 用户管理 → 可能查看角色
	if contains(currentPath, "/admin") {
		prediction.NextPath = "/roles"
		prediction.Probability = 0.6
		return prediction
	}

	// 规则3: 基于历史行为
	if session, exists := ai.userSessions[userID]; exists && len(session.Actions) > 1 {
		// 找到最常见的跳转路径
		nextPath := ai.findMostCommonNextPath(session.Actions, currentPath)
		if nextPath != "" {
			prediction.NextPath = nextPath
			prediction.Probability = 0.8
			return prediction
		}
	}

	return nil
}

// ShouldCache AI决定是否应该缓存（简化版）
func (ai *AIEnhancer) ShouldCache(path string) bool {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	stats, exists := ai.pathStats[path]
	if !exists {
		return false
	}

	// 缓存策略（基于规则，后续可用AI模型）
	return stats.AccessFrequency > 5 &&    // 访问频率 > 5次/分钟
		stats.AvgDuration > 10              // 平均耗时 > 10ms
}

// GetCacheDuration AI决定缓存时长
func (ai *AIEnhancer) GetCacheDuration(path string) time.Duration {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	stats, exists := ai.pathStats[path]
	if !exists {
		return 0
	}

	// 动态决定缓存时长
	if stats.AccessFrequency > 20 {
		return 10 * time.Minute // 高频访问
	} else if stats.AccessFrequency > 5 {
		return 5 * time.Minute // 中频访问
	}
	return 1 * time.Minute // 低频访问
}

// GetFromSmartCache 从智能缓存获取
func (ai *AIEnhancer) GetFromSmartCache(path string) interface{} {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	cached, exists := ai.smartCache[path]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Now().After(cached.ExpiresAt) {
		return nil
	}

	// 更新命中次数
	cached.HitCount++
	return cached.Data
}

// PutToSmartCache 放入智能缓存
func (ai *AIEnhancer) PutToSmartCache(path string, data interface{}) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	duration := ai.GetCacheDuration(path)
	if duration == 0 {
		return
	}

	ai.smartCache[path] = &CachedData{
		Data:      data,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
		HitCount:  0,
	}
}

// AnalyzeMatchingEfficiency AI分析前后端匹配效率
func (ai *AIEnhancer) AnalyzeMatchingEfficiency() map[string]interface{} {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	bottlenecks := []map[string]interface{}{}
	recommendations := []map[string]interface{}{}

	// 分析慢请求
	for path, stats := range ai.pathStats {
		if stats.AvgDuration > 100 {
			bottlenecks = append(bottlenecks, map[string]interface{}{
				"type":        "slow_response",
				"path":        path,
				"avg_time_ms": stats.AvgDuration,
				"impact":      "影响用户体验",
			})

			recommendations = append(recommendations, map[string]interface{}{
				"priority":    "high",
				"title":       "优化慢查询",
				"path":        path,
				"description": fmt.Sprintf("该接口平均响应时间 %dms，建议优化", stats.AvgDuration),
			})
		}

		// 分析缓存机会
		if stats.AccessFrequency > 10 && !ai.isCached(path) {
			recommendations = append(recommendations, map[string]interface{}{
				"priority":    "medium",
				"title":       "启用缓存",
				"path":        path,
				"description": fmt.Sprintf("该接口访问频率 %.1f次/分钟，建议缓存", stats.AccessFrequency),
			})
		}
	}

	// 计算总体评分
	overallScore := ai.calculateOverallScore()

	return map[string]interface{}{
		"overall_score":   overallScore,
		"bottlenecks":     bottlenecks,
		"recommendations": recommendations,
		"total_sessions":  len(ai.userSessions),
		"total_paths":     len(ai.pathStats),
	}
}

// calculateOverallScore 计算总体效率评分
func (ai *AIEnhancer) calculateOverallScore() float64 {
	if len(ai.pathStats) == 0 {
		return 0
	}

	score := 100.0

	// 扣分项1: 慢请求
	slowPaths := 0
	for _, stats := range ai.pathStats {
		if stats.AvgDuration > 100 {
			slowPaths++
		}
	}
	score -= float64(slowPaths) * 5

	// 扣分项2: 低缓存命中率
	cacheableButNotCached := 0
	for _, stats := range ai.pathStats {
		if stats.AccessFrequency > 10 && !ai.isCached(stats.Path) {
			cacheableButNotCached++
		}
	}
	score -= float64(cacheableButNotCached) * 3

	if score < 0 {
		score = 0
	}
	return score
}

// 辅助函数
func (ai *AIEnhancer) isCached(path string) bool {
	_, exists := ai.smartCache[path]
	return exists
}

func (ai *AIEnhancer) findMostCommonNextPath(actions []UserAction, currentPath string) string {
	nextPaths := make(map[string]int)

	for i := 0; i < len(actions)-1; i++ {
		if actions[i].Path == currentPath {
			nextPaths[actions[i+1].Path]++
		}
	}

	// 找到最常见的下一步
	maxCount := 0
	mostCommon := ""
	for path, count := range nextPaths {
		if count > maxCount {
			maxCount = count
			mostCommon = path
		}
	}

	return mostCommon
}

func containsKeyword(path string, keywords []string) bool {
	for _, keyword := range keywords {
		if contains(path, keyword) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func replaceLast(s, old, new string) string {
	// 简化实现：直接替换
	result := s
	for i := len(s) - len(old); i >= 0; i-- {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result = s[:i] + new + s[i+len(old):]
			break
		}
	}
	return result
}

// Middleware AI增强中间件
func (ai *AIEnhancer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		userID := c.GetInt("user_id")

		// 1. 检查智能缓存
		if cached := ai.GetFromSmartCache(path); cached != nil {
			c.JSON(200, cached)
			fmt.Printf("🎯 AI缓存命中: %s (< 1ms)\n", path)
			return
		}

		// 2. AI预测和预加载
		if userID > 0 {
			go ai.predictAndPreload(userID, path)
		}

		// 3. 正常处理请求
		c.Next()

		// 4. 记录操作（供AI分析）
		duration := time.Since(startTime).Milliseconds()
		success := c.Writer.Status() < 400
		ai.RecordAction(userID, "api_call", path, duration, success)

		// 5. 智能缓存决策
		if ai.ShouldCache(path) && success {
			// 从响应中提取数据并缓存
			// 注意：这是简化版，实际需要更复杂的实现
			fmt.Printf("🤖 AI建议缓存: %s\n", path)
		}
	}
}

// predictAndPreload AI预测并预加载（异步）
func (ai *AIEnhancer) predictAndPreload(userID int, currentPath string) {
	prediction := ai.PredictNextAction(userID, currentPath)
	if prediction != nil && prediction.Probability > 0.7 {
		fmt.Printf("🔮 AI预测: 用户 %d 可能访问 %s (概率%.0f%%)\n",
			userID, prediction.NextPath, prediction.Probability*100)

		// TODO: 实际的预加载逻辑
		// ai.preloadData(prediction.NextPath)
	}
}

// GetAnalysis 获取AI分析结果（供API调用）
func (ai *AIEnhancer) GetAnalysis() map[string]interface{} {
	return ai.AnalyzeMatchingEfficiency()
}

