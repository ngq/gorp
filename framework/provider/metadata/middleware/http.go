package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type HeaderCarrier struct {
	header http.Header
}

func NewHeaderCarrier(h http.Header) *HeaderCarrier {
	return &HeaderCarrier{header: h}
}

func (c *HeaderCarrier) Get(key string) string {
	return c.header.Get(key)
}

func (c *HeaderCarrier) Set(key, value string) {
	c.header.Set(key, value)
}

func (c *HeaderCarrier) Add(key, value string) {
	c.header.Add(key, value)
}

func (c *HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.header))
	for k := range c.header {
		keys = append(keys, k)
	}
	return keys
}

func (c *HeaderCarrier) Values(key string) []string {
	return c.header.Values(key)
}

func MetadataMiddleware(propagator transportcontract.MetadataPropagator) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			carrier := NewHeaderCarrier(c.Request().Header)
			ctx := propagator.Extract(c.Context(), carrier)
			c.SetContext(ctx)
			if next != nil {
				next(c)
			}
		}
	}
}

type MetadataInjector struct {
	propagator transportcontract.MetadataPropagator
}

func NewMetadataInjector(propagator transportcontract.MetadataPropagator) *MetadataInjector {
	return &MetadataInjector{propagator: propagator}
}

func (i *MetadataInjector) Inject(req *http.Request) {
	carrier := NewHeaderCarrier(req.Header)
	i.propagator.Inject(req.Context(), carrier)
}

func InjectToGinContext(ctx *gin.Context, md transportcontract.Metadata) {
	ctx.Set("metadata", md)
}

func ExtractFromGinContext(ctx *gin.Context) transportcontract.Metadata {
	if v, exists := ctx.Get("metadata"); exists {
		if md, ok := v.(transportcontract.Metadata); ok {
			return md
		}
	}
	return nil
}

func GetFromContext(ctx context.Context) (transportcontract.Metadata, bool) {
	if md, ok := transportcontract.FromServerContext(ctx); ok {
		return md, true
	}
	return transportcontract.FromClientContext(ctx)
}

func GetHeaderValue(ctx context.Context, key string) string {
	if md, ok := GetFromContext(ctx); ok {
		return md.Get(key)
	}
	return ""
}
