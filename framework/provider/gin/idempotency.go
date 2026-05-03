package gin

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// IdempotencyKeyHeader 是幂等 key 的请求头名称
	IdempotencyKeyHeader = "X-Idempotency-Key"
)

// IdempotencyStore 幂等 key 存储接口。
type IdempotencyStore interface {
	Check(key string) (exists bool, response *IdempotencyResponse)
	Set(key string, response *IdempotencyResponse, ttl time.Duration)
}

// IdempotencyResponse 幂等响应缓存。
type IdempotencyResponse struct {
	StatusCode int               `json:"status_code"`
	Body       interface{}       `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// MemoryIdempotencyStore 内存幂等 key 存储。
type MemoryIdempotencyStore struct {
	data sync.Map
}

type idempotencyEntry struct {
	response  *IdempotencyResponse
	expiresAt time.Time
}

// NewMemoryIdempotencyStore 创建内存幂等 key 存储。
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	store := &MemoryIdempotencyStore{}
	go store.cleanup()
	return store
}

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

func (s *MemoryIdempotencyStore) Set(key string, response *IdempotencyResponse, ttl time.Duration) {
	s.data.Store(key, &idempotencyEntry{response: response, expiresAt: time.Now().Add(ttl)})
}

func (s *MemoryIdempotencyStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		s.data.Range(func(key, value interface{}) bool {
			entry := value.(*idempotencyEntry)
			if time.Now().After(entry.expiresAt) {
				s.data.Delete(key)
			}
			return true
		})
	}
}

// IdempotencyMiddleware 创建幂等性中间件。
//
// 中文说明：
// - 这是 Gin provider 扩展层中间件，不属于默认 framework 主线契约；
// - 检查请求头中的 X-Idempotency-Key；
// - 如果 key 已存在，直接返回之前的响应；
// - 如果 key 不存在，执行请求并缓存响应；
// - 默认 TTL 为 24 小时。
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
			for k, v := range response.Headers {
				c.Header(k, v)
			}
			c.JSON(response.StatusCode, response.Body)
			c.Abort()
			return
		}

		c.Next()

		if c.Writer.Status() < 400 {
			store.Set(key, &IdempotencyResponse{StatusCode: c.Writer.Status(), Body: nil}, ttl)
		}
	}
}

// GenerateIdempotencyKey 生成幂等 key。
func GenerateIdempotencyKey(userID string, operation string, params string) string {
	data := userID + ":" + operation + ":" + params
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// IdempotencyKeyFromRequest 从请求中提取或生成幂等 key。
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
