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

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "dtm.sdk" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{integrationcontract.DTMKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.DTMKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getDTMConfig(c)
		if err != nil {
			return nil, err
		}
		return NewDTMClient(cfg)
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func getDTMConfig(c runtimecontract.Container) (*integrationcontract.DTMConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("dtm: invalid config service")
	}

	dtmCfg := &integrationcontract.DTMConfig{
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

var ErrDTMSDKNotImported = errors.New(`dtm: advanced transaction modes still require the official SDK, please run:
  go get github.com/dtm-labs/client

The lightweight framework adapter currently supports saga submit and query.
For more details, see https://dtm.pub`)

var (
	ErrSagaNoSteps            = errors.New("dtm: saga has no steps")
	ErrSagaStepRequired       = errors.New("dtm: saga step action and compensate are required")
	ErrTCCNoSteps             = errors.New("dtm: tcc has no steps")
	ErrTCCStepRequired        = errors.New("dtm: tcc step try/confirm/cancel are required")
	ErrXANoSteps              = errors.New("dtm: xa has no steps")
	ErrXAStepRequired         = errors.New("dtm: xa step url is required")
	ErrBarrierTransType       = errors.New("dtm: barrier transType is required")
	ErrBarrierUnsupportedType = errors.New("dtm: barrier transType is unsupported")
	ErrBarrierGID             = errors.New("dtm: barrier gid is required")
	ErrBarrierCallback        = errors.New("dtm: barrier callback is required")
)

type DTMClient struct {
	cfg        *integrationcontract.DTMConfig
	httpClient *http.Client
}

type HTTPClientProvider interface {
	HTTPClient() *http.Client
}

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

func (c *DTMClient) Underlying() any {
	return c
}

func (c *DTMClient) As(target any) bool {
	return internalnative.As(c, target)
}

func (c *DTMClient) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *DTMClient) SAGA(name string) integrationcontract.SAGABuilder {
	return &sagaBuilder{client: c, name: name}
}

func (c *DTMClient) TCC(name string) integrationcontract.TCCBuilder {
	return &tccBuilder{client: c, name: name}
}

func (c *DTMClient) XA(name string) integrationcontract.XABuilder {
	return &xaBuilder{client: c, name: name}
}

func (c *DTMClient) Barrier(transType, gid string) integrationcontract.BarrierHandler {
	return &barrierHandler{client: c, transType: transType, gid: gid}
}

func (c *DTMClient) Query(ctx context.Context, gid string) (*integrationcontract.TransactionInfo, error) {
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

func (b *sagaBuilder) Add(action string, compensate string, payload any) integrationcontract.SAGABuilder {
	b.steps = append(b.steps, sagaStep{action: action, compensate: compensate, payload: payload})
	return b
}

func (b *sagaBuilder) AddBranch(action string, compensate string, payload any, opts integrationcontract.BranchOptions) integrationcontract.SAGABuilder {
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
	steps, payloads, err := b.buildSteps()
	if err != nil {
		return err
	}
	gid, err := b.client.newGID(ctx)
	if err != nil {
		return err
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

func (b *sagaBuilder) Build() (*integrationcontract.SAGATransaction, error) {
	steps := make([]integrationcontract.SAGAStep, len(b.steps))
	for i, step := range b.steps {
		steps[i] = integrationcontract.SAGAStep{
			Action:        step.action,
			Compensate:    step.compensate,
			Payload:       step.payload,
			RetryCount:    step.retryCount,
			RetryInterval: step.retryInterval,
			Timeout:       step.timeout,
		}
	}
	return &integrationcontract.SAGATransaction{GID: b.gid, Steps: steps}, nil
}

func (b *sagaBuilder) buildSteps() ([]map[string]string, []string, error) {
	if len(b.steps) == 0 {
		return nil, nil, ErrSagaNoSteps
	}
	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.action) == "" || strings.TrimSpace(step.compensate) == "" {
			return nil, nil, ErrSagaStepRequired
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return nil, nil, fmt.Errorf("dtm: marshal saga payload failed: %w", err)
		}
		steps = append(steps, map[string]string{"action": step.action, "compensate": step.compensate})
		payloads = append(payloads, payload)
	}
	return steps, payloads, nil
}

type tccBuilder struct {
	client *DTMClient
	name   string
	gid    string
	steps  []tccStep
}

type tccStep struct {
	try     string
	confirm string
	cancel  string
	payload any
}

type TCCTransaction struct {
	GID   string
	Steps []TCCStep
}

type TCCStep struct {
	Try     string
	Confirm string
	Cancel  string
	Payload any
}

func (b *tccBuilder) Add(try string, confirm string, cancel string, payload any) integrationcontract.TCCBuilder {
	b.steps = append(b.steps, tccStep{try: try, confirm: confirm, cancel: cancel, payload: payload})
	return b
}

func (b *tccBuilder) Submit(ctx context.Context) error {
	steps, payloads, err := b.buildSteps()
	if err != nil {
		return err
	}
	gid, err := b.client.newGID(ctx)
	if err != nil {
		return err
	}
	body := map[string]any{
		"gid":             gid,
		"trans_type":      "tcc",
		"steps":           steps,
		"payloads":        payloads,
		"retry_interval":  int64(b.client.cfg.RetryInterval),
		"retry_count":     int64(b.client.cfg.RetryCount),
		"request_timeout": int64(b.client.cfg.Timeout),
	}
	if err := b.client.submit(ctx, body); err != nil {
		return err
	}
	b.gid = gid
	return nil
}

func (b *tccBuilder) Build() (*TCCTransaction, error) {
	if len(b.steps) == 0 {
		return nil, ErrTCCNoSteps
	}
	steps := make([]TCCStep, len(b.steps))
	for i, step := range b.steps {
		if strings.TrimSpace(step.try) == "" || strings.TrimSpace(step.confirm) == "" || strings.TrimSpace(step.cancel) == "" {
			return nil, ErrTCCStepRequired
		}
		steps[i] = TCCStep{Try: step.try, Confirm: step.confirm, Cancel: step.cancel, Payload: step.payload}
	}
	return &TCCTransaction{GID: b.gid, Steps: steps}, nil
}

func (b *tccBuilder) buildSteps() ([]map[string]string, []string, error) {
	if len(b.steps) == 0 {
		return nil, nil, ErrTCCNoSteps
	}
	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.try) == "" || strings.TrimSpace(step.confirm) == "" || strings.TrimSpace(step.cancel) == "" {
			return nil, nil, ErrTCCStepRequired
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return nil, nil, fmt.Errorf("dtm: marshal tcc payload failed: %w", err)
		}
		steps = append(steps, map[string]string{"try": step.try, "confirm": step.confirm, "cancel": step.cancel})
		payloads = append(payloads, payload)
	}
	return steps, payloads, nil
}

type xaBuilder struct {
	client *DTMClient
	name   string
	gid    string
	steps  []xaStep
}

type xaStep struct {
	url     string
	payload any
}

type XATransaction struct {
	GID   string
	Steps []XAStep
}

type XAStep struct {
	URL     string
	Payload any
}

func (b *xaBuilder) Add(url string, payload any) integrationcontract.XABuilder {
	b.steps = append(b.steps, xaStep{url: url, payload: payload})
	return b
}

func (b *xaBuilder) Submit(ctx context.Context) error {
	steps, payloads, err := b.buildSteps()
	if err != nil {
		return err
	}
	gid, err := b.client.newGID(ctx)
	if err != nil {
		return err
	}
	body := map[string]any{
		"gid":             gid,
		"trans_type":      "xa",
		"steps":           steps,
		"payloads":        payloads,
		"retry_interval":  int64(b.client.cfg.RetryInterval),
		"retry_count":     int64(b.client.cfg.RetryCount),
		"request_timeout": int64(b.client.cfg.Timeout),
	}
	if err := b.client.submit(ctx, body); err != nil {
		return err
	}
	b.gid = gid
	return nil
}

func (b *xaBuilder) Build() (*XATransaction, error) {
	if len(b.steps) == 0 {
		return nil, ErrXANoSteps
	}
	steps := make([]XAStep, len(b.steps))
	for i, step := range b.steps {
		if strings.TrimSpace(step.url) == "" {
			return nil, ErrXAStepRequired
		}
		steps[i] = XAStep{URL: step.url, Payload: step.payload}
	}
	return &XATransaction{GID: b.gid, Steps: steps}, nil
}

func (b *xaBuilder) buildSteps() ([]map[string]string, []string, error) {
	if len(b.steps) == 0 {
		return nil, nil, ErrXANoSteps
	}
	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.url) == "" {
			return nil, nil, ErrXAStepRequired
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return nil, nil, fmt.Errorf("dtm: marshal xa payload failed: %w", err)
		}
		steps = append(steps, map[string]string{"url": step.url})
		payloads = append(payloads, payload)
	}
	return steps, payloads, nil
}

type barrierHandler struct {
	client    *DTMClient
	transType string
	gid       string
}

type BarrierContext struct {
	TransType string
	GID       string
}

func (h *barrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	_ = ctx
	if strings.TrimSpace(h.transType) == "" {
		return ErrBarrierTransType
	}
	if !isSupportedBarrierType(h.transType) {
		return ErrBarrierUnsupportedType
	}
	if strings.TrimSpace(h.gid) == "" {
		return ErrBarrierGID
	}
	if fn == nil {
		return ErrBarrierCallback
	}
	return fn(&BarrierContext{TransType: h.transType, GID: h.gid})
}

func isSupportedBarrierType(transType string) bool {
	switch strings.ToLower(strings.TrimSpace(transType)) {
	case "saga", "tcc", "xa", "msg":
		return true
	default:
		return false
	}
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

func parseTransactionInfo(raw map[string]any, fallbackGID string) *integrationcontract.TransactionInfo {
	payload := unwrapTransactionPayload(raw)
	info := &integrationcontract.TransactionInfo{
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

func parseTransactionSteps(raw map[string]any) []integrationcontract.TransactionStep {
	values, ok := raw["steps"].([]any)
	if !ok {
		return nil
	}
	steps := make([]integrationcontract.TransactionStep, 0, len(values))
	for _, value := range values {
		stepMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		steps = append(steps, integrationcontract.TransactionStep{
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
