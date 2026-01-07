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

	_ "github.com/lib/pq"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "rls-demo-service"
	}

	// 连接PostgreSQL数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=szjason72 dbname=zervigo_mvp sslmode=disable")
	if err != nil {
		log.Fatalf("连接PostgreSQL失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("PostgreSQL连接测试失败: %v", err)
	}

	// 创建HTTP路由器
	mux := http.NewServeMux()

	// 设置路由
	setupRoutes(mux, db)

	// 启动服务
	log.Println("Starting RLS Demo Service on 0.0.0.0:8088")
	if err := http.ListenAndServe(":8088", mux); err != nil {
		log.Fatalf("RLS Demo Service启动失败: %v", err)
	}
}

// setupRoutes 设置路由
func setupRoutes(mux *http.ServeMux, db *sql.DB) {
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service":   "rls-demo-service",
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "3.1.0",
		}
		json.NewEncoder(w).Encode(response)
	})

	// 版本信息
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "rls-demo-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 测试RLS策略
	mux.HandleFunc("/api/v1/rls/test-policies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		policies := testRLSPolicies(db)
		response := map[string]interface{}{
			"code":      0,
			"message":   "RLS策略测试完成",
			"data":      policies,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 测试用户权限
	mux.HandleFunc("/api/v1/rls/test-permissions/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 从URL路径中提取用户ID
		path := r.URL.Path
		userIDStr := path[len("/api/v1/rls/test-permissions/"):]
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			response := map[string]interface{}{
				"code":      10001,
				"message":   "无效的用户ID",
				"timestamp": time.Now().UnixMilli(),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		permissions := testUserPermissions(db, uint(userID))
		response := map[string]interface{}{
			"code":      0,
			"message":   "用户权限测试完成",
			"data":      permissions,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})

	// 演示RLS效果
	mux.HandleFunc("/api/v1/rls/demo/rls-effect/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 从URL路径中提取用户ID
		path := r.URL.Path
		userIDStr := path[len("/api/v1/rls/demo/rls-effect/"):]
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			response := map[string]interface{}{
				"code":      10001,
				"message":   "无效的用户ID",
				"timestamp": time.Now().UnixMilli(),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		results := demonstrateRLSEffect(db, uint(userID))
		response := map[string]interface{}{
			"code":      0,
			"message":   "RLS效果演示完成",
			"data":      results,
			"timestamp": time.Now().UnixMilli(),
		}
		json.NewEncoder(w).Encode(response)
	})
}

// testRLSPolicies 测试RLS策略
func testRLSPolicies(db *sql.DB) map[string]interface{} {
	// 查询RLS策略状态
	query := `
		SELECT schemaname, tablename, policyname, permissive
		FROM pg_policies 
		WHERE schemaname = 'public'
		ORDER BY tablename, policyname
	`

	rows, err := db.Query(query)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	defer rows.Close()

	var policies []map[string]interface{}
	for rows.Next() {
		var schemaName, tableName, policyName string
		var permissive bool

		if err := rows.Scan(&schemaName, &tableName, &policyName, &permissive); err != nil {
			continue
		}

		policies = append(policies, map[string]interface{}{
			"schema":     schemaName,
			"table":      tableName,
			"policy":     policyName,
			"permissive": permissive,
		})
	}

	return map[string]interface{}{
		"total_policies": len(policies),
		"policies":       policies,
		"status":         "active",
	}
}

// testUserPermissions 测试用户权限
func testUserPermissions(db *sql.DB, userID uint) map[string]interface{} {
	// 设置用户上下文
	if err := setUserContext(db, userID, "test_user"); err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 测试权限检查函数
	var hasResumeView bool
	err := db.QueryRow("SELECT has_permission('resume:view')").Scan(&hasResumeView)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	var hasUserRole bool
	err = db.QueryRow("SELECT has_role('user')").Scan(&hasUserRole)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 获取用户权限列表
	permissionsQuery := `
		SELECT p.permission_code, p.permission_name
		FROM zervigo_auth_user_roles ur
		JOIN zervigo_auth_role_permissions rp ON ur.role_id = rp.role_id
		JOIN zervigo_auth_permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1
		ORDER BY p.permission_code
	`

	rows, err := db.Query(permissionsQuery, userID)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	defer rows.Close()

	var permissions []map[string]interface{}
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			continue
		}
		permissions = append(permissions, map[string]interface{}{
			"code": code,
			"name": name,
		})
	}

	return map[string]interface{}{
		"user_id":          userID,
		"has_resume_view":  hasResumeView,
		"has_user_role":    hasUserRole,
		"permissions":      permissions,
		"permission_count": len(permissions),
	}
}

// setUserContext 设置用户上下文
func setUserContext(db *sql.DB, userID uint, username string) error {
	query := "SELECT set_user_context($1, $2)"
	_, err := db.Exec(query, userID, username)
	return err
}

// demonstrateRLSEffect 演示RLS效果
func demonstrateRLSEffect(db *sql.DB, userID uint) map[string]interface{} {
	results := make(map[string]interface{})

	// 设置用户上下文
	if err := setUserContext(db, userID, "test_user"); err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// 查询不同表的数据量（受RLS影响）
	tables := []string{
		"zervigo_auth_users",
		"resume_permission",
		"approve_record",
		"points_bill",
		"view_history",
	}

	tableCounts := make(map[string]int)
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			tableCounts[table] = -1 // 表示查询失败
		} else {
			tableCounts[table] = count
		}
	}

	results["user_id"] = userID
	results["table_counts"] = tableCounts
	results["rls_active"] = true
	results["note"] = "数据量受RLS策略影响，不同用户看到的数据可能不同"

	return results
}
