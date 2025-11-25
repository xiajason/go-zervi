# VueCMF 字段完整对照表

## 📋 前端期望的数据结构

### 1. 登录响应存储

```javascript
// LoginService.ts:53-55
localStorage.setItem('vuecmf_token', res.data.data.token)
localStorage.setItem('vuecmf_user', JSON.stringify(res.data.data.user))
localStorage.setItem('vuecmf_server', JSON.stringify(res.data.data.server))
```

### 2. Welcome 页面使用的字段

#### 用户信息 (user_info)
```vue
<!-- Welcome.vue -->
{{ user_info.username }}         <!-- 登录账号 -->
{{ user_info.role }}             <!-- 角色 -->
{{ user_info.last_login_ip }}    <!-- 最后登录IP -->
{{ user_info.last_login_time }}  <!-- 最后登录时间 -->
```

#### 服务器信息 (server_info)
```vue
<!-- Welcome.vue -->
{{ server_info.version }}                    <!-- VueCMF版本 -->
{{ server_info.os }} {{ server_info.software }}  <!-- 服务器运行环境 -->
mysql {{ server_info.mysql }}                <!-- 服务器数据库 -->
{{ server_info.upload_max_size }}            <!-- 最大上传文件大小 -->
```

---

## 🗄️ 数据库字段

### zervigo_auth_users 表
```sql
- id
- username
- email
- phone
- password_hash
- status
- email_verified
- phone_verified
- subscription_status
- subscription_type
- subscription_expires_at
- last_login_at          -- 最后登录时间
- created_at
- updated_at
```

**注意**：数据库中**没有** `last_login_ip` 字段，需要从请求中获取！

---

## 🔄 完整字段映射表

### 登录响应必需字段

| 前端字段路径 | 后端返回字段 | 数据来源 | 状态 |
|-------------|-------------|---------|------|
| **Token** |||
| (root) | `token` | JWT生成 | ✅ |
| **User 对象** |||
| user.id | `user.id` | DB: id | ✅ |
| user.username | `user.username` | DB: username | ✅ |
| user.email | `user.email` | DB: email | ✅ |
| user.phone | `user.phone` | DB: phone | ✅ |
| user.status | `user.status` | DB: status | ✅ |
| user.role | `user.role` | 查询角色表 | ✅ |
| user.last_login_ip | `user.last_login_ip` | 请求 IP (getClientIP) | ✅ |
| user.last_login_time | `user.last_login_time` | DB: last_login_at | ✅ |
| **Server 对象** |||
| server.name | `server.name` | 配置常量 | ✅ |
| server.version | `server.version` | 配置常量 | ✅ |
| server.os | `server.os` | 系统信息 | ✅ |
| server.software | `server.software` | 软件栈 | ✅ |
| server.mysql | `server.mysql` | PostgreSQL版本 | ✅ |
| server.upload_max_size | `server.upload_max_size` | 配置 | ✅ |

### 菜单响应必需字段

| 前端字段路径 | 后端返回字段 | 数据来源 | 状态 |
|-------------|-------------|---------|------|
| data.nav_menu | `data.nav_menu` | vuecmf_menu表 | ✅ |
| data.api_maps | `data.api_maps` | vuecmf_api_map表 | ✅ |

#### nav_menu 对象结构（以 mid 为 key）
```json
{
  "/": {
    "id": 1,
    "pid": 0,
    "model_id": 1,
    "title": "首页",
    "path": "/dashboard",
    "icon": "HomeFilled",
    "mid": "/",
    "component_tpl": "Index",
    "path_name": ["dashboard-index"],
    "table_name": "dashboard",
    "default_action_type": "index",
    "sort_num": 1,
    "status": 10,
    "children": {...}  // 如果有子菜单
  }
}
```

#### api_maps 对象结构
```json
{
  "menu": {
    "nav": "/api/v1/menu/nav",
    "list": "/api/v1/menu/list"
  },
  "admin": {
    "login": "/api/v1/auth/login",
    "index": "/api/v1/users/list"
  }
}
```

---

## ✅ 当前实现状态

### 登录接口 (/api/v1/auth/login)

#### ✅ 完整返回（已修复）
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbG...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@zervigo.com",
      "phone": null,
      "status": "active",
      "role": "super_admin",
      "last_login_ip": "[::1]:51151",
      "last_login_time": "2025-11-05 13:56:51"
    },
    "server": {
      "name": "Zervigo MVP",
      "version": "1.0.0",
      "os": "macOS (darwin)",
      "software": "Go + Gin",
      "mysql": "PostgreSQL 14.19",
      "upload_max_size": "10MB"
    }
  }
}
```

### 菜单接口 (/api/v1/menu/nav)

#### ✅ 完整返回（已修复）
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "nav_menu": {
      "/": {...},
      "/system": {...},
      "/jobs": {...},
      ...
    },
    "api_maps": {
      "menu": {"nav": "/api/v1/menu/nav", "list": "/api/v1/menu/list"},
      "admin": {"login": "/api/v1/auth/login", "index": "/api/v1/users/list"},
      ...
    }
  }
}
```

---

## 🧪 完整验证测试

**请在浏览器控制台执行以下完整验证：**

```javascript
console.clear();
console.log('='.repeat(60));
console.log('📊 VueCMF 字段完整验证');
console.log('='.repeat(60));

// 1. 测试登录并检查所有字段
fetch('http://localhost:9000/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    data: {
      login_name: 'admin',
      password: 'admin123'
    }
  })
})
.then(res => res.json())
.then(loginData => {
  console.log('\n1️⃣ 登录响应验证\n' + '-'.repeat(40));
  
  // 检查顶层字段
  console.log('✓ code:', loginData.code === 0 ? '✅ 0' : '❌ ' + loginData.code);
  console.log('✓ message:', loginData.message);
  
  // 检查 data.token
  console.log('\n Token:');
  console.log('✓ token:', loginData.data.token ? '✅ 存在' : '❌ 缺失');
  
  // 检查 data.user 所有字段
  console.log('\n User 对象:');
  const userFields = ['id', 'username', 'email', 'phone', 'status', 'role', 'last_login_ip', 'last_login_time'];
  userFields.forEach(field => {
    const value = loginData.data.user[field];
    const status = value !== undefined && value !== null && value !== '' ? '✅' : '⚠️';
    console.log(`${status} user.${field}:`, value);
  });
  
  // 检查 data.server 所有字段
  console.log('\n Server 对象:');
  const serverFields = ['name', 'version', 'os', 'software', 'mysql', 'upload_max_size'];
  serverFields.forEach(field => {
    const value = loginData.data.server[field];
    const status = value !== undefined && value !== null && value !== '' ? '✅' : '⚠️';
    console.log(`${status} server.${field}:`, value);
  });
  
  // 测试存储
  console.log('\n' + '-'.repeat(40));
  console.log('存储到 LocalStorage...');
  localStorage.setItem('vuecmf_token', loginData.data.token);
  localStorage.setItem('vuecmf_user', JSON.stringify(loginData.data.user));
  localStorage.setItem('vuecmf_server', JSON.stringify(loginData.data.server));
  console.log('✅ 已存储');
  
  // 验证读取
  console.log('\n从 LocalStorage 读取验证:');
  const storedUser = JSON.parse(localStorage.getItem('vuecmf_user'));
  const storedServer = JSON.parse(localStorage.getItem('vuecmf_server'));
  
  console.log('user.role:', storedUser.role);
  console.log('user.last_login_ip:', storedUser.last_login_ip);
  console.log('user.last_login_time:', storedUser.last_login_time);
  console.log('server.os:', storedServer.os);
  console.log('server.mysql:', storedServer.mysql);
  console.log('server.upload_max_size:', storedServer.upload_max_size);
  
  // 测试菜单
  return fetch('http://localhost:9000/api/v1/menu/nav', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ data: { username: 'admin' } })
  });
})
.then(res => res.json())
.then(menuData => {
  console.log('\n2️⃣ 菜单响应验证\n' + '-'.repeat(40));
  console.log('✓ code:', menuData.code === 0 ? '✅ 0' : '❌ ' + menuData.code);
  console.log('✓ nav_menu:', menuData.data.nav_menu ? '✅ 存在' : '❌ 缺失');
  console.log('✓ api_maps:', menuData.data.api_maps ? '✅ 存在' : '❌ 缺失');
  
  if (menuData.data.nav_menu) {
    const menuKeys = Object.keys(menuData.data.nav_menu);
    console.log('\n菜单列表 (' + menuKeys.length + '个):');
    menuKeys.forEach(key => {
      const menu = menuData.data.nav_menu[key];
      console.log(`  ${menu.title} (${menu.mid})`);
      
      // 检查必需字段
      const required = ['component_tpl', 'path_name', 'table_name', 'default_action_type'];
      const missing = required.filter(f => !menu[f]);
      if (missing.length > 0) {
        console.warn(`    ⚠️ 缺少字段:`, missing);
      }
    });
  }
  
  if (menuData.data.api_maps) {
    const tableKeys = Object.keys(menuData.data.api_maps);
    console.log('\nAPI 映射 (' + tableKeys.length + '个表):');
    tableKeys.forEach(table => {
      const actions = Object.keys(menuData.data.api_maps[table]);
      console.log(`  ${table}: ${actions.join(', ')}`);
    });
  }
  
  console.log('\n' + '='.repeat(60));
  console.log('✅ 所有数据验证完成！');
  console.log('='.repeat(60));
  console.log('\n如果所有字段都显示✅，说明后端完全正常！');
  console.log('请刷新页面重新登录，数据应该能正常显示。');
  console.log('\n执行: location.reload();');
})
.catch(err => {
  console.error('❌ 验证失败:', err);
});
```

---

## 🚀 **请执行上述验证脚本，然后截图给我看：**

1. 所有字段的状态（✅ 或 ⚠️）
2. 是否有缺失的字段
3. LocalStorage 读取验证的结果

这样我就能确定：
- 后端是否返回了所有必需字段
- 字段名是否完全匹配
- 是否有其他问题

**然后我可以针对性地修复前端或后端！** 🎯
