package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"

	"github.com/szjason72/zervigo/shared/central-brain/vuecmf"
)

// VueCMFHandler VueCMF 处理器
type VueCMFHandler struct {
	apiMappingService *vuecmf.APIMappingService
	db                *sql.DB
	redis             *redis.Client
}

// NewVueCMFHandler 创建 VueCMF 处理器
func NewVueCMFHandler(config map[string]string) (*VueCMFHandler, error) {
	// 获取数据库配置
	dbHost := config["DB_HOST"]
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := config["DB_PORT"]
	if dbPort == "" {
		dbPort = "5432"
	}
	dbName := config["DB_NAME"]
	if dbName == "" {
		dbName = "zervigo_mvp"
	}
	dbUser := config["DB_USER"]
	if dbUser == "" {
		dbUser = "szjason72"
	}
	dbPassword := config["DB_PASSWORD"]

	// 连接数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName)
	if dbPassword != "" {
		connStr += fmt.Sprintf(" password=%s", dbPassword)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 ping 失败: %v", err)
	}

	log.Printf("✅ VueCMF 数据库连接成功: %s:%s/%s", dbHost, dbPort, dbName)

	// 连接 Redis（可选）
	redisAddr := config["REDIS_ADDR"]
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 测试 Redis 连接
	ctx := rdb.Context()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis 连接失败（将不使用缓存）: %v", err)
		rdb = nil
	} else {
		log.Printf("✅ VueCMF Redis 连接成功: %s", redisAddr)
	}

	// 创建 API 映射服务
	apiMappingService := vuecmf.NewAPIMappingService(db, rdb)

	return &VueCMFHandler{
		apiMappingService: apiMappingService,
		db:                db,
		redis:             rdb,
	}, nil
}

// GetAPIMap 获取 API 映射
func (h *VueCMFHandler) GetAPIMap(c *gin.Context) {
	var req struct {
		Data struct {
			TableName  string `json:"table_name"`
			ActionType string `json:"action_type"`
			AppID      int    `json:"app_id"`
		} `json:"data"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, vuecmf.Error(vuecmf.CodeValidation, "参数错误"))
		return
	}

	// 参数验证
	if req.Data.TableName == "" || req.Data.ActionType == "" {
		c.JSON(200, vuecmf.Error(vuecmf.CodeValidation, "table_name 和 action_type 不能为空"))
		return
	}

	// 查询 API 路径
	apiPath, err := h.apiMappingService.GetAPIPath(
		req.Data.TableName,
		req.Data.ActionType,
		req.Data.AppID,
	)

	if err != nil {
		log.Printf("❌ API 映射查询失败: %v", err)
		c.JSON(200, vuecmf.Error(vuecmf.CodeNotFound, err.Error()))
		return
	}

	log.Printf("✅ API 映射查询成功: %s.%s → %s", req.Data.TableName, req.Data.ActionType, apiPath)
	c.JSON(200, vuecmf.Success(apiPath))
}

// GetAllMappings 获取所有 API 映射（调试用）
func (h *VueCMFHandler) GetAllMappings(c *gin.Context) {
	mappings, err := h.apiMappingService.GetAllMappings()
	if err != nil {
		log.Printf("❌ 获取所有映射失败: %v", err)
		c.JSON(200, vuecmf.Error(vuecmf.CodeInternal, "查询失败"))
		return
	}

	log.Printf("✅ 获取所有映射成功: %d 条", len(mappings))
	c.JSON(200, vuecmf.Success(gin.H{
		"total":    len(mappings),
		"mappings": mappings,
	}))
}

// GetMenuNav 获取导航菜单
func (h *VueCMFHandler) GetMenuNav(c *gin.Context) {
	query := `
		SELECT menu_id, pid, COALESCE(model_id, 0), title, 
		       COALESCE(path, ''), COALESCE(icon, ''), sort_num, status,
		       COALESCE(component_tpl, ''), COALESCE(path_name, ''), COALESCE(mid, ''),
		       COALESCE(table_name, ''), COALESCE(default_action_type, 'index'),
		       COALESCE(id_path, ''), COALESCE(is_tree, 20)
		FROM vuecmf_menu
		WHERE status = 10
		ORDER BY sort_num ASC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("❌ 查询菜单失败: %v", err)
		// 返回临时静态菜单
		menus := []map[string]interface{}{
			{"id": 1, "pid": 0, "model_id": 1, "title": "首页", "path": "/dashboard", "icon": "HomeFilled", "sort_num": 1, "status": 10},
			{"id": 2, "pid": 0, "model_id": 2, "title": "系统管理", "path": "/system", "icon": "Setting", "sort_num": 2, "status": 10},
		}
		
		// VueCMF 前端期望的完整格式
		responseData := gin.H{
			"nav_menu": menus,
			"api_maps": gin.H{},
		}
		
		c.JSON(200, gin.H{"code": 0, "msg": "success", "data": responseData})
		return
	}
	defer rows.Close()

	// 先获取所有菜单，然后构建树形结构
	type MenuItem struct {
		MenuID             int
		Pid                int
		ModelID            int
		Title              string
		Path               string
		Icon               string
		SortNum            int
		Status             int
		ComponentTpl       string
		PathName           string
		Mid                string
		TableName          string
		DefaultActionType  string
		IDPath             string  // 新增：层级路径
		IsTree             int     // 新增：是否树形
	}
	
	var allMenus []MenuItem
	for rows.Next() {
		var item MenuItem
		err := rows.Scan(&item.MenuID, &item.Pid, &item.ModelID, &item.Title, 
			&item.Path, &item.Icon, &item.SortNum, &item.Status, 
			&item.ComponentTpl, &item.PathName, &item.Mid,
			&item.TableName, &item.DefaultActionType,
			&item.IDPath, &item.IsTree)  // 新增字段
		if err != nil {
			log.Printf("❌ 扫描菜单数据失败: %v", err)
			continue
		}
		allMenus = append(allMenus, item)
	}
	
	// 构建树形菜单结构（VueCMF 期望的对象格式，不是数组）
	menuMap := make(map[string]interface{})
	childrenMap := make(map[int][]map[string]interface{})
	
	// 第一步：收集所有子菜单
	for _, item := range allMenus {
		pathNameArray := []string{}
		if item.PathName != "" {
			pathNameArray = append(pathNameArray, item.PathName)
		}
		
		// 构建 id_path 数组（VueCMF 用于面包屑导航）
		idPathArray := []string{}
		if item.IDPath != "" {
			// id_path 是逗号分隔的字符串，需要转为数组
			idPathParts := strings.Split(item.IDPath, ",")
			for _, part := range idPathParts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					idPathArray = append(idPathArray, trimmed)
				}
			}
		}
		
		menuData := map[string]interface{}{
			"id":                  item.MenuID,
			"pid":                 item.Pid,
			"model_id":            item.ModelID,
			"title":               item.Title,
			"path":                item.Path,
			"icon":                item.Icon,
			"sort_num":            item.SortNum,
			"status":              item.Status,
			"component_tpl":       item.ComponentTpl,
			"path_name":           pathNameArray,
			"mid":                 item.Mid,
			"table_name":          item.TableName,
			"default_action_type": item.DefaultActionType,
			"id_path":             idPathArray,  // 层级路径数组
			"is_tree":             item.IsTree,   // 是否树形结构
		}
		
		if item.Pid == 0 {
			// 顶级菜单，用 mid 作为 key
			menuMap[item.Mid] = menuData
		} else {
			// 子菜单，按 pid 分组
			childrenMap[item.Pid] = append(childrenMap[item.Pid], menuData)
		}
	}
	
	// 第二步：将子菜单添加到父菜单
	for mid, menu := range menuMap {
		menuData := menu.(map[string]interface{})
		menuID := menuData["id"].(int)
		
		if children, ok := childrenMap[menuID]; ok {
			// 转换为对象格式（VueCMF 期望）
			childrenObj := make(map[string]interface{})
			for i, child := range children {
				key := fmt.Sprintf("%d", i)
				childrenObj[key] = child
			}
			menuData["children"] = childrenObj
		}
		
		menuMap[mid] = menuData
	}

	log.Printf("✅ 成功获取 %d 条菜单（%d 个顶级菜单）", len(allMenus), len(menuMap))
	
	// 构建 API 映射（从数据库读取）
	apiMapsData := h.buildAPIMaps()
	
	// 构建菜单顺序数组（按 sort_num 排序的 mid 列表）
	type MenuOrder struct {
		Mid     string
		SortNum int
	}
	var menuOrders []MenuOrder
	for mid, menu := range menuMap {
		menuData := menu.(map[string]interface{})
		sortNum := menuData["sort_num"].(int)
		menuOrders = append(menuOrders, MenuOrder{Mid: mid, SortNum: sortNum})
	}
	
	// 排序
	for i := 0; i < len(menuOrders)-1; i++ {
		for j := i + 1; j < len(menuOrders); j++ {
			if menuOrders[i].SortNum > menuOrders[j].SortNum {
				menuOrders[i], menuOrders[j] = menuOrders[j], menuOrders[i]
			}
		}
	}
	
	// 提取排序后的 mid 列表
	orderedMids := make([]string, len(menuOrders))
	for i, order := range menuOrders {
		orderedMids[i] = order.Mid
	}
	
	// VueCMF 前端期望的完整格式
	responseData := gin.H{
		"nav_menu":    menuMap,     // 菜单数据（对象格式，key 是 mid）
		"api_maps":    apiMapsData, // API 映射（完整的映射对象）
		"menu_order":  orderedMids, // 菜单顺序（按 sort_num 排序的 mid 数组）
	}
	
	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": responseData})
}

// buildAPIMaps 构建 API 映射对象
func (h *VueCMFHandler) buildAPIMaps() map[string]interface{} {
	query := `
		SELECT table_name, action_type, api_path
		FROM vuecmf_api_map
		ORDER BY table_name, action_type
	`

	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("❌ 查询 API 映射失败: %v", err)
		return make(map[string]interface{})
	}
	defer rows.Close()

	// 构建嵌套的映射对象：{ table_name: { action_type: api_path } }
	apiMaps := make(map[string]interface{})

	for rows.Next() {
		var tableName, actionType, apiPath string
		if err := rows.Scan(&tableName, &actionType, &apiPath); err != nil {
			continue
		}

		// 如果该表名还没有映射，创建一个新的
		if apiMaps[tableName] == nil {
			apiMaps[tableName] = make(map[string]interface{})
		}

		// 添加映射：api_maps[table_name][action_type] = api_path
		tableMap := apiMaps[tableName].(map[string]interface{})
		tableMap[actionType] = apiPath
	}

	log.Printf("✅ 构建了 %d 个表的 API 映射", len(apiMaps))
	return apiMaps
}

// ClearCache 清除缓存
func (h *VueCMFHandler) ClearCache(c *gin.Context) {
	if err := h.apiMappingService.ClearCache(); err != nil {
		c.JSON(200, vuecmf.Error(vuecmf.CodeInternal, "清除缓存失败"))
		return
	}

	log.Println("✅ API 映射缓存已清除")
	c.JSON(200, vuecmf.Success("缓存已清除"))
}

// Close 关闭连接
func (h *VueCMFHandler) Close() error {
	if h.db != nil {
		h.db.Close()
	}
	if h.redis != nil {
		h.redis.Close()
	}
	return nil
}

