// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements the provider registration for WebSocket capability.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件实现 WebSocket 能力的 provider 注册。
package websocket

import (
	"context"
	"runtime"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	// TextMessage indicates a text message frame.
	// TextMessage 表示文本消息帧。
	TextMessage = 1

	// BinaryMessage indicates a binary message frame.
	// BinaryMessage 表示二进制消息帧。
	BinaryMessage = 2

	// CloseMessage indicates a close frame.
	// CloseMessage 表示关闭帧。
	CloseMessage = 8
)

// Provider provides WebSocket capability.
//
// Provider 提供 WebSocket 能力。
type Provider struct {
	config        *transportcontract.WebSocketConfig
	clusterConfig *transportcontract.WebSocketClusterConfig
}

// NewProvider creates a new WebSocket provider with default config.
//
// NewProvider 使用默认配置创建新的 WebSocket provider。
func NewProvider() *Provider {
	return NewProviderWithConfig(nil)
}

// NewProviderWithConfig creates a new WebSocket provider with custom config.
//
// NewProviderWithConfig 使用自定义配置创建新的 WebSocket provider。
func NewProviderWithConfig(config *transportcontract.WebSocketConfig) *Provider {
	if config == nil {
		config = &transportcontract.WebSocketConfig{
			ParallelEnabled: true,
		}
	}
	return &Provider{config: config}
}

// NewProviderWithCluster creates a new WebSocket provider with cluster support.
//
// NewProviderWithCluster 创建支持集群的 WebSocket provider。
func NewProviderWithCluster(config *transportcontract.WebSocketConfig, clusterConfig *transportcontract.WebSocketClusterConfig) *Provider {
	if config == nil {
		config = &transportcontract.WebSocketConfig{
			ParallelEnabled: true,
		}
	}
	if clusterConfig == nil {
		clusterConfig = &transportcontract.WebSocketClusterConfig{Enabled: false}
	}
	return &Provider{config: config, clusterConfig: clusterConfig}
}

// Name returns the provider name.
//
// Name 返回 provider 名称。
func (p *Provider) Name() string {
	return "websocket"
}

// Register registers WebSocket service into container.
// Also registers a closer to gracefully shutdown connections on container destroy.
//
// Register 将 WebSocket 服务注册到容器。
// 同时注册 closer，在容器销毁时优雅关闭连接。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.WebSocketKey, func(c runtimecontract.Container) (any, error) {
		// Use cluster server if cluster config is provided and enabled
		// 如果提供了集群配置且启用则使用集群服务器
		if p.clusterConfig != nil && p.clusterConfig.Enabled {
			server, err := NewClusterServer(p.config, p.clusterConfig)
			if err != nil {
				return nil, err
			}
			return server, nil
		}
		return NewServerWithConfig(p.config), nil
	}, true)
	c.RegisterCloser(transportcontract.WebSocketKey, &websocketCloser{c: c})
	return nil
}

// Boot initializes the WebSocket service.
//
// Boot 初始化 WebSocket 服务。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// IsDefer reports whether the provider should be deferred.
//
// IsDefer 返回该 provider 是否应延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides returns the keys that this provider provides.
//
// Provides 返回该 provider 提供的 key 列表。
func (p *Provider) Provides() []string {
	return []string{string(transportcontract.WebSocketKey)}
}

// DependsOn returns the keys that this provider depends on.
//
// DependsOn 返回该 provider 依赖的 key 列表。
func (p *Provider) DependsOn() []string {
	return nil
}

// DefaultParallelGolimit returns the default parallel goroutine limit.
//
// DefaultParallelGolimit 返回默认并行 goroutine 限制。
func DefaultParallelGolimit() int {
	return runtime.NumCPU()
}

// websocketCloser implements io.Closer to gracefully shutdown WebSocket server on container destroy.
//
// websocketCloser 实现 io.Closer，在容器销毁时优雅关闭 WebSocket 服务器。
type websocketCloser struct {
	c runtimecontract.Container
}

func (wc *websocketCloser) Close() error {
	serverAny, err := wc.c.Make(transportcontract.WebSocketKey)
	if err != nil {
		return nil
	}
	// 先尝试 ClusterServer，再尝试 Server
	// ClusterServer 嵌入了 Server，所以可以调用 Shutdown
	// Try ClusterServer first, then Server
	// ClusterServer embeds Server, so Shutdown is available
	if cluster, ok := serverAny.(*ClusterServer); ok {
		return cluster.Shutdown(context.Background())
	}
	if server, ok := serverAny.(*Server); ok {
		return server.Shutdown(context.Background())
	}
	return nil
}
