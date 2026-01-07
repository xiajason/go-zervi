package main

import (
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceDiscovery Consul服务发现
type ServiceDiscovery struct {
	client *api.Client
	cache  map[string]bool // 服务可用性缓存
}

// NewServiceDiscovery 创建服务发现实例
func NewServiceDiscovery() *ServiceDiscovery {
	config := api.DefaultConfig()
	// Consul默认地址
	config.Address = "localhost:8500"

	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("⚠️  创建Consul客户端失败: %v，服务发现功能将不可用", err)
		return &ServiceDiscovery{
			client: nil,
			cache:  make(map[string]bool),
		}
	}

	sd := &ServiceDiscovery{
		client: client,
		cache:  make(map[string]bool),
	}

	log.Printf("✅ Consul服务发现已初始化")
	return sd
}

// GetAvailableServices 获取当前可用的服务列表
func (sd *ServiceDiscovery) GetAvailableServices() []string {
	if sd.client == nil {
		log.Printf("⚠️  Consul客户端未初始化，返回空服务列表")
		return []string{}
	}

	// 查询所有服务
	services, _, err := sd.client.Catalog().Services(nil)
	if err != nil {
		log.Printf("⚠️  查询Consul服务失败: %v", err)
		return []string{}
	}

	available := []string{}

	// P2业务服务列表
	p2Services := []string{
		"job-service",
		"resume-service",
		"company-service",
	}

	for _, serviceName := range p2Services {
		if _, exists := services[serviceName]; exists {
			// 检查服务健康状态
			if sd.IsServiceHealthy(serviceName) {
				available = append(available, serviceName)
				sd.cache[serviceName] = true
				log.Printf("✅ 发现可用服务: %s", serviceName)
			} else {
				sd.cache[serviceName] = false
				log.Printf("⚠️  服务不健康: %s", serviceName)
			}
		} else {
			sd.cache[serviceName] = false
		}
	}

	log.Printf("📊 当前可用的P2服务: %v", available)
	return available
}

// IsServiceHealthy 检查服务是否健康
func (sd *ServiceDiscovery) IsServiceHealthy(serviceName string) bool {
	if sd.client == nil {
		return false
	}

	// 查询服务健康状态
	health, _, err := sd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Printf("⚠️  查询服务健康状态失败 %s: %v", serviceName, err)
		return false
	}

	// 至少有一个健康的实例
	return len(health) > 0
}

// GetServiceCombination 获取当前的服务组合类型
func (sd *ServiceDiscovery) GetServiceCombination() string {
	available := sd.GetAvailableServices()

	hasJob := contains(available, "job-service")
	hasResume := contains(available, "resume-service")
	hasCompany := contains(available, "company-service")

	// 判断组合类型
	if hasJob && hasResume && hasCompany {
		return "all_services"
	} else if hasJob && hasResume {
		return "job_resume"
	} else if hasJob && hasCompany {
		return "job_company"
	} else if hasResume && hasCompany {
		return "resume_company"
	} else if hasJob {
		return "job_only"
	} else if hasResume {
		return "resume_only"
	} else if hasCompany {
		return "company_only"
	}

	return "minimal" // 只有基础设施
}

// RefreshCache 刷新服务缓存
func (sd *ServiceDiscovery) RefreshCache() {
	if sd.client == nil {
		return
	}

	log.Printf("🔄 刷新服务发现缓存...")
	sd.GetAvailableServices()
}

// StartAutoRefresh 启动自动刷新（后台任务）
func (sd *ServiceDiscovery) StartAutoRefresh(interval time.Duration) {
	if sd.client == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			sd.RefreshCache()
		}
	}()

	log.Printf("✅ 服务发现自动刷新已启动（间隔: %v）", interval)
}

// 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

