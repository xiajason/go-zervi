package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/central-brain/translator"
)

// VueCMFCRUDHandlerV2 使用翻译层的CRUD处理器
// 体现中央大脑的翻译和校验职责
type VueCMFCRUDHandlerV2 struct {
	db         *sql.DB
	translator *translator.DataTranslator
}

// NewVueCMFCRUDHandlerV2 创建V2处理器
func NewVueCMFCRUDHandlerV2(db *sql.DB) *VueCMFCRUDHandlerV2 {
	return &VueCMFCRUDHandlerV2{
		db:         db,
		translator: translator.NewDataTranslator(db),
	}
}

// HandleIndexV2 处理列表请求（使用翻译层）
func (h *VueCMFCRUDHandlerV2) HandleIndexV2(c *gin.Context) {
	// 1. 解析VueCMF格式的请求
	var vuecmfRequest map[string]interface{}
	if err := c.ShouldBindJSON(&vuecmfRequest); err != nil {
		c.JSON(200, h.translator.ErrorResponse(400, "请求参数错误"))
		return
	}
	
	// 2. 中央大脑：校验请求
	if err := h.translator.ValidateRequest(vuecmfRequest); err != nil {
		c.JSON(200, h.translator.ErrorResponse(400, err.Error()))
		return
	}
	
	// 3. 中央大脑：翻译请求格式
	standardRequest := h.translator.TranslateFromVueCMF(vuecmfRequest)
	tableName := standardRequest["table_name"].(string)
	page := standardRequest["page"].(int)
	pageSize := standardRequest["page_size"].(int)
	
	fmt.Printf("🧠 中央大脑翻译: VueCMF请求 → 标准请求 (table=%s, page=%d)\n", tableName, page)
	
	// 4. 从数据库获取原始数据（数据库只负责存储）
	rawData, total, err := h.queryRawData(tableName, page, pageSize)
	if err != nil {
		c.JSON(200, h.translator.ErrorResponse(500, err.Error()))
		return
	}
	
	// 5. 中央大脑：翻译响应格式（标准格式 → VueCMF格式）
	vuecmfResponse := h.translator.TranslateToVueCMF(tableName, rawData, total, page, pageSize)
	
	fmt.Printf("🧠 中央大脑翻译: 标准响应 → VueCMF响应 (total=%d)\n", total)
	
	c.JSON(200, vuecmfResponse)
}

// queryRawData 从数据库获取原始数据（通用查询，不关心格式）
func (h *VueCMFCRUDHandlerV2) queryRawData(tableName string, page int, pageSize int) (interface{}, int, error) {
	offset := (page - 1) * pageSize
	
	// 根据表名映射到实际数据库表
	dbTable, columns := h.getTableMapping(tableName)
	if dbTable == "" {
		return nil, 0, fmt.Errorf("不支持的表: %s", tableName)
	}
	
	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", dbTable)
	if err := h.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s DESC LIMIT $1 OFFSET $2", 
		columns, dbTable, h.getPrimaryKey(tableName))
	
	rows, err := h.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	// 动态解析行数据
	data := h.parseRows(rows, tableName)
	
	return data, total, nil
}

// getTableMapping 表名映射（中央大脑的翻译规则）
func (h *VueCMFCRUDHandlerV2) getTableMapping(tableName string) (string, string) {
	mappings := map[string]struct {
		table   string
		columns string
	}{
		"admin": {
			table:   "users",
			columns: "user_id as id, username, email, phone, status, role, last_login_at, created_at",
		},
		"roles": {
			table:   "roles",
			columns: "id, role_name, description, status, created_at",
		},
		"permissions": {
			table:   "zervigo_auth_permissions",
			columns: "id, permission_name, resource_type as resource, action, permission_description as description, CASE WHEN status THEN 10 ELSE 20 END as status, created_at",
		},
	}
	
	if mapping, ok := mappings[tableName]; ok {
		return mapping.table, mapping.columns
	}
	return "", ""
}

// getPrimaryKey 获取主键字段
func (h *VueCMFCRUDHandlerV2) getPrimaryKey(tableName string) string {
	keys := map[string]string{
		"admin":       "id",
		"roles":       "id",
		"permissions": "id",
	}
	if key, ok := keys[tableName]; ok {
		return key
	}
	return "id"
}

// parseRows 动态解析数据库行（通用方法）
func (h *VueCMFCRUDHandlerV2) parseRows(rows *sql.Rows, tableName string) []map[string]interface{} {
	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}
	}
	
	// 创建扫描目标
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	
	result := []map[string]interface{}{}
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			// 处理NULL值
			if val == nil {
				row[col] = ""
				continue
			}
			
			// 类型转换
			switch v := val.(type) {
			case []byte:
				row[col] = string(v)
			case time.Time:
				row[col] = v
			default:
				row[col] = v
			}
		}
		
		result = append(result, row)
	}
	
	return result
}

// ValidateAndTranslateRequest 校验并翻译请求（中央大脑的核心职责）
func (h *VueCMFCRUDHandlerV2) ValidateAndTranslateRequest(c *gin.Context) (string, int, int, error) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		return "", 0, 0, fmt.Errorf("请求格式错误: %v", err)
	}
	
	// 校验
	if err := h.translator.ValidateRequest(reqBody); err != nil {
		return "", 0, 0, err
	}
	
	// 翻译
	standardReq := h.translator.TranslateFromVueCMF(reqBody)
	
	tableName := standardReq["table_name"].(string)
	page := standardReq["page"].(int)
	pageSize := standardReq["page_size"].(int)
	
	// 从URL路径推断表名（备用方案）
	if tableName == "" {
		path := c.Request.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			tableName = parts[3]
		}
	}
	
	return tableName, page, pageSize, nil
}





