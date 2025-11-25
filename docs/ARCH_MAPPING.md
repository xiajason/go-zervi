# 🔁 旧服务 → 新服务/职责 映射对照表

> 本表用于对照早期 Eureka 体系与当前 Zervigo 分层架构（Central Brain/API Gateway + Consul/内置注册 + 分层编排）的服务职责变更与启动依赖。

## 核心变更摘要
- 服务发现：Eureka 被 Consul 或内置注册替代；在 Docker MVP 方案中使用 `consul`，在本地直启方案由“中央大脑”聚合注册能力。
- 网关：统一为 API Gateway 能力，分为两种形态：
  - 本地直启：`central-brain`（端口 9000）
  - Docker 微服务：`api-gateway`（端口 8080）
- 业务微服务：沿用并按领域重组为 `user`/`company`/`job`/`resume` 等；部分早期服务（resource/personal/enterprise）合并到现有领域服务。
- 支撑与扩展：补充 `notification`、`banner`、`template`、`statistics`、`dev-team`、`multi-database`、`ai-service`、`blockchain-service` 等。

## 映射表

| 旧服务 | 旧依赖 | 建议启动序 | 新服务/职责 | 新依赖 | 说明 |
| --- | --- | --- | --- | --- | --- |
| mysql | 无 | 1 | mysql | 无 | 基础设施保持不变 |
| redis | 无 | 1 | redis | 无 | 基础设施保持不变 |
| eureka-server | 无 | 1 | consul 或 中央大脑内置注册 | 无 | Docker MVP 使用 `consul`；本地直启由 `central-brain` 聚合注册/路由 |
| api-gateway | eureka-server | 2 | api-gateway（Docker）或 central-brain（本地） | auth-service | 统一入口；依赖认证与注册信息 |
| resource-service | eureka-server + mysql | 2 | 合并至 basic-server 或相应业务域服务 | auth-service + mysql | 资源/配置职责转入 `basic-server` 与各领域服务 |
| statistics-service | eureka-server + mysql | 2 | statistics-service | auth-service + redis + mysql | 统计与报表；缓存用 redis |
| points-service | eureka-server + mysql | 2 | （可选）并入 user/notification/statistics 组合 | auth-service + redis + mysql | 暂未单列，如需可恢复为独立服务 |
| personal-service | eureka-server + mysql | 2 | 合并至 user-service | auth-service + mysql | 个人中心/资料职责合并到用户域 |
| enterprise-service | eureka-server + mysql | 2 | company-service | auth-service + mysql | 企业域服务延续为 `company-service` |
| resume-service | eureka-server + mysql | 2 | resume-service | auth-service + mysql (+AI/链) | 简历域沿用；可联动 AI 与链服务 |
| open-api-service | eureka-server + redis | 2 | api-gateway + auth-service | auth-service + redis | 统一对外 API 能力由网关 + 认证提供 |
| blockchain-service | eureka-server | 2 | blockchain-service | mysql | 链上记录/存证等职责独立 |
| nginx | api-gateway | 3 | （可选）外层反向代理 | api-gateway/central-brain | 视对外暴露与静态资源需求选择 |

## 现行服务清单与端口（参考）
- 基础设施：`mysql:3306`、`postgres:5432`、`redis:6379`、（MVP：`consul:8500`）
- 核心：`auth-service:8207`、`api-gateway:8080`、`basic-server:8081`、（本地：`central-brain:9000`）
- 业务：`user-service:8082`、`company-service:8083`、`job-service:8084`、`resume-service:8085`
- 支持：`notification-service:8086`、`banner-service:8087`、`template-service:8088`、`statistics-service:8089`
- 扩展：`dev-team-service:8090`、`multi-database-service:8091`、`ai-service:8100`、`blockchain-service:8208`

## 启动顺序（建议）
1) 基础设施：mysql、postgres、redis（MVP 还包括 consul）
2) 核心：auth-service → api-gateway/basic-server/central-brain
3) 业务：user/company/job/resume（依据领域数据库与认证）
4) 支持：notification/statistics（依赖 redis 与认证），banner/template（依赖认证）
5) 扩展：ai-service（依赖认证 + pg/mysql）、blockchain-service（依赖 mysql）

## 术语说明
- 在“组合编排”的语境中：`ai-service` 作为可选组件参与 7 种组合的开关选择。
- 在“架构与部署”的语境中：`ai-service` 是独立可部署的微服务（有独立进程/端口/依赖/健康检查）。

> 备注：本地直启脚本参见 `scripts/start-local-services.sh`；Docker 全量微服务编排参见 `docker/docker-compose.microservices.yml` 与 `scripts/start-all-services.sh`；MVP 轻量集群参见 `docker/docker-compose.yml`。
