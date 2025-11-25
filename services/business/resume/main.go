package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	jobfirst "github.com/szjason72/zervigo/shared/core"
	"github.com/szjason72/zervigo/shared/core/auth"
	"github.com/szjason72/zervigo/shared/core/response"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "resume-service"
	}

	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 初始化Go-Zervi认证系统
	jwtSecret := "zervigo-mvp-secret-key-2025"

	// 智能识别：根据环境变量自动选择数据库
	var sqlDB *sql.DB

	if pgManager := core.Database.GetPostgreSQL(); pgManager != nil {
		var err error
		sqlDB, err = pgManager.GetSQLDB()
		if err != nil {
			log.Fatalf("获取PostgreSQL连接失败: %v", err)
		}
		log.Println("使用 PostgreSQL 数据库")
	} else if mysqlManager := core.Database.GetMySQL(); mysqlManager != nil {
		var err error
		sqlDB, err = mysqlManager.GetSQLDB()
		if err != nil {
			log.Fatalf("获取MySQL连接失败: %v", err)
		}
		log.Println("使用 MySQL 数据库")
	} else {
		log.Fatalf("未找到数据库配置（MySQL或PostgreSQL）")
	}

	// 创建认证系统
	authSystem := auth.NewUnifiedAuthSystem(sqlDB, jwtSecret)
	if err := authSystem.InitializeDatabase(); err != nil {
		log.Fatalf("初始化认证数据库失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 设置标准路由
	setupStandardRoutes(r, core)

	// 设置业务路由
	setupBusinessRoutes(r, core, sqlDB, jwtSecret)

	// 注册到Consul
	registerToConsul("resume-service", "127.0.0.1", 8085)

	// 启动服务
	log.Println("Starting Resume Service with Go-Zervi Framework on 0.0.0.0:8085")
	if err := r.Run(":8085"); err != nil {
		log.Fatalf("Resume Service启动失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "resume-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "resume-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "resume-service",
			"version":    "3.1.0",
			"port":       8085,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core, sqlDB *sql.DB, jwtSecret string) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 简历模板列表
		public.GET("/resume/templates", func(c *gin.Context) {
			templates := []gin.H{
				{"templateId": "template_001", "templateName": "经典模板", "previewUrl": "https://example.com/preview1.jpg"},
				{"templateId": "template_002", "templateName": "现代模板", "previewUrl": "https://example.com/preview2.jpg"},
				{"templateId": "template_003", "templateName": "创意模板", "previewUrl": "https://example.com/preview3.jpg"},
			}
			standardSuccessResponse(c, templates, "简历模板列表获取成功")
		})
	}

	// 需要认证的API路由
	zerviAuthAdapter := auth.NewZerviAuthAdapter(sqlDB, jwtSecret)
	authMiddleware := zerviAuthAdapter.RequireAuth()
	api := r.Group("/api/v1")
	api.Use(authMiddleware)
	{
		// 简历管理
		resume := api.Group("/resume")
		{
			// 获取简历列表摘要
			resume.GET("/list/summary", func(c *gin.Context) {
				userID := c.GetUint("user_id")
				summaries := getResumeSummaries(sqlDB, userID)
				standardSuccessResponse(c, summaries, "简历列表摘要获取成功")
			})

			// 获取简历详情
			resume.GET("/detail/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")
				detail := getResumeDetail(sqlDB, resumeID, userID)
				if detail == nil {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在", "")
					return
				}
				standardSuccessResponse(c, detail, "简历详情获取成功")
			})

			// 创建简历
			resume.POST("/create", func(c *gin.Context) {
				var req CreateResumeRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				userID := c.GetUint("user_id")
				resumeID := createResume(sqlDB, userID, req)
				if resumeID == "" {
					standardErrorResponse(c, http.StatusInternalServerError, "创建简历失败", "")
					return
				}

				result := gin.H{
					"resumeId": resumeID,
					"message":  "简历创建成功",
				}
				standardSuccessResponse(c, result, "简历创建成功")
			})

			// 更新简历
			resume.PUT("/update/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				var req UpdateResumeRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !updateResume(sqlDB, resumeID, userID, req) {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在或无权限", "")
					return
				}

				standardSuccessResponse(c, "简历已更新", "简历更新成功")
			})

			// 发布简历
			resume.POST("/publish/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				if !publishResume(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在或无权限", "")
					return
				}

				standardSuccessResponse(c, "简历已发布", "简历发布成功")
			})

			// 删除简历
			resume.DELETE("/publish/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				if !deleteResume(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在或无权限", "")
					return
				}

				standardSuccessResponse(c, "简历已删除", "简历删除成功")
			})

			// === 前端兼容层：/api/resume/** ===
			resume.GET("/current", func(c *gin.Context) {
				userID := c.GetUint("user_id")
				summaries := getResumeSummaries(sqlDB, userID)
				if len(summaries) == 0 {
					standardSuccessResponse(c, gin.H{
						"resumeId":     "",
						"resumeStatus": "EMPTY",
					}, "当前用户暂无简历")
					return
				}

				first := summaries[0]
				resumeID, _ := first["resumeId"].(string)
				if resumeID == "" {
					standardSuccessResponse(c, first, "获取简历摘要成功")
					return
				}

				detail := getResumeDetail(sqlDB, resumeID, userID)
				if detail == nil {
					standardSuccessResponse(c, first, "获取简历摘要成功")
					return
				}

				standardSuccessResponse(c, detail, "当前简历获取成功")
			})

			resume.GET("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				detail := getResumeDetail(sqlDB, resumeID, userID)
				if detail == nil {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在或无权限", "")
					return
				}

				standardSuccessResponse(c, detail, "简历详情获取成功")
			})

			resume.POST("", func(c *gin.Context) {
				var req CreateResumeRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				userID := c.GetUint("user_id")
				resumeID := createResume(sqlDB, userID, req)
				if resumeID == "" {
					standardErrorResponse(c, http.StatusInternalServerError, "创建简历失败", "")
					return
				}

				detail := getResumeDetail(sqlDB, resumeID, userID)
				if detail == nil {
					standardSuccessResponse(c, gin.H{"resumeId": resumeID}, "简历创建成功")
					return
				}

				standardSuccessResponse(c, detail, "简历创建成功")
			})

			resume.PUT("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				var req UpdateResumeRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !updateResume(sqlDB, resumeID, userID, req) {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在或无权限", "")
					return
				}

				detail := getResumeDetail(sqlDB, resumeID, userID)
				if detail == nil {
					standardSuccessResponse(c, gin.H{"resumeId": resumeID}, "简历更新成功")
					return
				}

				standardSuccessResponse(c, detail, "简历更新成功")
			})

			resume.DELETE("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				if !deleteResume(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusNotFound, "简历不存在或无权限", "")
					return
				}

				standardSuccessResponse(c, gin.H{
					"resumeId": resumeID,
					"deleted":  true,
				}, "简历删除成功")
			})
		}

		// 简历权限管理
		permission := api.Group("/resume/permission")
		{
			// 获取简历权限配置
			permission.GET("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				// 检查用户是否有权限查看该简历的权限配置
				if !checkResumeOwnership(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusForbidden, "无权限查看该简历的权限配置", "")
					return
				}

				config := getResumePermissionConfig(sqlDB, resumeID)
				if config == nil {
					// 如果没有权限配置，创建默认配置
					config = createDefaultPermissionConfig(sqlDB, resumeID)
				}

				standardSuccessResponse(c, config, "简历权限配置获取成功")
			})

			// 更新简历权限配置
			permission.PUT("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				// 检查用户是否有权限修改该简历的权限配置
				if !checkResumeOwnership(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusForbidden, "无权限修改该简历的权限配置", "")
					return
				}

				var req UpdatePermissionRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !updateResumePermissionConfig(sqlDB, resumeID, req) {
					standardErrorResponse(c, http.StatusInternalServerError, "更新权限配置失败", "")
					return
				}

				standardSuccessResponse(c, "权限配置已更新", "权限配置更新成功")
			})
		}

		// 简历黑名单管理
		blacklist := api.Group("/resume/blacklist")
		{
			// 获取简历黑名单
			blacklist.GET("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				// 检查用户是否有权限查看该简历的黑名单
				if !checkResumeOwnership(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusForbidden, "无权限查看该简历的黑名单", "")
					return
				}

				blacklist := getResumeBlacklist(sqlDB, resumeID)
				standardSuccessResponse(c, blacklist, "简历黑名单获取成功")
			})

			// 更新简历黑名单
			blacklist.PUT("/:resumeId", func(c *gin.Context) {
				resumeID := c.Param("resumeId")
				userID := c.GetUint("user_id")

				// 检查用户是否有权限修改该简历的黑名单
				if !checkResumeOwnership(sqlDB, resumeID, userID) {
					standardErrorResponse(c, http.StatusForbidden, "无权限修改该简历的黑名单", "")
					return
				}

				var req UpdateBlacklistRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				if !updateResumeBlacklist(sqlDB, resumeID, req) {
					standardErrorResponse(c, http.StatusInternalServerError, "更新黑名单失败", "")
					return
				}

				standardSuccessResponse(c, "黑名单已更新", "黑名单更新成功")
			})
		}

		// 审批管理
		approve := api.Group("/approve")
		{
			// 获取审批列表
			approve.GET("/list", func(c *gin.Context) {
				userID := c.GetUint("user_id")
				page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
				pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
				status := c.Query("status")

				approves := getApproveList(sqlDB, userID, page, pageSize, status)
				total := getApproveCount(sqlDB, userID, status)
				pendingCount := getPendingApproveCount(sqlDB, userID)

				result := gin.H{
					"list":         approves,
					"total":        total,
					"pageNum":      page,
					"pageSize":     pageSize,
					"pendingCount": pendingCount,
				}

				standardSuccessResponse(c, result, "审批列表获取成功")
			})

			// 处理审批
			approve.POST("/handle/:approveId", func(c *gin.Context) {
				approveID := c.Param("approveId")
				userID := c.GetUint("user_id")

				var req HandleApproveRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
					return
				}

				result := handleApprove(sqlDB, approveID, userID, req.Res)
				if result == "" {
					standardErrorResponse(c, http.StatusNotFound, "审批记录不存在或无权限", "")
					return
				}

				standardSuccessResponse(c, result, "审批处理成功")
			})
		}

		// 积分管理
		points := api.Group("/points")
		{
			// 获取用户积分
			points.GET("/user/:userId", func(c *gin.Context) {
				userIDStr := c.Param("userId")
				userID, err := strconv.ParseUint(userIDStr, 10, 64)
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "用户ID格式错误", err.Error())
					return
				}

				// 检查权限：只能查看自己的积分
				currentUserID := c.GetUint("user_id")
				if uint(userID) != currentUserID {
					standardErrorResponse(c, http.StatusForbidden, "无权限查看其他用户的积分", "")
					return
				}

				points := getUserPoints(sqlDB, uint(userID))
				standardSuccessResponse(c, points, "用户积分获取成功")
			})

			// 获取用户积分余额
			points.GET("/user/:userId/balance", func(c *gin.Context) {
				userIDStr := c.Param("userId")
				userID, err := strconv.ParseUint(userIDStr, 10, 64)
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "用户ID格式错误", err.Error())
					return
				}

				// 检查权限：只能查看自己的积分余额
				currentUserID := c.GetUint("user_id")
				if uint(userID) != currentUserID {
					standardErrorResponse(c, http.StatusForbidden, "无权限查看其他用户的积分余额", "")
					return
				}

				balance := getUserPointsBalance(sqlDB, uint(userID))
				result := gin.H{
					"userId":  userIDStr,
					"balance": balance,
				}
				standardSuccessResponse(c, result, "积分余额获取成功")
			})
		}
	}
}

// 数据模型定义
type CreateResumeRequest struct {
	ResumeName string `json:"resumeName" binding:"required"`
	TemplateID string `json:"templateId"`
}

type UpdateResumeRequest struct {
	ResumeName string `json:"resumeName"`
	Content    string `json:"content"`
}

type UpdatePermissionRequest struct {
	PrivacyLevel       string   `json:"privacyLevel"`
	AllowDownload      bool     `json:"allowDownload"`
	RequireApproval    bool     `json:"requireApproval"`
	AllowedEnterprises []string `json:"allowedEnterprises"`
	DeniedEnterprises  []string `json:"deniedEnterprises"`
}

type UpdateBlacklistRequest struct {
	Enterprises []BlacklistEnterprise `json:"enterprises"`
}

type BlacklistEnterprise struct {
	EnterpriseID   string `json:"enterpriseId"`
	EnterpriseName string `json:"enterpriseName"`
	Reason         string `json:"reason"`
}

type HandleApproveRequest struct {
	Res int `json:"res" binding:"required"` // 1-通过，2-拒绝
}

// 业务逻辑函数
func getResumeSummaries(sqlDB *sql.DB, userID uint) []gin.H {
	query := `
		SELECT resume_id, resume_name, resume_status, privacy_level, is_verified, view_count, update_time
		FROM resume 
		WHERE user_id = $1
		ORDER BY update_time DESC
	`
	rows, err := sqlDB.Query(query, userID)
	if err != nil {
		log.Printf("查询简历摘要失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var summaries []gin.H
	for rows.Next() {
		var resumeID, resumeName, resumeStatus, privacyLevel string
		var isVerified bool
		var viewCount int
		var updateTime time.Time

		err := rows.Scan(&resumeID, &resumeName, &resumeStatus, &privacyLevel, &isVerified, &viewCount, &updateTime)
		if err != nil {
			log.Printf("扫描简历摘要失败: %v", err)
			continue
		}

		summary := gin.H{
			"resumeId":     resumeID,
			"resumeName":   resumeName,
			"resumeStatus": resumeStatus,
			"privacyLevel": privacyLevel,
			"isVerified":   isVerified,
			"viewCount":    viewCount,
			"updateTime":   updateTime.UnixMilli(),
		}
		summaries = append(summaries, summary)
	}

	return summaries
}

func getResumeDetail(sqlDB *sql.DB, resumeID string, userID uint) *gin.H {
	query := `
		SELECT resume_id, resume_name, resume_status, privacy_level, user_id, user_name, user_phone, 
		       user_email, is_verified, view_count, download_count, create_time, update_time
		FROM resume 
		WHERE resume_id = $1 AND user_id = $2
	`
	row := sqlDB.QueryRow(query, resumeID, userID)

	var resumeIDResult, resumeName, resumeStatus, privacyLevel, userName, userPhone, userEmail string
	var userIDResult uint
	var isVerified bool
	var viewCount, downloadCount int
	var createTime, updateTime time.Time

	err := row.Scan(&resumeIDResult, &resumeName, &resumeStatus, &privacyLevel, &userIDResult,
		&userName, &userPhone, &userEmail, &isVerified, &viewCount, &downloadCount, &createTime, &updateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Printf("查询简历详情失败: %v", err)
		return nil
	}

	detail := &gin.H{
		"resumeId":      resumeIDResult,
		"resumeName":    resumeName,
		"resumeStatus":  resumeStatus,
		"privacyLevel":  privacyLevel,
		"userId":        strconv.FormatUint(uint64(userIDResult), 10),
		"userName":      userName,
		"userPhone":     userPhone,
		"userEmail":     userEmail,
		"isVerified":    isVerified,
		"viewCount":     viewCount,
		"downloadCount": downloadCount,
		"createTime":    createTime.UnixMilli(),
		"updateTime":    updateTime.UnixMilli(),
	}

	return detail
}

func createResume(sqlDB *sql.DB, userID uint, req CreateResumeRequest) string {
	resumeID := fmt.Sprintf("resume_%d_%d", userID, time.Now().Unix())

	query := `
		INSERT INTO resume (resume_id, resume_name, resume_status, privacy_level, user_id, user_name, 
		                   is_verified, view_count, download_count, create_time, update_time)
		VALUES ($1, $2, 'DRAFT', 'PRIVATE', $3, (SELECT username FROM zervigo_auth_users WHERE id = $3), 
		        false, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := sqlDB.Exec(query, resumeID, req.ResumeName, userID)
	if err != nil {
		log.Printf("创建简历失败: %v", err)
		return ""
	}

	// 创建默认权限配置
	createDefaultPermissionConfig(sqlDB, resumeID)

	return resumeID
}

func updateResume(sqlDB *sql.DB, resumeID string, userID uint, req UpdateResumeRequest) bool {
	query := `
		UPDATE resume 
		SET resume_name = $1, update_time = CURRENT_TIMESTAMP
		WHERE resume_id = $2 AND user_id = $3
	`

	result, err := sqlDB.Exec(query, req.ResumeName, resumeID, userID)
	if err != nil {
		log.Printf("更新简历失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func publishResume(sqlDB *sql.DB, resumeID string, userID uint) bool {
	query := `
		UPDATE resume 
		SET resume_status = 'PUBLISHED', update_time = CURRENT_TIMESTAMP
		WHERE resume_id = $1 AND user_id = $2
	`

	result, err := sqlDB.Exec(query, resumeID, userID)
	if err != nil {
		log.Printf("发布简历失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func deleteResume(sqlDB *sql.DB, resumeID string, userID uint) bool {
	query := `
		DELETE FROM resume 
		WHERE resume_id = $1 AND user_id = $2
	`

	result, err := sqlDB.Exec(query, resumeID, userID)
	if err != nil {
		log.Printf("删除简历失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func checkResumeOwnership(sqlDB *sql.DB, resumeID string, userID uint) bool {
	query := `SELECT 1 FROM resume WHERE resume_id = $1 AND user_id = $2`
	row := sqlDB.QueryRow(query, resumeID, userID)

	var exists int
	err := row.Scan(&exists)
	return err == nil
}

func getResumePermissionConfig(sqlDB *sql.DB, resumeID string) *gin.H {
	query := `
		SELECT resume_id, privacy_level, allow_download, require_approval, 
		       allowed_enterprises, denied_enterprises, update_time
		FROM resume_permission 
		WHERE resume_id = $1
	`
	row := sqlDB.QueryRow(query, resumeID)

	var resumeIDResult, privacyLevel, allowedEnterprises, deniedEnterprises string
	var allowDownload, requireApproval bool
	var updateTime time.Time

	err := row.Scan(&resumeIDResult, &privacyLevel, &allowDownload, &requireApproval,
		&allowedEnterprises, &deniedEnterprises, &updateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Printf("查询简历权限配置失败: %v", err)
		return nil
	}

	// 解析JSON数组
	var allowedList, deniedList []string
	if allowedEnterprises != "" {
		json.Unmarshal([]byte(allowedEnterprises), &allowedList)
	}
	if deniedEnterprises != "" {
		json.Unmarshal([]byte(deniedEnterprises), &deniedList)
	}

	config := &gin.H{
		"resumeId":           resumeIDResult,
		"privacyLevel":       privacyLevel,
		"allowDownload":      allowDownload,
		"requireApproval":    requireApproval,
		"allowedEnterprises": allowedList,
		"deniedEnterprises":  deniedList,
		"updateTime":         updateTime.UnixMilli(),
	}

	return config
}

func createDefaultPermissionConfig(sqlDB *sql.DB, resumeID string) *gin.H {
	query := `
		INSERT INTO resume_permission (resume_id, privacy_level, allow_download, require_approval, 
		                              allowed_enterprises, denied_enterprises, created_at, updated_at)
		VALUES ($1, 'PRIVATE', false, true, '[]', '[]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (resume_id) DO NOTHING
	`

	_, err := sqlDB.Exec(query, resumeID)
	if err != nil {
		log.Printf("创建默认权限配置失败: %v", err)
		return nil
	}

	return &gin.H{
		"resumeId":           resumeID,
		"privacyLevel":       "PRIVATE",
		"allowDownload":      false,
		"requireApproval":    true,
		"allowedEnterprises": []string{},
		"deniedEnterprises":  []string{},
		"updateTime":         time.Now().UnixMilli(),
	}
}

func updateResumePermissionConfig(sqlDB *sql.DB, resumeID string, req UpdatePermissionRequest) bool {
	// 将数组转换为JSON字符串
	allowedJSON, _ := json.Marshal(req.AllowedEnterprises)
	deniedJSON, _ := json.Marshal(req.DeniedEnterprises)

	query := `
		UPDATE resume_permission 
		SET privacy_level = $1, allow_download = $2, require_approval = $3, 
		    allowed_enterprises = $4, denied_enterprises = $5, updated_at = CURRENT_TIMESTAMP
		WHERE resume_id = $6
	`

	result, err := sqlDB.Exec(query, req.PrivacyLevel, req.AllowDownload, req.RequireApproval,
		string(allowedJSON), string(deniedJSON), resumeID)
	if err != nil {
		log.Printf("更新简历权限配置失败: %v", err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0
}

func getResumeBlacklist(sqlDB *sql.DB, resumeID string) []gin.H {
	query := `
		SELECT enterprise_id, enterprise_name, reason, add_time
		FROM resume_blacklist 
		WHERE resume_id = $1
		ORDER BY add_time DESC
	`
	rows, err := sqlDB.Query(query, resumeID)
	if err != nil {
		log.Printf("查询简历黑名单失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var blacklist []gin.H
	for rows.Next() {
		var enterpriseID, enterpriseName, reason string
		var addTime time.Time

		err := rows.Scan(&enterpriseID, &enterpriseName, &reason, &addTime)
		if err != nil {
			log.Printf("扫描简历黑名单失败: %v", err)
			continue
		}

		item := gin.H{
			"enterpriseId":   enterpriseID,
			"enterpriseName": enterpriseName,
			"reason":         reason,
			"addTime":        addTime.UnixMilli(),
		}
		blacklist = append(blacklist, item)
	}

	return blacklist
}

func updateResumeBlacklist(sqlDB *sql.DB, resumeID string, req UpdateBlacklistRequest) bool {
	// 先删除现有黑名单
	deleteQuery := `DELETE FROM resume_blacklist WHERE resume_id = $1`
	_, err := sqlDB.Exec(deleteQuery, resumeID)
	if err != nil {
		log.Printf("删除现有黑名单失败: %v", err)
		return false
	}

	// 插入新的黑名单
	insertQuery := `
		INSERT INTO resume_blacklist (resume_id, enterprise_id, enterprise_name, reason, add_time)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`

	for _, enterprise := range req.Enterprises {
		_, err := sqlDB.Exec(insertQuery, resumeID, enterprise.EnterpriseID,
			enterprise.EnterpriseName, enterprise.Reason)
		if err != nil {
			log.Printf("插入黑名单失败: %v", err)
			return false
		}
	}

	return true
}

func getApproveList(sqlDB *sql.DB, userID uint, page, pageSize int, status string) []gin.H {
	offset := (page - 1) * pageSize

	query := `
		SELECT approve_id, type, enterprise_name, resume_name, status, cost, create_time
		FROM approve_record 
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIndex := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY create_time DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := sqlDB.Query(query, args...)
	if err != nil {
		log.Printf("查询审批列表失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var approves []gin.H
	for rows.Next() {
		var approveID, approveType, enterpriseName, resumeName, approveStatus string
		var cost int
		var createTime time.Time

		err := rows.Scan(&approveID, &approveType, &enterpriseName, &resumeName, &approveStatus, &cost, &createTime)
		if err != nil {
			log.Printf("扫描审批列表失败: %v", err)
			continue
		}

		approve := gin.H{
			"approveId":      approveID,
			"type":           approveType,
			"enterpriseName": enterpriseName,
			"resumeName":     resumeName,
			"status":         approveStatus,
			"cost":           cost,
			"createTime":     createTime.UnixMilli(),
		}
		approves = append(approves, approve)
	}

	return approves
}

func getApproveCount(sqlDB *sql.DB, userID uint, status string) int64 {
	query := `SELECT COUNT(*) FROM approve_record WHERE user_id = $1`
	args := []interface{}{userID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	row := sqlDB.QueryRow(query, args...)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		log.Printf("查询审批数量失败: %v", err)
		return 0
	}

	return count
}

func getPendingApproveCount(sqlDB *sql.DB, userID uint) int64 {
	query := `SELECT COUNT(*) FROM approve_record WHERE user_id = $1 AND status = '待审批'`
	row := sqlDB.QueryRow(query, userID)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		log.Printf("查询待审批数量失败: %v", err)
		return 0
	}

	return count
}

func handleApprove(sqlDB *sql.DB, approveID string, userID uint, res int) string {
	var status string
	if res == 1 {
		status = "已通过"
	} else {
		status = "已拒绝"
	}

	query := `
		UPDATE approve_record 
		SET status = $1, handle_time = CURRENT_TIMESTAMP
		WHERE approve_id = $2 AND user_id = $3
	`

	result, err := sqlDB.Exec(query, status, approveID, userID)
	if err != nil {
		log.Printf("处理审批失败: %v", err)
		return ""
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ""
	}

	return status
}

func getUserPoints(sqlDB *sql.DB, userID uint) []gin.H {
	query := `
		SELECT type, amount, reason, balance, create_time
		FROM points_bill 
		WHERE user_id = $1
		ORDER BY create_time DESC
	`
	rows, err := sqlDB.Query(query, userID)
	if err != nil {
		log.Printf("查询用户积分失败: %v", err)
		return []gin.H{}
	}
	defer rows.Close()

	var points []gin.H
	for rows.Next() {
		var pointType, reason string
		var amount, balance int
		var createTime time.Time

		err := rows.Scan(&pointType, &amount, &reason, &balance, &createTime)
		if err != nil {
			log.Printf("扫描用户积分失败: %v", err)
			continue
		}

		point := gin.H{
			"type":       pointType,
			"amount":     amount,
			"reason":     reason,
			"balance":    balance,
			"createTime": createTime.UnixMilli(),
		}
		points = append(points, point)
	}

	return points
}

func getUserPointsBalance(sqlDB *sql.DB, userID uint) int {
	query := `
		SELECT COALESCE(SUM(CASE WHEN type = '收入' THEN amount ELSE -amount END), 0) as balance
		FROM points_bill 
		WHERE user_id = $1
	`
	row := sqlDB.QueryRow(query, userID)
	var balance int
	err := row.Scan(&balance)
	if err != nil {
		log.Printf("查询用户积分余额失败: %v", err)
		return 0
	}

	return balance
}

// 辅助函数
func registerToConsul(serviceName, serviceHost string, servicePort int) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Tags:    []string{"resume", "privacy", "permission", "jobfirst", "microservice", "version:3.1.0"},
		Port:    servicePort,
		Address: serviceHost,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Timeout:                        "3s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("%s registered with Consul successfully", serviceName)
	}
}

func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	msg := "操作成功"
	if len(message) > 0 {
		msg = message[0]
	}
	resp := response.Success(msg, data)
	c.JSON(http.StatusOK, resp)
}

func standardErrorResponse(c *gin.Context, statusCode int, message string, details ...string) {
	code := response.CodeInternalError
	switch statusCode {
	case http.StatusBadRequest:
		code = response.CodeInvalidParams
	case http.StatusUnauthorized:
		code = response.CodeUnauthorized
	case http.StatusForbidden:
		code = response.CodeForbidden
	case http.StatusNotFound:
		code = response.CodeNotFound
	}

	resp := response.Error(code, message)
	c.JSON(http.StatusOK, resp) // 使用200状态码，错误信息在响应体中
}
