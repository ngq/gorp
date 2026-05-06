// Application scenarios:
// - Prevent oversized request bodies from consuming too much memory or bandwidth.
// - Protect upload, callback, and external API endpoints from abusive payloads.
// - Provide a simple server-side request size guardrail.
//
// 适用场景：
// - 防止超大请求体占用过多内存或网络带宽。
// - 保护上传、回调和对外 API 接口免受恶意大包冲击。
// - 提供简单直接的服务端请求体大小护栏。
package middleware

import (
	"errors"
	"io"
	"net/http"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// BodyLimit constrains the incoming request body size for the HTTP mainline.
//
// BodyLimit 为 HTTP 主线限制请求体大小。
func BodyLimit(maxBytes int64) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if c == nil || maxBytes <= 0 {
				if next != nil {
					next(c)
				}
				return
			}

			req := c.Request()
			if req == nil {
				if next != nil {
					next(c)
				}
				return
			}

			if req.ContentLength > maxBytes {
				respondBodyTooLarge(c)
				return
			}

			if gc, ok := unwrapGinContext(c); ok && req.Body != nil {
				req.Body = http.MaxBytesReader(gc.Writer, req.Body, maxBytes)
				c.SetRequest(req)
			}

			if next != nil {
				next(c)
				if errors.Is(req.Context().Err(), io.EOF) {
					return
				}
			}
		}
	}
}

// respondBodyTooLarge writes the unified request-entity-too-large response.
//
// respondBodyTooLarge 输出统一的请求体过大响应。
func respondBodyTooLarge(c transportcontract.HTTPContext) {
	if gc, ok := unwrapGinContext(c); ok {
		writeGinResponseHeaders(gc)
		resp := Response{
			Code:    CodeBadRequest,
			Message: "request body too large",
		}
		gc.JSON(http.StatusRequestEntityTooLarge, resp)
		gc.Abort()
		return
	}

	c.JSON(http.StatusRequestEntityTooLarge, map[string]any{
		"code":    CodeBadRequest,
		"message": "request body too large",
	})
}
