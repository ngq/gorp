// Package grpc provides gRPC metadata carrier for propagation.
// This file implements transportcontract.MetadataCarrier using gRPC metadata.MD.
//
// 本包提供 gRPC metadata carrier 用于传播。
// 本文件使用 gRPC metadata.MD 实现 transportcontract.MetadataCarrier。
package grpc

import (
	"google.golang.org/grpc/metadata"
)

// grpcMetadataCarrier implements transportcontract.MetadataCarrier using gRPC metadata.
// Used for injecting and extracting metadata during RPC calls.
//
// grpcMetadataCarrier 使用 gRPC metadata 实现 transportcontract.MetadataCarrier。
// 用于 RPC 调用期间注入和提取 metadata。
type grpcMetadataCarrier struct {
	md metadata.MD
}

// newGRPCMetadataCarrier creates a new metadata carrier from gRPC metadata.MD.
// Initializes empty metadata if nil is passed.
//
// newGRPCMetadataCarrier 从 gRPC metadata.MD 创建新的 metadata carrier。
// 如果传入 nil 则初始化空 metadata。
func newGRPCMetadataCarrier(md metadata.MD) *grpcMetadataCarrier {
	if md == nil {
		md = metadata.New(nil)
	}
	return &grpcMetadataCarrier{md: md}
}

// Get returns the first value for the given key.
// Implements transportcontract.MetadataCarrier.Get.
//
// Get 返回给定 key 的第一个值。
// 实现 transportcontract.MetadataCarrier.Get。
func (c *grpcMetadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets the value for the given key, replacing any existing values.
// Implements transportcontract.MetadataCarrier.Set.
//
// Set 设置给定 key 的值，替换任何现有值。
// 实现 transportcontract.MetadataCarrier.Set。
func (c *grpcMetadataCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

// Add appends a value to the given key.
// Implements transportcontract.MetadataCarrier.Add.
//
// Add 向给定 key 添加值。
// 实现 transportcontract.MetadataCarrier.Add。
func (c *grpcMetadataCarrier) Add(key, value string) {
	c.md.Append(key, value)
}

// Keys returns all keys in the metadata.
// Implements transportcontract.MetadataCarrier.Keys.
//
// Keys 返回 metadata 中的所有 key。
// 实现 transportcontract.MetadataCarrier.Keys。
func (c *grpcMetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values for the given key.
// Implements transportcontract.MetadataCarrier.Values.
//
// Values 返回给定 key 的所有值。
// 实现 transportcontract.MetadataCarrier.Values。
func (c *grpcMetadataCarrier) Values(key string) []string {
	return c.md.Get(key)
}