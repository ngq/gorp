// Package polaris provides Polaris SDK client wrapper.
// This file wraps the official Polaris SDK for config retrieval and watching.
//
// 本包提供 Polaris SDK 客户端包装。
// 本文件包装官方 Polaris SDK 用于配置获取和监听。
package polaris

import (
	"context"
	"net/url"
	"strings"
	"sync"

	polarissdk "github.com/polarismesh/polaris-go"
	polarismodel "github.com/polarismesh/polaris-go/pkg/model"
)

// officialPolarisClient wraps the official Polaris SDK client.
//
// officialPolarisClient 包装官方 Polaris SDK 客户端。
type officialPolarisClient struct {
	mu      sync.Mutex
	context any
	api     polarissdk.ConfigAPI
}

// newOfficialPolarisClient creates a new official Polaris client wrapper.
//
// newOfficialPolarisClient 创建新的官方 Polaris 客户端包装。
func newOfficialPolarisClient() polarisConfigClient {
	return &officialPolarisClient{}
}

// Underlying returns the underlying Polaris SDK client.
//
// Underlying 返回底层 Polaris SDK 客户端。
func (c *officialPolarisClient) Underlying() any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.api != nil {
		return c.api
	}
	return nil
}

// Close closes the underlying Polaris client.
//
// Close 关闭底层 Polaris 客户端。
func (c *officialPolarisClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if destroyer, ok := c.context.(interface{ Destroy() }); ok {
		destroyer.Destroy()
	}
	c.context = nil
	c.api = nil
	return nil
}

// GetConfig retrieves config snapshot from Polaris server.
//
// GetConfig 从 Polaris 服务器获取配置快照。
func (c *officialPolarisClient) GetConfig(ctx context.Context, cfg *PolarisConfig) (polarisConfigSnapshot, error) {
	configFile, err := c.getConfigFile(cfg)
	if err != nil {
		return polarisConfigSnapshot{}, err
	}
	if configFile == nil || !configFile.HasContent() {
		return polarisConfigSnapshot{}, ErrConfigNotFound
	}
	content := configFile.GetContent()
	return polarisConfigSnapshot{
		Content:  content,
		Revision: normalizePolarisRevision(polarisConfigSnapshot{Content: content}),
	}, nil
}

// WatchConfig watches for config changes from Polaris server.
//
// WatchConfig 监听 Polaris 服务器配置变更。
func (c *officialPolarisClient) WatchConfig(ctx context.Context, cfg *PolarisConfig, lastRevision string, onUpdate func(snapshot polarisConfigSnapshot)) error {
	configFile, err := c.getConfigFile(cfg)
	if err != nil {
		return err
	}

	configFile.AddChangeListener(func(event polarismodel.ConfigFileChangeEvent) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		content := event.NewValue
		if strings.TrimSpace(content) == "" {
			snapshot, getErr := c.GetConfig(ctx, cfg)
			if getErr != nil {
				return
			}
			content = snapshot.Content
		}

		snapshot := polarisConfigSnapshot{
			Content:  content,
			Revision: normalizePolarisRevision(polarisConfigSnapshot{Content: content}),
		}
		if snapshot.Revision == "" || snapshot.Revision == lastRevision {
			return
		}
		lastRevision = snapshot.Revision
		onUpdate(snapshot)
	})

	<-ctx.Done()
	return nil
}

// getConfigFile retrieves the config file from Polaris.
//
// getConfigFile 从 Polaris 获取配置文件。
func (c *officialPolarisClient) getConfigFile(cfg *PolarisConfig) (polarissdk.ConfigFile, error) {
	api, err := c.ensureAPI(cfg)
	if err != nil {
		return nil, err
	}
	configFile, err := api.GetConfigFile(cfg.Namespace, cfg.FileGroup, cfg.FileName)
	if err != nil {
		return nil, translatePolarisSDKError(err)
	}
	return configFile, nil
}

// ensureAPI ensures the Polaris ConfigAPI is initialized.
//
// ensureAPI 确保 Polaris ConfigAPI 已初始化。
func (c *officialPolarisClient) ensureAPI(cfg *PolarisConfig) (polarissdk.ConfigAPI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.api != nil {
		return c.api, nil
	}

	addresses, err := normalizePolarisAddresses(cfg.ServerAddress)
	if err != nil {
		return nil, err
	}
	context, err := polarissdk.NewSDKContextByAddress(addresses...)
	if err != nil {
		return nil, translatePolarisSDKError(err)
	}
	api := polarissdk.NewConfigAPIByContext(context)
	c.context = context
	c.api = api
	return c.api, nil
}

// buildPolarisConfigURL constructs the Polaris config API URL.
//
// buildPolarisConfigURL 构造 Polaris 配置 API URL。
func buildPolarisConfigURL(cfg *PolarisConfig) string {
	base := strings.TrimRight(cfg.ServerAddress, "/")
	values := url.Values{}
	values.Set("namespace", cfg.Namespace)
	values.Set("group", cfg.FileGroup)
	values.Set("file", cfg.FileName)
	return base + "/config/v1/files?" + values.Encode()
}

// translatePolarisSDKError translates Polaris SDK errors to framework errors.
//
// translatePolarisSDKError 将 Polaris SDK 错误转换为框架错误。
func translatePolarisSDKError(err error) error {
	if err != nil {
		message := strings.ToLower(err.Error())
		switch {
		case strings.Contains(message, "401"), strings.Contains(message, "403"),
			strings.Contains(message, "forbidden"), strings.Contains(message, "unauthorized"):
			return ErrAuthFailed
		case strings.Contains(message, "not found"), strings.Contains(message, "404"):
			return ErrConfigNotFound
		case strings.Contains(message, "connection refused"),
			strings.Contains(message, "dial tcp"),
			strings.Contains(message, "timeout"),
			strings.Contains(message, "no such host"),
			strings.Contains(message, "unavailable"):
			return ErrSourceUnavailable
		}
	}
	return err
}
