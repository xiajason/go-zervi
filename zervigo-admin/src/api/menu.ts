/**
 * 菜单管理 API - 集成P1 Router Service
 */
import request from './request'

export interface MenuItem {
  id: number
  pid: number
  model_id?: number
  title: string
  path: string
  icon?: string
  component_path?: string
  sort_num?: number
  status?: number
}

export interface MenuResponse {
  code: number
  msg: string
  data: MenuItem[]
}

export interface ServiceCombination {
  combination: string
  available_services: string[]
  total_services: number
}

/**
 * 获取菜单列表（兼容旧版API）
 * 注意：路径已修正为与 Central Brain 注册的路由一致
 */
export function getMenuList() {
  return request.get<MenuResponse>('/api/v1/menu/list')
}

/**
 * 获取导航菜单（树形结构）
 */
export function getMenuNav() {
  return request.get<MenuResponse>('/api/v1/menu/nav')
}

/**
 * 【P1集成】从Router Service获取路由（智能过滤）
 * 这个API会根据当前运行的P2服务自动过滤路由
 */
export function getRouterRoutes() {
  return request.get('/api/v1/router/routes')
}

/**
 * 【P1集成】从Router Service获取页面配置
 */
export function getRouterPages() {
  return request.get('/api/v1/router/pages')
}

/**
 * 【P1集成】获取当前服务组合信息
 */
export function getServiceCombination() {
  return request.get<{code: number, data: ServiceCombination}>('/api/v1/router/service-combination')
}

/**
 * 根据用户权限获取菜单
 */
export function getUserMenus() {
  return request.get<MenuResponse>('/api/v1/router/user-pages')
}


