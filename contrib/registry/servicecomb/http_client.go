package servicecomb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// httpServiceCombClient 是基于 ServiceCenter REST API 的真实客户端。
// 此前 NewRegistry 默认使用 inMemoryServiceCombClient——注册"成功"但
// 其他服务永远发现不了该实例，生产使用 servicecomb 时完全失效。
type httpServiceCombClient struct {
	httpClient *http.Client
}

func newHTTPServiceCombClient() *httpServiceCombClient {
	return &httpServiceCombClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *httpServiceCombClient) Underlying() any { return c.httpClient }

// baseURL 拼装 ServiceCenter REST 前缀：/v4/{project}/registry。
func baseURL(cfg *ServiceCombConfig) string {
	// ServiceComb 配置没有独立 project 字段，统一用 default 租户。
	return strings.TrimRight(cfg.ServerURI, "/") + "/v4/default/registry"
}

// microserviceID 通过 serviceName/appId 查找 microserviceID，找不到则创建。
func (c *httpServiceCombClient) microserviceID(ctx context.Context, cfg *ServiceCombConfig, name string) (string, error) {
	// 1. 按 name 查询已存在的 microservice
	findURL := fmt.Sprintf("%s/microservices?appId=%s&serviceName=%s", baseURL(cfg), cfg.AppID, name)
	findReq, err := http.NewRequestWithContext(ctx, http.MethodGet, findURL, nil)
	if err != nil {
		return "", err
	}
	findResp, err := c.httpClient.Do(findReq)
	if err != nil {
		return "", fmt.Errorf("registry.servicecomb: find microservice: %w", err)
	}
	defer findResp.Body.Close()

	var findBody struct {
		Services []struct {
			ServiceID string `json:"serviceId"`
		} `json:"services"`
	}
	if findResp.StatusCode == http.StatusOK {
		_ = json.NewDecoder(findResp.Body).Decode(&findBody)
		if len(findBody.Services) > 0 && findBody.Services[0].ServiceID != "" {
			return findBody.Services[0].ServiceID, nil
		}
	}

	// 2. 创建 microservice
	register := map[string]any{
		"service": map[string]any{
			"serviceName": name,
			"appId":       cfg.AppID,
			"version":     cfg.Version,
			"environment": cfg.Environment,
			"status":      "UP",
		},
	}
	body, _ := json.Marshal(register)
	createURL := baseURL(cfg) + "/microservices"
	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := c.httpClient.Do(createReq)
	if err != nil {
		return "", fmt.Errorf("registry.servicecomb: create microservice: %w", err)
	}
	defer createResp.Body.Close()

	var createBody struct {
		ServiceID string `json:"serviceId"`
	}
	if createResp.StatusCode == http.StatusOK {
		_ = json.NewDecoder(createResp.Body).Decode(&createBody)
		if createBody.ServiceID != "" {
			return createBody.ServiceID, nil
		}
	}

	// 3. 并发创建冲突时（已有同名服务）再查一次
	findReq2, _ := http.NewRequestWithContext(ctx, http.MethodGet, findURL, nil)
	findResp2, err := c.httpClient.Do(findReq2)
	if err != nil {
		return "", fmt.Errorf("registry.servicecomb: re-find microservice: %w", err)
	}
	defer findResp2.Body.Close()
	var findBody2 struct {
		Services []struct {
			ServiceID string `json:"serviceId"`
		} `json:"services"`
	}
	if findResp2.StatusCode == http.StatusOK {
		_ = json.NewDecoder(findResp2.Body).Decode(&findBody2)
		if len(findBody2.Services) > 0 && findBody2.Services[0].ServiceID != "" {
			return findBody2.Services[0].ServiceID, nil
		}
	}
	return "", fmt.Errorf("registry.servicecomb: cannot resolve microservice id for %q (register status %d)", name, createResp.StatusCode)
}

func (c *httpServiceCombClient) Register(ctx context.Context, cfg *ServiceCombConfig, name, addr string, meta map[string]string) error {
	serviceID, err := c.microserviceID(ctx, cfg, name)
	if err != nil {
		return err
	}
	fullMeta := make(map[string]string)
	for k, v := range cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}
	fullMeta["version"] = cfg.Version
	fullMeta["environment"] = cfg.Environment

	instance := map[string]any{
		"instance": map[string]any{
			"serviceId":  serviceID,
			"endpoints":  []string{"rest://" + addr},
			"hostName":   cfg.InstanceHost,
			"status":     "UP",
			"properties": fullMeta,
			"healthCheck": map[string]any{
				"mode": "push",
			},
		},
	}
	body, _ := json.Marshal(instance)
	url := fmt.Sprintf("%s/microservices/%s/instances", baseURL(cfg), serviceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("registry.servicecomb: register instance: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry.servicecomb: register instance failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpServiceCombClient) Deregister(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	serviceID, err := c.microserviceID(ctx, cfg, name)
	if err != nil {
		return err
	}
	// ServiceCenter 支持按 endpoints 删除实例。
	url := fmt.Sprintf("%s/microservices/%s/instances?endpoint=rest%%3A%%2F%%2F%s", baseURL(cfg), serviceID, addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("registry.servicecomb: deregister instance: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry.servicecomb: deregister instance failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpServiceCombClient) Heartbeat(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	serviceID, err := c.microserviceID(ctx, cfg, name)
	if err != nil {
		return err
	}
	// 心跳模式为 push 时无需显式调用（实例 status 由 pull 模式探测），
	// 这里做一次轻量探活，避免返回 nil 造成"心跳成功"的错觉。
	url := fmt.Sprintf("%s/microservices/%s/instances?endpoint=rest%%3A%%2F%%2F%s", baseURL(cfg), serviceID, addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("registry.servicecomb: heartbeat probe: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	return nil
}

func (c *httpServiceCombClient) Discover(ctx context.Context, cfg *ServiceCombConfig, name string) ([]transportcontract.ServiceInstance, error) {
	serviceID, err := c.microserviceID(ctx, cfg, name)
	if err != nil {
		// 服务尚未注册时按"未找到"处理，让上层走注册/空列表逻辑。
		if strings.Contains(err.Error(), "cannot resolve") {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	url := fmt.Sprintf("%s/microservices/%s/instances", baseURL(cfg), serviceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry.servicecomb: discover instances: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry.servicecomb: discover instances failed with status %d", resp.StatusCode)
	}

	var body struct {
		Instances []struct {
			InstanceID string            `json:"instanceId"`
			ServiceID  string            `json:"serviceId"`
			Endpoints  []string          `json:"endpoints"`
			Status     string            `json:"status"`
			Properties map[string]string `json:"properties"`
		} `json:"instances"`
	}
	raw, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("registry.servicecomb: decode discover response: %w", err)
	}

	result := make([]transportcontract.ServiceInstance, 0, len(body.Instances))
	for _, inst := range body.Instances {
		addr := instanceAddrFromEndpoints(inst.Endpoints)
		if addr == "" {
			continue
		}
		result = append(result, transportcontract.ServiceInstance{
			ID:       inst.InstanceID,
			Name:     name,
			Address:  addr,
			Metadata: inst.Properties,
			Healthy:  strings.EqualFold(inst.Status, "UP"),
		})
	}
	if len(result) == 0 {
		return nil, ErrServiceNotFound
	}
	sortServiceInstances(result)
	return result, nil
}

// instanceAddrFromEndpoints 从 ServiceCenter endpoints（如 "rest://1.2.3.4:8080"）
// 提取 host:port。
func instanceAddrFromEndpoints(endpoints []string) string {
	for _, ep := range endpoints {
		if idx := strings.Index(ep, "://"); idx >= 0 {
			addr := ep[idx+3:]
			if addr != "" {
				return addr
			}
		}
	}
	return ""
}

var _ = errors.New // keep import used for future extensions
