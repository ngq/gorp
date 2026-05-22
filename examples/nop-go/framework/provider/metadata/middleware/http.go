// Package middleware provides HTTP metadata propagation middleware.
// Implements MetadataCarrier interface for HTTP headers.
// Supports automatic extraction and injection across HTTP boundaries.
//
// 中间件包提供 HTTP 元数据传播中间件。
// 为 HTTP headers 实现 MetadataCarrier 接口。
// 支持跨 HTTP 边界的自动提取和注入。
package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// unwrapGinContext extracts the raw gin.Context from a transport Context.
func unwrapGinContext(c transportcontract.Context) (*gin.Context, bool) {
	type ginContextProvider interface {
		GinContext() *gin.Context
	}
	provider, ok := c.(ginContextProvider)
	if !ok {
		return nil, false
	}
	gc := provider.GinContext()
	if gc == nil {
		return nil, false
	}
	return gc, true
}

// HeaderCarrier wraps http.Header to implement MetadataCarrier interface.
// Core logic: Delegate Get/Set/Add operations to underlying http.Header.
//
// HeaderCarrier 包装 http.Header，实现 MetadataCarrier 接口。
// 核心逻辑：将 Get/Set/Add 操作委托给底层 http.Header。
type HeaderCarrier struct {
	header http.Header
}

// NewHeaderCarrier creates a new HeaderCarrier with http.Header.
// Core logic: Wrap http.Header in carrier.
//
// NewHeaderCarrier 创建新的 HeaderCarrier，携带 http.Header。
// 核心逻辑：将 http.Header 包装为 carrier。
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

// MetadataMiddleware creates HTTP middleware for metadata extraction.
// Core logic: Extract metadata from request headers, inject into context.
//
// MetadataMiddleware 创建 HTTP 中间件，用于提取元数据。
// 核心逻辑：从请求头提取元数据、注入到 context。
func MetadataMiddleware(propagator transportcontract.MetadataPropagator) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			carrier := NewHeaderCarrier(c.Request().Header)
			ctx := propagator.Extract(c.Context(), carrier)
			// Check if metadata was extracted
			if md, ok := transportcontract.FromServerContext(ctx); ok && md != nil {
				c.Set("metadata", md)
			}
			// Update gin.Request.Context for context.Context value propagation
			if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
				gc.Request = gc.Request.WithContext(ctx)
			}
			if next != nil {
				next(c)
			}
		}
	}
}

// MetadataInjector injects metadata into HTTP requests.
// Core logic: Use propagator to inject metadata into request headers.
//
// MetadataInjector 将元数据注入 HTTP 请求。
// 核心逻辑：使用 propagator 将元数据注入到请求头。
type MetadataInjector struct {
	propagator transportcontract.MetadataPropagator
}

// NewMetadataInjector creates a new metadata injector with propagator.
// Core logic: Store propagator for later injection.
//
// NewMetadataInjector 创建新的元数据注入器，携带 propagator。
// 核心逻辑：存储 propagator 用于后续注入。
func NewMetadataInjector(propagator transportcontract.MetadataPropagator) *MetadataInjector {
	return &MetadataInjector{propagator: propagator}
}

// Inject injects metadata into HTTP request headers.
// Core logic: Create carrier from headers, call propagator.Inject.
//
// Inject 将元数据注入 HTTP 请求头。
// 核心逻辑：从 headers 创建 carrier、调用 propagator.Inject。
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
