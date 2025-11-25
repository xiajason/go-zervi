package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VueCMFModelHandler 处理VueCMF的模型配置
type VueCMFModelHandler struct {
	db *sql.DB
}

// NewVueCMFModelHandler 创建新的模型处理器
func NewVueCMFModelHandler(db *sql.DB) *VueCMFModelHandler {
	return &VueCMFModelHandler{db: db}
}

// GetModelConfig 获取模型配置
func (h *VueCMFModelHandler) GetModelConfig(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"msg":     "请求参数错误",
			"data":    nil,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 解析分页参数
	page := 1
	pageSize := 20
	if data, ok := reqBody["data"].(map[string]interface{}); ok {
		if p, ok := data["page"].(float64); ok {
			page = int(p)
		}
		if ps, ok := data["page_size"].(float64); ok {
			pageSize = int(ps)
		}
	}

	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err := h.db.QueryRow("SELECT COUNT(*) FROM model_config WHERE status = 10").Scan(&total)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"msg":     "查询失败",
			"data":    nil,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 查询列表
	query := `
		SELECT id, table_name, label, app_id, status, created_at, updated_at
		FROM model_config
		WHERE status = 10
		ORDER BY id ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := h.db.Query(query, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"msg":     "查询失败",
			"data":    nil,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, appID, status int
		var tableName, label string
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &tableName, &label, &appID, &status, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":         id,
			"table_name": tableName,
			"label":      label,
			"app_id":     appID,
			"status":     status,
		}

		if createdAt.Valid {
			item["created_at"] = createdAt.Time
		}
		if updatedAt.Valid {
			item["updated_at"] = updatedAt.Time
		}

		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"status":  "success",
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
			"limit": pageSize,
		},
	})
}

// GetModelField 获取模型字段配置
func (h *VueCMFModelHandler) GetModelField(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"msg":     "请求参数错误",
			"data":    nil,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 解析参数
	page := 1
	pageSize := 100
	var modelID int
	var tableName string

	if data, ok := reqBody["data"].(map[string]interface{}); ok {
		if p, ok := data["page"].(float64); ok {
			page = int(p)
		}
		if ps, ok := data["page_size"].(float64); ok {
			pageSize = int(ps)
		}
		if mid, ok := data["model_id"].(float64); ok {
			modelID = int(mid)
		}
		if tn, ok := data["table_name"].(string); ok {
			tableName = tn
		}
		// 支持filter格式
		if filters, ok := data["filter"].([]interface{}); ok {
			for _, f := range filters {
				if filter, ok := f.(map[string]interface{}); ok {
					if field, _ := filter["field"].(string); field == "table_name" {
						if value, ok := filter["value"].(string); ok {
							tableName = value
						}
					}
					if field, _ := filter["field"].(string); field == "model_id" {
						if value, ok := filter["value"].(float64); ok {
							modelID = int(value)
						}
					}
				}
			}
		}
	}

	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "WHERE mf.status = 10"
	args := []interface{}{}

	if modelID > 0 {
		whereClause += " AND mf.model_id = $1"
		args = append(args, modelID)
	} else if tableName != "" {
		// 通过table_name查找model_id
		var mid int
		err := h.db.QueryRow("SELECT id FROM model_config WHERE table_name = $1", tableName).Scan(&mid)
		if err == nil {
			whereClause += " AND mf.model_id = $1"
			args = append(args, mid)
		}
	}

	// 查询总数
	countQuery := `
		SELECT COUNT(*)
		FROM model_field mf
		` + whereClause
	var total int
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"msg":     "查询失败",
			"data":    nil,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 查询列表（使用固定的占位符）
	var query string
	if len(args) > 0 {
		query = `
			SELECT 
				mf.id, mf.model_id, mf.field_name, mf.label, mf.field_type,
				mf.default_value, mf.is_required, mf.is_show, mf.sort_num,
				mf.note, mf.status, mf.created_at, mf.updated_at,
				mc.table_name
			FROM model_field mf
			LEFT JOIN model_config mc ON mf.model_id = mc.id
			` + whereClause + `
			ORDER BY mf.sort_num ASC, mf.id ASC
			LIMIT $2 OFFSET $3
		`
		args = append(args, pageSize, offset)
	} else {
		query = `
			SELECT 
				mf.id, mf.model_id, mf.field_name, mf.label, mf.field_type,
				mf.default_value, mf.is_required, mf.is_show, mf.sort_num,
				mf.note, mf.status, mf.created_at, mf.updated_at,
				mc.table_name
			FROM model_field mf
			LEFT JOIN model_config mc ON mf.model_id = mc.id
			` + whereClause + `
			ORDER BY mf.sort_num ASC, mf.id ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{pageSize, offset}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"msg":     "查询失败",
			"data":    nil,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, modelID, isRequired, isShow, sortNum, status int
		var fieldName, label, fieldType string
		var defaultValue, note, tableNameVal sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&id, &modelID, &fieldName, &label, &fieldType,
			&defaultValue, &isRequired, &isShow, &sortNum,
			&note, &status, &createdAt, &updatedAt, &tableNameVal,
		)
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":          id,
			"model_id":    modelID,
			"field_name":  fieldName,
			"label":       label,
			"field_type":  fieldType,
			"is_required": isRequired,
			"is_show":     isShow,
			"sort_num":    sortNum,
			"status":      status,
		}

		if defaultValue.Valid {
			item["default_value"] = defaultValue.String
		}
		if note.Valid {
			item["note"] = note.String
		}
		if tableNameVal.Valid {
			item["table_name"] = tableNameVal.String
		}
		if createdAt.Valid {
			item["created_at"] = createdAt.Time
		}
		if updatedAt.Valid {
			item["updated_at"] = updatedAt.Time
		}

		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"status":  "success",
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
			"limit": pageSize,
		},
	})
}

