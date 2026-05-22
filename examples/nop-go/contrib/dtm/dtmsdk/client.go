// Package dtmsdk provides DTM distributed transaction client for the gorp framework.
// This file implements the DTMClient contract with HTTP-based transaction coordination.
//
// 本包提供 gorp 框架 DTM 分布式事务客户端实现。
// 本文件实现 DTMClient 契约，基于 HTTP 的事务协调。
package dtmsdk

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

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ErrDTMSDKNotImported indicates advanced transaction modes require the official SDK.
//
// ErrDTMSDKNotImported 表示高级事务模式仍需官方 SDK。
var ErrDTMSDKNotImported = errors.New(`dtm: advanced transaction modes still require the official SDK, please run:
  go get github.com/dtm-labs/client

The lightweight framework adapter currently supports saga submit and query.
For more details, see https://dtm.pub`)

// DTMClient implements integrationcontract.DTMClient using HTTP-based DTM API.
// Provides SAGA, TCC, XA transaction patterns and barrier execution.
//
// DTMClient 使用基于 HTTP 的 DTM API 实现 integrationcontract.DTMClient。
// 提供 SAGA、TCC、XA 事务模式和 barrier 执行。
type DTMClient struct {
	cfg        *integrationcontract.DTMConfig
	httpClient *http.Client
}

// HTTPClientProvider allows extracting the underlying HTTP client.
//
// HTTPClientProvider 允许提取底层 HTTP 客户端。
type HTTPClientProvider interface {
	HTTPClient() *http.Client
}

// NewDTMClient creates a new DTM client instance with configuration.
//
// NewDTMClient 根据配置创建新的 DTM 客户端实例。
func NewDTMClient(cfg *integrationcontract.DTMConfig) (*DTMClient, error) {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &DTMClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Underlying returns the DTMClient itself for native access.
//
// Underlying 返回 DTMClient 本身供原生访问。
func (c *DTMClient) Underlying() any {
	return c
}

// As attempts to cast the DTMClient to the target type.
//
// As 尝试将 DTMClient 转换为目标类型。
func (c *DTMClient) As(target any) bool {
	return As(c, target)
}

// Close closes the underlying HTTP client connections.
// Implements io.Closer for container lifecycle management.
//
// Close 关闭底层 HTTP 客户端连接。
// 实现 io.Closer 供容器生命周期管理。
func (c *DTMClient) Close() error {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
}

// HTTPClient returns the underlying HTTP client for advanced usage.
// Implements HTTPClientProvider interface.
//
// HTTPClient 返回底层 HTTP 客户端供高级使用。
// 实现 HTTPClientProvider 接口。
func (c *DTMClient) HTTPClient() *http.Client {
	return c.httpClient
}

// SAGA creates a new SAGA transaction builder.
// Implements integrationcontract.DTMClient.SAGA.
//
// SAGA 创建新的 SAGA 事务构建器。
// 实现 integrationcontract.DTMClient.SAGA。
func (c *DTMClient) SAGA(name string) integrationcontract.SAGABuilder {
	return &sagaBuilder{client: c, name: name}
}

// TCC creates a new TCC transaction builder.
// Implements integrationcontract.DTMClient.TCC.
//
// TCC 创建新的 TCC 事务构建器。
// 实现 integrationcontract.DTMClient.TCC。
func (c *DTMClient) TCC(name string) integrationcontract.TCCBuilder {
	return &tccBuilder{client: c, name: name}
}

// XA creates a new XA transaction builder.
// Implements integrationcontract.DTMClient.XA.
//
// XA 创建新的 XA 事务构建器。
// 实现 integrationcontract.DTMClient.XA。
func (c *DTMClient) XA(name string) integrationcontract.XABuilder {
	return &xaBuilder{client: c, name: name}
}

// Barrier creates a barrier handler for the given transaction type and gid.
// Implements integrationcontract.DTMClient.Barrier.
//
// Barrier 为给定事务类型和 gid 创建 barrier handler。
// 实现 integrationcontract.DTMClient.Barrier。
func (c *DTMClient) Barrier(transType, gid string) integrationcontract.BarrierHandler {
	return &barrierHandler{client: c, transType: transType, gid: gid}
}

// Query retrieves transaction information by gid.
// Implements integrationcontract.DTMClient.Query.
//
// Query 根据 gid 查询事务信息。
// 实现 integrationcontract.DTMClient.Query。
func (c *DTMClient) Query(ctx context.Context, gid string) (*integrationcontract.TransactionInfo, error) {
	gid = strings.TrimSpace(gid)
	if gid == "" {
		return nil, errors.New("dtm: gid is required")
	}

	queryURL := fmt.Sprintf("%s/query?gid=%s", c.apiBaseURL(), urlEncodeGID(gid))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dtm: build query request failed: %w", err)
	}

	var raw map[string]any
	if err := c.doJSON(req, &raw); err != nil {
		return nil, fmt.Errorf("dtm: query failed: %w", err)
	}

	return parseTransactionInfo(raw, gid), nil
}

// newGID generates a new global transaction ID from DTM server.
//
// newGID 从 DTM 服务器生成新的全局事务 ID。
func (c *DTMClient) newGID(ctx context.Context) (string, error) {
	newGIDURL := c.apiBaseURL() + "/newGid"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, newGIDURL, nil)
	if err != nil {
		return "", fmt.Errorf("dtm: build newGid request failed: %w", err)
	}
	var result struct {
		GID string `json:"gid"`
	}
	if err := c.doJSON(req, &result); err != nil {
		return "", fmt.Errorf("dtm: newGid failed: %w", err)
	}
	if strings.TrimSpace(result.GID) == "" {
		return "", errors.New("dtm: empty gid returned from server")
	}
	return result.GID, nil
}

// submit sends a transaction submit request to DTM server.
//
// submit 向 DTM 服务器发送事务提交请求。
func (c *DTMClient) submit(ctx context.Context, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("dtm: marshal submit request failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL()+"/submit", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("dtm: build submit request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.doJSON(req, nil); err != nil {
		return fmt.Errorf("dtm: submit failed: %w", err)
	}
	return nil
}

// apiBaseURL constructs the DTM API base URL from configuration.
//
// apiBaseURL 从配置构造 DTM API 基础 URL。
func (c *DTMClient) apiBaseURL() string {
	endpoint := strings.TrimRight(strings.TrimSpace(c.cfg.Endpoint), "/")
	if endpoint == "" {
		endpoint = "http://localhost:36789"
	}
	if strings.HasSuffix(endpoint, "/api/dtmsvr") {
		return endpoint
	}
	return endpoint + "/api/dtmsvr"
}

// doJSON executes an HTTP request with JSON handling and retry logic.
//
// doJSON 执行 HTTP 请求并处理 JSON 响应和重试逻辑。
func (c *DTMClient) doJSON(req *http.Request, out any) error {
	var lastErr error
	attempts := c.cfg.RetryCount + 1
	if attempts <= 0 {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		cloned := req.Clone(req.Context())
		if req.GetBody != nil && req.Body != nil {
			body, err := req.GetBody()
			if err != nil {
				return err
			}
			cloned.Body = body
		}

		resp, err := c.httpClient.Do(cloned)
		if err != nil {
			lastErr = err
		} else {
			lastErr = decodeResponse(resp, out)
			if lastErr == nil {
				return nil
			}
		}

		if attempt == attempts-1 || !shouldRetryDTM(lastErr) {
			break
		}
		if err := sleepRetry(req.Context(), time.Duration(c.cfg.RetryInterval)*time.Second); err != nil {
			return err
		}
	}
	return lastErr
}

// decodeResponse decodes HTTP response into target structure.
//
// decodeResponse 解码 HTTP 响应到目标结构。
func decodeResponse(resp *http.Response, out any) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return errors.New(message)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// shouldRetryDTM determines if the error is retryable for DTM operations.
//
// shouldRetryDTM 判断 DTM 操作错误是否可重试。
func shouldRetryDTM(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") ||
		strings.Contains(message, "tempor") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "bad gateway") ||
		strings.Contains(message, "service unavailable") ||
		strings.Contains(message, "gateway timeout") ||
		strings.Contains(message, "internal server error")
}

// sleepRetry waits for retry interval with context cancellation support.
//
// sleepRetry 等待重试间隔，支持 context 取消。
func sleepRetry(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = time.Second
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// marshalPayload converts payload to JSON string for DTM API.
//
// marshalPayload 将 payload 转换为 JSON 字符串供 DTM API 使用。
func marshalPayload(payload any) (string, error) {
	if payload == nil {
		return "null", nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// urlEncodeGID encodes gid for URL query parameter.
// Uses url.QueryEscape to prevent URL injection when gid contains special characters
// like &, ?, # that could alter the request path or parameters.
//
// urlEncodeGID 对 gid 进行 URL 编码。
// 使用 url.QueryEscape 防止 gid 中包含 &、?、# 等特殊字符导致的 URL 注入。
func urlEncodeGID(gid string) string {
	return url.QueryEscape(gid)
}
