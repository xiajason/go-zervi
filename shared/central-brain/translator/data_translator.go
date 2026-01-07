package translator

import (
	"database/sql"
	"fmt"
)

// DataTranslator 中央大脑的数据翻译层
// 负责在不同数据格式之间进行智能转换和校验
type DataTranslator struct {
	db *sql.DB
}

// NewDataTranslator 创建数据翻译器
func NewDataTranslator(db *sql.DB) *DataTranslator {
	return &DataTranslator{db: db}
}

// TranslateToVueCMF 将通用数据格式转换为VueCMF格式
// 输入: 标准的列表数据 {list: [], total: int}
// 输出: VueCMF格式 {code, msg, data: {data: {data: [], field_info, ...}}}
func (t *DataTranslator) TranslateToVueCMF(
	tableName string,
	listData interface{},
	total int,
	page int,
	pageSize int,
) map[string]interface{} {
	
	// 获取字段配置（动态从数据库读取）
	fieldInfo := t.getFieldInfo(tableName)
	
	// VueCMF标准格式（三层嵌套）
	return map[string]interface{}{
		"code":    0,
		"msg":     "success",
		"status":  "success",
		"message": "获取成功",
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"data":          listData,   // 列表数据
				"field_info":    fieldInfo,  // 字段配置
				"field_option":  map[string]interface{}{},
				"form_info":     map[string]interface{}{},
				"form_rules":    []interface{}{},
				"relation_info": map[string]interface{}{},
				"total":         total,
				"page":          page,
				"limit":         pageSize,
			},
		},
	}
}

// getFieldInfo 动态获取字段配置
func (t *DataTranslator) getFieldInfo(tableName string) []map[string]interface{} {
	query := `
		SELECT mf.field_name, mf.label, mf.field_type, mf.is_show, mf.sort_num
		FROM model_field mf
		INNER JOIN model_config mc ON mf.model_id = mc.id
		WHERE mc.table_name = $1 AND mf.status = 10
		ORDER BY mf.sort_num ASC
	`
	
	rows, err := t.db.Query(query, tableName)
	if err != nil {
		fmt.Printf("⚠️  获取字段配置失败: %v\n", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()
	
	fieldInfo := []map[string]interface{}{}
	for rows.Next() {
		var fieldName, label, fieldType string
		var isShow, sortNum int
		
		if err := rows.Scan(&fieldName, &label, &fieldType, &isShow, &sortNum); err != nil {
			continue
		}
		
		field := map[string]interface{}{
			"prop":   fieldName,
			"label":  label,
			"type":   fieldType,
			"show":   isShow == 10,
			"filter": false,
			"width":  "",
		}
		fieldInfo = append(fieldInfo, field)
	}
	
	return fieldInfo
}

// ValidateRequest 校验请求数据
func (t *DataTranslator) ValidateRequest(requestData map[string]interface{}) error {
	// 提取data字段
	data, ok := requestData["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("缺少data字段")
	}
	
	// 校验必需字段
	if tableName, ok := data["table_name"].(string); !ok || tableName == "" {
		return fmt.Errorf("缺少table_name字段")
	}
	
	return nil
}

// TranslateFromVueCMF 从VueCMF格式转换为标准格式
// 输入: VueCMF请求 {data: {table_name, page, ...}}
// 输出: 标准请求 {table_name, page, page_size, filters, ...}
func (t *DataTranslator) TranslateFromVueCMF(vuecmfRequest map[string]interface{}) map[string]interface{} {
	data, ok := vuecmfRequest["data"].(map[string]interface{})
	if !ok {
		data = vuecmfRequest
	}
	
	// 提取标准参数
	standardRequest := map[string]interface{}{
		"table_name": data["table_name"],
		"page":       getIntOrDefault(data, "page", 1),
		"page_size":  getIntOrDefault(data, "page_size", 20),
	}
	
	// 提取过滤条件
	if filter, ok := data["filter"].(map[string]interface{}); ok {
		standardRequest["filter"] = filter
	}
	
	return standardRequest
}

// getIntOrDefault 辅助函数：获取整数值或默认值
func getIntOrDefault(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return defaultValue
}

// ErrorResponse 生成VueCMF格式的错误响应
func (t *DataTranslator) ErrorResponse(code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"code":    code,
		"msg":     message,
		"status":  "error",
		"message": message,
		"data":    nil,
	}
}





