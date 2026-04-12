package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

// HeaderCarrier 实现 MetadataCarrier 接口。
//
// 中文说明：
// - 包装 http.Header；
// - 用于 HTTP 请求的 metadata 提取/注入。
type HeaderCarrier struct {
	header http.Header
}

// NewHeaderCarrier 创建 HeaderCarrier。
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

// MetadataMiddleware HTTP metadata 传递中间件。
//
// 中文说明：
// - 从 HTTP Header 提取 metadata 存入 context；
// - 支持前缀过滤（默认 x-md- 前缀）；
// - 自己实现，不抄袭 Kratos。
func MetadataMiddleware(propagator contract.MetadataPropagator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 HTTP Header 提取 metadata
		carrier := NewHeaderCarrier(c.Request.Header)
		ctx := propagator.Extract(c.Request.Context(), carrier)

		// 更新 request context
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// MetadataInjector HTTP metadata 注入器。
//
// 中文说明：
// - 用于客户端请求；
// - 从 context 读取 metadata 注入 HTTP Header。
type MetadataInjector struct {
	propagator contract.MetadataPropagator
}

// NewMetadataInjector 创建 metadata 注入器。
func NewMetadataInjector(propagator contract.MetadataPropagator) *MetadataInjector {
	return &MetadataInjector{propagator: propagator}
}

// Inject 注入 metadata 到 HTTP 请求。
//
// 中文说明：
// - 从 context 提取 metadata；
// - 注入到 HTTP Header。
func (i *MetadataInjector) Inject(req *http.Request) {
	carrier := NewHeaderCarrier(req.Header)
	i.propagator.Inject(req.Context(), carrier)
}

// InjectToGinContext 注入 metadata 到 Gin Context。
//
// 中文说明：
// - 用于 Gin 中间件链；
// - 从 context 提取 metadata 存入 Gin Context。
func InjectToGinContext(ctx *gin.Context, md contract.Metadata) {
	// 存入 Gin Context
	ctx.Set("metadata", md)
}

// ExtractFromGinContext 从 Gin Context 提取 metadata。
//
// 中文说明：
// - 从 Gin Context 读取 metadata；
// - 返回 Metadata 对象。
func ExtractFromGinContext(ctx *gin.Context) contract.Metadata {
	if v, exists := ctx.Get("metadata"); exists {
		if md, ok := v.(contract.Metadata); ok {
			return md
		}
	}
	return nil
}

// GetFromContext 从任意 context 获取 metadata。
//
// 中文说明：
// - 优先从 server context 获取；
// - 其次从 client context 获取；
// - 返回 metadata 和是否存在标志。
func GetFromContext(ctx context.Context) (contract.Metadata, bool) {
	// 先尝试 server context
	if md, ok := contract.FromServerContext(ctx); ok {
		return md, true
	}
	// 再尝试 client context
	return contract.FromClientContext(ctx)
}

// GetHeaderValue 从 context 获取指定 header 值。
//
// 中文说明：
// - 便捷方法，直接获取 metadata 中的值。
func GetHeaderValue(ctx context.Context, key string) string {
	if md, ok := GetFromContext(ctx); ok {
		return md.Get(key)
	}
	return ""
}