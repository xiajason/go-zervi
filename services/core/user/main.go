package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	jobfirst "github.com/szjason72/zervigo/shared/core"
	"github.com/szjason72/zervigo/shared/core/auth"
	"github.com/szjason72/zervigo/shared/core/response"
	"github.com/szjason72/zervigo/shared/core/service"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "user-service"
	}

	// 初始化JobFirst核心包 - 使用硬编码配置避免YAML读取问题
	core, err := jobfirst.NewCore("")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// auth-service URL
	authServiceURL := "http://localhost:8207"
	log.Printf("初始化集中式认证 - auth-service URL: %s", authServiceURL)

	// 创建认证服务客户端
	authClient := auth.NewAuthClient(authServiceURL)
	log.Println("认证客户端已创建")

	// 执行服务握手（获取Service Token用于内部服务间通信）
	log.Println("执行服务握手...")
	handshakeConfig := service.ServiceHandshake{
		ServiceID:       "user-service",
		ServiceName:     "User Service",
		ServiceSecret:   "userServiceSecret2025",
		ServiceType:     "core",
		CentralBrainURL: "http://localhost:9000",
		AuthServiceURL:  authServiceURL,
		Timeout:         10 * time.Second,
	}

	handshakeResult, err := service.Handshake(&handshakeConfig)
	if err != nil {
		log.Printf("警告: 服务握手失败: %v", err)
	} else if handshakeResult.Success {
		log.Printf("✅ 服务握手成功，Service Token: %s",
			func() string {
				if len(handshakeResult.ServiceToken) > 50 {
					return handshakeResult.ServiceToken[:50] + "..."
				}
				return handshakeResult.ServiceToken
			}())
	} else {
		log.Printf("警告: 服务握手失败: %s", handshakeResult.Error)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 设置标准路由 (使用jobfirst-core统一模板)
	setupStandardRoutes(r, core)

	// 设置业务路由 (保持现有API)
	setupBusinessRoutes(r, core, authClient)

	// 注册到Consul
	registerToConsul("user-service", "127.0.0.1", 8082)

	// 启动服务器
	log.Println("Starting User Service with jobfirst-core on 0.0.0.0:8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由 (使用jobfirst-core统一模板)
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查 (统一格式)
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "user-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "user-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "user-service",
			"version":    "3.1.0",
			"port":       8082,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core, authClient *auth.AuthClient) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 用户注册
		public.POST("/auth/register", func(c *gin.Context) {
			var req auth.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			// 使用核心包的认证管理器
			response, err := core.AuthManager.Register(req)
			if err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Registration failed", err.Error())
				return
			}

			standardSuccessResponse(c, response, "User registered successfully")
		})

		// 用户登录
		public.POST("/auth/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			// 使用核心包的认证管理器
			clientIP := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")
			response, err := core.AuthManager.Login(req, clientIP, userAgent)
			if err != nil {
				standardErrorResponse(c, http.StatusUnauthorized, "Login failed", err.Error())
				return
			}

			standardSuccessResponse(c, response, "Login successful")
		})

		// 刷新Token
		public.POST("/auth/refresh", func(c *gin.Context) {
			standardSuccessResponse(c, gin.H{"message": "Token refresh functionality"}, "Token refresh endpoint")
		})

		// 用户登出
		public.POST("/auth/logout", func(c *gin.Context) {
			standardSuccessResponse(c, gin.H{"message": "Logout successful"}, "Logout successful")
		})
	}

	// 需要认证的API路由 - 使用集中式认证（调用auth-service）
	authMiddleware := createCentralizedAuthMiddleware(authClient)
	api := r.Group("/api/v1")
	api.Use(authMiddleware)
	{
		// 用户管理API
		users := api.Group("/users")
		{
			// 获取用户资料 - 使用数据库中的最新信息
			users.GET("/profile", func(c *gin.Context) {
				userID := c.GetInt("user_id")
				role := c.GetString("role")

				if userID == 0 {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}

				db := core.GetDB()
				var userModel auth.User
				if err := db.First(&userModel, userID).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to load user profile", err.Error())
					return
				}

				standardSuccessResponse(c, buildUserProfileResponse(userModel, role), "User profile retrieved successfully")
			})

			// 更新用户资料（兼容前端 /api/user/profile）
			users.PUT("/profile", func(c *gin.Context) {
				userID := c.GetInt("user_id")
				role := c.GetString("role")

				if userID == 0 {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}

				var payload userProfileUpdate
				if err := c.ShouldBindJSON(&payload); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				db := core.GetDB()
				var userModel auth.User
				if err := db.First(&userModel, userID).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to load user profile", err.Error())
					return
				}

				updated := false

				if name := strings.TrimSpace(payload.Name); name != "" && name != userModel.Username {
					userModel.Username = name
					updated = true
				}

				if email := strings.TrimSpace(payload.Email); email != "" && email != userModel.Email {
					userModel.Email = email
					updated = true
				}

				if phone := strings.TrimSpace(payload.Phone); phone != "" {
					userModel.Phone = &phone
					updated = true
				}

				if updated {
					userModel.UpdatedAt = time.Now()
					if err := db.Save(&userModel).Error; err != nil {
						standardErrorResponse(c, http.StatusInternalServerError, "Update failed", err.Error())
						return
					}
				}

				standardSuccessResponse(c, buildUserProfileResponse(userModel, role), "User profile updated successfully")
			})

			// 模拟头像上传接口，返回可用的占位链接
			users.POST("/avatar", func(c *gin.Context) {
				userID := c.GetInt("user_id")
				if userID == 0 {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}

				var req struct {
					FilePath string `json:"filePath"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				fileName := filepath.Base(strings.TrimSpace(req.FilePath))
				if fileName == "" || fileName == "." {
					fileName = fmt.Sprintf("user-%d.png", userID)
				}

				avatarURL := fmt.Sprintf("https://static.local.zervigo/avatar/%s", fileName)
				standardSuccessResponse(c, gin.H{
					"avatar": avatarURL,
				}, "Avatar uploaded successfully (mock)")
			})

			// 修改密码
			users.PUT("/password", func(c *gin.Context) {
				standardSuccessResponse(c, gin.H{"message": "Password change functionality"}, "Password change endpoint")
			})

			// 获取用户列表（管理员功能）
			users.GET("/", func(c *gin.Context) {
				// 检查管理员权限
				role := c.GetString("role")
				if role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var users []auth.User
				offset := (page - 1) * pageSize

				if err := db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get user list", err.Error())
					return
				}

				var total int64
				db.Model(&auth.User{}).Count(&total)

				pageResp := response.NewPageResponse(users, total, page, pageSize)
				standardSuccessResponse(c, pageResp, "User list retrieved successfully")
			})

			// 获取单个用户（管理员功能）
			users.GET("/:id", func(c *gin.Context) {
				// 检查管理员权限
				role := c.GetString("role")
				if role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				userID, _ := strconv.Atoi(c.Param("id"))

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var user auth.User
				if err := db.First(&user, userID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
					return
				}

				standardSuccessResponse(c, user, "User retrieved successfully")
			})

			// 更新用户（管理员功能）
			users.PUT("/:id", func(c *gin.Context) {
				// 检查管理员权限
				role := c.GetString("role")
				if role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				userID, _ := strconv.Atoi(c.Param("id"))

				var updateData auth.User
				if err := c.ShouldBindJSON(&updateData); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				updateData.ID = uint(userID)
				updateData.UpdatedAt = time.Now()

				if err := db.Save(&updateData).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Update failed", err.Error())
					return
				}

				standardSuccessResponse(c, updateData, "User updated successfully")
			})

			// 删除用户（管理员功能）
			users.DELETE("/:id", func(c *gin.Context) {
				// 检查管理员权限
				role := c.GetString("role")
				if role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				userID, _ := strconv.Atoi(c.Param("id"))

				// 使用核心包的数据库管理器
				db := core.GetDB()
				if err := db.Delete(&auth.User{}, userID).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Delete failed", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{"deleted": true}, "User deleted successfully")
			})
		}

		// 角色管理
		roles := api.Group("/roles")
		{
			roles.GET("/", func(c *gin.Context) {
				// 使用核心包的数据库管理器
				db := core.GetDB()
				var roles []Role
				if err := db.Where("is_active = ?", true).Find(&roles).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get roles", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{
					"roles": roles,
					"total": len(roles),
				}, "Roles retrieved successfully")
			})

			// 获取单个角色
			roles.GET("/:id", func(c *gin.Context) {
				roleID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err.Error())
					return
				}

				db := core.GetDB()
				var role Role
				if err := db.First(&role, roleID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Role not found", err.Error())
					return
				}

				standardSuccessResponse(c, role, "Role retrieved successfully")
			})

			// 创建角色（管理员功能）
			roles.POST("/", func(c *gin.Context) {
				// 检查管理员权限
				role := c.GetString("role")
				if role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				var roleData Role
				if err := c.ShouldBindJSON(&roleData); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				db := core.GetDB()
				roleData.CreatedAt = time.Now()
				roleData.UpdatedAt = time.Now()

				if err := db.Create(&roleData).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create role", err.Error())
					return
				}

				standardSuccessResponse(c, roleData, "Role created successfully")
			})
		}

		// 权限管理
		permissions := api.Group("/permissions")
		{
			permissions.GET("/", func(c *gin.Context) {
				// 使用核心包的数据库管理器
				db := core.GetDB()
				var permissions []Permission
				if err := db.Where("is_active = ?", true).Find(&permissions).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get permissions", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{
					"permissions": permissions,
					"total":       len(permissions),
				}, "Permissions retrieved successfully")
			})

			// 获取单个权限
			permissions.GET("/:id", func(c *gin.Context) {
				permissionID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid permission ID", err.Error())
					return
				}

				db := core.GetDB()
				var permission Permission
				if err := db.First(&permission, permissionID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Permission not found", err.Error())
					return
				}

				standardSuccessResponse(c, permission, "Permission retrieved successfully")
			})

			// 创建权限（管理员功能）
			permissions.POST("/", func(c *gin.Context) {
				// 检查管理员权限
				role := c.GetString("role")
				if role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				var permissionData Permission
				if err := c.ShouldBindJSON(&permissionData); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				db := core.GetDB()
				permissionData.CreatedAt = time.Now()
				permissionData.UpdatedAt = time.Now()

				if err := db.Create(&permissionData).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create permission", err.Error())
					return
				}

				standardSuccessResponse(c, permissionData, "Permission created successfully")
			})
		}

		// 简历权限管理API
		resumePermissions := api.Group("/resume-permissions")
		{
			// 获取简历权限配置
			resumePermissions.GET("/:resume_id", func(c *gin.Context) {
				resumeID := c.Param("resume_id")

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var permissions []ResumePermissionConfig
				if err := db.Where("resume_id = ?", resumeID).Find(&permissions).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get permission config", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{
					"permissions": permissions,
					"resume_id":   resumeID,
					"total":       len(permissions),
				}, "Resume permissions retrieved successfully")
			})

			// 创建简历权限配置
			resumePermissions.POST("/", func(c *gin.Context) {
				var config ResumePermissionConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				config.CreatedAt = time.Now()
				config.UpdatedAt = time.Now()

				if err := db.Create(&config).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create permission config", err.Error())
					return
				}

				standardSuccessResponse(c, config, "Resume permission config created successfully")
			})
		}

		// 利益相关方管理API
		stakeholders := api.Group("/stakeholders")
		{
			// 获取利益相关方列表
			stakeholders.GET("/", func(c *gin.Context) {
				// 使用核心包的数据库管理器
				db := core.GetDB()
				var stakeholders []Stakeholder
				if err := db.Find(&stakeholders).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get stakeholders", err.Error())
					return
				}

				standardSuccessResponse(c, stakeholders, "Stakeholders retrieved successfully")
			})

			// 创建利益相关方
			stakeholders.POST("/", func(c *gin.Context) {
				var stakeholder Stakeholder
				if err := c.ShouldBindJSON(&stakeholder); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				stakeholder.CreatedAt = time.Now()
				stakeholder.UpdatedAt = time.Now()

				if err := db.Create(&stakeholder).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create stakeholder", err.Error())
					return
				}

				standardSuccessResponse(c, stakeholder, "Stakeholder created successfully")
			})
		}

		// 评论管理API
		comments := api.Group("/comments")
		{
			// 获取简历评论
			comments.GET("/resume/:resume_id", func(c *gin.Context) {
				resumeID := c.Param("resume_id")

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var comments []Comment
				if err := db.Where("resume_id = ?", resumeID).Find(&comments).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get comments", err.Error())
					return
				}

				standardSuccessResponse(c, comments, "Comments retrieved successfully")
			})

			// 创建评论
			comments.POST("/", func(c *gin.Context) {
				var comment Comment
				if err := c.ShouldBindJSON(&comment); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				comment.CreatedAt = time.Now()
				comment.UpdatedAt = time.Now()

				if err := db.Create(&comment).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create comment", err.Error())
					return
				}

				standardSuccessResponse(c, comment, "Comment created successfully")
			})
		}

		// 分享管理API
		shares := api.Group("/shares")
		{
			// 获取简历分享
			shares.GET("/resume/:resume_id", func(c *gin.Context) {
				resumeID := c.Param("resume_id")

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var shares []Share
				if err := db.Where("resume_id = ?", resumeID).Find(&shares).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get shares", err.Error())
					return
				}

				standardSuccessResponse(c, shares, "Shares retrieved successfully")
			})

			// 创建分享
			shares.POST("/", func(c *gin.Context) {
				var share Share
				if err := c.ShouldBindJSON(&share); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				share.CreatedAt = time.Now()
				share.UpdatedAt = time.Now()

				if err := db.Create(&share).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create share", err.Error())
					return
				}

				standardSuccessResponse(c, share, "Share created successfully")
			})
		}

		// 积分管理API
		points := api.Group("/points")
		{
			// 获取用户积分
			points.GET("/user/:user_id", func(c *gin.Context) {
				userID := c.Param("user_id")

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var points []Points
				if err := db.Where("user_id = ?", userID).Find(&points).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get points", err.Error())
					return
				}

				standardSuccessResponse(c, points, "User points retrieved successfully")
			})

			// 获取用户积分余额
			points.GET("/user/:user_id/balance", func(c *gin.Context) {
				userID := c.Param("user_id")

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var balance int
				if err := db.Model(&Points{}).Where("user_id = ?", userID).Select("COALESCE(SUM(points), 0)").Scan(&balance).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get points balance", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{
					"user_id": userID,
					"balance": balance,
				}, "Points balance retrieved successfully")
			})

			// 奖励积分
			points.POST("/award", func(c *gin.Context) {
				var points Points
				if err := c.ShouldBindJSON(&points); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				points.CreatedAt = time.Now()
				points.UpdatedAt = time.Now()

				if err := db.Create(&points).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to award points", err.Error())
					return
				}

				standardSuccessResponse(c, points, "Points awarded successfully")
			})
		}
	}
}

type userProfileUpdate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func buildUserProfileResponse(user auth.User, role string) gin.H {
	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	return gin.H{
		"id":        fmt.Sprintf("%d", user.ID),
		"name":      user.Username,
		"email":     user.Email,
		"phone":     phone,
		"role":      role,
		"status":    user.Status,
		"createdAt": user.CreatedAt.Format(time.RFC3339),
		"updatedAt": user.UpdatedAt.Format(time.RFC3339),
	}
}

// registerToConsul 注册服务到Consul
func registerToConsul(serviceName, serviceHost string, servicePort int) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Tags:    []string{"user", "auth", "rbac", "jobfirst", "microservice", "version:3.1.0"},
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

// standardSuccessResponse 标准成功响应
func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	msg := "操作成功"
	if len(message) > 0 {
		msg = message[0]
	}
	resp := response.Success(msg, data)
	c.JSON(http.StatusOK, resp)
}

// standardErrorResponse 标准错误响应
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

// 数据模型定义
type ResumePermissionConfig struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	UserID       int       `json:"user_id" gorm:"not null"`
	ResumeID     int       `json:"resume_id" gorm:"not null"`
	PermissionID int       `json:"permission_id" gorm:"not null"`
	RoleName     string    `json:"role_name" gorm:"size:50"`
	IsGranted    bool      `json:"is_granted" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Stakeholder struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	UserID      int       `json:"user_id" gorm:"not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Type        string    `json:"type" gorm:"size:50"` // 猎头顾问、职业技能评价机构、简历模板提供商、教育经历见证人、职业经历见证人等
	Description string    `json:"description" gorm:"type:text"`
	ContactInfo string    `json:"contact_info" gorm:"size:200"`
	Status      string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Comment struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null"`
	ResumeID  int       `json:"resume_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Points    int       `json:"points" gorm:"default:0"` // 积分奖励
	Status    string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Share struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null"`
	ResumeID  int       `json:"resume_id" gorm:"not null"`
	Platform  string    `json:"platform" gorm:"size:50"` // 分享平台
	Points    int       `json:"points" gorm:"default:0"` // 积分奖励
	Status    string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Points struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	UserID      int       `json:"user_id" gorm:"not null"`
	Points      int       `json:"points" gorm:"not null"`
	Type        string    `json:"type" gorm:"size:50"` // 积分类型：comment, share, create_resume等
	Description string    `json:"description" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role 角色模型
type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:50;uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Level       uint8     `json:"level" gorm:"default:1"`
	PID         uint      `json:"pid" gorm:"default:0"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Resource    string    `json:"resource" gorm:"size:100;not null"`
	Action      string    `json:"action" gorm:"size:50;not null"`
	Level       uint8     `json:"level" gorm:"default:1"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// createCentralizedAuthMiddleware 创建集中式认证中间件（调用auth-service）
func createCentralizedAuthMiddleware(authClient *auth.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: 集中式认证中间件 - 开始处理请求: %s %s\n", c.Request.Method, c.Request.URL.Path)
		// 打印原始Authorization头
		rawAuth := c.Request.Header.Get("Authorization")
		fmt.Printf("DEBUG: 中间件收到的Authorization: %s\n", func() string {
			if rawAuth == "" {
				return "<empty>"
			}
			if len(rawAuth) > 60 {
				return rawAuth[:60] + "..."
			}
			return rawAuth
		}())

		// 提取token
		token := extractTokenFromRequest(c)
		fmt.Printf("DEBUG: 集中式认证中间件 - 提取到的token: %s\n", func() string {
			if token == "" {
				return "空token"
			}
			if len(token) > 50 {
				return token[:50] + "..."
			}
			return token
		}())

		if token == "" {
			fmt.Printf("DEBUG: 集中式认证中间件 - token为空，返回未登录\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		// 调用auth-service验证token
		fmt.Printf("DEBUG: 集中式认证中间件 - 调用auth-service验证token\n")
		result, err := authClient.ValidateToken(token)
		if err != nil {
			fmt.Printf("DEBUG: 集中式认证中间件 - auth-service调用失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "认证服务不可用",
			})
			c.Abort()
			return
		}

		if !result.Success {
			fmt.Printf("DEBUG: 集中式认证中间件 - token验证失败: %s\n", result.Error)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   result.Error,
			})
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: 集中式认证中间件 - token验证成功，用户ID: %d, 用户名: %s, 角色: %s\n",
			result.User.ID, result.User.Username, result.User.Role)

		// 设置用户信息到上下文
		c.Set("user_id", result.User.ID)
		c.Set("username", result.User.Username)
		c.Set("role", result.User.Role)
		c.Set("email", result.User.Email)

		fmt.Printf("DEBUG: 集中式认证中间件 - 用户信息已设置到上下文，继续处理请求\n")
		c.Next()
	}
}

// extractTokenFromRequest 从请求中提取token
func extractTokenFromRequest(c *gin.Context) string {
	// 从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 兼容大小写并按空格分割
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
			return fields[1]
		}
	}

	// 从查询参数获取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从Cookie获取
	cookie, err := c.Cookie("token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}
