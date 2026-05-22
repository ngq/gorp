// Package zookeeper provides Zookeeper SDK client wrapper.
// This file wraps the official Zookeeper SDK for service registry operations.
//
// 本包提供 Zookeeper SDK 客户端包装。
// 本文件包装官方 Zookeeper SDK 用于服务注册操作。
package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

// errZKNoNode wraps Zookeeper ErrNoNode for use in registry.go.
//
// errZKNoNode 包装 Zookeeper ErrNoNode 错误供 registry.go 使用。
var errZKNoNode = zk.ErrNoNode

// realZKBackend implements zkBackend using the official Zookeeper SDK.
//
// realZKBackend 使用官方 Zookeeper SDK 实现 zkBackend。
type realZKBackend struct {
	conn *zk.Conn
}

// newZKBackend creates a new Zookeeper backend with official SDK connection.
//
// newZKBackend 使用官方 SDK 连接创建新的 Zookeeper 后端。
func newZKBackend(cfg *ZookeeperConfig) (zkBackend, error) {
	if len(cfg.Servers) == 0 {
		return nil, ErrNoServers
	}

	// 连接 Zookeeper 集群
	conn, _, err := zk.Connect(cfg.Servers, cfg.SessionTimeout)
	if err != nil {
		return nil, fmt.Errorf("registry.zookeeper: connect failed: %w", err)
	}
	return &realZKBackend{conn: conn}, nil
}

// EnsurePath ensures the target path exists in Zookeeper, creating parents if needed.
//
// EnsurePath 确保 Zookeeper 中目标路径存在，必要时创建父路径。
func (b *realZKBackend) EnsurePath(target string) error {
	// 分割路径并逐层创建
	parts := strings.Split(strings.Trim(target, "/"), "/")
	current := ""
	for _, part := range parts {
		current += "/" + part
		exists, _, err := b.conn.Exists(current)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		// 创建持久节点（非临时节点）
		_, err = b.conn.Create(current, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil && !errors.Is(err, zk.ErrNodeExists) {
			return err
		}
	}
	return nil
}

// CreateEphemeral creates an ephemeral ZNode at the target path with given data.
//
// CreateEphemeral 在目标路径创建临时 ZNode，携带给定数据。
func (b *realZKBackend) CreateEphemeral(target string, data []byte) error {
	// 创建临时节点（Ephemeral ZNode），节点随会话断开而删除
	_, err := b.conn.Create(target, data, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	return err
}

// Delete deletes the ZNode at the target path.
//
// Delete 删除目标路径的 ZNode。
func (b *realZKBackend) Delete(target string) error {
	return b.conn.Delete(target, -1)
}

// Children returns the list of child ZNode names under the target path.
//
// Children 返回目标路径下的子 ZNode 名称列表。
func (b *realZKBackend) Children(target string) ([]string, error) {
	children, _, err := b.conn.Children(target)
	return children, err
}

// Get returns the data stored in the ZNode at the target path.
//
// Get 返回目标路径 ZNode 中存储的数据。
func (b *realZKBackend) Get(target string) ([]byte, error) {
	data, _, err := b.conn.Get(target)
	return data, err
}

// WatchChildren watches for changes to child ZNodes under the target path.
//
// WatchChildren 监听目标路径下子 ZNode 的变更。
func (b *realZKBackend) WatchChildren(ctx context.Context, target string, onUpdate func()) error {
	for {
		// 设置子节点监听器
		_, _, events, err := b.conn.ChildrenW(target)
		if err != nil {
			if errors.Is(err, zk.ErrNoNode) {
				// 节点不存在时触发更新并重试
				onUpdate()
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(200 * time.Millisecond):
					continue
				}
			}
			return err
		}

		// 等待监听事件或上下文取消
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return zk.ErrConnectionClosed
			}
			if event.Err != nil {
				return event.Err
			}
			// 处理会话状态变更
			if event.Type == zk.EventSession {
				switch event.State {
				case zk.StateExpired:
					return zk.ErrSessionExpired
				case zk.StateAuthFailed:
					return errors.New("registry.zookeeper: watch session auth failed")
				case zk.StateDisconnected, zk.StateConnecting:
					return zk.ErrConnectionClosed
				}
			}
			// 触发更新回调
			onUpdate()
		}
	}
}

// Close closes the Zookeeper connection.
//
// Close 关闭 Zookeeper 连接。
func (b *realZKBackend) Close() error {
	b.conn.Close()
	return nil
}

// Underlying returns the underlying Zookeeper connection object.
//
// Underlying 返回底层 Zookeeper 连接对象。
func (b *realZKBackend) Underlying() any {
	return b.conn
}
