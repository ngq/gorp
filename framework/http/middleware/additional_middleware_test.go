// Package middleware_test provides unit tests for extended HTTP middleware catalog.
//
// 适用场景：
// - 验证除超时、限流和幂等之外的更完整主线中间件目录。
// - 锁定请求标识、过载保护、缓存、语言、鉴权、审计、校验和指标等行为。
// - 通过基于 Gin 的请求流，为 transport 中间件提供统一回归套件。
package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
	prometheus "github.com/prometheus/client_golang/prometheus"
)

type stubCircuitBreaker struct {
	allowErr         error
	allowedResources []string
	successResources []string
	failureResources []string
}

func (s *stubCircuitBreaker) Allow(context.Context, string) error { return s.allowErr }

func (s *stubCircuitBreaker) RecordSuccess(_ context.Context, resource string) {
	s.successResources = append(s.successResources, resource)
}

func (s *stubCircuitBreaker) RecordFailure(_ context.Context, resource string, _ error) {
	s.failureResources = append(s.failureResources, resource)
}

func (s *stubCircuitBreaker) Do(context.Context, string, func() error) error { return nil }

func (s *stubCircuitBreaker) State(context.Context, string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
}

// stubLogger is a lightweight in-memory logger used for logging, audit, and recovery assertions.
//
// stubLogger 是一个轻量内存 logger，用于 logging、audit 和 recovery 断言。
type logEntry struct {
	level  string
	msg    string
	fields []observabilitycontract.Field
}

type stubLogger struct {
	fields  []observabilitycontract.Field
	entries *[]logEntry
}

func newStubLogger() *stubLogger {
	entries := make([]logEntry, 0, 8)
	return &stubLogger{entries: &entries}
}

func (l *stubLogger) Debug(msg string, fields ...observabilitycontract.Field) {
	l.append("debug", msg, fields...)
}

func (l *stubLogger) Info(msg string, fields ...observabilitycontract.Field) {
	l.append("info", msg, fields...)
}

func (l *stubLogger) Warn(msg string, fields ...observabilitycontract.Field) {
	l.append("warn", msg, fields...)
}

func (l *stubLogger) Error(msg string, fields ...observabilitycontract.Field) {
	l.append("error", msg, fields...)
}

func (l *stubLogger) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	combined := append(append([]observabilitycontract.Field{}, l.fields...), fields...)
	return &stubLogger{fields: combined, entries: l.entries}
}

func (l *stubLogger) append(level string, msg string, fields ...observabilitycontract.Field) {
	combined := append(append([]observabilitycontract.Field{}, l.fields...), fields...)
	*l.entries = append(*l.entries, logEntry{level: level, msg: msg, fields: combined})
}

func (l *stubLogger) Entries() []logEntry {
	return append([]logEntry(nil), (*l.entries)...)
}

type stubValidator struct {
	validateFn func(context.Context, any) error
}

func (v *stubValidator) Validate(ctx context.Context, obj any) error {
	if v.validateFn != nil {
		return v.validateFn(ctx, obj)
	}
	return nil
}

func (v *stubValidator) ValidateVar(context.Context, any, string) error { return nil }
func (v *stubValidator) RegisterCustom(string, datacontract.CustomValidateFunc) error {
	return nil
}
func (v *stubValidator) SetLocale(string) error { return nil }
func (v *stubValidator) TranslateError(err error) resiliencecontract.AppError {
	if appErr, ok := err.(resiliencecontract.AppError); ok {
		return appErr
	}
	return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
}

func fieldValue(fields []observabilitycontract.Field, key string) any {
	for _, field := range fields {
		if field.Key == key {
			return field.Value
		}
	}
	return nil
}

func counterValue(metricName string, labels map[string]string) float64 {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0
	}
	for _, family := range families {
		if family.GetName() != metricName {
			continue
		}
		for _, metric := range family.GetMetric() {
			matched := true
			for _, pair := range metric.GetLabel() {
				expected, ok := labels[pair.GetName()]
				if !ok || expected != pair.GetValue() {
					matched = false
					break
				}
			}
			if matched && len(metric.GetLabel()) == len(labels) && metric.GetCounter() != nil {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
}

// TestRequestIdentityInjectsHeadersAndContext verifies automatic request identity generation and propagation.
//
// TestRequestIdentityInjectsHeadersAndContext 验证请求标识的自动生成与传播。
func TestRequestIdentityInjectsHeadersAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, RequestIdentity())
	router.GET("/identity", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"request_id": GetRequestID(c),
			"trace_id":   GetTraceID(c),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header")
	}
	if recorder.Header().Get("X-Trace-Id") == "" {
		t.Fatal("expected X-Trace-Id header")
	}
	if recorder.Header().Get("X-Request-Id") != recorder.Header().Get("X-Trace-Id") {
		t.Fatal("expected trace id to fall back to request id when no trace header is provided")
	}
}

// TestRequestIdentityReusesProvidedHeaders verifies that incoming request and trace ids are preserved.
//
// TestRequestIdentityReusesProvidedHeaders 验证传入的 request id 与 trace id 会被保留。
func TestRequestIdentityReusesProvidedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, RequestIdentity())
	router.GET("/identity", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	req.Header.Set("X-Request-Id", "req-123")
	req.Header.Set("X-Trace-Id", "trace-456")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("X-Request-Id") != "req-123" {
		t.Fatalf("expected request id passthrough, got %q", recorder.Header().Get("X-Request-Id"))
	}
	if recorder.Header().Get("X-Trace-Id") != "trace-456" {
		t.Fatalf("expected trace id passthrough, got %q", recorder.Header().Get("X-Trace-Id"))
	}
}

// =============================================================================
// 请求标识与过载保护
// =============================================================================

// TestLoadSheddingRejectsWhenConcurrentSlotsAreFull verifies fast failure when no concurrency slot is available.
//
// TestLoadSheddingRejectsWhenConcurrentSlotsAreFull 验证无可用并发槽位时会快速失败。
func TestLoadSheddingRejectsWhenConcurrentSlotsAreFull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, LoadShedding(1))

	started := make(chan struct{})
	release := make(chan struct{})
	router.GET("/busy", func(c *gin.Context) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-release
		c.Status(http.StatusNoContent)
	})

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/busy", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		firstDone <- recorder
	}()

	<-started

	req := httptest.NewRequest(http.MethodGet, "/busy", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}

	close(release)
	<-firstDone
}

// TestConcurrencyLimitWaitsForSlotWithinTimeout verifies queued waiting within the configured timeout budget.
//
// TestConcurrencyLimitWaitsForSlotWithinTimeout 验证会在配置的超时预算内等待空闲槽位。
func TestConcurrencyLimitWaitsForSlotWithinTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, ConcurrencyLimit(1, 100*time.Millisecond))

	started := make(chan struct{})
	release := make(chan struct{})
	router.GET("/wait", func(c *gin.Context) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-release
		c.Status(http.StatusNoContent)
	})

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/wait", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		firstDone <- recorder
	}()

	<-started

	secondDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/wait", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		secondDone <- recorder
	}()

	time.Sleep(20 * time.Millisecond)
	close(release)

	firstRecorder := <-firstDone
	secondRecorder := <-secondDone

	if firstRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected first request 204, got %d", firstRecorder.Code)
	}
	if secondRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected second request to wait and succeed with 204, got %d", secondRecorder.Code)
	}
}

// TestCircuitBreakerRejectsWhenOpen verifies that open breakers return the unified service-unavailable response.
//
// TestCircuitBreakerRejectsWhenOpen 验证熔断器打开时会返回统一的服务不可用响应。
func TestCircuitBreakerRejectsWhenOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	breaker := &stubCircuitBreaker{allowErr: errors.New("open")}
	applyTransportMiddleware(router, CircuitBreaker(breaker, ""))
	router.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

// TestCircuitBreakerRecordsSuccessAndFailureWithRouteResource verifies breaker success and failure accounting by route resource.
//
// TestCircuitBreakerRecordsSuccessAndFailureWithRouteResource 验证熔断器按路由资源记录成功与失败。
func TestCircuitBreakerRecordsSuccessAndFailureWithRouteResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	successBreaker := &stubCircuitBreaker{}
	successRouter := gin.New()
	applyTransportMiddleware(successRouter, CircuitBreaker(successBreaker, ""))
	successRouter.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	successReq := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	successRecorder := httptest.NewRecorder()
	successRouter.ServeHTTP(successRecorder, successReq)

	if len(successBreaker.successResources) != 1 {
		t.Fatalf("expected one success record, got %d", len(successBreaker.successResources))
	}
	if successBreaker.successResources[0] != "http.get.orders.id" {
		t.Fatalf("unexpected success resource %q", successBreaker.successResources[0])
	}

	failureBreaker := &stubCircuitBreaker{}
	failureRouter := gin.New()
	applyTransportMiddleware(failureRouter, CircuitBreaker(failureBreaker, ""))
	failureRouter.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	failureReq := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	failureRecorder := httptest.NewRecorder()
	failureRouter.ServeHTTP(failureRecorder, failureReq)

	if len(failureBreaker.failureResources) != 1 {
		t.Fatalf("expected one failure record, got %d", len(failureBreaker.failureResources))
	}
	if failureBreaker.failureResources[0] != "http.get.orders.id" {
		t.Fatalf("unexpected failure resource %q", failureBreaker.failureResources[0])
	}
}

// =============================================================================
// 缓存、压缩与 body 限制
// =============================================================================

// TestCacheControlWritesHeader verifies fixed Cache-Control output.
//
// TestCacheControlWritesHeader 验证固定 Cache-Control 头输出。
func TestCacheControlWritesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, CacheControl("public, max-age=60"))
	router.GET("/cache", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/cache", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Cache-Control") != "public, max-age=60" {
		t.Fatalf("unexpected Cache-Control header %q", recorder.Header().Get("Cache-Control"))
	}
}

// TestETagWritesHeaderAndReturnsNotModifiedOnMatch verifies ETag generation and conditional 304 responses.
//
// TestETagWritesHeaderAndReturnsNotModifiedOnMatch 验证 ETag 生成与条件 304 响应。
func TestETagWritesHeaderAndReturnsNotModifiedOnMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, ETag())
	router.GET("/etag", func(c *gin.Context) {
		c.String(http.StatusOK, "payload")
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/etag", nil)
	firstRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstRecorder, firstReq)

	etag := firstRecorder.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header")
	}
	if firstRecorder.Body.String() != "payload" {
		t.Fatalf("expected payload body, got %q", firstRecorder.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/etag", nil)
	secondReq.Header.Set("If-None-Match", etag)
	secondRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondRecorder, secondReq)

	if secondRecorder.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", secondRecorder.Code)
	}
	if secondRecorder.Body.Len() != 0 {
		t.Fatalf("expected empty 304 body, got %q", secondRecorder.Body.String())
	}
	if secondRecorder.Header().Get("ETag") != etag {
		t.Fatalf("expected same ETag header, got %q", secondRecorder.Header().Get("ETag"))
	}
}

// =============================================================================
// IP 黑白名单、本地化与安全头
// =============================================================================

// TestIPAllowlistAndDenylist verifies allowlist pass-through and denylist blocking behavior.
//
// TestIPAllowlistAndDenylist 验证 allowlist 放行与 denylist 拦截行为。
func TestIPAllowlistAndDenylist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowRouter := gin.New()
	applyTransportMiddleware(allowRouter, IPAllowlist("10.0.0.0/8"))
	allowRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	allowReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	allowReq.Header.Set("X-Forwarded-For", "10.1.2.3")
	allowRecorder := httptest.NewRecorder()
	allowRouter.ServeHTTP(allowRecorder, allowReq)

	if allowRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected allowlist request 204, got %d", allowRecorder.Code)
	}

	denyRouter := gin.New()
	applyTransportMiddleware(denyRouter, IPDenylist("10.0.0.0/8"))
	denyRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	denyReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	denyReq.Header.Set("X-Forwarded-For", "10.1.2.3")
	denyRecorder := httptest.NewRecorder()
	denyRouter.ServeHTTP(denyRecorder, denyReq)

	if denyRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected denylist request 403, got %d", denyRecorder.Code)
	}
}

// TestLocaleUsesQueryThenHeaderThenDefault verifies locale negotiation order and response header output.
//
// TestLocaleUsesQueryThenHeaderThenDefault 验证语言协商顺序与响应头输出。
func TestLocaleUsesQueryThenHeaderThenDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queryRouter := gin.New()
	applyTransportMiddleware(queryRouter, Locale(DefaultLocaleOptions()))
	queryRouter.GET("/locale", func(c *gin.Context) {
		c.String(http.StatusOK, GetLocale(c))
	})

	queryReq := httptest.NewRequest(http.MethodGet, "/locale?lang=en-US", nil)
	queryRecorder := httptest.NewRecorder()
	queryRouter.ServeHTTP(queryRecorder, queryReq)

	if queryRecorder.Body.String() != "en" {
		t.Fatalf("expected query locale en, got %q", queryRecorder.Body.String())
	}
	if queryRecorder.Header().Get("Content-Language") != "en" {
		t.Fatalf("expected Content-Language en, got %q", queryRecorder.Header().Get("Content-Language"))
	}

	headerReq := httptest.NewRequest(http.MethodGet, "/locale", nil)
	headerReq.Header.Set("Accept-Language", "en-GB,en;q=0.8,zh;q=0.7")
	headerRecorder := httptest.NewRecorder()
	queryRouter.ServeHTTP(headerRecorder, headerReq)

	if headerRecorder.Body.String() != "en" {
		t.Fatalf("expected header locale en, got %q", headerRecorder.Body.String())
	}

	defaultReq := httptest.NewRequest(http.MethodGet, "/locale", nil)
	defaultRecorder := httptest.NewRecorder()
	queryRouter.ServeHTTP(defaultRecorder, defaultReq)

	if defaultRecorder.Body.String() != "zh" {
		t.Fatalf("expected default locale zh, got %q", defaultRecorder.Body.String())
	}
}

// =============================================================================
// 安全头与 selector 条件中间件
// =============================================================================

// TestSecurityHeadersWritesDefaultsAndCustomValues verifies default security headers and custom overrides.
//
// TestSecurityHeadersWritesDefaultsAndCustomValues 验证默认安全头与自定义覆盖值。
func TestSecurityHeadersWritesDefaultsAndCustomValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	defaultRouter := gin.New()
	applyTransportMiddleware(defaultRouter, SecurityHeaders(SecurityHeadersOptions{}))
	defaultRouter.GET("/headers", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	defaultReq := httptest.NewRequest(http.MethodGet, "/headers", nil)
	defaultRecorder := httptest.NewRecorder()
	defaultRouter.ServeHTTP(defaultRecorder, defaultReq)

	if defaultRecorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected default X-Frame-Options, got %q", defaultRecorder.Header().Get("X-Frame-Options"))
	}
	if defaultRecorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected default X-Content-Type-Options, got %q", defaultRecorder.Header().Get("X-Content-Type-Options"))
	}
	if defaultRecorder.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Fatalf("unexpected Referrer-Policy %q", defaultRecorder.Header().Get("Referrer-Policy"))
	}

	customRouter := gin.New()
	applyTransportMiddleware(customRouter, SecurityHeaders(SecurityHeadersOptions{
		ContentSecurityPolicy: "default-src 'self'",
		PermissionsPolicy:     "geolocation=()",
	}))
	customRouter.GET("/headers", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	customReq := httptest.NewRequest(http.MethodGet, "/headers", nil)
	customRecorder := httptest.NewRecorder()
	customRouter.ServeHTTP(customRecorder, customReq)

	if customRecorder.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Fatalf("unexpected Content-Security-Policy %q", customRecorder.Header().Get("Content-Security-Policy"))
	}
	if customRecorder.Header().Get("Permissions-Policy") != "geolocation=()" {
		t.Fatalf("unexpected Permissions-Policy %q", customRecorder.Header().Get("Permissions-Policy"))
	}
}

// TestCompressionCompressesWhenAccepted verifies gzip compression for clients that advertise support.
//
// TestCompressionCompressesWhenAccepted 验证对声明支持 gzip 的客户端进行压缩。
func TestCompressionCompressesWhenAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, Compression())
	router.GET("/gzip", func(c *gin.Context) {
		c.String(http.StatusOK, "payload")
	})

	req := httptest.NewRequest(http.MethodGet, "/gzip", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("expected Vary Accept-Encoding, got %q", recorder.Header().Get("Vary"))
	}

	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}
	if string(body) != "payload" {
		t.Fatalf("expected decompressed payload, got %q", string(body))
	}
}

// TestCompressionSkipsWhenNotAccepted verifies that compression is bypassed when the client does not accept gzip.
//
// TestCompressionSkipsWhenNotAccepted 验证客户端不接受 gzip 时会跳过压缩。
func TestCompressionSkipsWhenNotAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, Compression())
	router.GET("/plain", func(c *gin.Context) {
		c.String(http.StatusOK, "payload")
	})

	req := httptest.NewRequest(http.MethodGet, "/plain", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no compression, got %q", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Body.String() != "payload" {
		t.Fatalf("expected plain payload, got %q", recorder.Body.String())
	}
}

// TestBodyLimitAllowsSmallBodyAndRejectsLargeBody verifies body-size guardrails for small and oversized payloads.
//
// TestBodyLimitAllowsSmallBodyAndRejectsLargeBody 验证请求体大小护栏对小包与超大包的处理。
func TestBodyLimitAllowsSmallBodyAndRejectsLargeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, BodyLimit(4))
	router.POST("/upload", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, string(body))
	})

	smallReq := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("ok"))
	smallRecorder := httptest.NewRecorder()
	router.ServeHTTP(smallRecorder, smallReq)

	if smallRecorder.Code != http.StatusOK {
		t.Fatalf("expected small body request 200, got %d", smallRecorder.Code)
	}
	if smallRecorder.Body.String() != "ok" {
		t.Fatalf("expected small body echo, got %q", smallRecorder.Body.String())
	}

	largeReq := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("large"))
	largeRecorder := httptest.NewRecorder()
	router.ServeHTTP(largeRecorder, largeReq)

	if largeRecorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected large body request 413, got %d", largeRecorder.Code)
	}
}

// TestSelectorWhenAppliesOnlyOnMatchedRoutes verifies predicate-based middleware selection.
//
// TestSelectorWhenAppliesOnlyOnMatchedRoutes 验证基于谓词的中间件选择行为。
func TestSelectorWhenAppliesOnlyOnMatchedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, When(MatchPrefix("/admin"), SecurityHeaders(SecurityHeadersOptions{
		XFrameOptions: "SAMEORIGIN",
	})))
	router.GET("/admin/panel", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/public/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	adminReq := httptest.NewRequest(http.MethodGet, "/admin/panel", nil)
	adminRecorder := httptest.NewRecorder()
	router.ServeHTTP(adminRecorder, adminReq)

	if adminRecorder.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatalf("expected selector-applied header, got %q", adminRecorder.Header().Get("X-Frame-Options"))
	}

	publicReq := httptest.NewRequest(http.MethodGet, "/public/ping", nil)
	publicRecorder := httptest.NewRecorder()
	router.ServeHTTP(publicRecorder, publicReq)

	if publicRecorder.Header().Get("X-Frame-Options") != "" {
		t.Fatalf("expected selector to skip public route, got %q", publicRecorder.Header().Get("X-Frame-Options"))
	}
}

// =============================================================================
// 鉴权、日志与审计
// =============================================================================

// TestAuthorizationMiddleware verifies unauthorized rejection, successful role checks, and forbidden rejection.
//
// TestAuthorizationMiddleware 验证未认证拒绝、角色校验成功以及无权限拒绝。
func TestAuthorizationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	unauthorizedRouter := gin.New()
	applyTransportMiddleware(unauthorizedRouter, RequireAuthorization())
	unauthorizedRouter.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	unauthorizedRecorder := httptest.NewRecorder()
	unauthorizedRouter.ServeHTTP(unauthorizedRecorder, unauthorizedReq)

	if unauthorizedRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthorizedRecorder.Code)
	}

	roleRouter := gin.New()
	applyTransportMiddleware(roleRouter, func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			ctx := securitycontract.NewJWTClaimsContext(c.Context(), &securitycontract.JWTClaims{
				SubjectID:   1,
				SubjectType: "user",
				Roles:       []string{"admin", "writer"},
			})
			c.SetContext(ctx)
			next(c)
		}
	}, RequireAnyRole("admin"), RequireAllRoles("admin", "writer"), RequireSubjectType("user"))
	roleRouter.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	roleReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	roleRecorder := httptest.NewRecorder()
	roleRouter.ServeHTTP(roleRecorder, roleReq)

	if roleRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected authorized request 204, got %d", roleRecorder.Code)
	}

	forbiddenRouter := gin.New()
	applyTransportMiddleware(forbiddenRouter, func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			ctx := securitycontract.NewJWTClaimsContext(c.Context(), &securitycontract.JWTClaims{
				SubjectID:   1,
				SubjectType: "service",
				Roles:       []string{"reader"},
			})
			c.SetContext(ctx)
			next(c)
		}
	}, RequireAnyRole("admin"))
	forbiddenRouter.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	forbiddenReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	forbiddenRecorder := httptest.NewRecorder()
	forbiddenRouter.ServeHTTP(forbiddenRecorder, forbiddenReq)

	if forbiddenRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", forbiddenRecorder.Code)
	}
}

// TestLoggingMiddlewareWritesAccessLogWithRequestIdentity verifies access log fields and propagated request identity.
//
// TestLoggingMiddlewareWritesAccessLogWithRequestIdentity 验证访问日志字段与传播的请求标识。
func TestLoggingMiddlewareWritesAccessLogWithRequestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := newStubLogger()
	router := gin.New()
	applyTransportMiddleware(router, RequestIdentity(), LoggingMiddleware(logger))
	router.GET("/logs/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/logs/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected log entries")
	}
	entry := entries[len(entries)-1]
	if entry.msg != "http request" {
		t.Fatalf("expected http request log, got %q", entry.msg)
	}
	if fieldValue(entry.fields, "method") != http.MethodGet {
		t.Fatalf("expected method GET, got %v", fieldValue(entry.fields, "method"))
	}
	if fieldValue(entry.fields, "path") != "/logs/42" {
		t.Fatalf("expected path /logs/42, got %v", fieldValue(entry.fields, "path"))
	}
	if fieldValue(entry.fields, "route") != "/logs/:id" {
		t.Fatalf("expected route /logs/:id, got %v", fieldValue(entry.fields, "route"))
	}
	if fieldValue(entry.fields, "status") != http.StatusNoContent {
		t.Fatalf("expected status 204, got %v", fieldValue(entry.fields, "status"))
	}
	if fieldValue(entry.fields, "request_id") == nil {
		t.Fatal("expected request_id field")
	}
	if fieldValue(entry.fields, "trace_id") == nil {
		t.Fatal("expected trace_id field")
	}
}

// =============================================================================
// Recovery panic 恢复
// =============================================================================

// TestRecoveryMiddlewareRecoversAndLogsPanic verifies panic recovery, unified 500 output, and panic logging.
//
// TestRecoveryMiddlewareRecoversAndLogsPanic 验证 panic 恢复、统一 500 输出和 panic 日志记录。
func TestRecoveryMiddlewareRecoversAndLogsPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := newStubLogger()
	router := gin.New()
	applyTransportMiddleware(router,
		func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
			return func(c transportcontract.HTTPContext) {
				c.SetContext(frameworkbizlog.WithContext(c.Context(), logger))
				next(c)
			}
		},
		RecoveryMiddleware(),
	)
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	var resp Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "internal server error" {
		t.Fatalf("expected internal server error message, got %q", resp.Message)
	}

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected panic log entry")
	}
	last := entries[len(entries)-1]
	if last.level != "error" || last.msg != "http panic recovered" {
		t.Fatalf("expected recovery error log, got level=%q msg=%q", last.level, last.msg)
	}
	if fieldValue(last.fields, "panic") != "boom" {
		t.Fatalf("expected panic field boom, got %v", fieldValue(last.fields, "panic"))
	}
}

// =============================================================================
// Audit 审计
// =============================================================================

// TestAuditMiddlewareWritesActorAndOutcomeFields verifies audit fields such as actor, locale, resource, and outcome.
//
// TestAuditMiddlewareWritesActorAndOutcomeFields 验证 actor、locale、resource 与 outcome 等审计字段。
func TestAuditMiddlewareWritesActorAndOutcomeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := newStubLogger()
	router := gin.New()
	applyTransportMiddleware(router,
		RequestIdentity(),
		Locale(DefaultLocaleOptions()),
		func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
			return func(c transportcontract.HTTPContext) {
				ctx := securitycontract.NewJWTClaimsContext(c.Context(), &securitycontract.JWTClaims{
					SubjectID:   7,
					SubjectType: "user",
					SubjectName: "alice",
					Roles:       []string{"admin", "writer"},
				})
				c.SetContext(ctx)
				next(c)
			}
		},
		AuditMiddleware(logger, AuditOptions{Event: "admin audit"}),
	)
	router.POST("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/orders/42?lang=en", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected audit log entry")
	}
	entry := entries[len(entries)-1]
	if entry.msg != "admin audit" {
		t.Fatalf("expected admin audit event, got %q", entry.msg)
	}
	if fieldValue(entry.fields, "kind") != "audit" {
		t.Fatalf("expected audit kind, got %v", fieldValue(entry.fields, "kind"))
	}
	if fieldValue(entry.fields, "action") != "POST" {
		t.Fatalf("expected action POST, got %v", fieldValue(entry.fields, "action"))
	}
	if fieldValue(entry.fields, "resource") != "/orders/:id" {
		t.Fatalf("expected resource /orders/:id, got %v", fieldValue(entry.fields, "resource"))
	}
	if fieldValue(entry.fields, "outcome") != "success" {
		t.Fatalf("expected outcome success, got %v", fieldValue(entry.fields, "outcome"))
	}
	if fieldValue(entry.fields, "actor_kind") != "user" {
		t.Fatalf("expected actor_kind user, got %v", fieldValue(entry.fields, "actor_kind"))
	}
	if fieldValue(entry.fields, "actor_id") != int64(7) {
		t.Fatalf("expected actor_id 7, got %v", fieldValue(entry.fields, "actor_id"))
	}
	if fieldValue(entry.fields, "actor_roles") != "admin,writer" {
		t.Fatalf("expected actor_roles admin,writer, got %v", fieldValue(entry.fields, "actor_roles"))
	}
	if fieldValue(entry.fields, "locale") != "en" {
		t.Fatalf("expected locale en, got %v", fieldValue(entry.fields, "locale"))
	}
}

// =============================================================================
// Bind 绑定与校验
// =============================================================================

// TestBindAndValidateJSONStoresValidatedBody verifies that successful validation stores the validated body in request context.
//
// TestBindAndValidateJSONStoresValidatedBody 验证成功校验后会把已校验对象写入请求上下文。
func TestBindAndValidateJSONStoresValidatedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	validator := &stubValidator{
		validateFn: func(ctx context.Context, obj any) error {
			input, ok := obj.(*struct {
				Name string `json:"name"`
			})
			if !ok || strings.TrimSpace(input.Name) == "" {
				return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "name is required")
			}
			return nil
		},
	}

	router.POST("/validate", func(c *gin.Context) {
		httpCtx := newHTTPContext(c)
		input := &struct {
			Name string `json:"name"`
		}{}
		if err := BindAndValidateJSON(httpCtx, validator, input); err != nil {
			return
		}
		validatedBody, ok := supportcontract.FromValidatedBodyContext(c.Request.Context())
		if !ok {
			c.String(http.StatusInternalServerError, "missing validated body")
			return
		}
		body := validatedBody.(*struct {
			Name string `json:"name"`
		})
		c.String(http.StatusOK, body.Name)
	})

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"name":"alice"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "alice" {
		t.Fatalf("expected validated body name alice, got %q", recorder.Body.String())
	}
}

// TestBindAndValidateJSONReturnsUnifiedError verifies unified validation error output and detail propagation.
//
// TestBindAndValidateJSONReturnsUnifiedError 验证统一校验错误输出与详情透传。
func TestBindAndValidateJSONReturnsUnifiedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	validator := &stubValidator{
		validateFn: func(context.Context, any) error {
			return resiliencecontract.BadRequest(
				resiliencecontract.ErrorReasonBadRequest,
				"validation failed",
			).WithMetadata(map[string]string{"validation_errors": `["name is required"]`})
		},
	}

	router.POST("/validate", func(c *gin.Context) {
		httpCtx := newHTTPContext(c)
		input := &struct {
			Name string `json:"name"`
		}{}
		_ = BindAndValidateJSON(httpCtx, validator, input)
	})

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var resp ValidateErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode validation response: %v", err)
	}
	if resp.Message != "validation failed" {
		t.Fatalf("expected validation failed message, got %q", resp.Message)
	}
	if resp.Details != `["name is required"]` {
		t.Fatalf("expected validation details, got %q", resp.Details)
	}
}

// =============================================================================
// Metrics 指标
// =============================================================================

// TestMetricsMiddlewareRecordsRequestCount verifies request counter increments with method, route, and status labels.
//
// TestMetricsMiddlewareRecordsRequestCount 验证请求计数器会按 method、route 和 status 标签递增。
func TestMetricsMiddlewareRecordsRequestCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	labels := map[string]string{
		"method": http.MethodGet,
		"path":   "/metrics/:id",
		"status": "204",
	}
	beforeCount := counterValue("gorp_http_requests_total", labels)

	router := gin.New()
	applyTransportMiddleware(router, MetricsMiddleware())
	router.GET("/metrics/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	afterCount := counterValue("gorp_http_requests_total", labels)
	if afterCount != beforeCount+1 {
		t.Fatalf("expected request counter to increase by 1, got before=%v after=%v", beforeCount, afterCount)
	}
}
