# 权限/路由 API 最小实现与 Mock 约定

所有路径经中央大脑网关 `/api/v1/*`。

## 用户档案
GET /api/v1/users/profile

示例返回:
```
{
  "code": 0,
  "data": {
    "id": 1,
    "username": "admin",
    "role": "super_admin",
    "permissions": ["resume:view","resume:edit","job:view"]
  }
}
```

## 权限服务（permission-service）
- GET /api/v1/permission/user/:userId/roles → ["super_admin"]
- GET /api/v1/permission/user/:userId/permissions → ["resume:view","resume:edit","job:view"]
- POST /api/v1/permission/check → { userId, permissionCode } → { code:0, data:{ allowed:true } }
- POST /api/v1/permission/batch-check → { userId, permissionCodes } → { code:0, data:{ "resume:view":true } }
- GET /api/v1/permission/all → 全量权限字典

## 路由服务（router-service）
- GET /api/v1/router/user-pages → ["pages/home/index","pages/resume/index","pages/mine/index"]
- GET /api/v1/router/user-routes →
```
{
  "code": 0,
  "data": [
    { "path": "/pages/home/index", "meta": { "roles": ["*"], "perms": [] } },
    { "path": "/pages/resume/index", "meta": { "roles": ["super_admin","candidate"], "perms": ["resume:view"] } }
  ]
}
```

## Mock 返回要求
- 统一 `code===0` 为成功
- 字段名与以上示例保持一致（role、permissions、pages、routes、allowed）
