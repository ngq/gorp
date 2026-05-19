// Package etcd provides etcd SDK client wrapper.
// This file wraps the official etcd SDK for service registry operations.
//
// 本包提供 etcd SDK 客户端包装。
// 本文件包装官方 etcd SDK 用于服务注册操作。
package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// liveEtcdRegistryClient wraps the official etcd SDK client.
//
// liveEtcdRegistryClient 包装官方 etcd SDK 客户端。
type liveEtcdRegistryClient struct {
	client *clientv3.Client
}

// newLiveEtcdRegistryClient creates a new live etcd client wrapper.
//
// newLiveEtcdRegistryClient 创建新的真实 etcd 客户端包装。
func newLiveEtcdRegistryClient(cfg *DiscoveryConfig) (*liveEtcdRegistryClient, error) {
	clientCfg := clientv3.Config{
		Endpoints:   cfg.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	}

	// 设置认证信息（如果配置了用户名和密码）
	if cfg.EtcdUsername != "" && cfg.EtcdPassword != "" {
		clientCfg.Username = cfg.EtcdUsername
		clientCfg.Password = cfg.EtcdPassword
	}

	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, err
	}

	return &liveEtcdRegistryClient{client: client}, nil
}

// Underlying returns the underlying etcd SDK client.
//
// Underlying 返回底层 etcd SDK 客户端。
func (c *liveEtcdRegistryClient) Underlying() any {
	return c.client
}

// Close closes the underlying etcd SDK client.
//
// Close 关闭底层 etcd SDK 客户端。
func (c *liveEtcdRegistryClient) Close() error {
	return c.client.Close()
}

// Grant creates a new lease with TTL.
//
// Grant 创建带有 TTL 的租约。
func (c *liveEtcdRegistryClient) Grant(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	resp, err := c.client.Lease.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

// Put writes a key-value pair with lease binding.
//
// Put 写入键值对并绑定租约。
func (c *liveEtcdRegistryClient) Put(ctx context.Context, key, value string, leaseID clientv3.LeaseID) error {
	_, err := c.client.KV.Put(ctx, key, value, clientv3.WithLease(leaseID))
	return err
}

// KeepAlive starts lease keepalive.
//
// KeepAlive 启动租约 keepalive。
func (c *liveEtcdRegistryClient) KeepAlive(ctx context.Context, leaseID clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return c.client.Lease.KeepAlive(ctx, leaseID)
}

// Revoke revokes a lease.
//
// Revoke 撤销租约。
func (c *liveEtcdRegistryClient) Revoke(ctx context.Context, leaseID clientv3.LeaseID) error {
	_, err := c.client.Lease.Revoke(ctx, leaseID)
	return err
}

// Get reads key-value pairs with options.
//
// Get 读取键值对（支持选项）。
func (c *liveEtcdRegistryClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return c.client.KV.Get(ctx, key, opts...)
}
