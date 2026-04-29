package gin

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ngq/gorp/framework/contract"
)

// RetryMiddleware 创建 HTTP 重试中间件。
//
// 中文说明：
// - 对可重试的 HTTP 状态码进行重试；
// - 使用指数退避延迟；
// - 仅对幂等方法（GET/HEAD/OPTIONS）重试；
// - 支持请求体重放。
//
// 注意：
// - POST/PUT/DELETE 等非幂等方法默认不重试；
// - 如需对非幂等方法重试，使用 RetryAllMethods 选项。
//
// 使用示例：
//
//	router.Use(gin.RetryMiddleware(retry, contract.DefaultRetryPolicy()))
func RetryMiddleware(retry contract.Retry, policy contract.RetryPolicy) gin.HandlerFunc {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(c *gin.Context) {
		// 仅对幂等方法重试
		method := c.Request.Method
		if !isIdempotentMethod(method) {
			c.Next()
			return
		}

		// 读取请求体（用于重试时恢复）
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 使用响应写入器捕获响应
		writer := &retryResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		baseWriter := c.Writer
		c.Writer = writer

		var lastErr error
		var lastStatusCode int

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			// 重置请求体
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// 重置响应写入器
			writer.body.Reset()
			writer.ResponseWriter = baseWriter

			// 执行请求
			c.Next()

			// 检查响应状态码
			statusCode := writer.Status()
			lastStatusCode = statusCode

			// 成功或客户端错误，不重试
			if statusCode < 500 {
				return
			}

			// 判断是否可重试
			if !isRetryableHTTPStatus(statusCode, policy.RetryableCodes) {
				return
			}

			// 构造错误
			lastErr = contract.NewError(statusCode, contract.ErrorReasonInternal, "request failed")

			// 最后一次尝试不等待
			if attempt == policy.MaxAttempts-1 {
				break
			}

			// 检查 context
			select {
			case <-c.Request.Context().Done():
				writeResponseHeaders(c)
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

			// 计算延迟
			jitter := rng.Float64()
			delay := policy.CalculateDelay(attempt, jitter)

			select {
			case <-c.Request.Context().Done():
				writeResponseHeaders(c)
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

		// 重试耗尽，返回错误
		if lastErr != nil {
			writeResponseHeaders(c)
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

// RetryAllMethodsMiddleware 创建对所有 HTTP 方法重试的中间件。
//
// 中文说明：
// - 与 RetryMiddleware 类似，但对所有方法都重试；
// - 注意：POST/PUT/DELETE 等非幂等方法可能导致重复操作。
func RetryAllMethodsMiddleware(retry contract.Retry, policy contract.RetryPolicy) gin.HandlerFunc {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(c *gin.Context) {
		// 读取请求体（用于重试时恢复）
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 使用响应写入器捕获响应
		writer := &retryResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		baseWriter := c.Writer
		c.Writer = writer

		var lastErr error
		var lastStatusCode int

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			// 重置请求体
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// 重置响应写入器
			writer.body.Reset()
			writer.ResponseWriter = baseWriter

			// 执行请求
			c.Next()

			// 检查响应状态码
			statusCode := writer.Status()
			lastStatusCode = statusCode

			// 成功或客户端错误，不重试
			if statusCode < 500 {
				return
			}

			// 判断是否可重试
			if !isRetryableHTTPStatus(statusCode, policy.RetryableCodes) {
				return
			}

			// 构造错误
			lastErr = contract.NewError(statusCode, contract.ErrorReasonInternal, "request failed")

			// 最后一次尝试不等待
			if attempt == policy.MaxAttempts-1 {
				break
			}

			// 检查 context
			select {
			case <-c.Request.Context().Done():
				writeResponseHeaders(c)
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

			// 计算延迟
			jitter := rng.Float64()
			delay := policy.CalculateDelay(attempt, jitter)

			select {
			case <-c.Request.Context().Done():
				writeResponseHeaders(c)
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

		// 重试耗尽，返回错误
		if lastErr != nil {
			writeResponseHeaders(c)
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

// isIdempotentMethod 判断是否为幂等 HTTP 方法。
func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// isRetryableHTTPStatus 判断 HTTP 状态码是否可重试。
func isRetryableHTTPStatus(status int, retryableCodes []int) bool {
	// 检查配置的可重试状态码
	for _, code := range retryableCodes {
		if status == code {
			return true
		}
	}

	// 默认 502, 503, 504 可重试
	return status == 502 || status == 503 || status == 504
}

// retryResponseWriter 用于捕获响应的写入器。
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

// DoWithRetry 辅助函数：在 handler 中执行带重试的操作。
//
// 中文说明：
// - 在业务逻辑中使用 Retry 服务；
// - 自动处理 context 取消。
//
// 使用示例：
//
//	func handler(c *gin.Context) {
//	    result, err := gin.DoWithRetry(c, retry, func(ctx context.Context) (any, error) {
//	        return callExternalService(ctx)
//	    })
//	    if err != nil {
//	        c.JSON(500, gin.H{"error": err.Error()})
//	        return
//	    }
//	    c.JSON(200, result)
//	}
func DoWithRetry(c *gin.Context, retry contract.Retry, fn func(ctx context.Context) (any, error)) (any, error) {
	return retry.DoWithResult(c.Request.Context(), func() (any, error) {
		return fn(c.Request.Context())
	})
}
