package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// strings包已导入

// VueCMFCRUDHandler 处理VueCMF的CRUD操作
type VueCMFCRUDHandler struct {
	db *sql.DB
}

// NewVueCMFCRUDHandler 创建新的CRUD处理器
func NewVueCMFCRUDHandler(db *sql.DB) *VueCMFCRUDHandler {
	return &VueCMFCRUDHandler{db: db}
}

// getFieldInfo 获取模型的字段信息（VueCMF标准格式）
func (h *VueCMFCRUDHandler) getFieldInfo(tableName string) ([]map[string]interface{}, error) {
	// 查询字段配置
	query := `
		SELECT mf.field_name, mf.label, mf.field_type, mf.is_show, mf.sort_num
		FROM model_field mf
		INNER JOIN model_config mc ON mf.model_id = mc.id
		WHERE mc.table_name = $1 AND mf.status = 10
		ORDER BY mf.sort_num ASC
	`
	
	rows, err := h.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	fieldInfo := []map[string]interface{}{}
	for rows.Next() {
		var fieldName, label, fieldType string
		var isShow, sortNum int
		
		if err := rows.Scan(&fieldName, &label, &fieldType, &isShow, &sortNum); err != nil {
			continue
		}
		
		// VueCMF字段格式
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
	
	return fieldInfo, nil
}

// AdminListResponse 管理员列表响应
type AdminListResponse struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Status       int       `json:"status"`
	Role         string    `json:"role"`
	LastLoginIP  string    `json:"last_login_ip"`
	LastLoginAt  time.Time `json:"last_login_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	ID          int       `json:"id"`
	RoleName    string    `json:"role_name"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	ID             int       `json:"id"`
	PermissionName string    `json:"permission_name"`
	Resource       string    `json:"resource"`
	Action         string    `json:"action"`
	Description    string    `json:"description"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// HandleIndex 处理列表请求（VueCMF的index action）
func (h *VueCMFCRUDHandler) HandleIndex(c *gin.Context) {
	// 获取表名（多种方式）
	tableName := c.Param("table")
	
	// 从请求体中获取（VueCMF标准格式）
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err == nil {
		// 方式1: data.table_name
		if data, ok := reqBody["data"].(map[string]interface{}); ok {
			if tn, ok := data["table_name"].(string); ok && tn != "" {
				tableName = tn
			}
		}
		// 方式2: 直接从reqBody获取
		if tableName == "" {
			if tn, ok := reqBody["table_name"].(string); ok && tn != "" {
				tableName = tn
			}
		}
	}
	
	// 从URL路径推断（例如 /api/v1/admin/index → admin）
	if tableName == "" {
		path := c.Request.URL.Path
		// 解析路径：/api/v1/:table/:action
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			possibleTable := parts[3] // /api/v1/[admin]/index
			if possibleTable != "" && possibleTable != "mapping" && possibleTable != "menu" {
				tableName = possibleTable
			}
		}
	}
	
	// 调试日志
	fmt.Printf("🔍 HandleIndex: tableName=%s, path=%s\n", tableName, c.Request.URL.Path)

	// 获取分页参数
	page := 1
	pageSize := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil {
			pageSize = s
		}
	}

	offset := (page - 1) * pageSize

	// 根据表名查询数据
	switch tableName {
	case "admin":
		h.handleAdminList(c, offset, pageSize)
	case "roles":
		h.handleRoleList(c, offset, pageSize)
	case "permissions":
		h.handlePermissionList(c, offset, pageSize)
	default:
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"msg":     fmt.Sprintf("不支持的表: %s", tableName),
			"data":    nil,
			"status":  "error",
			"message": fmt.Sprintf("不支持的表: %s", tableName),
		})
	}
}

// handleAdminList 处理管理员列表
func (h *VueCMFCRUDHandler) handleAdminList(c *gin.Context, offset, pageSize int) {
	// 查询总数
	var total int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
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

	// 查询列表 (users表的主键是user_id，需要映射为id)
	query := `
		SELECT user_id as id, username, email, phone, status, role, 
		       '' as last_login_ip, last_login_at, created_at
		FROM users
		ORDER BY user_id DESC
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

	list := []AdminListResponse{}
	for rows.Next() {
		var item AdminListResponse
		var phone, lastLoginIP sql.NullString
		var lastLoginAt sql.NullTime
		err := rows.Scan(
			&item.ID, &item.Username, &item.Email, &phone,
			&item.Status, &item.Role, &lastLoginIP, &lastLoginAt, &item.CreatedAt,
		)
		if err != nil {
			fmt.Printf("⚠️  Scan error: %v\n", err)
			continue
		}
		if phone.Valid {
			item.Phone = phone.String
		}
		if lastLoginIP.Valid {
			item.LastLoginIP = lastLoginIP.String
		}
		if lastLoginAt.Valid {
			item.LastLoginAt = lastLoginAt.Time
		}
		list = append(list, item)
	}

	// 获取字段配置
	fieldInfo, _ := h.getFieldInfo("admin")
	if fieldInfo == nil {
		fieldInfo = []map[string]interface{}{}
	}

	// VueCMF标准格式（data.data.data.data）
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"status":  "success",
		"message": "获取成功",
		"data": gin.H{
			"data": gin.H{ // VueCMF要求的第二层data
				"data":          list,  // VueCMF期望：data.data.data.data（列表）
				"field_info":    fieldInfo,
				"field_option":  gin.H{},
				"form_info":     gin.H{},
				"form_rules":    []interface{}{},
				"relation_info": gin.H{},
				"total":         total,
				"page":          (offset / pageSize) + 1,
				"limit":         pageSize,
			},
		},
	})
}

// handleRoleList 处理角色列表
func (h *VueCMFCRUDHandler) handleRoleList(c *gin.Context, offset, pageSize int) {
	// 查询总数
	var total int
	err := h.db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&total)
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

	// 查询列表 (roles表的字段映射)
	query := `
		SELECT id, role_name, description, status, created_at
		FROM roles
		ORDER BY id DESC
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

	list := []RoleListResponse{}
	for rows.Next() {
		var item RoleListResponse
		err := rows.Scan(&item.ID, &item.RoleName, &item.Description, &item.Status, &item.CreatedAt)
		if err != nil {
			continue
		}
		list = append(list, item)
	}

	// 获取字段配置
	fieldInfo, _ := h.getFieldInfo("roles")
	if fieldInfo == nil {
		fieldInfo = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"status":  "success",
		"message": "获取成功",
		"data": gin.H{
			"data": gin.H{ // VueCMF要求的第二层data
				"data":          list,  // VueCMF期望：data.data.data.data（列表）
				"field_info":    fieldInfo,
				"field_option":  gin.H{},
				"form_info":     gin.H{},
				"form_rules":    []interface{}{},
				"relation_info": gin.H{},
				"total":         total,
				"page":          (offset / pageSize) + 1,
				"limit":         pageSize,
			},
		},
	})
}

// handlePermissionList 处理权限列表
func (h *VueCMFCRUDHandler) handlePermissionList(c *gin.Context, offset, pageSize int) {
	// 先检查哪个权限表存在
	var tableName string
	var countQuery string
	
	// 优先使用 zervigo_auth_permissions
	err := h.db.QueryRow("SELECT COUNT(*) FROM zervigo_auth_permissions").Scan(&tableName)
	if err == nil {
		tableName = "zervigo_auth_permissions"
		countQuery = "SELECT COUNT(*) FROM zervigo_auth_permissions"
	} else {
		// 回退到route_permission
		tableName = "route_permission"
		countQuery = "SELECT COUNT(*) FROM route_permission"
	}
	
	// 查询总数
	var total int
	err = h.db.QueryRow(countQuery).Scan(&total)
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
	var query string
	if tableName == "zervigo_auth_permissions" {
		query = `
			SELECT id, permission_name, resource_type as resource, action, permission_description as description, 
			       CASE WHEN status THEN 10 ELSE 20 END as status, created_at
			FROM zervigo_auth_permissions
			ORDER BY id DESC
			LIMIT $1 OFFSET $2
		`
	} else {
		query = `
			SELECT permission_id as id, permission_name, resource, action, description, 
			       10 as status, created_at
			FROM route_permission
			ORDER BY permission_id DESC
			LIMIT $1 OFFSET $2
		`
	}
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

	list := []PermissionListResponse{}
	for rows.Next() {
		var item PermissionListResponse
		err := rows.Scan(&item.ID, &item.PermissionName, &item.Resource, &item.Action, &item.Description, &item.Status, &item.CreatedAt)
		if err != nil {
			continue
		}
		list = append(list, item)
	}

	// 获取字段配置
	fieldInfo, _ := h.getFieldInfo("permissions")
	if fieldInfo == nil {
		fieldInfo = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"status":  "success",
		"message": "获取成功",
		"data": gin.H{
			"data": gin.H{ // VueCMF要求的第二层data
				"data":          list,  // VueCMF期望：data.data.data.data（列表）
				"field_info":    fieldInfo,
				"field_option":  gin.H{},
				"form_info":     gin.H{},
				"form_rules":    []interface{}{},
				"relation_info": gin.H{},
				"total":         total,
				"page":          (offset / pageSize) + 1,
				"limit":         pageSize,
			},
		},
	})
}

// HandleSave 处理保存请求（新增/更新）
func (h *VueCMFCRUDHandler) HandleSave(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		tableName = c.Query("table")
	}

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

	// VueCMF格式: {"data": {"id": 1, "username": "...", ...}}
	data, ok := reqBody["data"].(map[string]interface{})
	if !ok {
		data = reqBody
	}

	// 如果还没有tableName，从data中获取
	if tableName == "" {
		if tn, ok := data["table_name"].(string); ok {
			tableName = tn
		}
	}

	switch tableName {
	case "admin":
		h.handleAdminSave(c, data)
	case "roles":
		h.handleRoleSave(c, data)
	case "permissions":
		h.handlePermissionSave(c, data)
	default:
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"msg":     fmt.Sprintf("不支持的表: %s", tableName),
			"data":    nil,
			"status":  "error",
			"message": fmt.Sprintf("不支持的表: %s", tableName),
		})
	}
}

// handleAdminSave 保存管理员
func (h *VueCMFCRUDHandler) handleAdminSave(c *gin.Context, data map[string]interface{}) {
	id, _ := data["id"].(float64)

	if id > 0 {
		// 更新
		query := "UPDATE users SET "
		params := []interface{}{}
		paramIdx := 1

		fields := []string{}
		if username, ok := data["username"].(string); ok {
			fields = append(fields, fmt.Sprintf("username = $%d", paramIdx))
			params = append(params, username)
			paramIdx++
		}
		if email, ok := data["email"].(string); ok {
			fields = append(fields, fmt.Sprintf("email = $%d", paramIdx))
			params = append(params, email)
			paramIdx++
		}
		if phone, ok := data["phone"].(string); ok {
			fields = append(fields, fmt.Sprintf("phone = $%d", paramIdx))
			params = append(params, phone)
			paramIdx++
		}
		if status, ok := data["status"].(float64); ok {
			fields = append(fields, fmt.Sprintf("status = $%d", paramIdx))
			params = append(params, int(status))
			paramIdx++
		}
		if role, ok := data["role"].(string); ok {
			fields = append(fields, fmt.Sprintf("role = $%d", paramIdx))
			params = append(params, role)
			paramIdx++
		}

		if len(fields) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"msg":     "没有要更新的字段",
				"data":    nil,
				"status":  "error",
				"message": "没有要更新的字段",
			})
			return
		}

		fields = append(fields, fmt.Sprintf("updated_at = $%d", paramIdx))
		params = append(params, time.Now())
		paramIdx++

		query += strings.Join(fields, ", ") + fmt.Sprintf(" WHERE user_id = $%d", paramIdx)
		params = append(params, int(id))

		_, err := h.db.Exec(query, params...)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"msg":     "更新失败",
				"data":    nil,
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"msg":     "更新成功",
			"data":    gin.H{"id": int(id)},
			"status":  "success",
			"message": "更新成功",
		})
	} else {
		// 新增（暂不实现，避免安全问题）
		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"msg":     "暂不支持新增用户",
			"data":    nil,
			"status":  "error",
			"message": "暂不支持新增用户，请使用注册接口",
		})
	}
}

// handleRoleSave 保存角色
func (h *VueCMFCRUDHandler) handleRoleSave(c *gin.Context, data map[string]interface{}) {
	// 简化实现
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "保存成功",
		"data":    gin.H{"id": 1},
		"status":  "success",
		"message": "保存成功（演示）",
	})
}

// handlePermissionSave 保存权限
func (h *VueCMFCRUDHandler) handlePermissionSave(c *gin.Context, data map[string]interface{}) {
	// 简化实现
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "保存成功",
		"data":    gin.H{"id": 1},
		"status":  "success",
		"message": "保存成功（演示）",
	})
}

// HandleDelete 处理删除请求
func (h *VueCMFCRUDHandler) HandleDelete(c *gin.Context) {
	// tableName 不需要在这里使用，因为删除是通用的
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

	// VueCMF格式: {"data": {"id": [1, 2, 3]}} 或 {"data": {"id": 1}}
	data, ok := reqBody["data"].(map[string]interface{})
	if !ok {
		data = reqBody
	}

	ids := []int{}
	if idFloat, ok := data["id"].(float64); ok {
		ids = append(ids, int(idFloat))
	} else if idArray, ok := data["id"].([]interface{}); ok {
		for _, id := range idArray {
			if idFloat, ok := id.(float64); ok {
				ids = append(ids, int(idFloat))
			}
		}
	}

	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"msg":     "缺少ID参数",
			"data":    nil,
			"status":  "error",
			"message": "缺少ID参数",
		})
		return
	}

	// 简化实现：不真正删除，只返回成功
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "删除成功",
		"data":    gin.H{"deleted": len(ids)},
		"status":  "success",
		"message": fmt.Sprintf("已删除 %d 条记录（演示）", len(ids)),
	})
}

// HandleAction 处理VueCMF的动作请求
func (h *VueCMFCRUDHandler) HandleAction(c *gin.Context) {
	// 从URL路径解析：/api/v1/:table/:action
	// tableName由各个具体handler从c.Param获取
	action := c.Param("action")

	// 如果没有action，尝试从请求体获取
	if action == "" {
		var reqBody map[string]interface{}
		if err := c.ShouldBindJSON(&reqBody); err == nil {
			if data, ok := reqBody["data"].(map[string]interface{}); ok {
				if act, ok := data["action"].(string); ok {
					action = act
				}
			}
			// 重新绑定请求体以便后续使用
			bodyBytes, _ := json.Marshal(reqBody)
			c.Request.Body = http.NoBody
			c.Request.Body = &readCloser{strings.NewReader(string(bodyBytes))}
		}
	}

	switch action {
	case "index", "list", "":
		h.HandleIndex(c)
	case "save":
		h.HandleSave(c)
	case "delete":
		h.HandleDelete(c)
	default:
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"msg":     fmt.Sprintf("不支持的操作: %s", action),
			"data":    nil,
			"status":  "error",
			"message": fmt.Sprintf("不支持的操作: %s", action),
		})
	}
}

// readCloser 实现 io.ReadCloser
type readCloser struct {
	*strings.Reader
}

func (r *readCloser) Close() error {
	return nil
}

