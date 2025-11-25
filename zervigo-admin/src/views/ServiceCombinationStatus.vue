<template>
  <div class="service-combination-status">
    <el-card class="status-card">
      <template #header>
        <div class="card-header">
          <el-icon><Monitor /></el-icon>
          <span>服务组合状态</span>
          <el-button size="small" @click="refresh" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <div class="status-content">
        <!-- 当前组合 -->
        <el-descriptions title="当前服务组合" :column="2" border>
          <el-descriptions-item label="组合类型">
            <el-tag :type="getCombinationTagType(combination)">
              {{ getCombinationName(combination) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="可用服务数">
            {{ availableServices.length }} 个
          </el-descriptions-item>
        </el-descriptions>

        <!-- 可用服务列表 -->
        <div class="services-list" v-if="availableServices.length > 0">
          <h3>可用的P2服务：</h3>
          <el-space wrap>
            <el-tag 
              v-for="service in availableServices" 
              :key="service"
              type="success"
              size="large"
            >
              <el-icon><CircleCheck /></el-icon>
              {{ getServiceDisplayName(service) }}
            </el-tag>
          </el-space>
        </div>

        <!-- P2服务状态 -->
        <div class="p2-services-status">
          <h3>P2业务服务状态：</h3>
          <el-row :gutter="20">
            <el-col :span="8">
              <el-card :class="['service-status-card', isServiceAvailable('job-service') ? 'available' : 'unavailable']">
                <div class="service-icon">
                  <el-icon :size="40"><BriefcaseFilled /></el-icon>
                </div>
                <div class="service-name">Job Service</div>
                <div class="service-status">
                  <el-tag :type="isServiceAvailable('job-service') ? 'success' : 'info'">
                    {{ isServiceAvailable('job-service') ? '运行中' : '未启动' }}
                  </el-tag>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card :class="['service-status-card', isServiceAvailable('resume-service') ? 'available' : 'unavailable']">
                <div class="service-icon">
                  <el-icon :size="40"><DocumentFilled /></el-icon>
                </div>
                <div class="service-name">Resume Service</div>
                <div class="service-status">
                  <el-tag :type="isServiceAvailable('resume-service') ? 'success' : 'info'">
                    {{ isServiceAvailable('resume-service') ? '运行中' : '未启动' }}
                  </el-tag>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card :class="['service-status-card', isServiceAvailable('company-service') ? 'available' : 'unavailable']">
                <div class="service-icon">
                  <el-icon :size="40"><OfficeBuilding /></el-icon>
                </div>
                <div class="service-name">Company Service</div>
                <div class="service-status">
                  <el-tag :type="isServiceAvailable('company-service') ? 'success' : 'info'">
                    {{ isServiceAvailable('company-service') ? '运行中' : '未启动' }}
                  </el-tag>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 组合说明 -->
        <el-alert 
          :title="getCombinationDescription(combination)"
          type="info"
          :closable="false"
          show-icon
          style="margin-top: 20px;"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getServiceCombination } from '@/api/menu'
import { 
  Monitor, Refresh, CircleCheck, 
  BriefcaseFilled, DocumentFilled, OfficeBuilding 
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const combination = ref('minimal')
const availableServices = ref<string[]>([])

const loadStatus = async () => {
  try {
    loading.value = true
    const response: any = await getServiceCombination()
    
    if (response.code === 0 && response.data) {
      combination.value = response.data.combination
      availableServices.value = response.data.available_services
      
      console.log('✅ 服务组合状态已更新:', combination.value)
    }
  } catch (error: any) {
    ElMessage.error('获取服务状态失败：' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const refresh = async () => {
  await loadStatus()
  ElMessage.success('服务状态已刷新')
}

const isServiceAvailable = (serviceName: string) => {
  return availableServices.value.includes(serviceName)
}

const getCombinationTagType = (comb: string) => {
  if (comb === 'all_services') return 'success'
  if (comb === 'minimal') return 'info'
  return 'warning'
}

const getCombinationName = (comb: string) => {
  const names: Record<string, string> = {
    'minimal': '基础模式',
    'job_only': '职位模式',
    'resume_only': '简历模式',
    'company_only': '企业模式',
    'job_resume': '职位+简历',
    'job_company': '职位+企业',
    'resume_company': '简历+企业',
    'all_services': '完整模式'
  }
  return names[comb] || '未知'
}

const getServiceDisplayName = (serviceName: string) => {
  const names: Record<string, string> = {
    'job-service': '职位服务',
    'resume-service': '简历服务',
    'company-service': '企业服务'
  }
  return names[serviceName] || serviceName
}

const getCombinationDescription = (comb: string) => {
  const descriptions: Record<string, string> = {
    'minimal': '仅启动了基础设施服务，业务功能暂不可用',
    'job_only': '已启动职位服务，可以进行职位管理',
    'resume_only': '已启动简历服务，可以进行简历管理',
    'company_only': '已启动企业服务，可以进行企业管理',
    'job_resume': '已启动职位和简历服务，可以进行职位-简历匹配',
    'job_company': '已启动职位和企业服务，可以进行企业招聘管理',
    'resume_company': '已启动简历和企业服务',
    'all_services': '所有业务服务已启动，系统功能完整'
  }
  return descriptions[comb] || '服务组合信息未知'
}

onMounted(() => {
  loadStatus()
})
</script>

<style scoped>
.service-combination-status {
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
}

.status-content {
  padding: 10px 0;
}

.services-list {
  margin-top: 20px;
}

.services-list h3,
.p2-services-status h3 {
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.p2-services-status {
  margin-top: 30px;
}

.service-status-card {
  text-align: center;
  padding: 20px;
  transition: all 0.3s;
}

.service-status-card.available {
  border-color: #67c23a;
  box-shadow: 0 2px 12px 0 rgba(103, 194, 58, 0.1);
}

.service-status-card.unavailable {
  border-color: #dcdfe6;
  opacity: 0.7;
}

.service-icon {
  margin-bottom: 15px;
  color: #409eff;
}

.service-status-card.available .service-icon {
  color: #67c23a;
}

.service-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
}

.service-status {
  margin-top: 10px;
}
</style>

