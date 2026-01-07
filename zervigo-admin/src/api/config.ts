// 配置API - 吸收VueCMF的低代码理念
import request from './request'

export interface FieldConfig {
  field_name: string
  label: string
  field_type: string
  is_show: number
  is_required: number
  sort_num: number
}

export interface ModelConfig {
  id: number
  table_name: string
  label: string
  status: number
}

// 获取模型配置（VueCMF的精华）
export function getModelConfig(tableName?: string) {
  return request.post('/api/v1/model_config/index', {
    data: {
      table_name: tableName,
      page: 1,
      page_size: 100
    }
  })
}

// 获取字段配置（VueCMF的精华）
export function getFieldConfig(tableName: string) {
  return request.post('/api/v1/model_field/index', {
    data: {
      table_name: tableName,
      page: 1,
      page_size: 100
    }
  })
}

// 获取API映射（VueCMF的精华，但我们简化了）
export function getApiMapping(tableName: string, actionType: string) {
  return request.post('/api/v1/mapping/get_api_map', {
    data: {
      table_name: tableName,
      action_type: actionType
    }
  })
}





