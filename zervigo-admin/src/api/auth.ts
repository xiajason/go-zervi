// 认证API - 简单直接
import request from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface UserInfo {
  id: number
  username: string
  email: string
  role: string
}

// 登录
export function login(data: LoginParams) {
  return request.post('/api/v1/auth/login', data)
}

// 登出
export function logout() {
  return request.post('/api/v1/auth/logout')
}

// 获取用户信息
export function getUserInfo() {
  return request.get('/api/v1/auth/user')
}


