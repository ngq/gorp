// Application scenarios:
// - Prevent duplicate writes caused by retries, refreshes, or client-side resubmission.
// - Replay successful write responses for the same idempotency key within a TTL window.
// - Provide a stable request de-duplication baseline for create, update, and callback endpoints.
//
// 适用场景：
// - 防止重试、刷新或客户端重复提交造成重复写入。
// - 在 TTL 窗口内对相同幂等键回放成功写请求的响应。
// - 为创建、更新和回调类接口提供稳定的请求去重基线。
package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	// IdempotencyKeyHeader is the standard request header used to carry the idempotency key.
	//
	// IdempotencyKeyHeader 是用于传递幂等键的标准请求头。
	IdempotencyKeyHeader = "X-Idempotency-Key"
)

// IdempotencyStore stores idempotent responses for later replay.
//
// IdempotencyStore 用于存储幂等响应，以便后续回放。
type IdempotencyStore interface {
	Check(key string) (exists bool, response *IdempotencyResponse)
	Set(key string, response *IdempotencyResponse, ttl time.Duration)
}

// IdempotencyResponse is the cached response payload used for replay.
//
// IdempotencyResponse 表示用于回放的缓存响应内容。
type IdempotencyResponse struct {
	StatusCode  int               `json:"status_code"`
	Body        []byte            `json:"body,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// MemoryIdempotencyStore is an in-memory idempotency response store.
//
// MemoryIdempotencyStore 是一个内存型幂等响应存储。
type MemoryIdempotencyStore struct {
	data sync.Map
}

type idempotencyEntry struct {
	response  *IdempotencyResponse
	expiresAt time.Time
}

type idempotencyResponseWriter struct {
	gin.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

// WriteHeader captures the final response status code.
//
// WriteHeader 记录最终响应状态码。
func (w *idempotencyResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write captures response bytes while forwarding them to the client.
//
// Write 在向客户端转发响应的同时捕获响应字节。
func (w *idempotencyResponseWriter) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// WriteString captures response text while forwarding it to the client.
//
// WriteString 在向客户端转发响应文本的同时完成捕获。
func (w *idempotencyResponseWriter) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Status returns the recorded response status code.
//
// Status 返回记录到的响应状态码。
func (w *idempotencyResponseWriter) Status() int {
	if w.statusCode != 0 {
		return w.statusCode
	}
	return w.ResponseWriter.Status()
}

// NewMemoryIdempotencyStore creates a memory-backed idempotency store.
//
// NewMemoryIdempotencyStore 创建一个基于内存的幂等存储。
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	store := &MemoryIdempotencyStore{}
	go store.cleanup()
	return store
}

// Check looks up a cached idempotent response by key.
//
// Check 按幂等键查询已缓存的响应。
func (s *MemoryIdempotencyStore) Check(key string) (bool, *IdempotencyResponse) {
	value, ok := s.data.Load(key)
	if !ok {
		return false, nil
	}
	entry := value.(*idempotencyEntry)
	if time.Now().After(entry.expiresAt) {
		s.data.Delete(key)
		return false, nil
	}
	return true, entry.response
}

// Set stores a cached idempotent response with TTL.
//
// Set 按 TTL 存储一条幂等响应记录。
func (s *MemoryIdempotencyStore) Set(key string, response *IdempotencyResponse, ttl time.Duration) {
	s.data.Store(key, &idempotencyEntry{response: response, expiresAt: time.Now().Add(ttl)})
}

// cleanup removes expired entries from the in-memory store.
//
// cleanup 清理内存存储中已过期的记录。
func (s *MemoryIdempotencyStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.data.Range(func(key, value any) bool {
			entry := value.(*idempotencyEntry)
			if time.Now().After(entry.expiresAt) {
				s.data.Delete(key)
			}
			return true
		})
	}
}

// IdempotencyMiddleware is the native Gin form of the idempotency middleware.
//
// IdempotencyMiddleware 是幂等中间件的原生 Gin 形态。
//
// Example:
//
//	router.Use(httpmiddleware.IdempotencyMiddleware(store, time.Hour))
func IdempotencyMiddleware(store IdempotencyStore, ttl time.Duration) gin.HandlerFunc {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
			c.Next()
			return
		}

		key := c.GetHeader(IdempotencyKeyHeader)
		if key == "" {
			c.Next()
			return
		}

		exists, response := store.Check(key)
		if exists {
			writeCachedIdempotencyResponse(newHTTPContext(c), response)
			c.Abort()
			return
		}

		recorder := &idempotencyResponseWriter{ResponseWriter: c.Writer}
		c.Writer = recorder
		c.Next()
		c.Writer = recorder.ResponseWriter

		if recorder.Status() < 400 {
			store.Set(key, captureIdempotencyResponse(c, recorder), ttl)
		}
	}
}

// Idempotency applies request de-duplication to transport-level HTTP middleware.
//
// Idempotency 将请求去重能力应用到 transport 层 HTTP 中间件链。
func Idempotency(store IdempotencyStore, ttl time.Duration) transportcontract.HTTPMiddleware {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if store == nil {
				if next != nil {
					next(c)
				}
				return
			}

			method := ""
			if req := c.Request(); req != nil {
				method = req.Method
			}
			if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
				if next != nil {
					next(c)
				}
				return
			}

			key := c.GetHeader(IdempotencyKeyHeader)
			if key == "" {
				if next != nil {
					next(c)
				}
				return
			}

			exists, response := store.Check(key)
			if exists && response != nil {
				writeCachedIdempotencyResponse(c, response)
				if gc, ok := unwrapGinContext(c); ok {
					gc.Abort()
				}
				return
			}

			if gc, ok := unwrapGinContext(c); ok {
				recorder := &idempotencyResponseWriter{ResponseWriter: gc.Writer}
				gc.Writer = recorder
				if next != nil {
					next(c)
				}
				gc.Writer = recorder.ResponseWriter
				if recorder.Status() > 0 && recorder.Status() < 400 {
					store.Set(key, captureIdempotencyResponse(gc, recorder), ttl)
				}
				return
			}

			if next != nil {
				next(c)
			}
			if status := c.ResponseStatus(); status > 0 && status < 400 {
				store.Set(key, &IdempotencyResponse{StatusCode: status}, ttl)
			}
		}
	}
}

// writeCachedIdempotencyResponse replays a previously cached idempotent response.
//
// writeCachedIdempotencyResponse 回放之前缓存的幂等响应。
func writeCachedIdempotencyResponse(c transportcontract.HTTPContext, response *IdempotencyResponse) {
	if response == nil {
		return
	}
	for k, v := range response.Headers {
		c.Header(k, v)
	}
	if len(response.Body) > 0 {
		c.Data(response.StatusCode, response.ContentType, response.Body)
		return
	}
	c.Status(response.StatusCode)
}

// captureIdempotencyResponse captures the current Gin response for later replay.
//
// captureIdempotencyResponse 捕获当前 Gin 响应，以便后续回放。
func captureIdempotencyResponse(c *gin.Context, recorder *idempotencyResponseWriter) *IdempotencyResponse {
	headers := make(map[string]string, len(c.Writer.Header()))
	for k, values := range c.Writer.Header() {
		if len(values) > 0 {
			headers[k] = values[0]
		}
	}
	return &IdempotencyResponse{
		StatusCode:  recorder.Status(),
		Body:        append([]byte(nil), recorder.body.Bytes()...),
		ContentType: c.Writer.Header().Get("Content-Type"),
		Headers:     headers,
	}
}

// GenerateIdempotencyKey derives a stable idempotency key from business dimensions.
//
// GenerateIdempotencyKey 根据业务维度生成稳定的幂等键。
func GenerateIdempotencyKey(userID string, operation string, params string) string {
	data := userID + ":" + operation + ":" + params
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// IdempotencyKeyFromRequest reads the request idempotency key or derives one from trace context.
//
// IdempotencyKeyFromRequest 读取请求幂等键，或根据链路上下文派生一个幂等键。
func IdempotencyKeyFromRequest(c *gin.Context, userID string) string {
	if key := c.GetHeader(IdempotencyKeyHeader); key != "" {
		return key
	}
	traceID := GetTraceID(c)
	if traceID != "" {
		return GenerateIdempotencyKey(userID, c.Request.Method+c.FullPath(), traceID)
	}
	return ""
}
