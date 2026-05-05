package gin

import (
	"context"
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

func RequestIdentity() transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			rid := resolveRequestID(c)
			c.Header(requestIDHeader, rid)
			ctx := supportcontract.NewRequestIDContext(c.Context(), rid)

			tid := resolveTraceID(c, ctx)
			c.Header(traceIDHeader, tid)
			ctx = supportcontract.NewTraceIDContext(ctx, tid)

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

func resolveRequestID(c transportcontract.HTTPContext) string {
	if c == nil {
		return generateIdentityID()
	}
	rid := c.GetHeader(requestIDHeader)
	if rid == "" {
		rid = generateIdentityID()
	}
	return rid
}

func resolveTraceID(c transportcontract.HTTPContext, ctx any) string {
	if c != nil {
		tid := c.GetHeader(traceIDHeader)
		if tid == "" {
			tid = c.GetHeader(traceIDHeaderAlt)
		}
		if tid != "" {
			return tid
		}
	}

	if baseCtx, ok := ctx.(context.Context); ok {
		if rid, ok := supportcontract.FromRequestIDContext(baseCtx); ok && rid != "" {
			return rid
		}
	}

	return generateIdentityID()
}

func generateIdentityID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
