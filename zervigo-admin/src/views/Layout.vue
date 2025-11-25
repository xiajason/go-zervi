<template>
  <el-container class="layout-container">
    <!-- 顶部导航 -->
    <el-header class="header">
      <div class="logo">
        <h1>🧠 Zervigo 管理平台</h1>
      </div>
      <div class="header-actions">
        <!-- 面包屑导航 -->
        <el-breadcrumb separator="/" class="breadcrumb">
          <el-breadcrumb-item :to="{ path: '/home' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item 
            v-for="(crumb, index) in menuService.breadcrumbs.value"
            :key="index"
          >
            {{ crumb }}
          </el-breadcrumb-item>
        </el-breadcrumb>
        
        <div class="user-info">
          <el-dropdown @command="handleCommand">
            <span class="user-name">
              <el-icon><User /></el-icon>
              {{ userInfo?.username || '管理员' }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="refresh">
                  <el-icon><Refresh /></el-icon>
                  刷新菜单
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>

    <el-container>
      <!-- 侧边栏 - 动态菜单 -->
      <el-aside width="200px" class="aside">
        <el-menu
          :default-active="activeMenu"
          class="menu"
          router
          v-loading="menuService.loading.value"
        >
          <!-- 固定首页 -->
          <el-menu-item index="/home">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </el-menu-item>

          <!-- 动态菜单 -->
          <template v-for="menu in menuService.menuTree.value" :key="menu.id">
            <!-- 有子菜单的项 -->
            <el-sub-menu v-if="menu.children && menu.children.length > 0" :index="String(menu.id)">
              <template #title>
                <el-icon v-if="menu.icon">
                  <component :is="getIconComponent(menu.icon)" />
                </el-icon>
                <span>{{ menu.title }}</span>
              </template>
              <el-menu-item 
                v-for="child in menu.children" 
                :key="child.id"
                :index="child.path"
              >
                <el-icon v-if="child.icon">
                  <component :is="getIconComponent(child.icon)" />
                </el-icon>
                <span>{{ child.title }}</span>
              </el-menu-item>
            </el-sub-menu>

            <!-- 没有子菜单的项 -->
            <el-menu-item v-else :index="menu.path">
              <el-icon v-if="menu.icon">
                <component :is="getIconComponent(menu.icon)" />
              </el-icon>
              <span>{{ menu.title }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  User, HomeFilled, Setting, UserFilled, Lock, 
  Refresh, SwitchButton, Document, List, Operation
} from '@element-plus/icons-vue'
import { menuService } from '@/services/MenuService'

const router = useRouter()
const route = useRoute()

const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || '{}'))
const activeMenu = computed(() => route.path)

// 图标映射表
const iconMap: Record<string, any> = {
  'HomeFilled': HomeFilled,
  'Setting': Setting,
  'User': User,
  'UserFilled': UserFilled,
  'Lock': Lock,
  'Document': Document,
  'List': List,
  'Operation': Operation,
  'BriefcaseFilled': Operation,
  'DocumentFilled': Document,
  'OfficeBuilding': Operation,
}

// 获取图标组件
const getIconComponent = (iconName: string) => {
  return iconMap[iconName] || Document
}

// 下拉菜单命令处理
const handleCommand = async (command: string) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      localStorage.clear()
      ElMessage.success('已退出登录')
      router.push('/login')
    })
  } else if (command === 'refresh') {
    await menuService.refresh()
    ElMessage.success('菜单已刷新')
  }
}

// 监听路由变化，更新面包屑
watch(() => route.path, (newPath) => {
  menuService.setActiveMenu(newPath)
  menuService.updateBreadcrumbs(newPath)
}, { immediate: true })

// 初始化
onMounted(async () => {
  // 加载菜单
  await menuService.loadMenu()
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
}

.logo h1 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 30px;
  flex: 1;
  justify-content: flex-end;
}

.breadcrumb {
  margin-right: auto;
  margin-left: 30px;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-name {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  color: #606266;
}

.user-name:hover {
  color: #409eff;
}

.aside {
  background: #fff;
  border-right: 1px solid #e4e7ed;
}

.menu {
  border-right: none;
}

.main {
  background: #f5f7fa;
  padding: 20px;
  overflow-y: auto;
}
</style>



