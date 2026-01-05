package blockchain

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BlockchainAPI 区块链API
type BlockchainAPI struct {
	service *BlockchainService
	port    int
}

// NewBlockchainAPI 创建区块链API
func NewBlockchainAPI(service *BlockchainService, port int) *BlockchainAPI {
	return &BlockchainAPI{
		service: service,
		port:    port,
	}
}

// Start 启动API服务器
func (api *BlockchainAPI) Start() error {
	router := gin.Default()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 健康检查
	router.GET("/health", api.healthCheck)

	// API路由组
	v1 := router.Group("/api/v1/blockchain")
	{
		// 版本状态相关API
		version := v1.Group("/version")
		{
			version.POST("/status/record", api.recordVersionStatusChange)
			version.GET("/status/history/:user_id", api.getUserStatusHistory)
		}

		// 权限变更相关API
		permission := v1.Group("/permission")
		{
			permission.POST("/change/record", api.recordPermissionChange)
			permission.GET("/change/history/:user_id", api.getUserPermissionHistory)
		}

		// 数据一致性校验API
		v1.POST("/consistency/validate", api.validateDataConsistency)

		// 交易查询API
		v1.GET("/transaction/list", api.getTransactionList)
	}

	return router.Run(":" + strconv.Itoa(api.port))
}

// healthCheck 健康检查
func (api *BlockchainAPI) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "区块链服务健康",
		"data": gin.H{
			"service":   "blockchain-service",
			"status":    "UP",
			"version":   "1.0.0",
			"timestamp": gin.H{},
		},
	})
}

// recordVersionStatusChange 记录版本状态变化
func (api *BlockchainAPI) recordVersionStatusChange(c *gin.Context) {
	var req VersionStatusChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if req.UserID == "" || req.VersionSource == "" || req.NewStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID、版本来源和新状态不能为空",
		})
		return
	}

	// 验证版本来源
	if !isValidVersionSource(req.VersionSource) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的版本来源，必须是 BASIC、PROFESSIONAL 或 FUTURE",
		})
		return
	}

	response, err := api.service.RecordVersionStatusChange(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "记录版本状态变化失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getUserStatusHistory 获取用户状态历史
func (api *BlockchainAPI) getUserStatusHistory(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID不能为空",
		})
		return
	}

	response, err := api.service.GetUserStatusHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询用户状态历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// recordPermissionChange 记录权限变更
func (api *BlockchainAPI) recordPermissionChange(c *gin.Context) {
	var req PermissionChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if req.UserID == "" || req.VersionSource == "" || req.NewPermission == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID、版本来源和新权限不能为空",
		})
		return
	}

	// 验证版本来源
	if !isValidVersionSource(req.VersionSource) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的版本来源，必须是 BASIC、PROFESSIONAL 或 FUTURE",
		})
		return
	}

	response, err := api.service.RecordPermissionChange(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "记录权限变更失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getUserPermissionHistory 获取用户权限历史
func (api *BlockchainAPI) getUserPermissionHistory(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID不能为空",
		})
		return
	}

	response, err := api.service.GetUserPermissionHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询用户权限历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// validateDataConsistency 数据一致性校验
func (api *BlockchainAPI) validateDataConsistency(c *gin.Context) {
	response, err := api.service.ValidateDataConsistency(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据一致性校验失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getTransactionList 获取交易列表
func (api *BlockchainAPI) getTransactionList(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	transactionType := c.Query("transaction_type")
	versionSource := c.Query("version_source")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	// 构建查询SQL
	query := `
	SELECT transaction_id, transaction_hash, transaction_type, version_source,
		   user_id, old_status, new_status, change_reason, operator_id,
		   status, block_height, create_time, confirm_time
	FROM blockchain_transaction
	WHERE 1=1`

	args := []interface{}{}

	if transactionType != "" {
		query += " AND transaction_type = ?"
		args = append(args, transactionType)
	}

	if versionSource != "" {
		query += " AND version_source = ?"
		args = append(args, versionSource)
	}

	query += " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	// 这里应该实现实际的数据库查询
	// 为了简化，返回模拟数据
	transactions := []map[string]interface{}{
		{
			"transaction_id":   "VS1234567890",
			"transaction_hash": "0xABCDEF1234567890",
			"transaction_type": "VERSION_STATUS",
			"version_source":   "BASIC",
			"user_id":          "user_001",
			"old_status":       "ACTIVE",
			"new_status":       "INACTIVE",
			"change_reason":    "用户注销",
			"operator_id":      "admin_001",
			"status":           "CONFIRMED",
			"block_height":     12345,
			"create_time":      "2025-01-10 10:00:00",
			"confirm_time":     "2025-01-10 10:00:05",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询交易列表成功",
		"data":    transactions,
		"total":   1,
		"page":    page,
		"size":    size,
	})
}

// isValidVersionSource 验证版本来源是否有效
func isValidVersionSource(versionSource string) bool {
	validVersions := []string{"BASIC", "PROFESSIONAL", "FUTURE"}
	for _, v := range validVersions {
		if versionSource == v {
			return true
		}
	}
	return false
}
