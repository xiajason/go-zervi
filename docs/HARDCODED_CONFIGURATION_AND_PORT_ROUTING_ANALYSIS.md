# 硬编码问题与端口路由架构分析

## 📋 问题概述

**问题**: 要改变硬编码问题，不仅仅是简单地读取配置文件，而是涉及到**内部三者之间的端口路由设置**。

**核心发现**: 硬编码问题涉及多个层面的架构设计：
1. **服务地址和端口配置**
2. **服务发现和动态路由**
3. **服务凭证管理**
4. **三者之间的协调机制**

---

## 🔍 当前硬编码问题分析

### 1. **Central Brain 的硬编码问题**

#### 问题1: 服务地址硬编码 ⚠️

**当前实现**:
```go
// shared/central-brain/centralbrain.go:41
authServiceURL: fmt.Sprintf("http://localhost:%d", config.AuthServicePort)

// shared/central-brain/centralbrain.go:78-114
services := map[string]ServiceProxy{
    "auth": {
        ServiceName: "auth-service",
        BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.AuthServicePort),  // ⚠️ 硬编码localhost
        PathPrefix:  "/api/v1/auth",
    },
    "ai": {
        ServiceName: "ai-service",
        BaseURL:     fmt.Sprintf("http://localhost:%d", cb.config.AIServicePort),
        PathPrefix:  "/api/v1/ai",
    },
    // ... 其他服务也是硬编码localhost
}
```

**问题**:
- ❌ **硬编码 `localhost`**: 无法支持Docker网络、Kubernetes等环境
- ❌ **硬编码端口**: 所有服务端口都在代码中硬编码
- ❌ **无法动态发现**: 不能从Consul动态获取服务地址

**影响**:
- 无法在不同环境（开发、测试、生产）灵活部署
- 无法支持服务动态扩容和故障转移
- 无法支持Docker Compose网络或其他容器网络

---

#### 问题2: 服务凭证硬编码 ⚠️

**当前实现**:
```go
// shared/central-brain/centralbrain.go:286-287
serviceID := "central-brain"
serviceSecret := "central-brain-secret-2025" // ⚠️ 临时硬编码，应该从配置读取
```

**问题**:
- ❌ **硬编码服务ID**: 无法灵活配置
- ❌ **硬编码服务密钥**: 安全风险，密钥暴露在代码中

**影响**:
- 安全性问题：密钥暴露在代码中
- 无法支持密钥轮换
- 无法支持多环境（不同环境使用不同密钥）

---

#### 问题3: 配置结构硬编码 ⚠️

**当前实现**:
```go
// shared/core/shared/config.go:16-27
func GetDefaultConfig() *Config {
    return &Config{
        CentralBrainPort:      9000,  // ⚠️ 硬编码默认值
        AuthServicePort:       8207,  // ⚠️ 硬编码默认值
        AIServicePort:         8100,  // ⚠️ 硬编码默认值
        BlockchainServicePort: 8208,  // ⚠️ 硬编码默认值
        UserServicePort:       8082,  // ⚠️ 硬编码默认值
        JobServicePort:        8084,  // ⚠️ 硬编码默认值
        ResumeServicePort:     8085,  // ⚠️ 硬编码默认值
        CompanyServicePort:    8083,  // ⚠️ 硬编码默认值
    }
}
```

**问题**:
- ❌ **所有端口硬编码**: 没有从环境变量读取
- ❌ **没有配置优先级**: 无法覆盖默认值
- ❌ **没有配置验证**: 无法验证配置的有效性

---

### 2. **Auth Service 的硬编码问题**

#### 问题1: 数据库连接硬编码 ⚠️

**当前实现**:
```go
// services/core/auth/main.go:16-19
dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
    dbURL = "postgres://szjason72@localhost:5432/zervigo_mvp?sslmode=disable"  // ⚠️ 硬编码默认值
}
```

**问题**:
- ❌ **硬编码数据库地址**: `localhost` 和端口 `5432`
- ❌ **硬编码数据库名**: `zervigo_mvp`
- ❌ **硬编码用户名**: `szjason72`

---

#### 问题2: 服务端口硬编码 ⚠️

**当前实现**:
```go
// services/core/auth/main.go:50-53
port := 8207  // ⚠️ 硬编码默认值
if portEnv := os.Getenv("AUTH_SERVICE_PORT"); portEnv != "" {
    fmt.Sscanf(portEnv, "%d", &port)
}
```

**问题**:
- ⚠️ **有环境变量支持，但默认值硬编码**
- ❌ **没有配置验证**: 无法验证端口是否可用

---

### 3. **三者之间的端口路由协调问题** 🔴

#### 问题1: 端口路由不一致 ⚠️

**当前情况**:
```
Consul (8500)
  ↓ 不知道服务端口
Central Brain (9000)
  ↓ 硬编码所有服务端口
Auth Service (8207)
  ↓ 自己知道自己的端口
PostgreSQL (5432)
```

**问题**:
- ❌ **Consul不知道服务端口**: 服务注册时端口信息可能不一致
- ❌ **Central Brain硬编码端口**: 无法动态适应端口变更
- ❌ **Auth Service端口独立配置**: 可能与Central Brain的配置不一致

**端口路由映射**:
```
Central Brain配置:
  AuthServicePort: 8207  → 硬编码
  UserServicePort: 8082  → 硬编码
  ...

Auth Service实际:
  AUTH_SERVICE_PORT: 8207  → 环境变量（如果设置）

问题: 如果两者不一致，路由失败！
```

---

#### 问题2: 服务发现缺失 ⚠️

**当前情况**:
```
Central Brain → 硬编码地址 → Auth Service
             ↓
         无法从Consul发现
```

**问题**:
- ❌ **无法动态发现**: Central Brain无法从Consul获取服务地址
- ❌ **无法负载均衡**: 无法支持多实例服务
- ❌ **无法故障转移**: 服务故障时无法自动切换

---

## 🏗️ 解决硬编码问题的完整架构设计

### **方案设计：三层配置体系**

```
┌─────────────────────────────────────────────────────────┐
│              配置层 (Configuration Layer)                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ 环境变量     │  │ 配置文件     │  │ 默认值       │ │
│  │ (优先级最高) │  │ (优先级中)   │  │ (优先级低)   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              服务发现层 (Service Discovery Layer)        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Consul     │  │ 服务注册     │  │ 动态发现     │ │
│  │ (服务地址)   │  │ (端口信息)   │  │ (路由更新)   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              路由层 (Routing Layer)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │Central Brain │  │ 服务代理     │  │ 负载均衡     │ │
│  │ (路由决策)   │  │ (请求转发)   │  │ (实例选择)   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

---

## 📋 详细解决方案

### **第一阶段：配置管理重构**

#### 1.1 统一配置结构

**目标**: 创建一个统一的配置管理机制

**实现**:
```go
// shared/core/shared/config.go
type Config struct {
    // 服务端口配置（支持环境变量）
    CentralBrainPort      int `env:"CENTRAL_BRAIN_PORT" default:"9000"`
    AuthServicePort       int `env:"AUTH_SERVICE_PORT" default:"8207"`
    AIServicePort         int `env:"AI_SERVICE_PORT" default:"8100"`
    // ... 其他服务端口
    
    // 服务地址配置（支持动态发现）
    ServiceDiscovery struct {
        Enabled    bool   `env:"SERVICE_DISCOVERY_ENABLED" default:"false"`
        ConsulURL  string `env:"CONSUL_URL" default:"http://localhost:8500"`
        ServiceHost string `env:"SERVICE_HOST" default:"localhost"`  // 本地开发用localhost，Docker用服务名
    }
    
    // 服务凭证配置（从环境变量或配置文件读取）
    ServiceCredentials struct {
        ServiceID     string `env:"SERVICE_ID" default:"central-brain"`
        ServiceSecret string `env:"SERVICE_SECRET"`  // 必须从环境变量读取，不能有默认值
    }
    
    // 数据库配置
    Database struct {
        Host     string `env:"DB_HOST" default:"localhost"`
        Port     int    `env:"DB_PORT" default:"5432"`
        Database string `env:"DB_NAME" default:"zervigo_mvp"`
        Username string `env:"DB_USER"`
        Password string `env:"DB_PASSWORD"`
    }
}
```

**优先级**:
1. **环境变量** (最高优先级)
2. **配置文件** (中等优先级)
3. **默认值** (最低优先级)

---

#### 1.2 配置加载机制

**实现**:
```go
// shared/core/shared/config.go
func LoadConfig(configPath string) (*Config, error) {
    config := &Config{}
    
    // 1. 设置默认值
    setDefaults(config)
    
    // 2. 从配置文件加载（如果存在）
    if configPath != "" {
        if err := loadFromFile(config, configPath); err != nil {
            return nil, fmt.Errorf("加载配置文件失败: %w", err)
        }
    }
    
    // 3. 从环境变量覆盖（最高优先级）
    loadFromEnv(config)
    
    // 4. 验证配置
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }
    
    return config, nil
}
```

---

### **第二阶段：服务发现集成**

#### 2.1 Consul服务发现集成

**目标**: Central Brain从Consul动态获取服务地址

**实现**:
```go
// shared/central-brain/service_discovery.go
type ServiceDiscovery struct {
    consulClient *consul.Client
    serviceCache map[string][]ServiceInstance  // 缓存服务实例
    updateCh     chan struct{}
}

type ServiceInstance struct {
    ServiceID   string
    ServiceName string
    Address     string  // 从Consul获取，不是硬编码localhost
    Port        int     // 从Consul获取，不是硬编码
    Healthy     bool
    LastSeen    time.Time
}

// 从Consul获取服务实例
func (sd *ServiceDiscovery) GetServiceInstances(serviceName string) ([]ServiceInstance, error) {
    // 1. 从Consul查询服务
    services, _, err := sd.consulClient.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return nil, fmt.Errorf("从Consul获取服务失败: %w", err)
    }
    
    // 2. 转换为ServiceInstance
    instances := make([]ServiceInstance, 0)
    for _, service := range services {
        instance := ServiceInstance{
            ServiceID:   service.Service.ID,
            ServiceName: service.Service.Service,
            Address:     service.Service.Address,  // ✅ 从Consul获取，不是硬编码
            Port:        service.Service.Port,      // ✅ 从Consul获取，不是硬编码
            Healthy:     isHealthy(service),
            LastSeen:    time.Now(),
        }
        instances = append(instances, instance)
    }
    
    // 3. 更新缓存
    sd.serviceCache[serviceName] = instances
    
    return instances, nil
}

// 动态更新服务路由
func (sd *ServiceDiscovery) WatchServices() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // 定期从Consul更新服务实例
            for serviceName := range sd.serviceCache {
                instances, err := sd.GetServiceInstances(serviceName)
                if err == nil {
                    sd.serviceCache[serviceName] = instances
                    sd.updateCh <- struct{}{}  // 通知路由更新
                }
            }
        }
    }
}
```

---

#### 2.2 动态路由更新

**实现**:
```go
// shared/central-brain/centralbrain.go
func (cb *CentralBrain) registerServiceProxies() {
    // 如果启用了服务发现，从Consul获取服务地址
    if cb.config.ServiceDiscovery.Enabled {
        cb.registerServiceProxiesFromConsul()
    } else {
        // 回退到硬编码配置（向后兼容）
        cb.registerServiceProxiesFromConfig()
    }
}

func (cb *CentralBrain) registerServiceProxiesFromConsul() {
    // 1. 从Consul发现服务
    authInstances, err := cb.serviceDiscovery.GetServiceInstances("auth-service")
    if err != nil || len(authInstances) == 0 {
        log.Printf("⚠️ 无法从Consul发现auth-service，使用配置中的地址")
        cb.registerServiceProxiesFromConfig()
        return
    }
    
    // 2. 选择健康的服务实例（负载均衡）
    authInstance := cb.selectHealthyInstance(authInstances)
    
    // 3. 构建服务代理（使用Consul中的地址和端口）
    services := map[string]ServiceProxy{
        "auth": {
            ServiceName: "auth-service",
            BaseURL:     fmt.Sprintf("http://%s:%d", authInstance.Address, authInstance.Port),  // ✅ 从Consul获取
            PathPrefix:  "/api/v1/auth",
        },
        // ... 其他服务
    }
    
    // 4. 注册服务代理
    for serviceKey, service := range services {
        cb.registerServiceProxy(serviceKey, service)
    }
    
    // 5. 启动服务监控（动态更新路由）
    go cb.watchServiceUpdates()
}
```

---

#### 2.3 负载均衡实现

**实现**:
```go
// shared/central-brain/load_balancer.go
type LoadBalancer struct {
    strategy LoadBalanceStrategy
}

type LoadBalanceStrategy interface {
    Select(instances []ServiceInstance) ServiceInstance
}

// 轮询策略
type RoundRobinStrategy struct {
    current int
    mutex   sync.Mutex
}

func (rr *RoundRobinStrategy) Select(instances []ServiceInstance) ServiceInstance {
    rr.mutex.Lock()
    defer rr.mutex.Unlock()
    
    // 过滤健康实例
    healthyInstances := filterHealthy(instances)
    if len(healthyInstances) == 0 {
        return instances[0]  // 如果没有健康实例，返回第一个
    }
    
    instance := healthyInstances[rr.current%len(healthyInstances)]
    rr.current++
    return instance
}

// 健康检查路由
func filterHealthy(instances []ServiceInstance) []ServiceInstance {
    healthy := make([]ServiceInstance, 0)
    for _, instance := range instances {
        if instance.Healthy {
            healthy = append(healthy, instance)
        }
    }
    return healthy
}
```

---

### **第三阶段：服务凭证管理**

#### 3.1 服务凭证配置化

**实现**:
```go
// shared/central-brain/centralbrain.go
func (cb *CentralBrain) requestServiceToken() (string, error) {
    // ✅ 从配置读取服务凭证（不再硬编码）
    serviceID := cb.config.ServiceCredentials.ServiceID
    serviceSecret := cb.config.ServiceCredentials.ServiceSecret
    
    // 验证服务凭证是否配置
    if serviceID == "" {
        return "", fmt.Errorf("SERVICE_ID未配置")
    }
    if serviceSecret == "" {
        return "", fmt.Errorf("SERVICE_SECRET未配置（必须从环境变量设置）")
    }
    
    // 调用Auth Service获取服务token
    // ...
}
```

**环境变量配置**:
```bash
# .env 文件（不提交到Git）
SERVICE_ID=central-brain
SERVICE_SECRET=central-brain-secret-2025  # 从安全的密钥管理系统读取

# Docker Compose环境变量
environment:
  - SERVICE_ID=central-brain
  - SERVICE_SECRET=${CENTRAL_BRAIN_SECRET}  # 从外部配置读取
```

---

#### 3.2 密钥管理集成

**方案**: 集成密钥管理系统（如Vault、AWS Secrets Manager等）

**实现**:
```go
// shared/core/secret/manager.go
type SecretManager interface {
    GetSecret(key string) (string, error)
}

// 从Vault获取密钥
type VaultSecretManager struct {
    client *vault.Client
}

func (vsm *VaultSecretManager) GetSecret(key string) (string, error) {
    secret, err := vsm.client.Logical().Read(fmt.Sprintf("secret/data/%s", key))
    if err != nil {
        return "", err
    }
    return secret.Data["value"].(string), nil
}

// 使用密钥管理器
func LoadConfigWithSecrets(configPath string) (*Config, error) {
    config, err := LoadConfig(configPath)
    if err != nil {
        return nil, err
    }
    
    // 从密钥管理器获取敏感信息
    secretManager := NewSecretManager()
    
    if config.ServiceCredentials.ServiceSecret == "" {
        secret, err := secretManager.GetSecret("central-brain-secret")
        if err != nil {
            return nil, fmt.Errorf("获取服务密钥失败: %w", err)
        }
        config.ServiceCredentials.ServiceSecret = secret
    }
    
    return config, nil
}
```

---

### **第四阶段：三者协调机制**

#### 4.1 服务注册标准化

**目标**: 确保所有服务注册到Consul时使用一致的端口信息

**实现**:
```go
// shared/core/service/registry/consul.go
func RegisterToConsul(serviceName string, port int, healthCheckURL string) error {
    // 1. 从环境变量获取服务地址（支持Docker网络）
    serviceHost := os.Getenv("SERVICE_HOST")
    if serviceHost == "" {
        serviceHost = "localhost"  // 本地开发默认值
    }
    
    // 2. 从环境变量获取服务端口（如果设置）
    if envPort := os.Getenv(fmt.Sprintf("%s_PORT", strings.ToUpper(serviceName))); envPort != "" {
        if p, err := strconv.Atoi(envPort); err == nil {
            port = p  // 使用环境变量中的端口
        }
    }
    
    // 3. 注册到Consul（使用实际的服务地址和端口）
    registration := &api.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%d", serviceName, port),
        Name:    serviceName,
        Tags:    []string{"zervigo", "microservice"},
        Port:    port,        // ✅ 使用实际配置的端口
        Address: serviceHost, // ✅ 使用实际配置的地址（支持Docker网络）
        Check: &api.AgentServiceCheck{
            HTTP:  fmt.Sprintf("http://%s:%d/health", serviceHost, port),
            Interval: "10s",
            Timeout: "3s",
        },
    }
    
    return client.Agent().ServiceRegister(registration)
}
```

---

#### 4.2 Central Brain服务发现集成

**实现**:
```go
// shared/central-brain/centralbrain.go
func NewCentralBrain(config *shared.Config) *CentralBrain {
    cb := &CentralBrain{
        config: config,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        router: gin.Default(),
    }
    
    // ✅ 如果启用服务发现，初始化Consul客户端
    if config.ServiceDiscovery.Enabled {
        consulClient, err := consul.NewClient(&consul.Config{
            Address: config.ServiceDiscovery.ConsulURL,
        })
        if err != nil {
            log.Printf("⚠️ Consul客户端初始化失败: %v，将使用配置中的地址", err)
            cb.authServiceURL = fmt.Sprintf("http://%s:%d", 
                config.ServiceDiscovery.ServiceHost, 
                config.AuthServicePort)
        } else {
            cb.serviceDiscovery = NewServiceDiscovery(consulClient)
            // ✅ 从Consul获取Auth Service地址（动态）
            go cb.updateAuthServiceURL()
        }
    } else {
        // 回退到配置中的地址
        cb.authServiceURL = fmt.Sprintf("http://%s:%d", 
            config.ServiceDiscovery.ServiceHost, 
            config.AuthServicePort)
    }
    
    return cb
}

// 动态更新Auth Service URL
func (cb *CentralBrain) updateAuthServiceURL() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            instances, err := cb.serviceDiscovery.GetServiceInstances("auth-service")
            if err == nil && len(instances) > 0 {
                instance := cb.selectHealthyInstance(instances)
                newURL := fmt.Sprintf("http://%s:%d", instance.Address, instance.Port)
                if newURL != cb.authServiceURL {
                    log.Printf("🔄 更新Auth Service URL: %s -> %s", cb.authServiceURL, newURL)
                    cb.authServiceURL = newURL
                }
            }
        }
    }
}
```

---

## 📊 端口路由协调机制

### **协调流程**

```
┌─────────────────────────────────────────────────────────┐
│              1. 服务启动阶段                              │
│                                                          │
│  Auth Service启动:                                       │
│    - 读取环境变量 AUTH_SERVICE_PORT (默认8207)          │
│    - 启动在端口 8207                                     │
│    - 注册到Consul:                                      │
│        service_id: "auth-service-8207"                  │
│        address: SERVICE_HOST (或localhost)               │
│        port: 8207                                       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              2. Consul服务注册阶段                        │
│                                                          │
│  Consul记录:                                             │
│    auth-service:                                        │
│      - Instance 1: localhost:8207 (healthy)             │
│      - Instance 2: auth-service:8207 (healthy) [Docker]  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              3. Central Brain发现阶段                     │
│                                                          │
│  Central Brain启动:                                       │
│    - 读取配置 SERVICE_DISCOVERY_ENABLED=true            │
│    - 连接到Consul: http://localhost:8500                │
│    - 查询服务: "auth-service"                           │
│    - 获取实例: [{address: "localhost", port: 8207}]     │
│    - 构建URL: http://localhost:8207                      │
│    - 注册路由: /api/v1/auth/** → http://localhost:8207  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              4. 动态更新阶段                              │
│                                                          │
│  Central Brain监控:                                      │
│    - 每30秒从Consul查询服务实例                          │
│    - 如果端口变更，自动更新路由                          │
│    - 如果服务故障，切换到健康实例                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🎯 配置优先级和协调机制

### **配置优先级**

```
1. 环境变量 (最高优先级)
   ├── AUTH_SERVICE_PORT=8207
   ├── SERVICE_HOST=localhost (或Docker服务名)
   └── SERVICE_SECRET=xxx
   
2. Consul配置 (如果启用服务发现)
   ├── 服务地址: 从Consul获取
   └── 服务端口: 从Consul获取
   
3. 配置文件 (中等优先级)
   ├── config.yaml
   └── config.json
   
4. 默认值 (最低优先级)
   ├── AuthServicePort: 8207
   └── ServiceHost: localhost
```

### **协调机制**

**问题**: 如何确保三者之间的端口路由一致？

**解决方案**:

1. **统一配置源**
   - 所有服务从同一个配置源读取端口配置
   - 使用环境变量作为单一配置源

2. **服务发现协调**
   - Auth Service注册到Consul时使用实际端口
   - Central Brain从Consul获取端口（不是硬编码）
   - 确保两者一致

3. **健康检查机制**
   - Central Brain定期检查服务健康状态
   - 如果端口变更，自动更新路由

4. **配置验证**
   - 启动时验证配置的一致性
   - 如果发现不一致，记录警告或错误

---

## 📋 实施步骤

### **阶段1: 配置管理重构** (1-2天)

1. ✅ 创建统一配置结构
2. ✅ 实现配置加载机制（环境变量 → 配置文件 → 默认值）
3. ✅ 添加配置验证
4. ✅ 更新Central Brain使用新配置

### **阶段2: 服务发现集成** (2-3天)

1. ✅ 实现Consul服务发现客户端
2. ✅ Central Brain集成Consul服务发现
3. ✅ 实现动态路由更新
4. ✅ 实现负载均衡

### **阶段3: 服务凭证管理** (1天)

1. ✅ 从环境变量读取服务凭证
2. ✅ 移除硬编码的服务凭证
3. ✅ 添加密钥管理集成（可选）

### **阶段4: 三者协调** (1-2天)

1. ✅ 统一服务注册机制
2. ✅ 实现配置一致性验证
3. ✅ 添加健康检查和自动恢复

---

## 🎯 关键实现点

### **1. 服务地址动态获取**

**当前（硬编码）**:
```go
BaseURL: fmt.Sprintf("http://localhost:%d", cb.config.AuthServicePort)
```

**改进后（动态发现）**:
```go
// 从Consul获取服务实例
instances, _ := consulClient.GetServiceInstances("auth-service")
instance := selectHealthyInstance(instances)
BaseURL: fmt.Sprintf("http://%s:%d", instance.Address, instance.Port)
```

---

### **2. 端口路由协调**

**配置一致性检查**:
```go
func ValidateServicePorts(config *Config) error {
    // 1. 检查环境变量
    authPort := os.Getenv("AUTH_SERVICE_PORT")
    if authPort != "" && config.AuthServicePort != parsePort(authPort) {
        return fmt.Errorf("配置不一致: 环境变量AUTH_SERVICE_PORT=%s, 配置AuthServicePort=%d", 
            authPort, config.AuthServicePort)
    }
    
    // 2. 检查Consul中的服务端口（如果启用服务发现）
    if config.ServiceDiscovery.Enabled {
        instances, _ := consulClient.GetServiceInstances("auth-service")
        if len(instances) > 0 {
            consulPort := instances[0].Port
            if config.AuthServicePort != consulPort {
                log.Printf("⚠️ 端口不一致: 配置=%d, Consul=%d，将使用Consul中的端口", 
                    config.AuthServicePort, consulPort)
                config.AuthServicePort = consulPort  // 使用Consul中的端口
            }
        }
    }
    
    return nil
}
```

---

### **3. 支持Docker网络**

**环境变量配置**:
```bash
# 本地开发
SERVICE_HOST=localhost
AUTH_SERVICE_PORT=8207

# Docker Compose
SERVICE_HOST=auth-service  # Docker服务名
AUTH_SERVICE_PORT=8207

# Kubernetes
SERVICE_HOST=auth-service.default.svc.cluster.local
AUTH_SERVICE_PORT=8207
```

**Central Brain自动适应**:
```go
// 根据环境自动选择服务地址
func getServiceHost() string {
    if host := os.Getenv("SERVICE_HOST"); host != "" {
        return host  // Docker/Kubernetes环境
    }
    return "localhost"  // 本地开发环境
}
```

---

## 🎯 总结

### **硬编码问题的本质**

不仅仅是简单的配置读取问题，而是涉及：

1. **配置管理架构**
   - 环境变量 → 配置文件 → 默认值的优先级
   - 配置验证和一致性检查

2. **服务发现机制**
   - 从Consul动态获取服务地址和端口
   - 动态路由更新和负载均衡

3. **三者协调机制**
   - Consul（服务注册）↔ Central Brain（服务发现）↔ Auth Service（服务提供）
   - 端口路由的一致性保证

4. **环境适配**
   - 本地开发（localhost）
   - Docker Compose（服务名）
   - Kubernetes（服务域名）

### **解决方案的核心**

**不是简单的"读取配置文件"**，而是：

1. ✅ **统一配置管理** - 三级优先级机制
2. ✅ **服务发现集成** - 从Consul动态获取
3. ✅ **动态路由更新** - 自动适应服务变更
4. ✅ **三者协调** - 确保配置一致性

**这是一个完整的微服务架构优化，而不仅仅是配置读取！**

---

**文档生成时间**: 2025-01-29  
**下次审查**: 实现服务发现集成后

