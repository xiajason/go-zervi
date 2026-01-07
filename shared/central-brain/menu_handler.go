package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	db *sql.DB
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(db *sql.DB) *MenuHandler {
	return &MenuHandler{db: db}
}

// GetMenuNav 获取导航菜单
func (h *MenuHandler) GetMenuNav(c *gin.Context) {
	query := `
		SELECT menu_id, pid, model_id, title, path, icon, sort_num, status
		FROM vuecmf_menu
		WHERE status = 10
		ORDER BY sort_num ASC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("❌ 查询菜单失败: %v", err)
		c.JSON(200, gin.H{
			"code": 500,
			"msg":  "查询菜单失败",
			"data": nil,
		})
		return
	}
	defer rows.Close()

	menus := []map[string]interface{}{}
	for rows.Next() {
		var menuID, pid, modelID, sortNum, status int
		var title, path, icon sql.NullString

		err := rows.Scan(&menuID, &pid, &modelID, &title, &path, &icon, &sortNum, &status)
		if err != nil {
			log.Printf("❌ 扫描菜单数据失败: %v", err)
			continue
		}

		menu := map[string]interface{}{
			"id":       menuID,
			"pid":      pid,
			"model_id": modelID,
			"title":    title.String,
			"path":     path.String,
			"icon":     icon.String,
			"sort_num": sortNum,
			"status":   status,
		}
		menus = append(menus, menu)
	}

	log.Printf("✅ 成功获取 %d 条菜单", len(menus))

	// VueCMF 格式响应
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": menus,
	})
}

// RegisterMenuRoutes 注册菜单路由
func RegisterMenuRoutes(r *gin.Engine, db *sql.DB) {
	handler := NewMenuHandler(db)
	
	// 公开访问的菜单路由
	r.GET("/api/v1/router/menu/nav", handler.GetMenuNav)
	r.GET("/api/v1/router/menu/list", handler.GetMenuNav) // 使用相同的处理器
	
	log.Println("✅ 菜单路由已注册:")
	log.Println("   GET /api/v1/router/menu/nav")
	log.Println("   GET /api/v1/router/menu/list")
}

