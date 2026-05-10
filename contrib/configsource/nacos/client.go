// Package nacos provides Nacos SDK client wrapper.
// This file wraps the official Nacos SDK for config retrieval and watching.
//
// 本包提供 Nacos SDK 客户端包装。
// 本文件包装官方 Nacos SDK 用于配置获取和监听。
package nacos

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	configclient "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// sdkNacosClient wraps the official Nacos SDK client.
//
// sdkNacosClient 包装官方 Nacos SDK 客户端。
// 该结构体实现了 nacosConfigClient 接口，提供配置获取、发布和监听能力。
// 所有 SDK 调用都在后台 goroutine 中执行，通过 channel 将结果传递回主线程，
// 确保调用方可以通过 context 控制超时和取消。
type sdkNacosClient struct {
	// client 是官方 Nacos SDK 的配置客户端实例。
	client configclient.IConfigClient
}

// newOfficialNacosClient creates a new official Nacos client wrapper.
//
// newOfficialNacosClient 创建新的官方 Nacos 客户端包装。
// 该函数使用配置初始化 Nacos SDK 客户端，建立与服务器的连接。
// 参数：
//   - cfg: Nacos 配置，包含服务器地址、端口、命名空间等
// 返回：
//   - 实现了 nacosConfigClient 接口的客户端包装
//   - 错误信息（如果创建失败）
func newOfficialNacosClient(cfg *NacosConfig) (nacosConfigClient, error) {
	// 配置 Nacos 客户端参数
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		NotLoadCacheAtStart: true,
	}
	// 如果配置了用户名和密码，添加认证信息
	if cfg.Username != "" && cfg.Password != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	// 创建 Nacos 配置客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig: &clientConfig,
			ServerConfigs: []constant.ServerConfig{
				{
					IpAddr:      parseNacosHost(cfg.ServerAddr),
					Port:        uint64(cfg.Port),
					ContextPath: "/nacos",
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("nacos: create config client failed: %w", err)
	}

	return &sdkNacosClient{client: client}, nil
}

// GetConfig retrieves config content from Nacos server.
//
// GetConfig 从 Nacos 服务器获取配置内容。
// 该方法在后台 goroutine 中调用 SDK，通过 channel 返回结果，
// 支持通过 context 控制超时和取消。
func (c *sdkNacosClient) GetConfig(ctx context.Context, cfg *NacosConfig) (string, error) {
	// 定义结果结构，用于 channel 传递
	type result struct {
		content string
		err     error
	}
	done := make(chan result, 1)

	// 在后台 goroutine 中执行 SDK 调用
	go func() {
		content, err := c.client.GetConfig(toNacosConfigParam(cfg))
		done <- result{content: content, err: translateNacosError("load config", err)}
	}()

	// 等待结果或 context 取消
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-done:
		return res.content, res.err
	}
}

// PublishConfig publishes config content to Nacos server.
//
// PublishConfig 发布配置内容到 Nacos 服务器。
// 该方法在后台 goroutine 中调用 SDK，支持 context 控制。
func (c *sdkNacosClient) PublishConfig(ctx context.Context, cfg *NacosConfig, content string) error {
	// 定义结果结构
	type result struct {
		ok  bool
		err error
	}
	done := make(chan result, 1)

	// 在后台 goroutine 中执行 SDK 调用
	go func() {
		param := toNacosConfigParam(cfg)
		param.Content = content
		ok, err := c.client.PublishConfig(param)
		done <- result{ok: ok, err: err}
	}()

	// 等待结果或 context 取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-done:
		if res.err != nil {
			return translateNacosError("publish config", res.err)
		}
		if !res.ok {
			return errors.New("nacos: publish config failed")
		}
		return nil
	}
}

// WatchConfig watches for config changes from Nacos server.
//
// WatchConfig 监听 Nacos 服务器配置变更。
// 该方法使用 SDK 的 ListenConfig API 注册监听器。
// 监听器在 context 取消时自动取消监听。
func (c *sdkNacosClient) WatchConfig(ctx context.Context, cfg *NacosConfig, onUpdate func(string)) error {
	// 构造监听参数
	param := toNacosConfigParam(cfg)
	param.OnChange = func(namespace, group, dataID, data string) {
		// 检查 context 是否已取消，避免在取消后触发回调
		select {
		case <-ctx.Done():
			return
		default:
			onUpdate(data)
		}
	}

	// 注册监听
	if err := c.client.ListenConfig(param); err != nil {
		return translateNacosError("listen config", err)
	}

	// 等待 context 取消
	<-ctx.Done()
	// 取消监听
	_ = c.client.CancelListenConfig(param)
	return nil
}

// CloseClient closes the underlying Nacos SDK client.
//
// CloseClient 关闭底层 Nacos SDK 客户端。
// 该方法释放 SDK 客户端持有的资源，如连接池等。
func (c *sdkNacosClient) CloseClient() {
	c.client.CloseClient()
}

// Underlying returns the underlying Nacos SDK client.
//
// Underlying 返回底层 Nacos SDK 客户端。
// 该方法用于下探机制，暴露原生 SDK 客户端实例。
func (c *sdkNacosClient) Underlying() any {
	return c.client
}

// toNacosConfigParam converts NacosConfig to SDK config parameter.
//
// toNacosConfigParam 将 NacosConfig 转换为 SDK 配置参数。
// 该辅助函数将框架配置结构转换为 Nacos SDK 所需的参数结构。
func toNacosConfigParam(cfg *NacosConfig) vo.ConfigParam {
	return vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	}
}

// parseNacosHost extracts host from server address string.
//
// parseNacosHost 从服务器地址字符串中提取主机名。
// 该函数处理以下格式：
//   - "http://host:port" -> "host"
//   - "https://host:port" -> "host"
//   - "host:port" -> "host"
//   - "host" -> "host"
func parseNacosHost(addr string) string {
	// 移除协议前缀
	trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(addr, "http://"), "https://"))
	// 移除路径部分
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	// 移除端口部分
	if host, _, found := strings.Cut(trimmed, ":"); found {
		return host
	}
	return trimmed
}

// translateNacosError translates Nacos SDK errors to framework errors.
//
// translateNacosError 将 Nacos SDK 错误转换为框架错误。
// 该函数将 SDK 的原始错误转换为框架定义的错误类型，
// 如将 404 错误转换为 ErrConfigNotFound。
func translateNacosError(action string, err error) error {
	if err == nil {
		return nil
	}
	// 检查错误消息中的特征
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "config data not exist"),
		strings.Contains(message, "data not exist"),
		strings.Contains(message, "404"):
		return ErrConfigNotFound
	default:
		return fmt.Errorf("nacos: %s failed: %w", action, err)
	}
}

// unwrapNacosNativeClient unwraps the native client from the wrapper.
//
// unwrapNacosNativeClient 从包装中解包原生客户端。
// 该函数用于 Underlying 和 As 方法，获取底层的 SDK 客户端实例。
// 如果客户端包装实现了 Underlying() 方法，则返回其结果；
// 否则直接返回客户端包装本身。
func unwrapNacosNativeClient(client nacosConfigClient) any {
	// 检查客户端是否实现了 Underlying() 方法
	if provider, ok := client.(interface{ Underlying() any }); ok {
		return provider.Underlying()
	}
	return client
}