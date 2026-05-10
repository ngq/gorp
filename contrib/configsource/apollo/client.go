// Package apollo provides Apollo SDK client wrapper.
// This file wraps the official Apollo SDK for config retrieval and watching.
//
// 本包提供 Apollo SDK 客户端包装。
// 本文件包装官方 Apollo SDK 用于配置获取和监听。
package apollo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/apolloconfig/agollo/v4"
	apolloagcache "github.com/apolloconfig/agollo/v4/agcache"
	apolloconfig "github.com/apolloconfig/agollo/v4/env/config"
	apollostorage "github.com/apolloconfig/agollo/v4/storage"
	"gopkg.in/yaml.v3"
)

// officialApolloClient wraps the official Apollo SDK client.
//
// officialApolloClient 包装官方 Apollo SDK 客户端。
type officialApolloClient struct {
	mu     sync.Mutex
	client agollo.Client
}

// newOfficialApolloClient creates a new official Apollo client wrapper.
//
// newOfficialApolloClient 创建新的官方 Apollo 客户端包装。
func newOfficialApolloClient() apolloConfigClient {
	return &officialApolloClient{}
}

// Underlying returns the underlying Apollo SDK client.
//
// Underlying 返回底层 Apollo SDK 客户端。
func (c *officialApolloClient) Underlying() any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client
}

// Close closes the underlying Apollo client.
//
// Close 关闭底层 Apollo 客户端。
func (c *officialApolloClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	return nil
}

// GetConfig retrieves config snapshot from Apollo server.
//
// GetConfig 从 Apollo 服务器获取配置快照。
func (c *officialApolloClient) GetConfig(ctx context.Context, cfg *ApolloConfig) (apolloConfigSnapshot, error) {
	client, err := c.ensureClient(cfg)
	if err != nil {
		return apolloConfigSnapshot{}, err
	}
	cache := client.GetConfigCache(cfg.Namespace)
	if cache == nil {
		return apolloConfigSnapshot{}, ErrConfigNotFound
	}
	content, err := buildApolloSnapshotContent(cache)
	if err != nil {
		return apolloConfigSnapshot{}, err
	}
	return apolloConfigSnapshot{
		Content:  content,
		Revision: normalizeApolloRevision(apolloConfigSnapshot{Content: content}),
	}, nil
}

// WatchConfig watches for config changes from Apollo server.
//
// WatchConfig 监听 Apollo 服务器配置变更。
func (c *officialApolloClient) WatchConfig(ctx context.Context, cfg *ApolloConfig, lastRevision string, onUpdate func(snapshot apolloConfigSnapshot)) error {
	client, err := c.ensureClient(cfg)
	if err != nil {
		return err
	}

	listener := &apolloChangeListener{
		namespace: cfg.Namespace,
		onEvent: func() {
			snapshot, getErr := c.GetConfig(ctx, cfg)
			if getErr != nil {
				return
			}
			revision := normalizeApolloRevision(snapshot)
			if revision == "" || revision == lastRevision {
				return
			}
			lastRevision = revision
			onUpdate(snapshot)
		},
	}
	client.AddChangeListener(listener)
	defer client.RemoveChangeListener(listener)

	<-ctx.Done()
	return nil
}

// ensureClient ensures the Apollo client is initialized.
//
// ensureClient 确保 Apollo 客户端已初始化。
func (c *officialApolloClient) ensureClient(cfg *ApolloConfig) (agollo.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	client, err := agollo.StartWithConfig(func() (*apolloconfig.AppConfig, error) {
		return &apolloconfig.AppConfig{
			AppID:             cfg.AppID,
			Cluster:           cfg.Cluster,
			NamespaceName:     cfg.Namespace,
			IP:                cfg.MetaServer,
			Secret:            cfg.AccessKey,
			IsBackupConfig:    true,
			MustStart:         false,
			SyncServerTimeout: int((10 * time.Second) / time.Millisecond),
		}, nil
	})
	if err != nil {
		return nil, translateApolloSDKError(err)
	}
	c.client = client
	return c.client, nil
}

// apolloChangeListener implements Apollo change listener interface.
//
// apolloChangeListener 实现 Apollo 变更监听器接口。
type apolloChangeListener struct {
	namespace string
	onEvent   func()
}

// OnChange handles Apollo change events.
//
// OnChange 处理 Apollo 变更事件。
func (l *apolloChangeListener) OnChange(event *apollostorage.ChangeEvent) {
	if event == nil || event.Namespace != l.namespace {
		return
	}
	l.onEvent()
}

// OnNewestChange handles Apollo newest change events.
//
// OnNewestChange 处理 Apollo 最新变更事件。
func (l *apolloChangeListener) OnNewestChange(event *apollostorage.FullChangeEvent) {
	if event == nil || event.Namespace != l.namespace {
		return
	}
	l.onEvent()
}

// buildApolloConfigURL constructs the Apollo config API URL.
//
// buildApolloConfigURL 构造 Apollo 配置 API URL。
func buildApolloConfigURL(cfg *ApolloConfig) string {
	base := strings.TrimRight(cfg.MetaServer, "/")
	return fmt.Sprintf("%s/configs/%s/%s/%s", base, cfg.AppID, cfg.Cluster, cfg.Namespace)
}

// buildApolloSnapshotContent builds YAML content from Apollo cache.
//
// buildApolloSnapshotContent 从 Apollo 缓存构建 YAML 内容。
func buildApolloSnapshotContent(cache apolloagcache.CacheInterface) (string, error) {
	if cache == nil || cache.EntryCount() == 0 {
		return "", ErrConfigNotFound
	}

	loaded := make(map[string]any)
	cache.Range(func(key, value interface{}) bool {
		keyString, ok := key.(string)
		if !ok || keyString == "" {
			return true
		}
		assignNestedValue(loaded, keyString, value)
		return true
	})
	if len(loaded) == 0 {
		return "", ErrConfigNotFound
	}

	content, err := yaml.Marshal(loaded)
	if err != nil {
		return "", fmt.Errorf("apollo: marshal config failed: %w", err)
	}
	return string(content), nil
}

// translateApolloSDKError translates Apollo SDK errors to framework errors.
//
// translateApolloSDKError 将 Apollo SDK 错误转换为框架错误。
func translateApolloSDKError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "401"), strings.Contains(message, "403"),
		strings.Contains(message, "forbidden"), strings.Contains(message, "unauthorized"):
		return ErrAuthFailed
	case strings.Contains(message, "404"), strings.Contains(message, "not found"):
		return ErrConfigNotFound
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "timeout"),
		strings.Contains(message, "dial tcp"),
		strings.Contains(message, "server unavailable"):
		return ErrSourceUnavailable
	default:
		return err
	}
}