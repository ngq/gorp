// Package governance provides outbound RPC governance middleware functions.
// This file implements a gRPC metadata carrier for TextMapCarrier compatibility.
//
// Package governance 提供出站 RPC 治理中间件函数。
// 本文件实现 gRPC metadata carrier，用于 TextMapCarrier 兼容性。
package governance

import (
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"google.golang.org/grpc/metadata"
)

// grpcMetadataCarrier wraps gRPC metadata.MD to implement MetadataCarrier interface.
// Used for injecting and extracting metadata from gRPC outgoing context.
//
// grpcMetadataCarrier 包装 gRPC metadata.MD，实现 MetadataCarrier 接口。
// 用于从 gRPC outgoing context 注入和提取 metadata。
type grpcMetadataCarrier struct {
	md metadata.MD
}

// NewGRPCMetadataCarrier creates a new carrier wrapping the given gRPC metadata.
// Creates empty metadata if nil is provided.
//
// NewGRPCMetadataCarrier 创建包装给定 gRPC metadata 的 carrier。
// 如果传入 nil，创建空 metadata。
func NewGRPCMetadataCarrier(md metadata.MD) transportcontract.MetadataCarrier {
	if md == nil {
		md = metadata.New(nil)
	}
	return &grpcMetadataCarrier{md: md}
}

// GRPCMetadataFromCarrier extracts the underlying gRPC metadata from a MetadataCarrier.
// Returns nil if the carrier is not a grpcMetadataCarrier.
//
// GRPCMetadataFromCarrier 从 MetadataCarrier 提取底层 gRPC metadata。
// 如果 carrier 不是 grpcMetadataCarrier，返回 nil。
func GRPCMetadataFromCarrier(carrier transportcontract.MetadataCarrier) metadata.MD {
	if c, ok := carrier.(*grpcMetadataCarrier); ok {
		return c.md
	}
	return nil
}

// Get returns the first value for the given key from the metadata.
//
// Get 从 metadata 返回给定键的第一个值。
func (c *grpcMetadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets the value for the given key in the metadata, replacing any existing values.
//
// Set 在 metadata 中为给定键设置值，替换任何现有值。
func (c *grpcMetadataCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

// Add appends a value to the given key in the metadata.
//
// Add 在 metadata 中为给定键追加值。
func (c *grpcMetadataCarrier) Add(key, value string) {
	c.md.Append(key, value)
}

// Keys returns all keys present in the metadata.
//
// Keys 返回 metadata 中存在的所有键。
func (c *grpcMetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values for the given key from the metadata.
//
// Values 从 metadata 返回给定键的所有值。
func (c *grpcMetadataCarrier) Values(key string) []string {
	return c.md.Get(key)
}