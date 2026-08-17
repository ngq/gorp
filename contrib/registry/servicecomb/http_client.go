package servicecomb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
// 注意：ServiceCenter 的 find 查询对 serviceName 的过滤不可靠（实测会返回
// 同 appId 下的无关服务），因此必须在客户端侧按 name/appId 精确匹配，
// 否则会拿到错误服务的 serviceId，注册实例时返回 "Micro-service does not exist"。
func (c *httpServiceCombClient) microserviceID(ctx context.Context, cfg *ServiceCombConfig, name string) (string, error) {
	// 1. 按 name 查询已存在的 microservice
	findURL := fmt.Sprintf("%s/microservices?appId=%s&serviceName=%s",
		baseURL(cfg), url.QueryEscape(cfg.AppID), url.QueryEscape(name))
	findReq, err := http.NewRequestWithContext(ctx, http.MethodGet, findURL, nil)
	if err != nil {
		return "", err
	}
	findResp, err := c.httpClient.Do(findReq)
	if err != nil {
		return "", fmt.Errorf("registry.servicecomb: find microservice: %w", err)
	}
	if serviceID, ok := c.matchMicroservice(findResp, cfg.AppID, name); ok {
		return serviceID, nil
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

	// 3. 并发创建冲突时（已有同名服务）再查一次，同样做客户端侧匹配。
	findReq2, _ := http.NewRequestWithContext(ctx, http.MethodGet, findURL, nil)
	findResp2, err := c.httpClient.Do(findReq2)
	if err != nil {
		return "", fmt.Errorf("registry.servicecomb: re-find microservice: %w", err)
	}
	if serviceID, ok := c.matchMicroservice(findResp2, cfg.AppID, name); ok {
		return serviceID, nil
	}
	return "", fmt.Errorf("registry.servicecomb: cannot resolve microservice id for %q (register status %d)", name, createResp.StatusCode)
}

// matchMicroservice 从 find 响应中按 appId+serviceName 精确匹配 microservice。
func (c *httpServiceCombClient) matchMicroservice(resp *http.Response, appID, name string) (string, bool) {
	defer resp.Body.Close()
	var body struct {
		Services []struct {
			ServiceID  string `json:"serviceId"`
			ServiceName string `json:"serviceName"`
			AppID      string `json:"appId"`
		} `json:"services"`
	}
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	for _, svc := range body.Services {
		if svc.ServiceID != "" && svc.ServiceName == name && svc.AppID == appID {
			return svc.ServiceID, true
		}
	}
	return "", false
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
				// ServiceCenter 对 push 模式强制要求 interval/times 合法
				// （interval >= 1），缺省会以 400 拒绝注册。
				"mode":     "push",
				"interval": heartbeatIntervalSeconds(cfg),
				"times":    3,
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry.servicecomb: register instance failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *httpServiceCombClient) Deregister(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	serviceID, err := c.microserviceID(ctx, cfg, name)
	if err != nil {
		return err
	}
	// ServiceCenter 只能按 instanceId 删除实例（DELETE by endpoint 查询参数
	// 会返回 405）。先查 endpoint 对应的 instanceId，再删除。
	instanceID, err := c.findInstanceID(ctx, cfg, serviceID, addr)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/microservices/%s/instances/%s", baseURL(cfg), serviceID, instanceID)
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry.servicecomb: deregister instance failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// findInstanceID 按 endpoint 查找实例的 instanceId。
func (c *httpServiceCombClient) findInstanceID(ctx context.Context, cfg *ServiceCombConfig, serviceID, addr string) (string, error) {
	instances, err := c.discoverInstances(ctx, cfg, serviceID)
	if err != nil {
		return "", err
	}
	want := "rest://" + addr
	for _, inst := range instances {
		for _, ep := range inst.endpoints {
			if ep == want {
				if inst.instanceID == "" {
					return "", fmt.Errorf("registry.servicecomb: instance %s has empty instanceId", addr)
				}
				return inst.instanceID, nil
			}
		}
	}
	return "", ErrServiceNotFound
}

// Heartbeat 真正续租 ServiceCenter 的实例 lease。
// push 模式下 ServiceCenter 按 healthCheck.interval×times 判定租约过期：
// 仅做 GET 探活不续租，实例会变成"已过期但仍在列表"的僵尸节点
// （DELETE 报不存在、GET 仍列出，直到被 GC）。
func (c *httpServiceCombClient) Heartbeat(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	serviceID, err := c.microserviceID(ctx, cfg, name)
	if err != nil {
		return err
	}
	instanceID, err := c.findInstanceID(ctx, cfg, serviceID, addr)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/microservices/%s/instances/%s/heartbeat", baseURL(cfg), serviceID, instanceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("registry.servicecomb: heartbeat: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry.servicecomb: heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// serviceCombInstance 是 ServiceCenter 实例的原始响应结构。
type serviceCombInstance struct {
	instanceID string
	serviceID  string
	endpoints  []string
	status     string
	properties map[string]string
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
	instances, err := c.discoverInstances(ctx, cfg, serviceID)
	if err != nil {
		return nil, err
	}
	result := make([]transportcontract.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		addr := instanceAddrFromEndpoints(inst.endpoints)
		if addr == "" {
			continue
		}
		result = append(result, transportcontract.ServiceInstance{
			ID:       inst.instanceID,
			Name:     name,
			Address:  addr,
			Metadata: inst.properties,
			Healthy:  strings.EqualFold(inst.status, "UP"),
		})
	}
	if len(result) == 0 {
		return nil, ErrServiceNotFound
	}
	sortServiceInstances(result)
	return result, nil
}

// discoverInstances 拉取某个 microservice 的全部实例原始数据。
func (c *httpServiceCombClient) discoverInstances(ctx context.Context, cfg *ServiceCombConfig, serviceID string) ([]serviceCombInstance, error) {
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

	out := make([]serviceCombInstance, 0, len(body.Instances))
	for _, inst := range body.Instances {
		out = append(out, serviceCombInstance{
			instanceID: inst.InstanceID,
			serviceID:  inst.ServiceID,
			endpoints:  inst.Endpoints,
			status:     inst.Status,
			properties: inst.Properties,
		})
	}
	return out, nil
}

// heartbeatIntervalSeconds 返回健康检查间隔秒数（>=1）。
func heartbeatIntervalSeconds(cfg *ServiceCombConfig) int {
	sec := int(cfg.HeartbeatInterval.Seconds())
	if sec < 1 {
		sec = 30
	}
	return sec
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
