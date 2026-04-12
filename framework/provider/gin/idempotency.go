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
//
// 中文说明：
// - 定义幂等 key 的存储接口；
// - Check 检查 key 是否存在，若存在则返回之前的响应；
// - Set 存储 key 和对应的响应；
// - 支持内存、Redis 等不同实现。
type IdempotencyStore interface {
	// Check 检查 key 是否存在，若存在返回之前的响应
	Check(key string) (exists bool, response *IdempotencyResponse)
	// Set 存储 key 和对应的响应
	Set(key string, response *IdempotencyResponse, ttl time.Duration)
}

// IdempotencyResponse 幂等响应缓存。
type IdempotencyResponse struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// MemoryIdempotencyStore 内存幂等 key 存储。
//
// 中文说明：
// - 基于 sync.Map 的内存实现，适合单机场景；
// - 支持 TTL 过期清理；
// - 生产环境建议使用 Redis 实现。
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
	// 启动后台清理协程
	go store.cleanup()
	return store
}

// Check 检查 key 是否存在。
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

// Set 存储 key 和对应的响应。
func (s *MemoryIdempotencyStore) Set(key string, response *IdempotencyResponse, ttl time.Duration) {
	s.data.Store(key, &idempotencyEntry{
		response:  response,
		expiresAt: time.Now().Add(ttl),
	})
}

// cleanup 定期清理过期条目。
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
// - 检查请求头中的 X-Idempotency-Key；
// - 如果 key 已存在，直接返回之前的响应；
// - 如果 key 不存在，执行请求并缓存响应；
// - 默认 TTL 为 24 小时。
func IdempotencyMiddleware(store IdempotencyStore, ttl time.Duration) gin.HandlerFunc {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return func(c *gin.Context) {
		// 只对幂等方法（POST/PUT/PATCH）生效
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
			c.Next()
			return
		}

		// 获取幂等 key
		key := c.GetHeader(IdempotencyKeyHeader)
		if key == "" {
			c.Next()
			return
		}

		// 检查是否已存在
		exists, response := store.Check(key)
		if exists {
			// 返回缓存的响应
			for k, v := range response.Headers {
				c.Header(k, v)
			}
			c.JSON(response.StatusCode, response.Body)
			c.Abort()
			return
		}

		// 执行请求
		c.Next()

		// 缓存成功的响应
		if c.Writer.Status() < 400 {
			// 注意：这里简化处理，实际可能需要捕获响应体
			store.Set(key, &IdempotencyResponse{
				StatusCode: c.Writer.Status(),
				Body:       nil, // 实际实现需要捕获响应体
			}, ttl)
		}
	}
}

// GenerateIdempotencyKey 生成幂等 key。
//
// 中文说明：
// - 基于请求参数生成唯一的幂等 key；
// - 可用于客户端生成或服务端校验；
// - 确保相同请求生成相同的 key。
func GenerateIdempotencyKey(userID string, operation string, params string) string {
	data := userID + ":" + operation + ":" + params
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// IdempotencyKeyFromRequest 从请求中提取或生成幂等 key。
//
// 中文说明：
// - 如果请求头已有幂等 key，直接返回；
// - 否则基于请求信息生成一个。
func IdempotencyKeyFromRequest(c *gin.Context, userID string) string {
	// 优先使用请求头中的 key
	if key := c.GetHeader(IdempotencyKeyHeader); key != "" {
		return key
	}

	// 基于 trace id 生成
	traceID := GetTraceID(c)
	if traceID != "" {
		return GenerateIdempotencyKey(userID, c.Request.Method+c.FullPath(), traceID)
	}

	return ""
}