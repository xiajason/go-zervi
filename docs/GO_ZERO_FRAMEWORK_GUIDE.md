# Zervigo MVP Go-Zero微服务架构

## 📁 标准Go-Zero目录结构

```
MVPDEMO/
├── api/                         # API定义文件
│   ├── auth.api                 # 认证服务API定义
│   ├── user.api                 # 用户服务API定义
│   ├── job.api                  # 职位服务API定义
│   ├── resume.api               # 简历服务API定义
│   ├── company.api              # 企业服务API定义
│   ├── ai.api                   # AI服务API定义
│   └── blockchain.api           # 区块链服务API定义
├── rpc/                         # RPC服务
│   ├── auth/                    # 认证RPC服务
│   │   ├── auth.proto           # Protobuf定义
│   │   ├── authclient/          # 客户端代码
│   │   └── authserver/          # 服务端代码
│   ├── user/                    # 用户RPC服务
│   ├── job/                     # 职位RPC服务
│   ├── resume/                  # 简历RPC服务
│   ├── company/                 # 企业RPC服务
│   ├── ai/                      # AI RPC服务
│   └── blockchain/              # 区块链RPC服务
├── model/                       # 数据模型
│   ├── authmodel/               # 认证数据模型
│   ├── usermodel/               # 用户数据模型
│   ├── jobmodel/                # 职位数据模型
│   ├── resumemodel/             # 简历数据模型
│   ├── companymodel/            # 企业数据模型
│   └── blockchainmodel/         # 区块链数据模型
├── service/                     # HTTP服务
│   ├── auth/                    # 认证HTTP服务
│   │   ├── api/                 # API处理
│   │   ├── config/              # 配置
│   │   ├── handler/             # 处理器
│   │   ├── logic/               # 业务逻辑
│   │   ├── svc/                 # 服务上下文
│   │   ├── types/               # 类型定义
│   │   └── main.go              # 主入口
│   ├── user/                    # 用户HTTP服务
│   ├── job/                     # 职位HTTP服务
│   ├── resume/                  # 简历HTTP服务
│   ├── company/                 # 企业HTTP服务
│   ├── ai/                      # AI HTTP服务
│   └── blockchain/              # 区块链HTTP服务
├── gateway/                     # API网关
│   ├── config/                  # 网关配置
│   ├── handler/                 # 网关处理器
│   ├── middleware/              # 中间件
│   └── main.go                  # 网关主入口
├── common/                      # 公共组件
│   ├── config/                  # 公共配置
│   ├── middleware/              # 公共中间件
│   ├── utils/                   # 工具函数
│   └── types/                   # 公共类型
├── deploy/                      # 部署配置
│   ├── docker/                  # Docker配置
│   ├── k8s/                     # Kubernetes配置
│   └── scripts/                 # 部署脚本
└── tools/                       # 工具
    ├── goctl/                   # 代码生成工具
    └── scripts/                 # 开发脚本
```

## 🎯 Go-Zero核心特性

### 1. API定义文件（.api）
```go
// api/auth.api
syntax = "v1"

info(
    title: "认证服务API"
    desc: "用户认证、登录、注册等接口"
    author: "Zervigo Team"
    version: "v1.0.0"
)

type (
    LoginRequest {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    LoginResponse {
        Token string `json:"token"`
        UserId int64 `json:"user_id"`
        Username string `json:"username"`
    }
)

@server(
    group: auth
    middleware: Auth
)
service auth-api {
    @handler login
    post /api/v1/auth/login (LoginRequest) returns (LoginResponse)
    
    @handler logout
    post /api/v1/auth/logout returns (Response)
    
    @handler refresh
    post /api/v1/auth/refresh returns (LoginResponse)
}
```

### 2. RPC服务定义（.proto）
```protobuf
// rpc/auth/auth.proto
syntax = "proto3";

package auth;
option go_package = "./auth";

message LoginRequest {
    string username = 1;
    string password = 2;
}

message LoginResponse {
    string token = 1;
    int64 user_id = 2;
    string username = 3;
}

service Auth {
    rpc Login(LoginRequest) returns(LoginResponse);
    rpc Logout(Empty) returns(Empty);
    rpc Refresh(Empty) returns(LoginResponse);
}
```

### 3. 数据模型
```go
// model/usermodel/usermodel.go
package usermodel

import (
    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound

type UserModel interface {
    Insert(data *User) error
    FindOne(id int64) (*User, error)
    FindOneByUsername(username string) (*User, error)
    Update(data *User) error
    Delete(id int64) error
}

type defaultUserModel struct {
    conn  sqlx.SqlConn
    table string
}

func NewUserModel(conn sqlx.SqlConn) UserModel {
    return &defaultUserModel{
        conn:  conn,
        table: "`user`",
    }
}
```

### 4. 服务配置
```go
// service/auth/config/config.go
package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
    rest.RestConf
    Auth struct {
        AccessSecret string
        AccessExpire int64
    }
    Database struct {
        DataSource string
    }
    Redis struct {
        Host string
        Pass string
        DB   int
    }
}
```

## 🚀 代码生成工具

### 使用goctl生成代码
```bash
# 生成API服务
goctl api go -api api/auth.api -dir service/auth

# 生成RPC服务
goctl rpc protoc rpc/auth/auth.proto --go_out=./rpc/auth --go-grpc_out=./rpc/auth --zrpc_out=./rpc/auth

# 生成数据模型
goctl model mysql datasource -url="root:password@tcp(localhost:3306)/zervigo" -table="user" -dir="./model/usermodel"

# 生成Dockerfile
goctl docker -go service/auth/main.go
```

## 📊 服务端口分配

| 服务 | 端口 | 类型 | 说明 |
|------|------|------|------|
| Gateway | 8888 | HTTP | API网关 |
| Auth API | 8001 | HTTP | 认证服务 |
| User API | 8002 | HTTP | 用户服务 |
| Job API | 8003 | HTTP | 职位服务 |
| Resume API | 8004 | HTTP | 简历服务 |
| Company API | 8005 | HTTP | 企业服务 |
| AI API | 8006 | HTTP | AI服务 |
| Blockchain API | 8007 | HTTP | 区块链服务 |
| Auth RPC | 9001 | RPC | 认证RPC |
| User RPC | 9002 | RPC | 用户RPC |
| Job RPC | 9003 | RPC | 职位RPC |
| Resume RPC | 9004 | RPC | 简历RPC |
| Company RPC | 9005 | RPC | 企业RPC |
| AI RPC | 9006 | RPC | AI RPC |
| Blockchain RPC | 9007 | RPC | 区块链RPC |

## 🔧 中间件支持

### 1. 认证中间件
```go
// common/middleware/auth.go
func AuthMiddleware(secretKey string) rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // JWT验证逻辑
            token := r.Header.Get("Authorization")
            // 验证token
            // ...
            next(w, r)
        }
    }
}
```

### 2. 日志中间件
```go
// common/middleware/log.go
func LogMiddleware() rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 记录请求日志
            logx.Infof("Request: %s %s", r.Method, r.URL.Path)
            next(w, r)
        }
    }
}
```

### 3. 限流中间件
```go
// common/middleware/ratelimit.go
func RateLimitMiddleware() rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 限流逻辑
            // ...
            next(w, r)
        }
    }
}
```

## 📈 监控和链路追踪

### 1. Prometheus监控
```go
// 在main.go中启用监控
func main() {
    flag.Parse()
    
    var c config.Config
    conf.MustLoad(*configFile, &c)
    
    server := rest.MustNewServer(c.RestConf)
    defer server.Stop()
    
    // 启用Prometheus监控
    prometheus.MustRegister()
    
    ctx := svc.NewServiceContext(c)
    handler.RegisterHandlers(server, ctx)
    
    server.Start()
}
```

### 2. 链路追踪
```go
// 使用Jaeger进行链路追踪
import "github.com/zeromicro/go-zero/core/trace"

func main() {
    // 初始化链路追踪
    trace.StartAgent(trace.Config{
        Name: "auth-service",
        Endpoint: "http://jaeger:14268/api/traces",
    })
    defer trace.StopAgent()
}
```

## 🎯 最佳实践

### 1. 错误处理
```go
// common/types/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}) Response {
    return Response{
        Code:    0,
        Message: "success",
        Data:    data,
    }
}

func Error(code int, message string) Response {
    return Response{
        Code:    code,
        Message: message,
    }
}
```

### 2. 配置管理
```go
// 支持多环境配置
// config.yaml
Name: auth-api
Host: 0.0.0.0
Port: 8001
Auth:
  AccessSecret: "your-secret-key"
  AccessExpire: 7200
Database:
  DataSource: "root:password@tcp(localhost:3306)/zervigo?charset=utf8mb4&parseTime=True&loc=Local"
Redis:
  Host: localhost:6379
  Pass: ""
  DB: 0
```

### 3. 服务发现
```go
// 使用etcd进行服务发现
import "github.com/zeromicro/go-zero/core/discov"

func main() {
    // 服务注册
    discov.RegisterService("auth-rpc", "localhost:9001")
    
    // 服务发现
    client := discov.NewEtcdClient([]string{"localhost:2379"})
    services, err := client.GetServices("auth-rpc")
}
```

## ✅ 总结

Go-Zero框架提供了完整的微服务解决方案：

1. **API定义**：使用.api文件定义REST API
2. **RPC服务**：使用.proto文件定义RPC服务
3. **数据模型**：自动生成数据库操作代码
4. **代码生成**：使用goctl工具自动生成代码
5. **中间件支持**：内置多种中间件
6. **监控追踪**：支持Prometheus和Jaeger
7. **服务发现**：支持etcd服务发现
8. **配置管理**：支持多环境配置

通过使用Go-Zero框架，可以大大提高开发效率，确保代码质量和一致性。
