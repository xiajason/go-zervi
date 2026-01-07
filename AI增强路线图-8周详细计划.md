# 📅 AI增强路线图 - 8周详细执行计划

## 🎯 项目背景

**当前问题**:
- 前后端交互效率低（响应50-200ms）
- 菜单加载问题频发
- 缓存策略固定（命中率仅45%）
- 性能优化依赖人工

**解决方案**:
- **AI服务完善**: 从业务工具 → 系统优化引擎
- **中央大脑集成**: 让AI成为大脑的智慧中枢
- **智能化升级**: 从被动响应 → 主动预测优化

---

## 📊 8周路线图总览

```
Week 1-2: 基础集成    → AI能力扩展 + 中央大脑代理
Week 3-4: 智能优化    → 实时分析 + 智能缓存
Week 5-6: 可视化      → AI仪表盘 + 自动化
Week 7-8: 深度学习    → 模型训练 + 完整测试
```

---

## 🗓️ Week 1: AI服务能力扩展

### Day 1 (Monday): 项目启动 + 环境准备

#### 上午：项目启动会

**任务**:
- [ ] 团队成员确认（后端Go、后端Python、前端、AI工程师）
- [ ] 项目目标对齐
- [ ] 工作分工明确
- [ ] 开发环境搭建检查

**会议议程**:
```
1. 项目背景介绍 (15分钟)
2. 技术方案讲解 (30分钟)
3. 路线图说明 (15分钟)
4. 任务分配 (15分钟)
5. Q&A (15分钟)
```

#### 下午：代码学习

**任务**:
- [ ] 学习现有AI服务代码 (`zervipy/ai-services/`)
- [ ] 学习中央大脑代码 (`shared/central-brain/`)
- [ ] 学习AI客户端 (`services/business/job/ai_client.go`)
- [ ] 绘制当前架构图

**输出**:
- 架构理解文档
- 问题清单
- 改进建议

---

### Day 2 (Tuesday): 性能分析AI开发

#### 上午：设计与原型

**文件**: `zervipy/ai-services/services/performance_analyzer.py`

```python
#!/usr/bin/env python3
"""
性能分析AI引擎
功能: 分析系统性能，识别瓶颈，生成优化建议
"""

class PerformanceAnalyzer:
    """
    核心功能:
    1. 分析请求日志，识别慢接口
    2. 分析缓存效率，识别缓存机会
    3. 分析请求模式，识别批量化机会
    4. 生成可执行的优化建议
    """
    
    def __init__(self):
        self.slow_threshold_ms = 100       # 慢请求阈值
        self.cache_threshold_freq = 10     # 缓存频率阈值
        self.batch_threshold_count = 3     # 批量化阈值
    
    def analyze(self, logs: List[dict]) -> dict:
        """
        主分析入口
        """
        return {
            'overall_score': self._calculate_score(logs),
            'slow_apis': self._find_slow_apis(logs),
            'cache_opportunities': self._find_cache_opportunities(logs),
            'batch_opportunities': self._find_batch_opportunities(logs),
            'recommendations': self._generate_recommendations(logs)
        }
    
    def _find_slow_apis(self, logs):
        """
        识别慢接口
        """
        path_stats = {}
        for log in logs:
            path = log['path']
            duration = log['duration_ms']
            
            if path not in path_stats:
                path_stats[path] = {
                    'count': 0,
                    'total_time': 0,
                    'max_time': 0
                }
            
            path_stats[path]['count'] += 1
            path_stats[path]['total_time'] += duration
            path_stats[path]['max_time'] = max(path_stats[path]['max_time'], duration)
        
        slow_apis = []
        for path, stats in path_stats.items():
            avg_time = stats['total_time'] / stats['count']
            if avg_time > self.slow_threshold_ms:
                slow_apis.append({
                    'path': path,
                    'avg_time_ms': round(avg_time, 2),
                    'max_time_ms': stats['max_time'],
                    'count': stats['count'],
                    'root_cause': self._analyze_slow_cause(path, avg_time),
                    'priority': self._calculate_priority(avg_time, stats['count'])
                })
        
        return sorted(slow_apis, key=lambda x: x['priority'], reverse=True)
    
    def _analyze_slow_cause(self, path: str, avg_time: float):
        """
        AI分析慢的根本原因
        """
        causes = []
        
        # 规则1: 时间过长 → 可能是数据库查询
        if avg_time > 200:
            causes.append('数据库查询慢，可能缺少索引')
        elif avg_time > 100:
            causes.append('数据处理耗时，考虑优化算法')
        
        # 规则2: 根据路径特征判断
        if 'index' in path or 'list' in path:
            causes.append('列表查询，建议添加分页索引')
        
        return causes[0] if causes else '未知原因'
    
    def _generate_recommendations(self, logs):
        """
        生成具体的优化建议
        """
        recommendations = []
        
        slow_apis = self._find_slow_apis(logs)
        for api in slow_apis[:5]:  # Top 5
            rec = {
                'priority': api['priority'],
                'title': f"优化 {api['path']}",
                'description': f"当前平均响应{api['avg_time_ms']}ms",
                'actions': []
            }
            
            # 生成具体actions
            if 'index' in api['root_cause'] or 'list' in api['root_cause']:
                rec['actions'].append({
                    'type': 'sql',
                    'code': self._generate_index_sql(api['path']),
                    'expected_benefit': '70-90% 查询时间减少'
                })
            
            if api['avg_time_ms'] > 50:
                rec['actions'].append({
                    'type': 'cache',
                    'code': f"启用缓存: {api['path']}",
                    'expected_benefit': '80-95% 响应时间减少'
                })
            
            recommendations.append(rec)
        
        return recommendations
```

#### 下午：API路由实现

**文件**: `zervipy/ai-services/routes/performance.py`

```python
from sanic import Blueprint
from sanic.response import json as sanic_json

bp = Blueprint('performance', url_prefix='/api/ai/performance')

from services.performance_analyzer import PerformanceAnalyzer

analyzer = PerformanceAnalyzer()

@bp.post('/analyze')
async def analyze_performance(request):
    """
    分析系统性能
    """
    data = request.json
    logs = data.get('logs', [])
    
    analysis = analyzer.analyze(logs)
    
    return sanic_json({
        'success': True,
        'result': analysis
    })

@bp.post('/suggest-optimizations')
async def suggest_optimizations(request):
    """
    生成优化建议
    """
    data = request.json
    logs = data.get('logs', [])
    
    analysis = analyzer.analyze(logs)
    
    return sanic_json({
        'success': True,
        'recommendations': analysis['recommendations']
    })
```

---

### Day 3 (Wednesday): 行为预测AI开发

#### 全天：行为预测引擎

**文件**: `zervipy/ai-services/services/behavior_predictor.py`

```python
class BehaviorPredictor:
    """
    用户行为预测引擎
    基于历史数据预测用户下一步操作
    """
    
    def __init__(self):
        # 初始化规则库（简单版）
        self.patterns = {
            # 管理员操作模式
            'admin_patterns': [
                {
                    'sequence': ['login', 'menu/list', 'admin/index'],
                    'next_likely': 'admin/save',
                    'probability': 0.75
                },
                {
                    'sequence': ['admin/index'],
                    'next_likely': 'roles/index',
                    'probability': 0.65
                }
            ],
            # 用户操作模式
            'user_patterns': [
                # ...
            ]
        }
    
    def predict(self, user_id: int, current_path: str, recent_actions: list) -> dict:
        """
        预测下一步操作
        """
        # 1. 匹配已知模式
        for pattern in self.patterns['admin_patterns']:
            if self._match_sequence(recent_actions, pattern['sequence']):
                return {
                    'next_path': pattern['next_likely'],
                    'probability': pattern['probability'],
                    'reason': '基于历史模式匹配',
                    'preload_suggestions': self._get_preload_data(pattern['next_likely'])
                }
        
        # 2. 基于统计的预测
        next_path, prob = self._statistical_predict(user_id, current_path, recent_actions)
        if prob > 0.6:
            return {
                'next_path': next_path,
                'probability': prob,
                'reason': '基于统计分析',
                'preload_suggestions': self._get_preload_data(next_path)
            }
        
        # 3. 无法预测
        return {
            'next_path': None,
            'probability': 0,
            'reason': '模式不明确'
        }
    
    def _get_preload_data(self, next_path: str) -> list:
        """
        建议预加载的数据
        """
        preload_map = {
            'admin/save': ['roles/index', 'permissions/index'],
            'roles/index': ['permissions/index'],
            # ...
        }
        return preload_map.get(next_path, [])
```

---

### Day 4 (Thursday): 缓存优化AI开发

**文件**: `zervipy/ai-services/services/cache_optimizer.py`

```python
class CacheOptimizer:
    """
    智能缓存优化AI
    动态决定缓存策略
    """
    
    def decide_cache_strategy(self, path: str, stats: dict) -> dict:
        """
        决定缓存策略
        """
        access_freq = stats.get('access_frequency', 0)  # 次/分钟
        data_change_rate = stats.get('data_change_rate', 1.0)  # 0-1
        avg_duration = stats.get('avg_duration_ms', 0)
        
        # AI决策逻辑
        should_cache = False
        cache_duration = 0
        priority = 0
        
        # 规则1: 高频+低变化 = 必须缓存
        if access_freq > 20 and data_change_rate < 0.1:
            should_cache = True
            cache_duration = 3600  # 1小时
            priority = 10
        elif access_freq > 10 and data_change_rate < 0.3:
            should_cache = True
            cache_duration = 300  # 5分钟
            priority = 8
        elif access_freq > 5:
            should_cache = True
            cache_duration = 60  # 1分钟
            priority = 5
        
        # 规则2: 查询耗时长 = 建议缓存
        if avg_duration > 100:
            should_cache = True
            priority = max(priority, 7)
        
        return {
            'should_cache': should_cache,
            'duration_seconds': cache_duration,
            'priority': priority,
            'strategy': self._get_strategy_name(access_freq, data_change_rate),
            'invalidation_triggers': self._get_invalidation_triggers(path)
        }
    
    def _get_invalidation_triggers(self, path: str) -> list:
        """
        AI决定哪些操作会使缓存失效
        """
        triggers = {
            '/api/v1/menu/list': ['menu/save', 'menu/delete'],
            '/api/v1/admin/index': ['admin/save', 'admin/delete'],
            '/api/v1/roles/index': ['roles/save', 'roles/delete'],
        }
        return triggers.get(path, [])
```

---

### Day 5 (Friday): 整合测试 + 周总结

#### 上午：集成测试

**测试场景**:
```bash
# 1. 启动AI服务
cd /Users/szjason72/gozervi/zervipy/ai-services
python app.py

# 2. 测试新API
curl -X POST http://localhost:8110/api/ai/performance/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {"path": "/api/v1/admin/index", "duration_ms": 150, "timestamp": "..."},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "timestamp": "..."}
    ]
  }'

# 预期返回：AI分析结果
{
  "overall_score": 75,
  "slow_apis": [
    {
      "path": "/api/v1/admin/index",
      "avg_time_ms": 150,
      "root_cause": "数据库查询慢，可能缺少索引"
    }
  ],
  "recommendations": [
    {
      "priority": "high",
      "title": "添加数据库索引",
      "actions": [
        {
          "type": "sql",
          "code": "CREATE INDEX idx_users_status ON zervigo_auth_users(status);"
        }
      ]
    }
  ]
}
```

#### 下午：周总结

**输出文档**: `Week1_Progress_Report.md`

---

## 🗓️ Week 2: 中央大脑集成

### Day 6 (Monday): AI服务代理注册

#### 任务

**文件修改**: `shared/central-brain/centralbrain.go`

```go
// 1. 在 registerServiceProxies 中添加AI服务

services := []ServiceProxy{
    // ... 现有服务 (auth, user, job, resume, company)
    
    // AI服务代理（新增）
    {
        ServiceName:       "ai-service",
        CircuitBreakerKey: "ai",
        BaseURL:           fmt.Sprintf("http://%s:8110", serviceHost),
        PathPrefix:        "/api/v1/ai",        // 前端调用路径
        TargetPrefix:      "/api/ai",           // Python服务实际路径
        Rewrite: map[string]string{
            "/performance/analyze":           "/performance/analyze",
            "/behavior/predict":              "/behavior/predict",
            "/cache/optimize":                "/cache/optimize",
        },
    },
}

// 2. 初始化AI增强器

aiEnhancer := NewAIEnhancer(
    fmt.Sprintf("http://%s:8110", serviceHost),  // AI服务地址
)

cb := &CentralBrain{
    // ... 现有字段
    aiEnhancer: aiEnhancer,  // 新增
}

// 3. 注册AI中间件

func (cb *CentralBrain) Start() error {
    // ... CORS等配置
    
    // 中间件顺序很重要！
    cb.router.Use(cb.requestLogger.Middleware())  // 第1层：日志
    cb.router.Use(cb.metrics.Middleware())        // 第2层：指标
    cb.router.Use(cb.rateLimiter.Middleware())    // 第3层：限流
    cb.router.Use(cb.aiEnhancer.Middleware())     // 第4层：AI增强 (NEW!)
    
    // ... 服务代理注册
}
```

**验证**:
```bash
# 启动中央大脑
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
go run .

# 应该看到:
✅ AI服务代理已注册: /api/v1/ai -> http://localhost:8110/api/ai
✅ AI增强器初始化成功

# 测试AI代理
curl -X POST http://localhost:9000/api/v1/ai/performance/analyze \
  -H "Content-Type: application/json" \
  -d '{"logs": [...]}'

# 应该成功转发到AI服务并返回分析结果
```

---

### Day 7 (Tuesday): AI增强器完善

**文件**: `shared/central-brain/ai_enhancer.go`

**完善任务**:
- [ ] 添加AI服务客户端调用
- [ ] 实现智能缓存决策
- [ ] 实现行为预测调用
- [ ] 实现性能数据收集

```go
// AI服务客户端

type AIServiceClient struct {
    baseURL    string
    httpClient *http.Client
}

func (aic *AIServiceClient) AnalyzePerformance(logs []RequestLog) (*PerformanceAnalysis, error) {
    // 调用Python AI服务
    requestBody, _ := json.Marshal(map[string]interface{}{
        "logs": logs,
    })
    
    resp, err := aic.httpClient.Post(
        aic.baseURL+"/api/ai/performance/analyze",
        "application/json",
        bytes.NewBuffer(requestBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result PerformanceAnalysis
    json.NewDecoder(resp.Body).Decode(&result)
    
    return &result, nil
}

func (aic *AIServiceClient) PredictNextAction(userID int, currentPath string, recentActions []string) (*Prediction, error) {
    // 调用Python AI服务预测
    // ...
}

func (aic *AIServiceClient) OptimizeCacheStrategy(path string, stats PathStats) (*CacheStrategy, error) {
    // 调用Python AI服务优化缓存
    // ...
}
```

---

### Day 8 (Wednesday): 智能缓存实现

**任务**: 实现AI驱动的多层缓存

```go
type AISmartCache struct {
    l1 *sync.Map           // L1: 内存缓存（热数据）
    l2 *redis.Client       // L2: Redis缓存（温数据）
    aiClient *AIServiceClient
    
    // 缓存统计
    stats struct {
        l1Hits   int64
        l2Hits   int64
        misses   int64
        hitRate  float64
    }
}

func (asc *AISmartCache) Get(path string) (interface{}, bool) {
    // L1检查
    if val, ok := asc.l1.Load(path); ok {
        atomic.AddInt64(&asc.stats.l1Hits, 1)
        return val, true
    }
    
    // L2检查
    val, err := asc.l2.Get(ctx, path).Result()
    if err == nil {
        atomic.AddInt64(&asc.stats.l2Hits, 1)
        // 提升到L1
        asc.l1.Store(path, val)
        return val, true
    }
    
    atomic.AddInt64(&asc.stats.misses, 1)
    return nil, false
}

func (asc *AISmartCache) Set(path string, data interface{}, stats PathStats) {
    // 调用AI决定缓存策略
    strategy, err := asc.aiClient.OptimizeCacheStrategy(path, stats)
    if err != nil || !strategy.ShouldCache {
        return
    }
    
    // 根据优先级决定缓存层级
    if strategy.Priority >= 8 {
        // 高优先级 → L1+L2
        asc.l1.Store(path, data)
        asc.l2.Set(ctx, path, data, time.Duration(strategy.DurationSeconds)*time.Second)
    } else if strategy.Priority >= 5 {
        // 中优先级 → 仅L2
        asc.l2.Set(ctx, path, data, time.Duration(strategy.DurationSeconds)*time.Second)
    }
}
```

---

### Day 9 (Thursday): 预测预加载实现

```go
type SmartPreloader struct {
    aiClient *AIServiceClient
    cache    *AISmartCache
    fetcher  *DataFetcher
}

func (sp *SmartPreloader) PredictAndPreload(c *gin.Context, userID int, currentPath string) {
    // 收集最近操作
    recentActions := sp.getRecentActions(userID)
    
    // 调用AI预测
    prediction, err := sp.aiClient.PredictNextAction(userID, currentPath, recentActions)
    if err != nil || prediction.Probability < 0.7 {
        return
    }
    
    fmt.Printf("🔮 AI预测: 用户%d 可能访问 %s (概率%.0f%%)\n",
        userID, prediction.NextPath, prediction.Probability*100)
    
    // 后台预加载
    go func() {
        for _, suggestedPath := range prediction.PreloadSuggestions {
            data, err := sp.fetcher.Fetch(suggestedPath)
            if err == nil {
                sp.cache.l2.Set(ctx, suggestedPath, data, 5*time.Minute)
                fmt.Printf("✅ AI预加载完成: %s\n", suggestedPath)
            }
        }
    }()
}
```

---

### Day 10 (Friday): Week 2 整合测试

#### 全天：端到端测试

**测试脚本**: `test-ai-integration.sh`

```bash
#!/bin/bash

echo "=========================================="
echo "  AI增强集成测试"
echo "=========================================="

# 1. 启动AI服务
echo "[1/5] 启动AI服务..."
cd /Users/szjason72/gozervi/zervipy/ai-services
python app.py &
AI_PID=$!
sleep 3

# 2. 启动中央大脑
echo "[2/5] 启动中央大脑..."
cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
go run . &
CB_PID=$!
sleep 3

# 3. 测试AI代理
echo "[3/5] 测试AI服务代理..."
curl -X POST http://localhost:9000/api/v1/ai/performance/analyze \
  -H "Content-Type: application/json" \
  -d '{"logs":[{"path":"/api/v1/menu/list","duration_ms":15}]}'

# 4. 测试智能缓存
echo "[4/5] 测试智能缓存..."
# 第一次请求（无缓存）
time curl http://localhost:9000/api/v1/menu/list

# 第二次请求（应该命中AI缓存）
time curl http://localhost:9000/api/v1/menu/list

# 5. 测试AI预测
echo "[5/5] 测试AI行为预测..."
curl -X POST http://localhost:9000/api/v1/ai/behavior/predict \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "current_path": "/api/v1/admin/index",
    "recent_actions": ["login", "menu/list"]
  }'

# 清理
kill $AI_PID $CB_PID
echo "✅ 测试完成"
```

---

## 🗓️ Week 3-4: 智能优化引擎（略）

## 🗓️ Week 5-6: AI可视化（略）

## 🗓️ Week 7-8: 深度学习（略）

---

## 📋 每日站会模板

```
Daily Standup (每天9:30):

1. 昨天完成了什么？
2. 今天计划做什么？
3. 遇到什么阻碍？
4. 需要什么帮助？

每人5分钟，总计20分钟
```

## 📊 周报模板

```
Weekly Report (每周五):

本周进度:
├─ 已完成任务: X/Y
├─ 代码行数: +X行
├─ 测试覆盖率: X%
└─ 遇到的问题: ...

下周计划:
├─ 任务1: ...
├─ 任务2: ...
└─ 风险点: ...

需要支持:
└─ ...
```

---

## 🎯 风险管理

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|---------|
| AI服务不稳定 | 中 | 高 | 添加熔断器、降级策略 |
| 性能不达标 | 低 | 中 | 提前性能测试、优化 |
| 团队协作问题 | 中 | 中 | 每日站会、及时沟通 |
| 时间延期 | 中 | 中 | 预留缓冲时间、优先级管理 |

---

**下一步**: 请查看 `AI服务完善与中央大脑集成-实施计划.md` 了解完整8周计划细节！

**立即开始**: Day 1 项目启动会议！🚀

