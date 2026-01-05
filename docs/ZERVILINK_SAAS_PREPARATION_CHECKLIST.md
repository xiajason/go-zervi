# ZerviLinkSaaS 平台准备清单

**目标**: 在阿里云服务器搭建自研的 ZerviLinkSaaS 平台  
**对标**: OpenLinkSaaS  
**预计时间**: 4-6 周完整上线

---

## ✅ 准备清单

### 一、阿里云资源（预算：500-2000元/月）

#### 1.1 服务器资源 ☐

**推荐配置（分布式部署）**:
```yaml
□ ECS 实例 x 3:
  - 管理平台: 4核8G, 100G SSD (ZerviLinkSaaS 后台)
  - 微服务集群: 4核8G, 100G SSD x 2 (Zervigo 服务)
  
□ RDS PostgreSQL:
  - 规格: 2核4G
  - 存储: 100GB
  - 版本: 14.x
  
□ Redis 标准版:
  - 规格: 1GB
  - 版本: 7.0
  
□ OSS 对象存储:
  - Bucket: zervilink-saas
  - 用途: 文件、日志、备份
```

**最小配置（单机测试）**:
```yaml
□ ECS 实例 x 1:
  - 规格: 8核16G, 200G SSD
  - 操作系统: Ubuntu 22.04 或 CentOS 8
  - 用途: 所有服务单机部署
  
□ 本地 PostgreSQL + Redis:
  - 跑在 ECS 上（节省成本）
```

#### 1.2 网络与安全 ☐
```yaml
□ 弹性公网 IP (EIP): 1个
  - 带宽: 5M 起步（按需升级）
  
□ 域名注册:
  - 示例: zervilink.com
  - 备案: 如需国内访问必须备案
  
□ SSL 证书:
  - 阿里云免费证书 或 Let's Encrypt
  
□ 安全组配置:
  - 入站规则:
    * 80/443 (HTTP/HTTPS) - 全球开放
    * 9100 (ZerviLinkSaaS 管理后台) - 内网或白名单
    * 22 (SSH) - 白名单 IP
    * 8207-8208,8082-8085,8083,8100,9000 (微服务) - 内网
    * 5432 (PostgreSQL) - 内网
    * 6379 (Redis) - 内网
    * 8500 (Consul) - 内网
```

#### 1.3 阿里云产品订阅 ☐
```yaml
核心产品:
  □ ECS (云服务器)
  □ RDS (PostgreSQL)
  □ Redis
  □ OSS (对象存储)
  
可选产品:
  □ SLB (负载均衡) - 高可用场景
  □ SLS (日志服务) - 日志聚合
  □ ACK (容器服务) - Kubernetes
  □ CDN - 静态资源加速
```

---

### 二、账号与权限（安全关键）

#### 2.1 阿里云 RAM 配置 ☐
```yaml
□ 创建 RAM 用户: zervilink-service
  
□ 授予权限策略:
  - AliyunECSFullAccess (ECS 管理)
  - AliyunRDSFullAccess (数据库管理)
  - AliyunOSSFullAccess (对象存储)
  - AliyunLogFullAccess (日志服务)
  - AliyunCloudMonitorFullAccess (监控)
  
□ 创建 AccessKey:
  - AccessKeyId: LTAI5t...
  - AccessKeySecret: (妥善保管)
  
□ 启用 MFA (多因素认证):
  - 提升账号安全性
```

#### 2.2 环境变量配置 ☐
```bash
# configs/saas.env (新建)
# 阿里云配置
ALIYUN_ACCESS_KEY_ID=LTAI5t...
ALIYUN_ACCESS_KEY_SECRET=xxx...
ALIYUN_REGION=cn-hangzhou

# 数据库配置
SAAS_DB_HOST=rm-xxx.mysql.rds.aliyuncs.com
SAAS_DB_PORT=5432
SAAS_DB_USER=zervilink
SAAS_DB_PASSWORD=xxx...
SAAS_DB_NAME=zervilink_saas

# Redis 配置
SAAS_REDIS_HOST=r-xxx.redis.rds.aliyuncs.com
SAAS_REDIS_PORT=6379
SAAS_REDIS_PASSWORD=xxx...

# OSS 配置
SAAS_OSS_BUCKET=zervilink-saas
SAAS_OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com

# 平台配置
SAAS_ADMIN_USERNAME=admin
SAAS_ADMIN_PASSWORD=xxx...
SAAS_JWT_SECRET=zervilink-saas-secret-2025
SAAS_PLATFORM_PORT=9100

# GitHub/GitLab
GITHUB_TOKEN=ghp_xxx...
GITLAB_TOKEN=glpat_xxx...

# 通知配置
DINGTALK_WEBHOOK=https://oapi.dingtalk.com/robot/send?access_token=xxx
WECHAT_WORK_WEBHOOK=https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx
```

---

### 三、技术栈与工具

#### 3.1 后端开发环境 ☐
```yaml
□ Go 1.23+ 安装与配置
□ Python 3.9+ (AI 服务)
□ Git 配置（SSH Key）
□ Docker 安装
□ Docker Compose 安装
□ Make 工具
```

#### 3.2 前端开发环境 ☐
```yaml
□ Node.js 22+ (已有)
□ pnpm (包管理器)
□ Vue CLI / Vite
□ Element Plus UI 库
□ VS Code + Volar 插件
```

#### 3.3 数据库工具 ☐
```yaml
□ pgAdmin / DBeaver (PostgreSQL 客户端)
□ Redis Desktop Manager
□ Consul UI (http://localhost:8500)
```

#### 3.4 DevOps 工具 ☐
```yaml
□ Prometheus (监控)
□ Grafana (可视化)
□ Nginx (反向代理)
□ Ansible (可选 - 自动化运维)
```

---

### 四、代码与文档准备

#### 4.1 项目结构 ☐
```bash
□ 创建平台目录:
mkdir -p services/platform/zervilink-saas
mkdir -p services/platform/zervilink-saas/api
mkdir -p services/platform/zervilink-saas/service
mkdir -p services/platform/zervilink-saas/models
mkdir -p services/platform/zervilink-saas/web

□ 初始化 Go 模块:
cd services/platform/zervilink-saas
go mod init github.com/xiajason/zervilink-saas

□ 初始化前端项目:
cd web
npm create vite@latest . -- --template vue-ts
```

#### 4.2 文档准备 ☐
```yaml
□ 创建核心文档:
  - docs/ZERVILINK_SAAS_ARCHITECTURE.md (架构设计)
  - docs/ZERVILINK_SAAS_API_REFERENCE.md (API 文档)
  - docs/ZERVILINK_SAAS_DEPLOYMENT_GUIDE.md (部署指南)
  - docs/ZERVILINK_SAAS_USER_MANUAL.md (用户手册)
  - docs/ZERVILINK_SAAS_DEVELOPMENT_GUIDE.md (开发指南)
```

---

### 五、数据与脚本准备

#### 5.1 数据库初始化 ☐
```bash
□ 创建初始化脚本:
databases/postgres/init/10-zervilink-saas-schema.sql

□ 包含内容:
  - 服务管理表
  - 部署记录表
  - 配置中心表
  - 审计日志表
  - CI/CD 流水线表
  - 用户权限表
```

#### 5.2 部署脚本 ☐
```bash
□ 创建部署脚本:
scripts/deploy-zervilink-saas.sh
scripts/start-zervilink-saas.sh
scripts/stop-zervilink-saas.sh

□ 创建监控脚本:
scripts/health-check-all.sh
scripts/backup-saas-data.sh
```

---

### 六、第三方集成准备

#### 6.1 GitHub / GitLab ☐
```yaml
□ GitHub:
  - 创建 Personal Access Token
  - 权限: repo, workflow, admin:org
  - 配置 Webhook (推送触发 CI/CD)

□ GitLab (可选):
  - 创建 Access Token
  - 配置 CI/CD Runner
```

#### 6.2 通知渠道 ☐
```yaml
□ 钉钉机器人:
  - 创建群机器人
  - 获取 Webhook URL
  - 配置关键词（可选）

□ 企业微信机器人:
  - 创建群机器人
  - 获取 Webhook URL

□ 邮件 SMTP:
  - 配置 SMTP 服务器
  - 测试邮件发送
```

#### 6.3 监控告警 ☐
```yaml
□ Prometheus:
  - 安装配置
  - 配置采集规则
  - 数据持久化

□ Grafana:
  - 安装配置
  - 创建 Dashboard
  - 配置告警规则
  
□ 阿里云监控:
  - 配置 ECS 监控
  - 配置 RDS 监控
  - 配置告警策略
```

---

### 七、安全加固准备

#### 7.1 网络安全 ☐
```yaml
□ 防火墙规则:
  - 仅开放必要端口
  - 白名单 IP 访问
  
□ HTTPS 配置:
  - 申请 SSL 证书
  - 配置 Nginx HTTPS
  - 强制 HTTPS 跳转
  
□ VPN / 堡垒机:
  - 内网服务仅通过 VPN 访问
  - 或使用阿里云堡垒机
```

#### 7.2 应用安全 ☐
```yaml
□ JWT 认证:
  - 超级管理员 Token
  - Token 过期策略
  - Refresh Token
  
□ API 鉴权:
  - 所有 API 需要认证
  - 操作权限校验
  - 审计日志记录
  
□ 数据加密:
  - 敏感配置加密存储
  - 数据库密码加密
  - API Key 加密
```

---

### 八、测试与验证准备

#### 8.1 功能测试清单 ☐
```yaml
□ 服务管理:
  - [ ] 服务列表展示
  - [ ] 服务启动/停止
  - [ ] 健康检查
  - [ ] 日志查看
  
□ 部署管理:
  - [ ] 手动部署触发
  - [ ] 自动部署（Webhook）
  - [ ] 部署日志
  - [ ] 回滚操作
  
□ 配置管理:
  - [ ] 配置查看/编辑
  - [ ] 配置版本管理
  - [ ] 配置热更新
  
□ 监控告警:
  - [ ] 实时监控大盘
  - [ ] 告警规则配置
  - [ ] 告警通知发送
```

#### 8.2 性能测试 ☐
```yaml
□ 压力测试:
  - 管理后台并发访问
  - API 接口压测
  - 数据库连接池测试
  
□ 稳定性测试:
  - 7x24 小时运行测试
  - 故障恢复测试
  - 自动重启验证
```

---

### 九、运维准备

#### 9.1 监控与告警 ☐
```yaml
□ 配置监控指标:
  - CPU 使用率 > 80% 告警
  - 内存使用率 > 85% 告警
  - 磁盘使用率 > 90% 告警
  - 服务挂掉立即告警
  - 数据库连接数告警
  
□ 配置告警通知:
  - 钉钉群通知
  - 企业微信通知
  - 邮件通知
  - 短信通知（紧急）
```

#### 9.2 备份策略 ☐
```yaml
□ 数据库备份:
  - 每天凌晨 2 点自动备份
  - 保留 30 天
  - 异地备份（OSS）
  
□ 配置文件备份:
  - 版本化管理（Git）
  - 每次修改前自动备份
  
□ 日志备份:
  - 归档到 OSS
  - 保留 90 天
```

---

## 📋 分阶段实施计划

### 第一阶段：基础环境搭建（1 周）

#### Week 1: 服务器与基础设施
```yaml
Day 1-2: 阿里云资源采购与配置
  □ 购买 ECS 实例
  □ 购买 RDS / Redis（或本地部署）
  □ 配置安全组
  □ 域名解析
  □ SSL 证书申请

Day 3-4: 服务器环境初始化
  □ 安装 Docker
  □ 安装 PostgreSQL（如本地）
  □ 安装 Redis（如本地）
  □ 安装 Consul
  □ 安装 Nginx
  □ 配置防火墙

Day 5-7: 部署现有 Zervigo 服务
  □ 迁移数据库到阿里云
  □ 部署所有微服务
  □ 配置 Consul 服务发现
  □ 验证服务健康
```

### 第二阶段：ZerviLinkSaaS 后端开发（2 周）

#### Week 2: 核心功能开发
```yaml
Day 1-3: 服务管理模块
  □ 服务注册与发现（基于 Consul）
  □ 服务健康检查
  □ 服务启动/停止/重启 API
  □ 服务日志查询 API

Day 4-5: 配置管理模块
  □ 配置 CRUD API
  □ 配置版本管理
  □ 配置热更新机制
  □ 配置审计日志

Day 6-7: 监控模块
  □ Prometheus 数据采集
  □ 监控指标 API
  □ 告警规则配置
  □ 告警通知发送
```

#### Week 3: 部署与 CI/CD
```yaml
Day 1-3: CI/CD 流水线
  □ GitHub Webhook 接收
  □ 构建任务队列
  □ Docker 镜像构建
  □ 自动部署脚本

Day 4-5: 部署管理
  □ 部署记录管理
  □ 回滚功能
  □ 蓝绿部署
  □ 灰度发布

Day 6-7: 权限与审计
  □ 超级管理员认证
  □ 操作权限控制
  □ 审计日志记录
  □ 敏感操作二次确认
```

### 第三阶段：前端开发（2 周）

#### Week 4: 管理界面
```yaml
Day 1-3: 核心页面
  □ 登录页面
  □ 服务管理页面
  □ 监控大盘
  □ 配置中心页面

Day 4-5: CI/CD 界面
  □ 流水线列表
  □ 构建日志展示
  □ 部署历史
  □ 一键部署

Day 6-7: 辅助功能
  □ 日志查询界面
  □ 告警管理
  □ 用户权限管理
  □ 系统设置
```

### 第四阶段：测试与上线（1 周）

#### Week 5-6: 测试与优化
```yaml
Day 1-2: 功能测试
  □ 所有功能冒烟测试
  □ 用户流程测试
  □ 边界情况测试

Day 3-4: 性能测试
  □ 压力测试
  □ 稳定性测试
  □ 并发测试

Day 5-6: 安全测试
  □ 权限测试
  □ SQL 注入测试
  □ XSS 测试
  □ CSRF 测试

Day 7: 上线准备
  □ 数据备份
  □ 文档完善
  □ 培训材料
  □ 正式上线
```

---

## 🛠️ 立即行动项（本周）

### 今天需要完成
1. ☐ **确定部署模式**：单机 or 分布式？
2. ☐ **购买阿里云资源**：ECS + RDS + Redis（或单机）
3. ☐ **配置 RAM 权限**：创建 AccessKey
4. ☐ **域名与证书**：注册域名、申请 SSL

### 本周需要完成
1. ☐ **初始化服务器环境**：Docker + Nginx + 数据库
2. ☐ **迁移现有服务到阿里云**：验证功能正常
3. ☐ **创建项目结构**：`services/platform/zervilink-saas`
4. ☐ **编写数据库 Schema**：平台核心表设计

---

## 📦 交付清单

### 代码交付
```yaml
□ ZerviLinkSaaS 后端代码
  - 服务管理 API
  - 配置管理 API
  - 监控 API
  - CI/CD API

□ ZerviLinkSaaS 前端代码
  - 管理后台界面
  - 响应式设计
  - 实时更新

□ 部署脚本
  - Docker Compose 文件
  - Nginx 配置
  - 启动/停止脚本
```

### 文档交付
```yaml
□ 架构文档
□ API 文档
□ 部署文档
□ 用户手册
□ 运维手册
□ 故障排查指南
```

### 测试交付
```yaml
□ 功能测试报告
□ 性能测试报告
□ 安全测试报告
□ 用户验收测试
```

---

## 💰 预算估算

### 阿里云成本（月）
```yaml
单机部署（测试）:
  - ECS 8核16G: ~300元
  - 带宽 5M: ~50元
  - OSS 存储 100GB: ~10元
  - 域名: ~60元/年
  合计: ~400元/月

分布式部署（生产）:
  - ECS x 3: ~900元
  - RDS PostgreSQL 2核4G: ~500元
  - Redis 1GB: ~200元
  - SLB: ~200元
  - 带宽 10M: ~100元
  - OSS: ~20元
  合计: ~2000元/月

高可用部署（大规模）:
  - ACK 集群: ~3000元
  - RDS 高可用: ~1000元
  - Redis 集群: ~500元
  - CDN: ~200元
  - 日志服务: ~300元
  合计: ~5000元/月
```

### 开发成本（人天）
```yaml
后端开发: 15-20 人天
前端开发: 10-15 人天
测试与优化: 5-7 人天
部署与运维: 3-5 人天

合计: 35-50 人天
```

---

## 🎯 成功标准

### 功能标准
- [ ] 可视化管理所有 Zervigo 微服务
- [ ] 一键部署到阿里云
- [ ] 实时监控所有服务健康
- [ ] 配置在线修改并热更新
- [ ] CI/CD 自动化流水线

### 性能标准
- [ ] 管理后台响应时间 < 200ms
- [ ] 支持管理 50+ 微服务实例
- [ ] 部署操作 < 5 分钟完成
- [ ] 监控数据延迟 < 10 秒

### 安全标准
- [ ] 超级管理员认证
- [ ] 操作审计 100% 覆盖
- [ ] 敏感信息加密存储
- [ ] HTTPS 强制加密传输

---

## 📞 下一步行动

### 需要你提供的信息
1. **部署规模**：测试环境（单机）还是生产环境（分布式）？
2. **预算范围**：月度预算是多少？
3. **时间要求**：希望多久上线第一版？
4. **团队规模**：有多少开发人员参与？
5. **阿里云账号**：是否已有阿里云账号？

### 我可以立即帮你做的
1. ✅ 创建项目结构 `services/platform/zervilink-saas`
2. ✅ 编写数据库初始化脚本
3. ✅ 实现服务管理 API（第一版）
4. ✅ 创建部署脚本模板
5. ✅ 编写阿里云部署指南

**准备好了吗？告诉我你的选择，我们马上开始！** 🚀

