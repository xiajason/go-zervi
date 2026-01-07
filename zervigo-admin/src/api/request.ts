// 简洁的API请求封装 - 完全自主可控
import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: 'http://localhost:9000',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 兼容多种后端响应格式
request.interceptors.response.use(
  (response) => {
    const res = response.data
    
    // 兼容多种响应格式：
    // 1. {code: 0, data: {}, message: ""}  - VueCMF 格式
    // 2. {status: "success", data: {}, message: ""} - Zervigo 格式
    // 3. {code: 200, data: {}, msg: ""} - 其他格式
    
    // 检查是否成功
    const isSuccess = 
      res.code === 0 || 
      res.code === 200 || 
      res.status === 'success' || 
      res.success === true
    
    if (isSuccess) {
      return res.data || res  // 返回 data 字段，如果没有则返回整个响应
    } else {
      const errorMsg = res.message || res.msg || res.error || '请求失败'
      ElMessage.error(errorMsg)
      return Promise.reject(new Error(errorMsg))
    }
  },
  (error) => {
    const errorMsg = error.response?.data?.message || error.message || '网络错误'
    ElMessage.error(errorMsg)
    return Promise.reject(error)
  }
)

export default request




