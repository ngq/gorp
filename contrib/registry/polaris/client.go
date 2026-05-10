// Package polaris provides Polaris SDK client wrapper.
// This file wraps the official Polaris SDK for service registry operations.
//
// 本包提供 Polaris SDK 客户端包装。
// 本文件包装官方 Polaris SDK 用于服务注册操作。
package polaris

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	polarissdk "github.com/polarismesh/polaris-go"
	polarismodel "github.com/polarismesh/polaris-go/pkg/model"
)

// officialPolarisRegistryClient wraps the official Polaris SDK client.
//
// officialPolarisRegistryClient 包装官方 Polaris SDK 客户端。
type officialPolarisRegistryClient struct {
	mu       sync.Mutex
	context  any
	provider polarissdk.ProviderAPI
	consumer polarissdk.ConsumerAPI
}

// polarisRegistryNative holds the native Polaris SDK components.
//
// polarisRegistryNative 持有原生 Polaris SDK 组件。
type polarisRegistryNative struct {
	Context  any
	Provider polarissdk.ProviderAPI
	Consumer polarissdk.ConsumerAPI
}

// newOfficialPolarisRegistryClient creates a new official Polaris client wrapper.
//
// newOfficialPolarisRegistryClient 创建新的官方 Polaris 客户端包装。
func newOfficialPolarisRegistryClient() polarisRegistryClient {
	return &officialPolarisRegistryClient{}
}

// Underlying returns the underlying Polaris SDK components.
//
// Underlying 返回底层 Polaris SDK 组件。
func (c *officialPolarisRegistryClient) Underlying() any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.provider == nil && c.consumer == nil {
		return nil
	}
	return &polarisRegistryNative{
		Context:  c.context,
		Provider: c.provider,
		Consumer: c.consumer,
	}
}

// Close closes the underlying Polaris SDK components.
//
// Close 关闭底层 Polaris SDK 组件。
func (c *officialPolarisRegistryClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.provider != nil {
		c.provider.Destroy()
	}
	if c.consumer != nil {
		c.consumer.Destroy()
	}
	if destroyer, ok := c.context.(interface{ Destroy() }); ok {
		destroyer.Destroy()
	}
	c.context = nil
	c.provider = nil
	c.consumer = nil
	return nil
}

// Register registers a service instance to Polaris.
//
// Register 将服务实例注册到 Polaris。
func (c *officialPolarisRegistryClient) Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error {
	provider, _, err := c.ensureClients(cfg)
	if err != nil {
		return err
	}
	host, port, err := splitHostPort(addr)
	if err != nil {
		return err
	}

	request := &polarissdk.InstanceRegisterRequest{}
	request.Service = name
	request.ServiceToken = cfg.Token
	request.Namespace = cfg.Namespace
	request.Host = host
	request.Port = port
	request.Metadata = mergePolarisMetadata(cfg.ServiceMeta, meta)
	request.SetHealthy(true)
	_, err = provider.RegisterInstance(request)
	return translatePolarisRegistryError(err)
}

// Deregister deregisters a service instance from Polaris.
//
// Deregister 将服务实例从 Polaris 注销。
func (c *officialPolarisRegistryClient) Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error {
	provider, _, err := c.ensureClients(cfg)
	if err != nil {
		return err
	}
	host, port, err := splitHostPort(addr)
	if err != nil {
		return err
	}

	request := &polarissdk.InstanceDeRegisterRequest{}
	request.Service = name
	request.ServiceToken = cfg.Token
	request.Namespace = cfg.Namespace
	request.Host = host
	request.Port = port
	err = provider.Deregister(request)
	return translatePolarisRegistryError(err)
}

// Discover discovers service instances from Polaris.
//
// Discover 从 Polaris 发现服务实例。
func (c *officialPolarisRegistryClient) Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]transportcontract.ServiceInstance, error) {
	_, consumer, err := c.ensureClients(cfg)
	if err != nil {
		return nil, err
	}
	request := &polarissdk.GetInstancesRequest{}
	request.Namespace = cfg.Namespace
	request.Service = name
	response, err := consumer.GetInstances(request)
	if err != nil {
		return nil, translatePolarisRegistryError(err)
	}
	instances := polarisInstancesToContract(response)
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}
	return instances, nil
}

// Watch watches service instance changes from Polaris.
//
// Watch 监听 Polaris 服务实例变更。
func (c *officialPolarisRegistryClient) Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	_, consumer, err := c.ensureClients(cfg)
	if err != nil {
		return err
	}
	request := &polarissdk.WatchServiceRequest{}
	request.Key = polarismodel.ServiceKey{
		Namespace: cfg.Namespace,
		Service:   name,
	}
	response, err := consumer.WatchService(request)
	if err != nil {
		return translatePolarisRegistryError(err)
	}
	if response != nil && response.GetAllInstancesResp != nil {
		onUpdate(polarisInstancesToContract(response.GetAllInstancesResp))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-response.EventChannel:
			if !ok {
				return nil
			}
			if event == nil {
				continue
			}
			instances, convErr := c.Discover(ctx, cfg, name)
			if convErr != nil {
				if errors.Is(convErr, ErrServiceNotFound) {
					onUpdate([]transportcontract.ServiceInstance{})
					continue
				}
				return convErr
			}
			onUpdate(instances)
		}
	}
}

// ensureClients ensures the Polaris SDK clients are initialized.
//
// ensureClients 确保 Polaris SDK 客户端已初始化。
func (c *officialPolarisRegistryClient) ensureClients(cfg *PolarisConfig) (polarissdk.ProviderAPI, polarissdk.ConsumerAPI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.provider != nil && c.consumer != nil {
		return c.provider, c.consumer, nil
	}

	addresses, err := normalizeRegistryPolarisAddresses(cfg.Address)
	if err != nil {
		return nil, nil, err
	}
	context, err := polarissdk.NewSDKContextByAddress(addresses...)
	if err != nil {
		return nil, nil, translatePolarisRegistryError(err)
	}
	c.context = context
	c.provider = polarissdk.NewProviderAPIByContext(context)
	c.consumer = polarissdk.NewConsumerAPIByContext(context)
	return c.provider, c.consumer, nil
}

// splitHostPort splits address into host and port.
//
// splitHostPort 将地址拆分为 host 和 port。
func splitHostPort(addr string) (string, int, error) {
	host, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("polaris: invalid address %q: %w", addr, err)
	}
	port, err := net.LookupPort("tcp", portString)
	if err != nil {
		return "", 0, fmt.Errorf("polaris: invalid port %q: %w", portString, err)
	}
	return host, port, nil
}

// mergePolarisMetadata merges base and override metadata.
//
// mergePolarisMetadata 合并基础和覆盖的 metadata。
func mergePolarisMetadata(base map[string]string, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

// polarisInstancesToContract converts Polaris instances to contract format.
//
// polarisInstancesToContract 将 Polaris 实例转换为契约格式。
func polarisInstancesToContract(response *polarismodel.InstancesResponse) []transportcontract.ServiceInstance {
	if response == nil {
		return nil
	}
	sourceInstances := response.GetInstances()
	instances := make([]transportcontract.ServiceInstance, 0, len(sourceInstances))
	for _, instance := range sourceInstances {
		if instance == nil {
			continue
		}
		instances = append(instances, transportcontract.ServiceInstance{
			ID:       instance.GetId(),
			Name:     instance.GetService(),
			Address:  fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
			Metadata: instance.GetMetadata(),
			Healthy:  instance.IsHealthy(),
		})
	}
	sortServiceInstances(instances)
	return instances
}

// normalizeRegistryPolarisAddresses normalizes Polaris addresses.
//
// normalizeRegistryPolarisAddresses 规范化 Polaris 地址。
func normalizeRegistryPolarisAddresses(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, "://") {
			parsed, err := url.Parse(candidate)
			if err != nil {
				return nil, fmt.Errorf("polaris: invalid address: %w", err)
			}
			if parsed.Host != "" {
				candidate = parsed.Host
			}
		}
		addresses = append(addresses, candidate)
	}
	if len(addresses) == 0 {
		return nil, ErrNoAddress
	}
	return addresses, nil
}

// translatePolarisRegistryError translates Polaris SDK errors to framework errors.
//
// translatePolarisRegistryError 将 Polaris SDK 错误转换为框架错误。
func translatePolarisRegistryError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "not found"), strings.Contains(message, "404"):
		return ErrServiceNotFound
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "dial tcp"),
		strings.Contains(message, "timeout"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "unavailable"):
		return fmt.Errorf("polaris: source unavailable: %w", err)
	default:
		return err
	}
}