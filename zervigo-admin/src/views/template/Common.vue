<!-- 
  通用模板组件
  借鉴 VueCMF 的设计理念
  当菜单指定的组件模板不存在时，使用此组件作为降级方案
-->
<template>
  <div class="common-template">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Document /></el-icon>
          <span>{{ title }}</span>
        </div>
      </template>
      
      <div class="template-info">
        <el-alert
          type="info"
          :closable="false"
          show-icon
        >
          <template #title>
            页面模板开发中
          </template>
          <p>此页面的专用模板正在开发中，当前显示通用模板。</p>
        </el-alert>

        <el-descriptions 
          title="页面信息" 
          :column="2" 
          border 
          style="margin-top: 20px"
        >
          <el-descriptions-item label="页面标题">
            {{ title }}
          </el-descriptions-item>
          <el-descriptions-item label="路由路径">
            {{ routePath }}
          </el-descriptions-item>
          <el-descriptions-item label="菜单ID">
            {{ menuId || '未配置' }}
          </el-descriptions-item>
          <el-descriptions-item label="图标">
            <el-icon v-if="icon">
              <component :is="icon" />
            </el-icon>
            <span v-else>未配置</span>
          </el-descriptions-item>
        </el-descriptions>

        <div class="development-tips" style="margin-top: 20px">
          <h3>🔧 开发提示</h3>
          <el-alert type="warning" :closable="false">
            <p><strong>如需为此页面创建专用模板，请按以下步骤操作：</strong></p>
            <ol>
              <li>在后端菜单配置中设置 <code>component_path</code> 字段</li>
              <li>在 <code>src/views/template/</code> 目录下创建对应的 Vue 组件</li>
              <li>组件文件路径应与 <code>component_path</code> 值匹配</li>
              <li>刷新页面后即可看到新的模板</li>
            </ol>
            <p style="margin-top: 10px;">
              <strong>推荐的组件路径格式：</strong><br>
              例如：<code>system/Users</code> → <code>src/views/template/system/Users.vue</code>
            </p>
          </el-alert>
        </div>

        <div class="quick-actions" style="margin-top: 20px">
          <h3>⚡ 快捷操作</h3>
          <el-space wrap>
            <el-button type="primary" @click="handleRefresh">
              <el-icon><Refresh /></el-icon>
              刷新页面
            </el-button>
            <el-button @click="handleBack">
              <el-icon><Back /></el-icon>
              返回上一页
            </el-button>
            <el-button @click="handleGoHome">
              <el-icon><HomeFilled /></el-icon>
              返回首页
            </el-button>
          </el-space>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Document, Refresh, Back, HomeFilled } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

// 从路由元数据中获取信息
const title = computed(() => route.meta.title as string || '未命名页面')
const routePath = computed(() => route.path)
const menuId = computed(() => route.meta.menuId)
const icon = computed(() => route.meta.icon)

// 刷新页面
const handleRefresh = () => {
  router.go(0)
}

// 返回上一页
const handleBack = () => {
  router.back()
}

// 返回首页
const handleGoHome = () => {
  router.push('/home')
}
</script>

<style scoped>
.common-template {
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
}

.template-info {
  padding: 10px;
}

.development-tips h3,
.quick-actions h3 {
  margin-bottom: 10px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.development-tips ol {
  margin: 10px 0 0 20px;
  line-height: 1.8;
}

.development-tips code {
  padding: 2px 6px;
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}
</style>


