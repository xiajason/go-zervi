<!--
  实时监控页面
  展示中央大脑的监控数据
-->
<template>
  <div class="monitoring-page">
    <el-row :gutter="20">
      <!-- 总体统计 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ metrics.total_requests || 0 }}</div>
          <div class="stat-label">总请求数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card success">
          <div class="stat-value">{{ metrics.success_rate || 0 }}%</div>
          <div class="stat-label">成功率</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ metrics.avg_duration_ms || 0 }}ms</div>
          <div class="stat-label">平均响应时间</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ metrics.max_duration_ms || 0 }}ms</div>
          <div class="stat-label">最大响应时间</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 路径统计 -->
    <el-card style="margin-top: 20px" title="API路径性能统计">
      <template #header>
        <div class="card-header">
          <span>API路径性能统计</span>
          <el-button size="small" @click="loadMetrics">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-table :data="pathStatsTable" stripe>
        <el-table-column prop="path" label="API路径" min-width="200" />
        <el-table-column prop="count" label="请求次数" width="100" />
        <el-table-column label="平均耗时" width="120">
          <template #default="{ row }">
            <el-tag :type="getTimeType(row.avg_time_ms)">
              {{ row.avg_time_ms }}ms
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="min_time_ms" label="最小耗时" width="100" />
        <el-table-column prop="max_time_ms" label="最大耗时" width="100" />
        <el-table-column label="成功率" width="100">
          <template #default="{ row }">
            <el-progress 
              :percentage="row.success_rate" 
              :color="row.success_rate >= 95 ? '#67c23a' : '#e6a23c'"
            />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 熔断器状态 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>服务熔断器状态</span>
          <el-button size="small" @click="loadCircuitBreakers">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-table :data="circuitBreakersTable" stripe>
        <el-table-column prop="service" label="服务名称" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStateType(row.state)">
              {{ getStateText(row.state) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="failure_count" label="失败次数" width="100" />
        <el-table-column prop="last_failure_time" label="最后失败时间" width="180" />
      </el-table>
    </el-card>

    <!-- 实时说明 -->
    <el-alert 
      type="info" 
      :closable="false"
      style="margin-top: 20px"
    >
      <template #title>
        🧠 中央大脑实时监控说明
      </template>
      <div style="line-height: 1.8">
        <p><strong>监控范围：</strong></p>
        <ul>
          <li>✅ 所有HTTP API请求（100%实时监控）</li>
          <li>✅ 每个请求的响应时间、状态码</li>
          <li>✅ 每个API路径的性能统计</li>
          <li>✅ 后端服务的健康状态</li>
          <li>⚠️ 前端路由跳转（需要埋点才能监控）</li>
        </ul>
        <p style="margin-top: 10px"><strong>数据更新：</strong></p>
        <ul>
          <li>实时性：每个请求立即记录（0延迟）</li>
          <li>统计粒度：毫秒级</li>
          <li>数据保存：内存 + 日志文件</li>
        </ul>
      </div>
    </el-alert>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'
import { ElMessage } from 'element-plus'

const metrics = reactive<any>({})
const pathStatsTable = ref<any[]>([])
const circuitBreakersTable = ref<any[]>([])

// 加载监控数据
const loadMetrics = async () => {
  try {
    const res = await request.get('/api/v1/metrics')
    Object.assign(metrics, res)
    
    // 转换路径统计为表格数据
    if (res.path_stats) {
      pathStatsTable.value = Object.entries(res.path_stats).map(([path, stats]: [string, any]) => ({
        path,
        count: stats.Count || stats.count,
        avg_time_ms: stats.TotalTime && stats.Count 
          ? Math.round(stats.TotalTime / stats.Count / 1000000)  // 纳秒转毫秒
          : 0,
        min_time_ms: Math.round((stats.MinTime || 0) / 1000000),
        max_time_ms: Math.round((stats.MaxTime || 0) / 1000000),
        success_rate: stats.SuccessRate || stats.success_rate || 100
      })).sort((a, b) => b.count - a.count)
    }
  } catch (error: any) {
    ElMessage.error('加载监控数据失败: ' + error.message)
  }
}

// 加载熔断器状态
const loadCircuitBreakers = async () => {
  try {
    const res = await request.get('/api/v1/circuit-breakers')
    circuitBreakersTable.value = Object.entries(res || {}).map(([service, state]: [string, any]) => ({
      service,
      state: state.state || state.State || 'unknown',
      failure_count: state.failure_count || state.FailureCount || 0,
      last_failure_time: state.last_failure_time || state.LastFailureTime || '-'
    }))
  } catch (error: any) {
    ElMessage.error('加载熔断器状态失败: ' + error.message)
  }
}

// 获取耗时类型
const getTimeType = (ms: number) => {
  if (ms < 50) return 'success'
  if (ms < 200) return ''
  return 'warning'
}

// 获取状态类型
const getStateType = (state: string) => {
  if (state === 'closed') return 'success'
  if (state === 'half-open') return 'warning'
  return 'danger'
}

// 获取状态文本
const getStateText = (state: string) => {
  const map: Record<string, string> = {
    'closed': '正常',
    'open': '熔断',
    'half-open': '半开'
  }
  return map[state] || state
}

// 自动刷新
let refreshTimer: any = null

onMounted(() => {
  loadMetrics()
  loadCircuitBreakers()
  
  // 每5秒自动刷新
  refreshTimer = setInterval(() => {
    loadMetrics()
    loadCircuitBreakers()
  }, 5000)
})

// 组件卸载时清理定时器
onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<script setup lang="ts">
import { onBeforeUnmount } from 'vue'
</script>

<style scoped>
.monitoring-page {
  padding: 20px;
}

.stat-card {
  text-align: center;
  padding: 20px 0;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 10px;
}

.stat-card.success .stat-value {
  color: #67c23a;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

