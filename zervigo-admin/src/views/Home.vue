<template>
  <div class="home">
    <el-card>
      <h1>🎉 欢迎使用 Zervigo 管理平台</h1>
      <el-divider />
      
      <el-descriptions title="系统信息" :column="2" border>
        <el-descriptions-item label="登录账号">{{ userInfo?.username }}</el-descriptions-item>
        <el-descriptions-item label="角色">{{ userInfo?.role }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ userInfo?.email }}</el-descriptions-item>
        <el-descriptions-item label="最后登录时间">{{ userInfo?.last_login_time }}</el-descriptions-item>
      </el-descriptions>

      <el-divider />

      <el-alert
        title="✨ 完全自主可控"
        type="success"
        :closable="false"
      >
        <p>这是我们自己开发的管理后台，不依赖任何第三方低代码框架</p>
        <ul>
          <li>✅ 代码简洁清晰，易于维护</li>
          <li>✅ 完全掌控数据格式和业务逻辑</li>
          <li>✅ 基于成熟技术栈：Vue 3 + Element Plus</li>
          <li>✅ 想加什么功能就加什么功能</li>
        </ul>
      </el-alert>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || '{}'))
</script>

<style scoped>
.home {
  max-width: 1200px;
}

.home h1 {
  color: #303133;
  margin-bottom: 20px;
}

.home ul {
  margin: 10px 0;
  padding-left: 20px;
}

.home li {
  margin: 8px 0;
  color: #606266;
}
</style>





