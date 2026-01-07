<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <h2>🧠 Zervigo 管理平台</h2>
          <p>自主可控 · 简洁高效</p>
        </div>
      </template>

      <el-form :model="loginForm" :rules="rules" ref="formRef">
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            style="width: 100%"
            :loading="loading"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="footer">
        <p>© 2025 Go-Zervi · Powered by Zervigo Team</p>
        <p class="tip">完全自主可控的管理系统</p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const loginForm = reactive({
  username: 'admin',
  password: 'admin123'
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  try {
    await formRef.value.validate()
    loading.value = true
    
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 500))
    
    // 本地演示版本：直接登录成功
    if (loginForm.username === 'admin' && loginForm.password === 'admin123') {
      // 生成模拟token
      const mockToken = 'demo_token_' + Date.now()
      const mockUserInfo = {
        username: 'admin',
        role: '超级管理员',
        email: 'admin@zervigo.com',
        last_login_time: new Date().toLocaleString()
      }
      
      localStorage.setItem('token', mockToken)
      localStorage.setItem('userInfo', JSON.stringify(mockUserInfo))
      
      console.log('🔐 本地演示登录成功')
      console.log('📱 用户信息:', mockUserInfo)
      
      ElMessage.success('登录成功！欢迎使用 Zervigo 管理平台')
      router.push('/home')
    } else {
      ElMessage.error('用户名或密码错误，请使用 admin / admin123')
    }
  } catch (error: any) {
    console.error('登录失败:', error)
    ElMessage.error(error.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 420px;
  border-radius: 12px;
}

.card-header {
  text-align: center;
  padding: 20px 0;
}

.card-header h2 {
  margin: 0 0 10px 0;
  color: #303133;
  font-size: 28px;
}

.card-header p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.footer {
  text-align: center;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.footer p {
  margin: 5px 0;
  color: #909399;
  font-size: 12px;
}

.tip {
  color: #409eff !important;
  font-weight: 500;
}
</style>


