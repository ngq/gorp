package gin

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	requestIDHeader  = "X-Request-Id"
	traceIDHeader    = "X-Trace-Id"
	traceIDHeaderAlt = "X-Trace-Id"
)

func RequestID() transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			rid := c.GetHeader(requestIDHeader)
			if rid == "" {
				buf := make([]byte, 16)
				_, _ = rand.Read(buf)
				rid = hex.EncodeToString(buf)
			}
			c.Header(requestIDHeader, rid)
			ctx := supportcontract.NewRequestIDContext(c.Context(), rid)
			c.SetContext(ctx)
			if next != nil {
				next(c)
			}
		}
	}
}

func TraceID() transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			tid := c.GetHeader(traceIDHeader)
			if tid == "" {
				tid = c.GetHeader(traceIDHeaderAlt)
			}
			if tid == "" {
				if rid, ok := supportcontract.FromRequestIDContext(c.Context()); ok && rid != "" {
					tid = rid
				} else {
					buf := make([]byte, 16)
					_, _ = rand.Read(buf)
					tid = hex.EncodeToString(buf)
				}
			}
			c.Header(traceIDHeader, tid)
			ctx := supportcontract.NewTraceIDContext(c.Context(), tid)
			c.SetContext(ctx)
			if next != nil {
				next(c)
			}
		}
	}
}

func GetTraceID(c *gin.Context) string {
	if c != nil && c.Request != nil {
		if tid, ok := supportcontract.FromTraceIDContext(c.Request.Context()); ok {
			return tid
		}
	}
	return ""
}

func GetRequestID(c *gin.Context) string {
	if c != nil && c.Request != nil {
		if rid, ok := supportcontract.FromRequestIDContext(c.Request.Context()); ok {
			return rid
		}
	}
	return ""
}
