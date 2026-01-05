package orchestrator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// ServiceDependency 服务依赖定义
type ServiceDependency struct {
	Name         string   `yaml:"name"`
	DisplayName  string   `yaml:"display_name"`
	Port         int      `yaml:"port"`
	Type         string   `yaml:"type"`
	Dependencies []string `yaml:"dependencies"`
	Description  string   `yaml:"description"`
}

// ServiceComposition 服务组合定义
type ServiceComposition struct {
	Name             string   `yaml:"name"`
	DisplayName      string   `yaml:"display_name"`
	Description      string   `yaml:"description"`
	TargetServices   []string `yaml:"target_services"`
	AutoStartRequired bool   `yaml:"auto_start_required"`
	Enabled          bool    `yaml:"enabled"`
}

// ServiceDependenciesConfig 服务依赖配置
type ServiceDependenciesConfig struct {
	Services map[string]ServiceDependency `yaml:"services"`
}

// ServiceCompositionsConfig 服务组合配置
type ServiceCompositionsConfig struct {
	Compositions map[string]ServiceComposition `yaml:"compositions"`
}

// DependencyResolver 依赖解析器
type DependencyResolver struct {
	dependencies map[string]ServiceDependency
	compositions map[string]ServiceComposition
}

// NewDependencyResolver 创建依赖解析器
func NewDependencyResolver() (*DependencyResolver, error) {
	resolver := &DependencyResolver{
		dependencies: make(map[string]ServiceDependency),
		compositions: make(map[string]ServiceComposition),
	}

	// 加载服务依赖配置
	if err := resolver.loadDependencies(); err != nil {
		return nil, fmt.Errorf("加载服务依赖配置失败: %w", err)
	}

	// 加载服务组合配置
	if err := resolver.loadCompositions(); err != nil {
		return nil, fmt.Errorf("加载服务组合配置失败: %w", err)
	}

	return resolver, nil
}

// loadDependencies 加载服务依赖配置
func (dr *DependencyResolver) loadDependencies() error {
	configFile := "configs/service-dependencies.yaml"
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config ServiceDependenciesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	dr.dependencies = config.Services
	return nil
}

// loadCompositions 加载服务组合配置
func (dr *DependencyResolver) loadCompositions() error {
	configFile := "configs/service-compositions.yaml"
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config ServiceCompositionsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	dr.compositions = config.Compositions
	return nil
}

// ResolveDependencies 解析服务依赖
func (dr *DependencyResolver) ResolveDependencies(serviceNames []string) ([]string, error) {
	allServices := make(map[string]bool)
	
	// 递归解析依赖
	for _, serviceName := range serviceNames {
		if err := dr.resolveServiceDependencies(serviceName, allServices); err != nil {
			return nil, err
		}
	}

	// 转换为切片
	result := make([]string, 0, len(allServices))
	for service := range allServices {
		result = append(result, service)
	}

	return result, nil
}

// resolveServiceDependencies 递归解析单个服务的依赖
func (dr *DependencyResolver) resolveServiceDependencies(serviceName string, visited map[string]bool) error {
	// 检查服务是否存在
	service, exists := dr.dependencies[serviceName]
	if !exists {
		return fmt.Errorf("服务 %s 不存在", serviceName)
	}

	// 添加当前服务
	visited[serviceName] = true

	// 递归解析依赖
	for _, dep := range service.Dependencies {
		if !visited[dep] {
			if err := dr.resolveServiceDependencies(dep, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetComposition 获取服务组合
func (dr *DependencyResolver) GetComposition(name string) (*ServiceComposition, error) {
	composition, exists := dr.compositions[name]
	if !exists {
		return nil, fmt.Errorf("服务组合 %s 不存在", name)
	}

	if !composition.Enabled {
		return nil, fmt.Errorf("服务组合 %s 未启用", name)
	}

	return &composition, nil
}

// GetService 获取服务信息
func (dr *DependencyResolver) GetService(name string) (*ServiceDependency, error) {
	service, exists := dr.dependencies[name]
	if !exists {
		return nil, fmt.Errorf("服务 %s 不存在", name)
	}

	return &service, nil
}

// GetAllServices 获取所有服务
func (dr *DependencyResolver) GetAllServices() map[string]ServiceDependency {
	return dr.dependencies
}

// GetAllCompositions 获取所有组合
func (dr *DependencyResolver) GetAllCompositions() map[string]ServiceComposition {
	return dr.compositions
}

// SortServicesByDependencies 按依赖关系排序服务
func (dr *DependencyResolver) SortServicesByDependencies(services []string) ([]string, error) {
	// 使用拓扑排序
	sorted := make([]string, 0)
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	// 初始化
	for _, service := range services {
		inDegree[service] = 0
		graph[service] = []string{}
	}

	// 构建图和入度
	for _, service := range services {
		deps, exists := dr.dependencies[service]
		if !exists {
			continue
		}

		for _, dep := range deps.Dependencies {
			// 只考虑在services列表中的依赖
			for _, s := range services {
				if s == dep {
					graph[dep] = append(graph[dep], service)
					inDegree[service]++
				}
			}
		}
	}

	// 拓扑排序
	queue := []string{}
	for service, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, service)
		}
	}

	for len(queue) > 0 {
		service := queue[0]
		queue = queue[1:]
		sorted = append(sorted, service)

		for _, neighbor := range graph[service] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(services) {
		return nil, fmt.Errorf("检测到循环依赖")
	}

	return sorted, nil
}

