// 系统管理API - 我们自己的简洁格式
import request from './request'

export interface PageParams {
  page: number
  pageSize: number
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// 用户管理
export function getUserList(params: PageParams) {
  return request.post<PageResult<any>>('/api/v1/admin/index', {
    data: {
      table_name: 'admin',
      page: params.page,
      page_size: params.pageSize
    }
  })
}

// 角色管理
export function getRoleList(params: PageParams) {
  return request.post<PageResult<any>>('/api/v1/roles/index', {
    data: {
      table_name: 'roles',
      page: params.page,
      page_size: params.pageSize
    }
  })
}

// 权限管理
export function getPermissionList(params: PageParams) {
  return request.post<PageResult<any>>('/api/v1/permissions/index', {
    data: {
      table_name: 'permissions',
      page: params.page,
      page_size: params.pageSize
    }
  })
}





