import Taro from '@tarojs/taro'

// API基础配置
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:9000'

// 请求拦截器
const request = (options: any) => {
  return new Promise((resolve, reject) => {
    // 添加认证token
    const token = Taro.getStorageSync('token')
    if (token) {
      options.header = {
        ...options.header,
        'Authorization': `Bearer ${token}`
      }
    }

    Taro.request({
      ...options,
      url: `${API_BASE_URL}${options.url}`,
      success: (res) => {
        if (res.statusCode === 200) {
          resolve(res.data)
        } else {
          reject(res)
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}

// API服务类
export class ApiService {
  // 认证相关
  static async login(username: string, password: string) {
    return request({
      url: '/api/v1/auth/login',
      method: 'POST',
      data: { username, password }
    })
  }

  static async register(userData: any) {
    return request({
      url: '/api/v1/auth/register',
      method: 'POST',
      data: userData
    })
  }

  static async logout() {
    return request({
      url: '/api/v1/auth/logout',
      method: 'POST'
    })
  }

  // 用户相关
  static async getUserInfo() {
    return request({
      url: '/api/v1/user/info',
      method: 'GET'
    })
  }

  static async updateUserInfo(userData: any) {
    return request({
      url: '/api/v1/user/info',
      method: 'PUT',
      data: userData
    })
  }

  // 职位相关
  static async getJobList(params: any) {
    return request({
      url: '/api/v1/job/list',
      method: 'GET',
      data: params
    })
  }

  static async getJobDetail(jobId: string) {
    return request({
      url: `/api/v1/job/${jobId}`,
      method: 'GET'
    })
  }

  static async searchJobs(params: any) {
    return request({
      url: '/api/v1/job/search',
      method: 'POST',
      data: params
    })
  }

  // 简历相关
  static async getResumeList() {
    return request({
      url: '/api/v1/resume/user/me',
      method: 'GET'
    })
  }

  static async createResume(resumeData: any) {
    return request({
      url: '/api/v1/resume',
      method: 'POST',
      data: resumeData
    })
  }

  static async updateResume(resumeId: string, resumeData: any) {
    return request({
      url: `/api/v1/resume/${resumeId}`,
      method: 'PUT',
      data: resumeData
    })
  }

  // AI相关
  static async aiMatch(params: any) {
    return request({
      url: '/api/v1/ai/match',
      method: 'POST',
      data: params
    })
  }

  static async aiChat(message: string, sessionId?: string) {
    return request({
      url: '/api/v1/ai/chat',
      method: 'POST',
      data: { message, sessionId }
    })
  }

  static async analyzeResume(resumeId: string, jobId?: string) {
    return request({
      url: '/api/v1/ai/resume/analyze',
      method: 'POST',
      data: { resumeId, jobId }
    })
  }

  // 企业相关
  static async getCompanyList(params: any) {
    return request({
      url: '/api/v1/company/list',
      method: 'GET',
      data: params
    })
  }

  static async getCompanyDetail(companyId: string) {
    return request({
      url: `/api/v1/company/${companyId}`,
      method: 'GET'
    })
  }

  // 区块链相关
  static async getBlockchainStats() {
    return request({
      url: '/api/v1/blockchain/stats',
      method: 'GET'
    })
  }

  static async validateDataConsistency(params: any) {
    return request({
      url: '/api/v1/blockchain/validate',
      method: 'POST',
      data: params
    })
  }
}

export default ApiService
