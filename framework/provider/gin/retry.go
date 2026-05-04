package gin

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

func RetryMiddleware(retry resiliencecontract.Retry, policy resiliencecontract.RetryPolicy) gin.HandlerFunc {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(c *gin.Context) {
		method := c.Request.Method
		if !isIdempotentMethod(method) {
			c.Next()
			return
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		writer := &retryResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		baseWriter := c.Writer
		c.Writer = writer

		var lastErr error
		var lastStatusCode int

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			writer.body.Reset()
			writer.ResponseWriter = baseWriter
			c.Next()

			statusCode := writer.Status()
			lastStatusCode = statusCode
			if statusCode < 500 {
				return
			}
			if !isRetryableHTTPStatus(statusCode, policy.RetryableCodes) {
				return
			}

			lastErr = resiliencecontract.NewError(statusCode, resiliencecontract.ErrorReasonInternal, "request failed")
			if attempt == policy.MaxAttempts-1 {
				break
			}

			select {
			case <-c.Request.Context().Done():
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			default:
			}

			jitter := rng.Float64()
			delay := policy.CalculateDelay(attempt, jitter)
			select {
			case <-c.Request.Context().Done():
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			case <-time.After(delay):
			}
		}

		if lastErr != nil {
			writeGinResponseHeaders(c)
			resp := Response{
				Code:    CodeServiceUnavailable,
				Message: lastErr.Error(),
				Data: map[string]any{
					"attempts": policy.MaxAttempts,
					"reason":   "SERVICE_UNAVAILABLE",
				},
			}
			c.JSON(lastStatusCode, resp)
		}
	}
}

func RetryAllMethodsMiddleware(retry resiliencecontract.Retry, policy resiliencecontract.RetryPolicy) gin.HandlerFunc {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		writer := &retryResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		baseWriter := c.Writer
		c.Writer = writer

		var lastErr error
		var lastStatusCode int

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			writer.body.Reset()
			writer.ResponseWriter = baseWriter
			c.Next()

			statusCode := writer.Status()
			lastStatusCode = statusCode
			if statusCode < 500 {
				return
			}
			if !isRetryableHTTPStatus(statusCode, policy.RetryableCodes) {
				return
			}

			lastErr = resiliencecontract.NewError(statusCode, resiliencecontract.ErrorReasonInternal, "request failed")
			if attempt == policy.MaxAttempts-1 {
				break
			}

			select {
			case <-c.Request.Context().Done():
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			default:
			}

			jitter := rng.Float64()
			delay := policy.CalculateDelay(attempt, jitter)
			select {
			case <-c.Request.Context().Done():
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			case <-time.After(delay):
			}
		}

		if lastErr != nil {
			writeGinResponseHeaders(c)
			resp := Response{
				Code:    CodeServiceUnavailable,
				Message: lastErr.Error(),
				Data: map[string]any{
					"attempts": policy.MaxAttempts,
					"reason":   "SERVICE_UNAVAILABLE",
				},
			}
			c.JSON(lastStatusCode, resp)
		}
	}
}

func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func isRetryableHTTPStatus(status int, retryableCodes []int) bool {
	for _, code := range retryableCodes {
		if status == code {
			return true
		}
	}
	return status == 502 || status == 503 || status == 504
}

type retryResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *retryResponseWriter) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *retryResponseWriter) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *retryResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *retryResponseWriter) Status() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func DoWithRetry(c *gin.Context, retry resiliencecontract.Retry, fn func(ctx context.Context) (any, error)) (any, error) {
	return retry.DoWithResult(c.Request.Context(), func() (any, error) {
		return fn(c.Request.Context())
	})
}
