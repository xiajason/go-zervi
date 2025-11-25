#!/bin/bash
#
# 性能分析AI测试脚本
# 用于测试性能分析AI的所有功能
#

echo "=========================================="
echo "  性能分析AI - 功能测试"
echo "=========================================="
echo

# 基础URL
BASE_URL="http://localhost:8110"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试1: 健康检查
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 1: 健康检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

curl -s -X GET "$BASE_URL/api/ai/performance/health" | jq .
echo
echo

# 测试2: 简单性能分析
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 2: 简单性能分析"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

curl -s -X POST "$BASE_URL/api/ai/performance/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {
        "path": "/api/v1/menu/list",
        "duration_ms": 15,
        "status_code": 200,
        "timestamp": "2024-11-06T10:00:00Z",
        "user_id": 1
      },
      {
        "path": "/api/v1/admin/index",
        "duration_ms": 150,
        "status_code": 200,
        "timestamp": "2024-11-06T10:00:01Z",
        "user_id": 1
      }
    ]
  }' | jq '.result | {overall_score, total_requests, slow_apis: .slow_apis | length, cache_opportunities: .cache_opportunities | length, recommendations: .recommendations | length}'

echo
echo

# 测试3: 菜单加载场景分析
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 3: 菜单加载场景分析（高频访问）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

curl -s -X POST "$BASE_URL/api/ai/performance/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:00Z", "user_id": 1},
      {"path": "/api/v1/menu/list", "duration_ms": 14, "status_code": 200, "timestamp": "2024-11-06T10:00:05Z", "user_id": 1},
      {"path": "/api/v1/menu/list", "duration_ms": 16, "status_code": 200, "timestamp": "2024-11-06T10:00:10Z", "user_id": 2},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:15Z", "user_id": 3},
      {"path": "/api/v1/menu/list", "duration_ms": 14, "status_code": 200, "timestamp": "2024-11-06T10:00:20Z", "user_id": 1},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:25Z", "user_id": 2},
      {"path": "/api/v1/menu/list", "duration_ms": 16, "status_code": 200, "timestamp": "2024-11-06T10:00:30Z", "user_id": 4},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:35Z", "user_id": 1},
      {"path": "/api/v1/menu/list", "duration_ms": 14, "status_code": 200, "timestamp": "2024-11-06T10:00:40Z", "user_id": 5},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:45Z", "user_id": 3},
      {"path": "/api/v1/menu/list", "duration_ms": 16, "status_code": 200, "timestamp": "2024-11-06T10:00:50Z", "user_id": 1},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:55Z", "user_id": 2}
    ]
  }' | jq '.result.cache_opportunities[0]'

echo
echo

# 测试4: 慢接口检测
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 4: 慢接口检测"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

curl -s -X POST "$BASE_URL/api/ai/performance/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {"path": "/api/v1/admin/index", "duration_ms": 350, "status_code": 200, "timestamp": "2024-11-06T10:00:00Z", "user_id": 1},
      {"path": "/api/v1/admin/index", "duration_ms": 320, "status_code": 200, "timestamp": "2024-11-06T10:00:05Z", "user_id": 1},
      {"path": "/api/v1/admin/index", "duration_ms": 380, "status_code": 200, "timestamp": "2024-11-06T10:00:10Z", "user_id": 2},
      {"path": "/api/v1/admin/index", "duration_ms": 340, "status_code": 200, "timestamp": "2024-11-06T10:00:15Z", "user_id": 3}
    ]
  }' | jq '.result.slow_apis[0] | {path, avg_time_ms, root_cause, priority}'

echo
echo

# 测试5: 优化建议生成
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 5: 优化建议生成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

curl -s -X POST "$BASE_URL/api/ai/performance/suggest-optimizations" \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {"path": "/api/v1/admin/index", "duration_ms": 250, "status_code": 200, "timestamp": "2024-11-06T10:00:00Z", "user_id": 1},
      {"path": "/api/v1/admin/index", "duration_ms": 240, "status_code": 200, "timestamp": "2024-11-06T10:00:05Z", "user_id": 1},
      {"path": "/api/v1/menu/list", "duration_ms": 15, "status_code": 200, "timestamp": "2024-11-06T10:00:10Z", "user_id": 1}
    ],
    "scope": "all",
    "auto_apply": false
  }' | jq '.recommendations[0] | {title, priority, expected_benefit, actions: .actions | length}'

echo
echo

# 测试6: 快速扫描
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 6: 快速扫描"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

curl -s -X POST "$BASE_URL/api/ai/performance/quick-scan" \
  -H "Content-Type: application/json" \
  -d '{
    "time_range": "last_hour",
    "focus": "response_time"
  }' | jq .

echo
echo

echo "=========================================="
echo -e "${GREEN}✅ 所有测试完成！${NC}"
echo "=========================================="
echo
echo "下一步："
echo "1. 检查AI服务日志"
echo "2. 查看分析结果是否符合预期"
echo "3. 调整AI参数（如slow_threshold_ms）"
echo

