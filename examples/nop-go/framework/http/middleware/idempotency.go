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
// The Reserve/Commit pattern eliminates the TOCTOU race between Check and Set:
// Reserve atomically reserves a key (preventing duplicate processing) or returns
// the previously stored response; Commit finalizes the stored response.
//
// IdempotencyStore 用于存储幂等响应，以便后续回放。
// Reserve/Commit 模式消除了 Check 和 Set 之间的 TOCTOU 竞态：
// Reserve 原子地预留一个 key（防止重复处理）或返回已存储的响应；
// Commit 最终确认存储的响应。
type IdempotencyStore interface {
	// Reserve atomically attempts to reserve the key for processing.
	// Returns (false, cachedResponse) if the key already exists — the caller
	// should replay cachedResponse to the client.
	// Returns (true, nil) if the key was successfully reserved — the caller
	// should proceed with processing and then call Commit.
	//
	// Reserve 原子地尝试预留 key 进行处理。
	// 如果 key 已存在则返回 (false, cachedResponse) —— 调用方应回放 cachedResponse。
	// 如果 key 成功预留则返回 (true, nil) —— 调用方应继续处理并调用 Commit。
	Reserve(key string, ttl time.Duration) (reserved bool, response *IdempotencyResponse)

	// Commit stores the final response for a previously reserved key.
	//
	// Commit 为之前预留的 key 存储最终响应。
	Commit(key string, response *IdempotencyResponse, ttl time.Duration)
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
// Uses sync.Map.LoadOrStore for atomic Reserve, eliminating TOCTOU races.
//
// MemoryIdempotencyStore 是一个内存型幂等响应存储。
// 使用 sync.Map.LoadOrStore 实现原子 Reserve，消除 TOCTOU 竞态。
type MemoryIdempotencyStore struct {
	data   sync.Map
	stopCh chan struct{} // stopCh 用于通知 cleanup goroutine 退出
	//
	// stopCh signals the cleanup goroutine to exit.
}

type idempotencyEntry struct {
	response  *IdempotencyResponse
	expiresAt time.Time
	committed bool // false = reserved (placeholder), true = final response stored
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
	store := &MemoryIdempotencyStore{
		stopCh: make(chan struct{}),
	}
	go store.cleanup()
	return store
}

// Close 停止 cleanup goroutine，释放资源。
// 调用后 store 不应再被使用。
func (s *MemoryIdempotencyStore) Close() {
	select {
	case <-s.stopCh:
		// 已关闭
	default:
		close(s.stopCh)
	}
}

// Reserve atomically reserves a key or returns the cached response.
// Uses sync.Map.LoadOrStore to guarantee that only one goroutine wins
// the reservation for a given key, eliminating TOCTOU races.
//
// Reserve 原子地预留 key 或返回已缓存的响应。
// 使用 sync.Map.LoadOrStore 保证只有一个 goroutine 赢得给定 key 的预留，
// 消除 TOCTOU 竞态。
func (s *MemoryIdempotencyStore) Reserve(key string, ttl time.Duration) (bool, *IdempotencyResponse) {
	placeholder := &idempotencyEntry{
		response:  nil,
		expiresAt: time.Now().Add(ttl),
		committed: false,
	}
	actual, loaded := s.data.LoadOrStore(key, placeholder)
	entry := actual.(*idempotencyEntry)

	if !loaded {
		// We successfully reserved this key.
		// 我们成功预留了此 key。
		return true, nil
	}

	// Key already exists. Check if expired.
	// Key 已存在。检查是否过期。
	if time.Now().After(entry.expiresAt) {
		// Entry is expired. Try to replace it with our placeholder.
		// 条目已过期。尝试用我们的占位符替换。
		s.data.Delete(key)
		// Retry reservation (could lose to another goroutine, so use LoadOrStore again).
		actual2, loaded2 := s.data.LoadOrStore(key, placeholder)
		if !loaded2 {
			return true, nil
		}
		entry = actual2.(*idempotencyEntry)
	}

	if !entry.committed {
		// Key is reserved but not yet committed. Another request is processing.
		// Return a pending indicator so the caller can wait or reject.
		// Key 已预留但尚未提交。另一个请求正在处理中。
		// 返回 pending 指示，调用方可等待或拒绝。
		return false, nil
	}

	return false, entry.response
}

// Commit stores the final response for a previously reserved key.
//
// Commit 为之前预留的 key 存储最终响应。
func (s *MemoryIdempotencyStore) Commit(key string, response *IdempotencyResponse, ttl time.Duration) {
	s.data.Store(key, &idempotencyEntry{
		response:  response,
		expiresAt: time.Now().Add(ttl),
		committed: true,
	})
}

// cleanup removes expired entries from the in-memory store.
//
// cleanup 清理内存存储中已过期的记录。
func (s *MemoryIdempotencyStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			// 收到退出信号，停止 cleanup goroutine
			return
		case <-ticker.C:
			now := time.Now()
			s.data.Range(func(key, value any) bool {
				entry := value.(*idempotencyEntry)
				if now.After(entry.expiresAt) {
					s.data.Delete(key)
				}
				return true
			})
		}
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

		// Atomic reserve: only one goroutine wins for each key.
		// 原子预留：每个 key 只有一个 goroutine 胜出。
		reserved, response := store.Reserve(key, ttl)
		if !reserved {
			if response != nil {
				writeCachedIdempotencyResponse(newContext(c), response)
			} else {
				// Another request is processing this key (reserved but not committed).
				// Return 409 Conflict to tell the client to retry later.
				// 另一个请求正在处理此 key（已预留但未提交）。
				// 返回 409 Conflict 告知客户端稍后重试。
				c.JSON(http.StatusConflict, map[string]string{
					"error": "idempotency key is being processed",
				})
			}
			c.Abort()
			return
		}

		recorder := &idempotencyResponseWriter{ResponseWriter: c.Writer}
		c.Writer = recorder
		c.Next()
		c.Writer = recorder.ResponseWriter

		// Only cache successful responses (status < 400).
		// Failed responses are NOT cached so that the client can retry
		// and potentially succeed on the next attempt.
		// If you need to cache all responses regardless of status,
		// implement a custom IdempotencyStore with different Commit logic.
		// 仅缓存成功响应（状态码 < 400）。
		// 失败响应不缓存，以便客户端重试后可能成功。
		// 如果需要缓存所有响应（无论状态码），请实现自定义 IdempotencyStore。
		if recorder.Status() < 400 {
			store.Commit(key, captureIdempotencyResponse(c, recorder), ttl)
		}
	}
}

// Idempotency applies request de-duplication to transport-level HTTP middleware.
//
// Idempotency 将请求去重能力应用到 transport 层 HTTP 中间件链。
func Idempotency(store IdempotencyStore, ttl time.Duration) transportcontract.Middleware {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
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

			// Atomic reserve: only one goroutine wins for each key.
			// 原子预留：每个 key 只有一个 goroutine 胜出。
			reserved, response := store.Reserve(key, ttl)
			if !reserved {
				if response != nil {
					writeCachedIdempotencyResponse(c, response)
				} else {
					// Another request is processing this key.
					// 另一个请求正在处理此 key。
					c.JSON(http.StatusConflict, map[string]string{
						"error": "idempotency key is being processed",
					})
				}
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
					store.Commit(key, captureIdempotencyResponse(gc, recorder), ttl)
				}
				return
			}

			if next != nil {
				next(c)
			}
			if status := c.ResponseStatus(); status > 0 && status < 400 {
				store.Commit(key, &IdempotencyResponse{StatusCode: status}, ttl)
			}
		}
	}
}

// writeCachedIdempotencyResponse replays a previously cached idempotent response.
//
// writeCachedIdempotencyResponse 回放之前缓存的幂等响应。
func writeCachedIdempotencyResponse(c transportcontract.Context, response *IdempotencyResponse) {
	if response == nil {
		return
	}
	for k, v := range response.Headers {
		c.SetHeader(k, v)
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
