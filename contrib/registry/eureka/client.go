// Package eureka provides Eureka REST API client wrapper.
// This file implements the eurekaClient interface with HTTP calls.
//
// 本包提供 Eureka REST API 客户端包装。
// 本文件使用 HTTP 调用实现 eurekaClient 接口。
package eureka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// eurekaClient defines the internal client interface for Eureka operations.
//
// eurekaClient 定义 Eureka 操作的内部客户端接口。
// 该接口抽象底层 HTTP 调用，便于测试时注入 fake client。
type eurekaClient interface {
	Register(ctx context.Context, cfg *EurekaConfig, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, cfg *EurekaConfig, name, addr string) error
	Heartbeat(ctx context.Context, cfg *EurekaConfig, name, addr string) error
	Discover(ctx context.Context, cfg *EurekaConfig, name string) ([]transportcontract.ServiceInstance, error)
	Watch(ctx context.Context, cfg *EurekaConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error
}

// HTTPClientProvider exposes the current HTTP transport object for native down-dive.
//
// HTTPClientProvider 暴露当前 HTTP transport 对象，用于原生下探。
// 实现该接口允许业务通过 As 方法获取 http.Client。
type HTTPClientProvider interface {
	HTTPClient() *http.Client
}

// httpEurekaClient 使用标准 http.Client 实现 eurekaClient。
//
// 特性：
//   - 使用 REST API 与 Eureka 服务端通信
//   - 支持 JSON 格式请求/响应
//   - 内置 10 秒超时
type httpEurekaClient struct {
	httpClient *http.Client
}

// newHTTPEurekaClient creates a new HTTP Eureka client.
//
// newHTTPEurekaClient 创建新的 HTTP Eureka 客户端。
func newHTTPEurekaClient() eurekaClient {
	return &httpEurekaClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// HTTPClient returns the underlying http.Client.
// Implements HTTPClientProvider interface.
//
// HTTPClient 返回底层 http.Client。
// 实现 HTTPClientProvider 接口。
func (c *httpEurekaClient) HTTPClient() *http.Client {
	return c.httpClient
}

// Register registers a service instance to Eureka via REST API.
//
// Register 通过 REST API 将服务实例注册到 Eureka。
//
// Eureka REST API:
//   - POST /eureka/apps/{appName}
//   - Body: JSON instance payload
//   - Success: 204 No Content
func (c *httpEurekaClient) Register(ctx context.Context, cfg *EurekaConfig, name, addr string, meta map[string]string) error {
	// 构建 Eureka instance payload
	payload := map[string]any{
		"instance": map[string]any{
			"app":      name,
			"hostName": cfg.InstanceHost,
			"ipAddr":   hostFromAddr(addr),
			"port": map[string]any{
				"$":        portFromAddr(addr, cfg.InstancePort),
				"@enabled": true,
			},
			"vipAddress": name,
			"metadata":   mergeMeta(cfg.ServiceMeta, meta),
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 构建请求 URL: POST /eureka/apps/{APP_NAME}
	appURL := strings.TrimRight(cfg.ServerURL, "/") + "/eureka/apps/" + url.PathEscape(strings.ToUpper(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("eureka: register failed: %w", err)
	}
	defer resp.Body.Close()

	// Eureka 注册成功返回 204 No Content 或 201 Created 或 200 OK
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("eureka: register failed with status %d", resp.StatusCode)
	}
	return nil
}

// Deregister deregisters a service instance from Eureka via REST API.
//
// Deregister 通过 REST API 将服务实例从 Eureka 注销。
//
// Eureka REST API:
//   - DELETE /eureka/apps/{appName}/{instanceId}
//   - Success: 200 OK or 204 No Content
//   - Not found: 404
func (c *httpEurekaClient) Deregister(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	// 构建请求 URL: DELETE /eureka/apps/{APP_NAME}/{INSTANCE_ID}
	instanceURL := strings.TrimRight(cfg.ServerURL, "/") +
		"/eureka/apps/" + url.PathEscape(strings.ToUpper(name)) +
		"/" + url.PathEscape(instanceID(name, addr))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, instanceURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("eureka: deregister failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("eureka: deregister failed with status %d", resp.StatusCode)
	}
	return nil
}

// Heartbeat sends a heartbeat renewal to Eureka via REST API.
//
// Heartbeat 通过 REST API 发送心跳续租到 Eureka。
//
// Eureka REST API:
//   - PUT /eureka/apps/{appName}/{instanceId}
//   - Success: 200 OK or 204 No Content
//   - Not found: 404（表示实例已过期，需要重新注册）
func (c *httpEurekaClient) Heartbeat(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	// 构建请求 URL: PUT /eureka/apps/{APP_NAME}/{INSTANCE_ID}
	instanceURL := strings.TrimRight(cfg.ServerURL, "/") +
		"/eureka/apps/" + url.PathEscape(strings.ToUpper(name)) +
		"/" + url.PathEscape(instanceID(name, addr))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, instanceURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("eureka: heartbeat failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("eureka: heartbeat failed with status %d", resp.StatusCode)
	}
	return nil
}

// Discover discovers service instances from Eureka via REST API.
//
// Discover 通过 REST API 从 Eureka 发现服务实例。
//
// Eureka REST API:
//   - GET /eureka/apps/{appName}
//   - Response: JSON application payload
//   - Not found: 404
func (c *httpEurekaClient) Discover(ctx context.Context, cfg *EurekaConfig, name string) ([]transportcontract.ServiceInstance, error) {
	// 构建请求 URL: GET /eureka/apps/{APP_NAME}
	appURL := strings.TrimRight(cfg.ServerURL, "/") + "/eureka/apps/" + url.PathEscape(strings.ToUpper(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("eureka: discover failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eureka: discover failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("eureka: read discover response failed: %w", err)
	}

	var payload eurekaDiscoverResponse
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("eureka: decode discover response failed: %w", err)
	}

	instances := payload.Application.Instance
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}

	// 转换为契约格式
	result := make([]transportcontract.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		port := instance.Port.Value
		if port == 0 {
			port = cfg.InstancePort
		}
		result = append(result, transportcontract.ServiceInstance{
			ID:       instance.InstanceID,
			Name:     strings.ToLower(instance.App),
			Address:  fmt.Sprintf("%s:%d", instance.IPAddr, port),
			Metadata: instance.Metadata,
			Healthy:  strings.EqualFold(instance.Status, "UP"),
		})
	}
	sortServiceInstances(result)
	return result, nil
}

// Watch watches service instance changes via periodic polling.
//
// Watch 通过周期性轮询监听服务实例变更。
// Eureka REST API 不支持真正的长连接 watch，因此采用轮询实现。
//
// 行为说明：
//   - 定时调用 Discover 获取实例列表
//   - 通过 JSON 序列化比较去重
//   - 变化时调用 onUpdate 推送新列表
func (c *httpEurekaClient) Watch(ctx context.Context, cfg *EurekaConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	interval := cfg.WatchInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var last string
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			instances, err := c.Discover(ctx, cfg, name)
			if err != nil {
				if errors.Is(err, ErrServiceNotFound) {
					if last != "[]" {
						last = "[]"
						onUpdate([]transportcontract.ServiceInstance{})
					}
					continue
				}
				continue
			}

			// JSON 序列化比较，去重
			payload, marshalErr := json.Marshal(instances)
			if marshalErr != nil {
				continue
			}
			current := string(payload)
			if current == last {
				continue
			}
			last = current
			onUpdate(instances)
		}
	}
}

// eurekaDiscoverResponse 定义 Eureka discover API 响应结构。
//
// JSON 结构示例：
// {
//   "application": {
//     "instance": [
//       {
//         "instanceId": "MYAPP:192.168.1.1:8080",
//         "app": "MYAPP",
//         "ipAddr": "192.168.1.1",
//         "status": "UP",
//         "metadata": {},
//         "port": {"$": 8080}
//       }
//     ]
//   }
// }
type eurekaDiscoverResponse struct {
	Application struct {
		Instance []struct {
			InstanceID string            `json:"instanceId"`
			App        string            `json:"app"`
			IPAddr     string            `json:"ipAddr"`
			Status     string            `json:"status"`
			Metadata   map[string]string `json:"metadata"`
			Port       struct {
				Value int `json:"$"`
			} `json:"port"`
		} `json:"instance"`
	} `json:"application"`
}

// sortServiceInstances sorts service instances by ID and address.
//
// sortServiceInstances 按 ID 和地址排序服务实例。
func sortServiceInstances(instances []transportcontract.ServiceInstance) {
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ID != instances[j].ID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Address < instances[j].Address
	})
}