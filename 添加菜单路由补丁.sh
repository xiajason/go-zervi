#!/bin/bash

# 快速补丁：在 Central Brain 的 main.go 中添加菜单路由

MAIN_FILE="/Users/szjason72/gozervi/zervigo.demo/shared/central-brain/main.go"
BACKUP_FILE="/Users/szjason72/gozervi/zervigo.demo/shared/central-brain/main.go.backup.$(date +%Y%m%d%H%M%S)"

echo "=== 备份 main.go ==="
cp "$MAIN_FILE" "$BACKUP_FILE"
echo "✅ 已备份到: $BACKUP_FILE"

echo ""
echo "=== 添加菜单路由注册代码 ==="

# 在 main 函数中找到适当位置添加代码
# 这需要手动编辑，因为自动化可能会有风险

cat << 'EOF'

请手动在 main.go 中添加以下代码：

1. 在 import 部分确保包含：
   "database/sql"

2. 在路由注册部分（通常在 registerRoutes 或类似函数中）添加：

// 菜单路由
if vuecmfDB != nil {
    RegisterMenuRoutes(r, vuecmfDB)
} else {
    // 临时静态菜单
    r.GET("/api/v1/router/menu/nav", func(c *gin.Context) {
        menus := []map[string]interface{}{
            {"id": 1, "pid": 0, "model_id": 1, "title": "首页", "path": "/dashboard", "icon": "HomeFilled", "sort_num": 1, "status": 10},
            {"id": 2, "pid": 0, "model_id": 2, "title": "系统管理", "path": "/system", "icon": "Setting", "sort_num": 2, "status": 10},
        }
        c.JSON(200, gin.H{"code": 200, "msg": "success", "data": menus})
    })
}

3. 保存文件

4. 重启 Central Brain：
   cd /Users/szjason72/gozervi/zervigo.demo/shared/central-brain
   go run .

EOF

echo""
echo "✅ 补丁说明已显示"
echo "📝 请按照上述说明手动编辑 main.go"

