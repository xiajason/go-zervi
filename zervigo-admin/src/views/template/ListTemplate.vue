<!--
  通用列表模板组件
  可用于用户、角色、权限等各种列表页面
  借鉴 VueCMF 的设计模式
-->
<template>
  <div class="list-template">
    <el-card>
      <template #header>
        <div class="header-container">
          <h3>{{ title }}</h3>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增
          </el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="关键词">
          <el-input 
            v-model="searchForm.keyword" 
            placeholder="请输入关键词"
            clearable
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格区域 -->
      <el-table 
        :data="tableData" 
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column 
          v-for="column in columns"
          :key="column.prop"
          :prop="column.prop"
          :label="column.label"
          :width="column.width"
          :formatter="column.formatter"
        />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              size="small" 
              @click="handleEdit(row)"
              link
            >
              编辑
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="handleDelete(row)"
              link
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        style="margin-top: 20px; justify-content: center"
      />
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="dialogTitle"
      width="600px"
    >
      <el-form :model="formData" label-width="100px">
        <el-form-item 
          v-for="field in formFields"
          :key="field.prop"
          :label="field.label"
          :required="field.required"
        >
          <el-input 
            v-if="field.type === 'input' || !field.type"
            v-model="formData[field.prop]"
            :placeholder="`请输入${field.label}`"
          />
          <el-select 
            v-else-if="field.type === 'select'"
            v-model="formData[field.prop]"
            :placeholder="`请选择${field.label}`"
          >
            <el-option 
              v-for="option in field.options"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'

const route = useRoute()

// Props（可以通过路由 meta 传递配置）
const title = computed(() => route.meta.title as string || '数据列表')

// 状态
const loading = ref(false)
const tableData = ref<any[]>([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增')
const formData = reactive<any>({})

// 搜索表单
const searchForm = reactive({
  keyword: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 表格列配置（可以从 meta 或 API 获取）
const columns = ref([
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'name', label: '名称' },
  { prop: 'created_at', label: '创建时间' },
  { prop: 'status', label: '状态', formatter: (row: any) => row.status === 10 ? '启用' : '禁用' }
])

// 表单字段配置
const formFields = ref([
  { prop: 'name', label: '名称', type: 'input', required: true },
  { 
    prop: 'status', 
    label: '状态', 
    type: 'select', 
    required: true,
    options: [
      { label: '启用', value: 10 },
      { label: '禁用', value: 20 }
    ]
  }
])

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    // TODO: 调用实际的 API
    // const response = await getSomeList({
    //   page: pagination.page,
    //   pageSize: pagination.pageSize,
    //   keyword: searchForm.keyword
    // })
    
    // 模拟数据
    tableData.value = [
      { id: 1, name: '示例数据1', created_at: '2024-01-01', status: 10 },
      { id: 2, name: '示例数据2', created_at: '2024-01-02', status: 20 }
    ]
    pagination.total = 2
    
    ElMessage.success('数据加载成功')
  } catch (error: any) {
    ElMessage.error(error.message || '数据加载失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadData()
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  pagination.page = 1
  loadData()
}

// 新增
const handleAdd = () => {
  dialogTitle.value = '新增'
  Object.keys(formData).forEach(key => {
    formData[key] = ''
  })
  dialogVisible.value = true
}

// 编辑
const handleEdit = (row: any) => {
  dialogTitle.value = '编辑'
  Object.assign(formData, row)
  dialogVisible.value = true
}

// 删除
const handleDelete = (row: any) => {
  ElMessageBox.confirm('确定要删除此项吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      // TODO: 调用删除 API
      ElMessage.success('删除成功')
      loadData()
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  })
}

// 提交
const handleSubmit = async () => {
  try {
    // TODO: 调用保存 API
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  }
}

// 分页变化
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  loadData()
}

const handlePageChange = (val: number) => {
  pagination.page = val
  loadData()
}

// 初始化
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.list-template {
  padding: 20px;
}

.header-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-container h3 {
  margin: 0;
}

.search-form {
  margin-bottom: 20px;
}
</style>


