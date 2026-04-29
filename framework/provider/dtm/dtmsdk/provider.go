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

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 DTM 分布式事务客户端适配器。
//
// 中文说明：
// - DTM 是成熟的分布式事务管理器（https://dtm.pub）；
// - framework 不自研事务协调器，而是对官方 HTTP API 做轻量适配；
// - 当前已经支持最小可用的 SAGA submit 与 Query 闭环；
// - TCC / XA / Barrier 仍保留轻量骨架，后续若需要更完整能力，再接官方 SDK。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "dtm.sdk" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.DTMKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.DTMKey, func(c contract.Container) (any, error) {
		cfg, err := getDTMConfig(c)
		if err != nil {
			return nil, err
		}
		return NewDTMClient(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getDTMConfig 从容器获取 DTM 配置。
func getDTMConfig(c contract.Container) (*contract.DTMConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("dtm: invalid config service")
	}

	dtmCfg := &contract.DTMConfig{
		Enabled:         true,
		Endpoint:        "http://localhost:36789",
		Timeout:         10,
		RetryCount:      3,
		RetryInterval:   5,
		CallbackPort:    8080,
		CallbackAddress: "localhost",
	}

	if endpoint := cfg.GetString("dtm.endpoint"); endpoint != "" {
		dtmCfg.Endpoint = endpoint
	}
	if enabled := cfg.GetBool("dtm.enabled"); !enabled {
		dtmCfg.Enabled = false
	}
	if timeout := cfg.GetInt("dtm.timeout"); timeout > 0 {
		dtmCfg.Timeout = timeout
	}
	if port := cfg.GetInt("dtm.callback_port"); port > 0 {
		dtmCfg.CallbackPort = port
	}
	if addr := cfg.GetString("dtm.callback_address"); addr != "" {
		dtmCfg.CallbackAddress = addr
	}
	if retryCount := cfg.GetInt("dtm.retry_count"); retryCount > 0 {
		dtmCfg.RetryCount = retryCount
	}
	if retryInterval := cfg.GetInt("dtm.retry_interval"); retryInterval > 0 {
		dtmCfg.RetryInterval = retryInterval
	}

	return dtmCfg, nil
}

// ErrDTMSDKNotImported 表示更完整的高级事务模式仍需引入官方 SDK。
var ErrDTMSDKNotImported = errors.New(`dtm: advanced transaction modes still require the official SDK, please run:
  go get github.com/dtm-labs/client

The lightweight framework adapter currently supports saga submit and query.
For more details, see https://dtm.pub`)

// DTMClient 是 DTM 客户端适配器。
//
// 中文说明：
// - 这是面向 framework 的轻量实现，而不是完整复刻官方 SDK；
// - 当前聚焦“真正最小可用链路”：newGid / saga submit / query；
// - 这样 starter 和样板项目已经可以验证 DTM 基础接线路径。
type DTMClient struct {
	cfg        *contract.DTMConfig
	httpClient *http.Client
}

// NewDTMClient 创建 DTM 客户端。
func NewDTMClient(cfg *contract.DTMConfig) (*DTMClient, error) {
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

// SAGA 创建 SAGA 事务。
func (c *DTMClient) SAGA(name string) contract.SAGABuilder {
	return &sagaBuilder{client: c, name: name}
}

// TCC 创建 TCC 事务。
func (c *DTMClient) TCC(name string) contract.TCCBuilder {
	return &tccBuilder{client: c, name: name}
}

// XA 创建 XA 事务。
func (c *DTMClient) XA(name string) contract.XABuilder {
	return &xaBuilder{client: c, name: name}
}

// Barrier 创建 Barrier 事务。
func (c *DTMClient) Barrier(transType, gid string) contract.BarrierHandler {
	return &barrierHandler{client: c, transType: transType, gid: gid}
}

// Query 查询事务状态。
//
// 中文说明：
// - 通过 DTM HTTP API `GET /api/dtmsvr/query?gid=...` 查询；
// - 返回字段做了宽松兼容，避免后续 DTM 响应细节轻微变化时直接失效；
// - 对当前 framework 来说，只要能拿到 gid / status / trans_type / steps 就足够支撑最小验证链路。
func (c *DTMClient) Query(ctx context.Context, gid string) (*contract.TransactionInfo, error) {
	gid = strings.TrimSpace(gid)
	if gid == "" {
		return nil, errors.New("dtm: gid is required")
	}

	queryURL := fmt.Sprintf("%s/query?gid=%s", c.apiBaseURL(), url.QueryEscape(gid))
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

type sagaSubmitRequest struct {
	GID            string              `json:"gid"`
	TransType      string              `json:"trans_type"`
	Steps          []map[string]string `json:"steps,omitempty"`
	Payloads       []string            `json:"payloads,omitempty"`
	RetryInterval  int64               `json:"retry_interval,omitempty"`
	RetryCount     int64               `json:"retry_count,omitempty"`
	RequestTimeout int64               `json:"request_timeout,omitempty"`
}

// sagaBuilder 是 SAGA 事务构建器。
type sagaBuilder struct {
	client *DTMClient
	name   string
	gid    string
	steps  []sagaStep
}

type sagaStep struct {
	action        string
	compensate    string
	payload       any
	retryCount    int
	retryInterval int
	timeout       int
}

func (b *sagaBuilder) Add(action string, compensate string, payload any) contract.SAGABuilder {
	b.steps = append(b.steps, sagaStep{
		action:     action,
		compensate: compensate,
		payload:    payload,
	})
	return b
}

func (b *sagaBuilder) AddBranch(action string, compensate string, payload any, opts contract.BranchOptions) contract.SAGABuilder {
	b.steps = append(b.steps, sagaStep{
		action:        action,
		compensate:    compensate,
		payload:       payload,
		retryCount:    opts.RetryCount,
		retryInterval: opts.RetryInterval,
		timeout:       opts.Timeout,
	})
	return b
}

func (b *sagaBuilder) Submit(ctx context.Context) error {
	if len(b.steps) == 0 {
		return errors.New("dtm: saga has no steps")
	}

	gid, err := b.client.newGID(ctx)
	if err != nil {
		return err
	}

	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.action) == "" || strings.TrimSpace(step.compensate) == "" {
			return errors.New("dtm: saga step action and compensate are required")
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return fmt.Errorf("dtm: marshal saga payload failed: %w", err)
		}
		steps = append(steps, map[string]string{
			"action":     step.action,
			"compensate": step.compensate,
		})
		payloads = append(payloads, payload)
	}

	body := sagaSubmitRequest{
		GID:            gid,
		TransType:      "saga",
		Steps:          steps,
		Payloads:       payloads,
		RetryInterval:  int64(b.client.cfg.RetryInterval),
		RetryCount:     int64(b.client.cfg.RetryCount),
		RequestTimeout: int64(b.client.cfg.Timeout),
	}
	if err := b.client.submit(ctx, body); err != nil {
		return err
	}

	b.gid = gid
	return nil
}

func (b *sagaBuilder) Build() (*contract.SAGATransaction, error) {
	steps := make([]contract.SAGAStep, len(b.steps))
	for i, step := range b.steps {
		steps[i] = contract.SAGAStep{
			Action:        step.action,
			Compensate:    step.compensate,
			Payload:       step.payload,
			RetryCount:    step.retryCount,
			RetryInterval: step.retryInterval,
			Timeout:       step.timeout,
		}
	}
	return &contract.SAGATransaction{GID: b.gid, Steps: steps}, nil
}

// tccBuilder 是 TCC 事务构建器骨架。
type tccBuilder struct {
	client *DTMClient
	name   string
	steps  []tccStep
}

type tccStep struct {
	try     string
	confirm string
	cancel  string
	payload any
}

func (b *tccBuilder) Add(try string, confirm string, cancel string, payload any) contract.TCCBuilder {
	b.steps = append(b.steps, tccStep{
		try:     try,
		confirm: confirm,
		cancel:  cancel,
		payload: payload,
	})
	return b
}

func (b *tccBuilder) Submit(ctx context.Context) error {
	return ErrDTMSDKNotImported
}

// xaBuilder 是 XA 事务构建器骨架。
type xaBuilder struct {
	client *DTMClient
	name   string
	steps  []xaStep
}

type xaStep struct {
	url     string
	payload any
}

func (b *xaBuilder) Add(url string, payload any) contract.XABuilder {
	b.steps = append(b.steps, xaStep{url: url, payload: payload})
	return b
}

func (b *xaBuilder) Submit(ctx context.Context) error {
	return ErrDTMSDKNotImported
}

// barrierHandler 是 Barrier 事务处理器骨架。
type barrierHandler struct {
	client    *DTMClient
	transType string
	gid       string
}

func (h *barrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	return fn(nil)
}

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

func (c *DTMClient) doJSON(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
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

func parseTransactionInfo(raw map[string]any, fallbackGID string) *contract.TransactionInfo {
	payload := unwrapTransactionPayload(raw)
	info := &contract.TransactionInfo{
		GID:             fallbackGID,
		Status:          stringFromMap(payload, "status"),
		TransactionType: stringFromMap(payload, "trans_type", "transaction_type"),
		CreateTime:      int64FromMap(payload, "create_time", "createTime"),
		UpdateTime:      int64FromMap(payload, "update_time", "updateTime"),
		Steps:           parseTransactionSteps(payload),
	}
	if gid := stringFromMap(payload, "gid"); gid != "" {
		info.GID = gid
	}
	if info.Status == "" {
		info.Status = "unknown"
	}
	if info.TransactionType == "" {
		info.TransactionType = "unknown"
	}
	return info
}

func unwrapTransactionPayload(raw map[string]any) map[string]any {
	for _, key := range []string{"transaction", "data", "result"} {
		if nested, ok := raw[key].(map[string]any); ok && len(nested) > 0 {
			return nested
		}
	}
	return raw
}

func parseTransactionSteps(raw map[string]any) []contract.TransactionStep {
	values, ok := raw["steps"].([]any)
	if !ok {
		return nil
	}
	steps := make([]contract.TransactionStep, 0, len(values))
	for _, value := range values {
		stepMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		steps = append(steps, contract.TransactionStep{
			BranchID: stringFromMap(stepMap, "branch_id", "branchID"),
			Status:   stringFromMap(stepMap, "status"),
			Op:       stringFromMap(stepMap, "op"),
			URL:      stringFromMap(stepMap, "url", "action", "try", "confirm", "cancel"),
		})
	}
	return steps
}

func stringFromMap(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if s := stringFromAny(value); s != "" {
				return s
			}
		}
	}
	return ""
}

func int64FromMap(values map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if n, ok := int64FromAny(value); ok {
				return n
			}
		}
	}
	return 0
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}

func int64FromAny(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	case string:
		if v == "" {
			return 0, false
		}
		parsed, err := json.Number(v).Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}
