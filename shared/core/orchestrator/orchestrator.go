package orchestrator

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Orchestrator 服务编排器
type Orchestrator struct {
	resolver       *DependencyResolver
	projectRoot    string
	logDir         string
	runningServices map[string]*ServiceProcess
}

// ServiceProcess 服务进程信息
type ServiceProcess struct {
	Name      string
	Port      int
	PID       int
	StartTime time.Time
	Status    string
}

// NewOrchestrator 创建服务编排器
func NewOrchestrator(projectRoot string) (*Orchestrator, error) {
	resolver, err := NewDependencyResolver()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(projectRoot, "logs")
	
	return &Orchestrator{
		resolver:        resolver,
		projectRoot:     projectRoot,
		logDir:          logDir,
		runningServices: make(map[string]*ServiceProcess),
	}, nil
}

// StartServices 启动服务（支持服务名列表或组合名）
func (o *Orchestrator) StartServices(targets []string, composition string) error {
	var serviceNames []string

	// 如果指定了组合，使用组合配置
	if composition != "" {
		comp, err := o.resolver.GetComposition(composition)
		if err != nil {
			return err
		}
		serviceNames = comp.TargetServices
	} else {
		// 否则使用指定的服务名列表
		serviceNames = targets
	}

	// 解析依赖
	allServices, err := o.resolver.ResolveDependencies(serviceNames)
	if err != nil {
		return err
	}

	// 按依赖关系排序
	sortedServices, err := o.resolver.SortServicesByDependencies(allServices)
	if err != nil {
		return err
	}

	// 打印启动计划
	log.Println("=" + "=" + "=" + "=" + "=" + "=")
	log.Println("智能中央大脑 - 服务编排计划")
	log.Println("=" + "=" + "=" + "=" + "=" + "=")
	log.Printf("目标服务: %v", serviceNames)
	log.Printf("完整服务列表: %v", allServices)
	log.Printf("启动顺序: %v", sortedServices)
	log.Println("=" + "=" + "=" + "=" + "=" + "=")

	// 启动服务
	for _, serviceName := range sortedServices {
		if err := o.StartService(serviceName); err != nil {
			log.Printf("启动服务 %s 失败: %v", serviceName, err)
			return err
		}
		// 等待服务启动
		time.Sleep(3 * time.Second)
	}

	// 健康检查
	if err := o.HealthCheck(allServices); err != nil {
		return err
	}

	return nil
}

// StartService 启动单个服务
func (o *Orchestrator) StartService(serviceName string) error {
	service, err := o.resolver.GetService(serviceName)
	if err != nil {
		return err
	}

	// 检查服务是否已经在运行
	if _, exists := o.runningServices[serviceName]; exists {
		log.Printf("服务 %s 已在运行，跳过", serviceName)
		return nil
	}

	log.Printf("正在启动服务: %s (端口: %d)", service.DisplayName, service.Port)

	// 根据服务类型选择启动方式
	servicePath := o.getServicePath(serviceName)
	
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = servicePath
	
	// 设置环境变量
	cmd.Env = o.getServiceEnv()
	
	// 输出到日志文件
	logFile := filepath.Join(o.logDir, serviceName+".log")
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}
	
	cmd.Stdout = file
	cmd.Stderr = file
	
	// 启动服务
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	// 保存进程信息
	o.runningServices[serviceName] = &ServiceProcess{
		Name:      serviceName,
		Port:      service.Port,
		PID:       cmd.Process.Pid,
		StartTime: time.Now(),
		Status:    "running",
	}

	log.Printf("✓ 服务 %s 启动成功 (PID: %d)", service.DisplayName, cmd.Process.Pid)
	return nil
}

// StopService 停止单个服务
func (o *Orchestrator) StopService(serviceName string) error {
	proc, exists := o.runningServices[serviceName]
	if !exists {
		return fmt.Errorf("服务 %s 未运行", serviceName)
	}

	log.Printf("正在停止服务: %s (PID: %d)", serviceName, proc.PID)
	
	// 使用 kill 命令停止进程
	cmd := exec.Command("kill", fmt.Sprintf("%d", proc.PID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("停止服务失败: %w", err)
	}

	delete(o.runningServices, serviceName)
	log.Printf("✓ 服务 %s 已停止", serviceName)
	return nil
}

// HealthCheck 健康检查
func (o *Orchestrator) HealthCheck(serviceNames []string) error {
	log.Println("执行健康检查...")
	
	for _, serviceName := range serviceNames {
		service, err := o.resolver.GetService(serviceName)
		if err != nil {
			continue
		}

		// 检查端口是否被占用（简单的健康检查）
		healthURL := fmt.Sprintf("http://localhost:%d/health", service.Port)
		
		cmd := exec.Command("curl", "-s", healthURL)
		if err := cmd.Run(); err != nil {
			log.Printf("⚠ 服务 %s 健康检查失败", service.DisplayName)
		} else {
			log.Printf("✓ 服务 %s 健康检查通过", service.DisplayName)
		}
	}

	return nil
}

// getServicePath 获取服务路径
func (o *Orchestrator) getServicePath(serviceName string) string {
	// 根据服务名返回对应的路径
	paths := map[string]string{
		"auth-service":    filepath.Join(o.projectRoot, "services/core/auth"),
		"user-service":    filepath.Join(o.projectRoot, "services/core/user"),
		"job-service":     filepath.Join(o.projectRoot, "services/business/job"),
		"resume-service":  filepath.Join(o.projectRoot, "services/business/resume"),
		"company-service": filepath.Join(o.projectRoot, "services/business/company"),
	}

	path, exists := paths[serviceName]
	if !exists {
		return filepath.Join(o.projectRoot, "services", serviceName)
	}

	return path
}

// getServiceEnv 获取服务环境变量
func (o *Orchestrator) getServiceEnv() []string {
	// 从环境变量中读取配置
	env := os.Environ()
	
	// 添加数据库环境变量
	env = append(env, "MYSQL_HOST=localhost")
	env = append(env, "POSTGRESQL_HOST=localhost")
	env = append(env, "REDIS_HOST=localhost")
	
	return env
}

// GetRunningServices 获取运行中的服务
func (o *Orchestrator) GetRunningServices() map[string]*ServiceProcess {
	return o.runningServices
}

