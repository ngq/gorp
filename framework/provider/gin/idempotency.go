package gin

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	IdempotencyKeyHeader = "X-Idempotency-Key"
)

type IdempotencyStore interface {
	Check(key string) (exists bool, response *IdempotencyResponse)
	Set(key string, response *IdempotencyResponse, ttl time.Duration)
}

type IdempotencyResponse struct {
	StatusCode int               `json:"status_code"`
	Body       interface{}       `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type responseFuncConfigurer interface {
	SetResponseFuncs(
		json func(int, any),
		str func(int, string),
		xml func(int, any),
		data func(int, string, []byte),
		redirect func(int, string),
		status func(int),
		statusRead func() int,
	)
}

type MemoryIdempotencyStore struct {
	data sync.Map
}

type idempotencyEntry struct {
	response  *IdempotencyResponse
	expiresAt time.Time
}

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
				for k, v := range response.Headers {
					c.Header(k, v)
				}
				if response.Body != nil {
					c.JSON(response.StatusCode, response.Body)
				} else {
					c.Status(response.StatusCode)
				}
				if gc, ok := unwrapGinContext(c); ok {
					gc.Abort()
				}
				return
			}

			var (
				capturedStatus  int
				capturedBody    any
				capturedHeaders map[string]string
			)
			origJSON := c.JSON
			origStatus := c.Status
			if cfg, ok := c.(responseFuncConfigurer); ok {
				statusReader := func() int {
					if gc, ok := unwrapGinContext(c); ok {
						return gc.Writer.Status()
					}
					if capturedStatus != 0 {
						return capturedStatus
					}
					return 0
				}
				cfg.SetResponseFuncs(func(status int, body any) {
					capturedStatus = status
					capturedBody = body
					origJSON(status, body)
				}, c.String, c.XML, c.Data, c.Redirect, func(code int) {
					capturedStatus = code
					origStatus(code)
				}, statusReader)
			}

			if next != nil {
				next(c)
			}

			if capturedStatus == 0 {
				capturedStatus = c.ResponseStatus()
			}
			if capturedStatus > 0 && capturedStatus < 400 {
				if gc, ok := unwrapGinContext(c); ok {
					capturedHeaders = make(map[string]string, len(gc.Writer.Header()))
					for k, values := range gc.Writer.Header() {
						if len(values) > 0 {
							capturedHeaders[k] = values[0]
						}
					}
				}
				store.Set(key, &IdempotencyResponse{
					StatusCode: capturedStatus,
					Body:       capturedBody,
					Headers:    capturedHeaders,
				}, ttl)
			}
		}
	}
}

func GenerateIdempotencyKey(userID string, operation string, params string) string {
	data := userID + ":" + operation + ":" + params
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

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
