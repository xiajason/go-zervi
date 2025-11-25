<template>
  <div class="dynamic-table">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ modelLabel }}</span>
          <el-button v-if="canCreate" type="primary" :icon="Plus" @click="handleAdd">
            新增{{ modelLabel }}
          </el-button>
        </div>
      </template>

      <!-- 动态生成的表格 - 基于字段配置 -->
      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column
          v-for="field in visibleFields"
          :key="field.field_name"
          :prop="field.field_name"
          :label="field.label"
          :width="getColumnWidth(field)"
        >
          <template #default="{ row }" v-if="field.field_type === 'switch'">
            <el-tag :type="row[field.field_name] === 10 ? 'success' : 'danger'">
              {{ row[field.field_name] === 10 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" :width="operationWidth" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canUpdate" size="small" type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button v-if="canDelete" size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadData"
        @current-change="loadData"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getFieldConfig } from '../api/config'
import request from '../api/request'

// Props - 配置驱动（VueCMF的精华）
interface Props {
  tableName: string        // 表名
  modelLabel: string       // 模型标签
  apiPath: string          // API路径
  canCreate?: boolean      // 是否可创建
  canUpdate?: boolean      // 是否可更新
  canDelete?: boolean      // 是否可删除
}

const props = withDefaults(defineProps<Props>(), {
  canCreate: true,
  canUpdate: true,
  canDelete: true
})

const loading = ref(false)
const tableData = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const fieldConfigs = ref<any[]>([])

// 动态计算可见字段（基于配置）
const visibleFields = computed(() => {
  return fieldConfigs.value.filter(f => f.is_show === 10)
})

// 动态计算操作列宽度
const operationWidth = computed(() => {
  let width = 80
  if (props.canUpdate) width += 80
  if (props.canDelete) width += 80
  return width
})

// 动态计算列宽
const getColumnWidth = (field: any) => {
  if (field.field_type === 'datetime') return 180
  if (field.field_type === 'number' || field.field_name === 'id') return 100
  if (field.field_type === 'switch') return 100
  return undefined
}

// 加载字段配置（VueCMF的精华）
const loadFieldConfig = async () => {
  try {
    const res: any = await getFieldConfig(props.tableName)
    // 兼容多种响应格式：res.data.list (标准格式) 或 res.data.data (嵌套格式) 或 res.list
    fieldConfigs.value = res.data?.list || res.data?.data || res.list || []
    
    console.log(`✅ ${props.tableName} 字段配置加载成功:`, fieldConfigs.value.length, '个字段')
  } catch (error) {
    console.warn('⚠️ 字段配置加载失败，使用默认配置')
  }
}

// 加载数据
const loadData = async () => {
  try {
    loading.value = true
    
    const res: any = await request.post(props.apiPath, {
      data: {
        table_name: props.tableName,
        page: currentPage.value,
        page_size: pageSize.value
      }
    })
    
    // 兼容多种格式（智能适配）
    tableData.value = res.data?.list || res.data?.data || res.list || []
    total.value = res.data?.total || res.total || 0
    
    console.log(`✅ ${props.modelLabel}数据加载成功:`, tableData.value.length, '/', total.value)
    
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  ElMessage.info('新增功能开发中...')
}

const handleEdit = (row: any) => {
  ElMessage.info(`编辑 ${row.id} - 功能开发中...`)
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除吗？`, '提示', {
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功（演示）')
  })
}

onMounted(async () => {
  await loadFieldConfig()  // 先加载配置
  await loadData()         // 再加载数据
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>




