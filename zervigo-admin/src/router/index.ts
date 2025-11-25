import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/',
      name: 'Layout',  // 重要：设置 name 为 'Layout'，供动态路由使用
      component: () => import('../views/Layout.vue'),
      redirect: '/home',
      children: [
        {
          path: 'home',
          name: 'Home',
          component: () => import('../views/Home.vue'),
          meta: {
            title: '首页',
            requiresAuth: true
          }
        },
        {
          path: 'service-status',
          name: 'ServiceStatus',
          component: () => import('../views/ServiceCombinationStatus.vue'),
          meta: {
            title: '服务状态',
            requiresAuth: true
          }
        },
        // 保留静态路由作为备份
        {
          path: 'system/users',
          name: 'Users',
          component: () => import('../views/system/Users.vue'),
          meta: {
            title: '用户管理',
            requiresAuth: true
          }
        },
        {
          path: 'system/roles',
          name: 'Roles',
          component: () => import('../views/system/Roles.vue'),
          meta: {
            title: '角色管理',
            requiresAuth: true
          }
        },
        {
          path: 'system/permissions',
          name: 'Permissions',
          component: () => import('../views/system/Permissions.vue'),
          meta: {
            title: '权限管理',
            requiresAuth: true
          }
        }
        // 动态路由将在运行时通过 MenuService 添加到这里
      ]
    }
  ]
})

// 路由守卫 - 简单直接
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  if (to.path === '/login') {
    next()
  } else {
    if (!token) {
      ElMessage.warning('请先登录')
      next('/login')
    } else {
      next()
    }
  }
})

export default router



