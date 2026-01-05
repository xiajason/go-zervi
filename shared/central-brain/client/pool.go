package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

// HTTPClientPool HTTP客户端连接池
type HTTPClientPool struct {
	clients       map[string]*http.Client
	mu            sync.RWMutex
	defaultClient *http.Client
}

// NewHTTPClientPool 创建HTTP客户端连接池
func NewHTTPClientPool() *HTTPClientPool {
	// 创建默认客户端（优化配置）
	defaultTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	defaultClient := &http.Client{
		Transport: defaultTransport,
		Timeout:   30 * time.Second,
	}

	return &HTTPClientPool{
		clients:       make(map[string]*http.Client),
		defaultClient: defaultClient,
	}
}

// GetClient 获取客户端（如果不存在则创建）
func (p *HTTPClientPool) GetClient(serviceName string) *http.Client {
	p.mu.RLock()
	client, exists := p.clients[serviceName]
	p.mu.RUnlock()

	if exists {
		return client
	}

	// 创建新客户端
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查
	client, exists = p.clients[serviceName]
	if exists {
		return client
	}

	// 创建服务专用客户端
	transport := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	p.clients[serviceName] = client
	return client
}

// GetDefaultClient 获取默认客户端
func (p *HTTPClientPool) GetDefaultClient() *http.Client {
	return p.defaultClient
}

// Close 关闭连接池
func (p *HTTPClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}
