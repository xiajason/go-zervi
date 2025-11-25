# 📋 后端前端交付检查清单

> 本文档列出了后端团队在交付给前端团队之前需要完成的所有检查和准备工作

## 🎯 交付目标

确保后端API接口稳定、文档完善、易于集成、安全可靠，让前端开发团队能够快速、高效地进行开发和对接工作。

---

## 📚 一、API接口文档准备

### 1.1 完整API清单

- [ ] **生成所有API接口清单文档**
  - 列出所有微服务的API端点
  - 标注每个接口的HTTP方法、路径、用途
  - 区分核心接口和可选接口

- [ ] **API接口说明文档**
  ```bash
  # 需要包含的信息：
  # - 请求方法 (GET/POST/PUT/DELETE)
  # - 完整URL路径
  # - 请求参数（必填/可选）
  # - 请求体示例
  # - 响应数据格式
  # - 错误码说明
  # - 是否需要认证
  # - 使用示例
  ```

- [ ] **生成交互式API文档（推荐）**
  ```bash
  # 使用工具：
  # - Swagger/OpenAPI 文档
  # - Postman Collection
  # - Postman API Documentation
  # - 或使用项目自带的 .api 文件
  ```

### 1.2 当前API概览

根据项目当前的API定义文件：

| 服务 | API文件 | 主要功能 | 端口 |
|------|---------|---------|------|
| 认证服务 | auth.api | 登录、注册、Token刷新、权限管理 | 8207 |
| 用户服务 | user.api | 用户信息、个人资料管理 | 8082 |
| 职位服务 | job.api | 职位CRUD、搜索、推荐 | 8084 |
| 简历服务 | resume.api | 简历CRUD、分析、匹配 | 8085 |
| 企业服务 | company.api | 企业CRUD、搜索、认证 | 8083 |
| AI服务 | ai.api | 智能匹配、简历分析、AI聊天 | 8100 |
| 区块链服务 | blockchain.api | 数据审计、不可篡改记录 | 8208 |

---

## 🔐 二、认证和授权准备

### 2.1 Token认证机制

- [ ] **JWT Token生成和验证**
  - 确保Token生成逻辑正确
  - Token有效期设置合理
  - 提供Token刷新机制

- [ ] **Token传递方式说明**
  ```javascript
  // 方式1: Header传递（推荐）
  headers: {
    'Authorization': 'Bearer {token}'
  }
  
  // 方式2: Query参数（备用）
  ?token={token}
  
  // 方式3: Cookie（如果启用）
  ```

- [ ] **测试账号准备**
  ```bash
  # 准备不同角色的测试账号
  - 超级管理员账号
  - 普通用户账号
  - 企业账号
  - 游客账号（仅查看权限）
  ```

### 2.2 权限控制说明

- [ ] **角色权限矩阵表**
  | 接口 | 游客 | 普通用户 | 企业用户 | 管理员 |
  |------|------|---------|---------|-------|
  | GET /api/v1/job/list | ✅ | ✅ | ✅ | ✅ |
  | POST /api/v1/job | ❌ | ❌ | ✅ | ✅ |
  | DELETE /api/v1/job/:id | ❌ | ❌ | ✅ | ✅ |
  | POST /api/v1/resume | ✅ | ✅ | ❌ | ✅ |

- [ ] **中间件配置说明**
  - 哪些接口需要Auth中间件
  - 哪些接口需要特定角色权限
  - 权限验证失败的错误响应格式

---

## 🌐 三、CORS和安全配置

### 3.1 跨域配置

- [ ] **CORS中间件配置正确**
  ```go:158:178:shared/core/middleware/error_handler.go
  // CORS 跨域中间件
  func CORS() gin.HandlerFunc {
  	return func(c *gin.Context) {
  		origin := c.GetHeader("Origin")

  		// 设置CORS头
  		c.Header("Access-Control-Allow-Origin", origin)
  		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
  		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
  		c.Header("Access-Control-Allow-Credentials", "true")
  		c.Header("Access-Control-Max-Age", "86400")

  		// 处理预检请求
  		if c.Request.Method == "OPTIONS" {
  			c.AbortWithStatus(http.StatusNoContent)
  			return
  		}

  		c.Next()
  	}
  }
  ```

- [ ] **生产环境CORS白名单设置**
  - 配置允许的前端域名
  - 限制Origin来源（生产环境）

- [ ] **OPTIONS预检请求测试**
  ```bash
  # 测试预检请求
  curl -X OPTIONS http://localhost:9000/api/v1/job/list \
    -H "Origin: http://localhost:10086" \
    -H "Access-Control-Request-Method: GET" \
    -i
  ```

### 3.2 安全头配置

- [ ] **Security中间件已启用**
  ```go:180:192:shared/core/middleware/error_handler.go
  // Security 安全中间件
  func Security() gin.HandlerFunc {
  	return func(c *gin.Context) {
  		// 设置安全头
  		c.Header("X-Content-Type-Options", "nosniff")
  		c.Header("X-Frame-Options", "DENY")
  		c.Header("X-XSS-Protection", "1; mode=block")
  		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
  		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

  		c.Next()
  	}
  }
  ```

---

## 🔌 四、API网关和服务发现

### 4.1 Central Brain配置

- [ ] **Central Brain（API网关）正常启动**
  - 端口：9000
  - 作为所有请求的统一入口
  - 路由配置正确

- [ ] **路由规则文档**
  ```bash
  # 前端应使用的基址
  API_BASE_URL = http://localhost:9000  # 开发环境
  API_BASE_URL = https://api.production.com  # 生产环境
  
  # 示例接口调用
  GET  http://localhost:9000/api/v1/auth/login
  POST http://localhost:9000/api/v1/job
  PUT  http://localhost:9000/api/v1/user/info
  ```

- [ ] **服务健康检查端点**
  ```bash
  # Central Brain
  curl http://localhost:9000/health
  
  # 各个微服务
  curl http://localhost:8207/health  # Auth Service
  curl http://localhost:8082/health  # User Service
  curl http://localhost:8084/health  # Job Service
  ```

### 4.2 端口列表

- [ ] **提供完整的端口列表**
  ```bash
  9000  - Central Brain (API Gateway)
  8207  - Auth Service
  8082  - User Service
  8083  - Company Service
  8084  - Job Service
  8085  - Resume Service
  8100  - AI Service
  8208  - Blockchain Service
  8086  - Notification Service
  8087  - Statistics Service
  8088  - Banner Service
  8089  - Template Service
  ```

---

## 📊 五、数据格式和错误处理

### 5.1 统一响应格式

- [ ] **定义标准响应格式**
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      // 业务数据
    }
  }
  ```

- [ ] **定义错误响应格式**
  ```json
  {
    "code": 400,
    "message": "参数错误：用户名不能为空",
    "error": "INVALID_PARAMETER",
    "details": {
      "field": "username",
      "reason": "required"
    }
  }
  ```

- [ ] **HTTP状态码映射表**
  | 业务码 | HTTP状态码 | 说明 | 示例 |
  |-------|-----------|------|------|
  | 200 | 200 | 成功 | 查询成功 |
  | 400 | 400 | 参数错误 | 必填参数缺失 |
  | 401 | 401 | 未认证 | Token无效 |
  | 403 | 403 | 无权限 | 缺少权限 |
  | 404 | 404 | 资源不存在 | ID不存在 |
  | 500 | 500 | 服务器错误 | 数据库连接失败 |

### 5.2 分页格式

- [ ] **统一分页响应格式**
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "list": [...],
      "total": 100,
      "page": 1,
      "page_size": 20
    }
  }
  ```

### 5.3 字段命名规范

- [ ] **统一命名规范**
  - 使用 `snake_case`（下划线命名）
  - 时间戳使用 `create_time`, `update_time`
  - ID使用 `user_id`, `job_id`
  - 状态使用 `status`（整型或枚举）

---

## 🧪 六、测试数据和测试环境

### 6.1 测试数据准备

- [ ] **提供完整的测试数据**
  ```bash
  # 使用快速测试数据生成器
  python scripts/quick_test_data_generator.py
  ```

- [ ] **准备测试数据类型**
  - ✅ 用户数据（不同角色）
  - ✅ 职位数据（不同状态）
  - ✅ 简历数据
  - ✅ 企业数据
  - ✅ 测试图片资源

### 6.2 测试环境配置

- [ ] **开发环境配置**
  - 数据库连接配置（dev.env）
  - Redis配置
  - JWT密钥配置
  - 上传文件路径配置

- [ ] **Mock数据接口**
  - 如果某些服务未完成，提供Mock接口
  - 或者使用测试数据模拟返回

### 6.3 Postman Collection

- [ ] **提供Postman测试集合**
  ```bash
  # 包含内容：
  - 所有API接口的请求示例
  - 环境变量配置（base_url, token）
  - 测试脚本（自动设置token）
  - 响应断言
  ```

---

## 🚀 七、启动和部署

### 7.1 本地启动文档

- [ ] **提供启动脚本和说明**
  ```bash
  # 使用统一启动脚本
  ./scripts/start-local-services.sh
  ```

- [ ] **启动顺序**
  ```bash
  1. PostgreSQL 数据库
  2. Redis 缓存
  3. Consul 服务发现（如启用）
  4. Central Brain（API Gateway）
  5. 核心服务（Auth, User）
  6. 业务服务（Job, Resume, Company）
  7. 基础设施服务（AI, Blockchain等）
  ```

- [ ] **数据库初始化**
  ```bash
  # 运行数据库迁移脚本
  ./scripts/migrate-microservices-db.sh
  ```

### 7.2 部署文档

- [ ] **提供部署配置文档**
  - 开发环境部署步骤
  - 测试环境部署步骤
  - 生产环境部署注意事项

- [ ] **Docker部署（如果适用）**
  ```bash
  docker-compose -f docker/docker-compose.yml up -d
  ```

---

## 📝 八、日志和监控

### 8.1 日志规范

- [ ] **定义日志级别**
  - DEBUG: 详细调试信息
  - INFO: 一般信息
  - WARN: 警告信息
  - ERROR: 错误信息

- [ ] **提供日志查看方式**
  ```bash
  # 查看服务日志
  tail -f logs/user-service.log
  tail -f logs/job-service.log
  ```

### 8.2 监控和健康检查

- [ ] **健康检查端点可用**
  ```bash
  curl http://localhost:9000/health
  curl http://localhost:8207/health
  ```

- [ ] **使用健康检查脚本**
  ```bash
  ./scripts/comprehensive_health_check.sh
  ```

---

## 🔄 九、接口版本管理

### 9.1 版本控制

- [ ] **API版本管理**
  - 当前版本：`/api/v1/`
  - 如需升级：`/api/v2/`
  - 同时维护多个版本（向后兼容）

- [ ] **版本变更日志**
  - 记录每个版本的变更内容
  - 废弃接口公告
  - 迁移指南

### 9.2 兼容性

- [ ] **避免破坏性变更**
  - 不要修改现有响应结构
  - 新增字段保持向后兼容
  - 废弃字段提前通知

---

## 🔧 十、工具和脚本

### 10.1 开发工具

- [ ] **提供常用脚本**
  ```bash
  # 启动服务
  ./scripts/start-local-services.sh
  
  # 停止服务
  ./scripts/stop-local-services.sh
  
  # 数据库迁移
  ./scripts/migrate-microservices-db.sh
  
  # 健康检查
  ./scripts/comprehensive_health_check.sh
  
  # 测试数据生成
  python scripts/quick_test_data_generator.py
  ```

### 10.2 调试工具

- [ ] **提供调试方法**
  ```bash
  # 查看API日志
  tail -f logs/*.log
  
  # 数据库连接测试
  psql -h localhost -U postgres -d zervigo_mvp
  
  # Redis连接测试
  redis-cli -h localhost -p 6379
  ```

---

## 📞 十一、沟通和协作

### 11.1 交付会议

- [ ] **召开交付会议**
  - API接口演示
  - 常见问题解答
  - 开发环境搭建指导
  - 联调计划制定

### 11.2 联系人和支持

- [ ] **提供支持渠道**
  - 技术负责人联系方式
  - 问题反馈渠道（如GitHub Issues, Slack群等）
  - 文档更新通知机制

### 11.3 问题跟踪

- [ ] **建立问题跟踪机制**
  - API问题反馈模板
  - Bug优先级分类
  - 响应时间承诺

---

## ✅ 十二、最终验收检查

### 12.1 端到端测试

- [ ] **完成核心业务流程测试**
  ```
  1. 用户注册 → 登录 → 获取Token
  2. 使用Token访问受保护接口
  3. 创建职位/简历
  4. 搜索和查询功能
  5. 文件上传功能
  6. 权限验证
  ```

### 12.2 性能测试

- [ ] **基础性能指标**
  ```bash
  - API响应时间 < 500ms（常规查询）
  - API响应时间 < 2s（复杂查询、AI接口）
  - 并发支持至少100 req/s
  - 数据库连接池配置合理
  ```

### 12.3 安全检查

- [ ] **安全扫描**
  - SQL注入防护
  - XSS防护
  - CSRF防护（如果适用）
  - Token泄露防护
  - 敏感数据加密

---

## 📦 十三、交付物清单

### 13.1 必需交付物

- [ ] **技术文档**
  - ✅ API接口文档（本文档 + API文件）
  - ✅ 数据库设计文档
  - ✅ 架构设计文档
  - ✅ 部署运维文档

- [ ] **代码和配置**
  - ✅ 完整源代码
  - ✅ 配置文件模板
  - ✅ 启动脚本
  - ✅ 数据库初始化脚本

- [ ] **测试资料**
  - ✅ Postman Collection
  - ✅ 测试数据脚本
  - ✅ 测试账号列表

### 13.2 可交付物（推荐）

- [ ] **补充资料**
  - Swagger/OpenAPI文档
  - API Mock服务器（如未完成）
  - 性能测试报告
  - 安全测试报告
  - 视频演示

---

## 🎓 十四、前端对接指南

### 14.1 前端需要了解的信息

- [ ] **提供前端对接指南**
  ```markdown
  ## 前端对接指南
  
  ### 1. 环境配置
  ```javascript
  // config/dev.js
  export default {
    baseURL: 'http://localhost:9000',
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json'
    }
  }
  ```
  
  ### 2. 请求拦截器
  ```javascript
  // 自动添加Token
  request.interceptors.request.use(config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  })
  
  // 错误处理
  request.interceptors.response.use(
    response => response,
    error => {
      if (error.response.status === 401) {
        // 跳转登录页
      }
      return Promise.reject(error)
    }
  )
  ```
  
  ### 3. 常用接口封装
  ```javascript
  // api/user.js
  export const getUserInfo = () => {
    return request.get('/api/v1/user/info')
  }
  
  export const updateUserInfo = (data) => {
    return request.put('/api/v1/user/info', data)
  }
  ```
  ```

### 14.2 联调计划

- [ ] **制定联调计划**
  - 第一周：核心功能联调（登录、用户信息、基础CRUD）
  - 第二周：业务功能联调（职位、简历、企业）
  - 第三周：高级功能联调（AI、搜索、推荐）
  - 第四周：全流程测试和优化

---

## 📈 十五、后续支持

### 15.1 迭代计划

- [ ] **明确后续迭代计划**
  - 阶段一完成哪些功能
  - 阶段二规划哪些功能
  - 里程碑时间节点

### 15.2 持续优化

- [ ] **建立持续优化机制**
  - 定期API性能监控
  - 错误日志分析
  - 用户反馈收集
  - 迭代优化计划

---

## 📝 附：检查清单模板

### 使用说明

1. **逐步完成检查项**
   - 在每一项前的 `[ ]` 打勾 `[x]`
   - 添加完成时间和备注

2. **问题记录**
   - 发现的问题记录在"问题跟踪表"中
   - 及时修复并验证

3. **签字确认**
   - 后端负责人签字
   - 前端负责人签字
   - 项目负责人签字

### 问题跟踪表

| 问题ID | 问题描述 | 优先级 | 负责人 | 状态 | 解决时间 |
|-------|---------|-------|-------|------|---------|
| #1 | | P0/P1/P2 | | 待处理/处理中/已解决 | |

---

## 🎯 总结

完成以上所有检查项后，后端团队可以：

1. ✅ **确保API接口稳定可用**
2. ✅ **提供完善的文档和示例**
3. ✅ **建立良好的沟通机制**
4. ✅ **支持前端快速开发**

---

## 📚 相关文档

- [项目README](README.md)
- [数据库设计文档](docs/MICROSERVICE_DATABASE_DESIGN.md)
- [部署文档](docs/DEPLOYMENT.md)
- [框架设计文档](docs/GO_ZERVI_FRAMEWORK_NAMING_PLAN.md)

---

**最后更新**: 2025-01-XX  
**文档版本**: v1.0  
**维护团队**: Zervigo 后端团队

