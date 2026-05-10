// Package ssh provides SSH service for gorp framework.
// Supports remote command execution through SSH connections.
// Configurable hosts, authentication (password/key), and known_hosts verification.
//
// SSH 包提供 SSH 服务，用于 gorp 框架。
// 支持通过 SSH 连接执行远程命令。
// 可配置主机、认证（密码/密钥）、known_hosts 验证。
package ssh

import (
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers SSH service.
// Core logic: Bind SSH service factory to container.
//
// Provider 注册 SSH 服务。
// 核心逻辑：将 SSH 服务工厂绑定到容器。
type Provider struct{}

// NewProvider creates a new SSH provider.
//
// NewProvider 创建新的 SSH provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string       { return "ssh" }

// IsDefer indicates SSH provider should defer loading.
// SSH connections are typically established on-demand.
//
// IsDefer 表示 SSH provider 应延迟加载。
// SSH 连接通常按需建立。
func (p *Provider) IsDefer() bool      { return true }

// Provides returns the capability keys this provider exposes.
// Exposes SSHKey for SSH service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 SSHKey 用于 SSH 服务。
func (p *Provider) Provides() []string { return []string{integrationcontract.SSHKey} }

// Register binds the SSH service factory to the container.
// Core logic: Create Service instance, bind to container.
//
// Register 将 SSH 服务工厂绑定到容器。
// 核心逻辑：创建 Service 实例、绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.SSHKey, func(c runtimecontract.Container) (any, error) {
		return NewService(c)
	}, true)
	return nil
}

// Boot initializes the SSH provider.
// No additional startup logic required.
//
// Boot 初始化 SSH provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
