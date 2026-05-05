package gin

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
	IdempotencyKeyHeader = "X-Idempotency-Key"
)

type IdempotencyStore interface {
	Check(key string) (exists bool, response *IdempotencyResponse)
	Set(key string, response *IdempotencyResponse, ttl time.Duration)
}

type IdempotencyResponse struct {
	StatusCode  int               `json:"status_code"`
	Body        []byte            `json:"body,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

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

func (w *idempotencyResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseWriter) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *idempotencyResponseWriter) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *idempotencyResponseWriter) Status() int {
	if w.statusCode != 0 {
		return w.statusCode
	}
	return w.ResponseWriter.Status()
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
