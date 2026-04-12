package contract

import (
	"context"
)

const (
	// MetadataKey 是 Metadata 在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于服务间元数据传递；
	// - noop 实现返回空 metadata，单体项目零依赖；
	// - 微服务项目可启用 metadata 传递机制。
	MetadataKey = "framework.metadata"

	// MetadataPropagatorKey 是 Metadata 传播器在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于 HTTP / gRPC 客户端和服务端注入/提取 metadata；
	// - 默认实现可按前缀传播 metadata，并注入常量 metadata；
	// - noop 模式下返回空传播器。
	MetadataPropagatorKey = "framework.metadata.propagator"
)

// Metadata 定义服务间元数据传递接口。
//
// 中文说明：
// - 用于 HTTP Header / gRPC Metadata 的统一传递；
// - 服务端提取 metadata 存入 context；
// - 客户端从 context 读取 metadata 注入请求；
// - noop 实现空操作，单体项目零依赖。
type Metadata interface {
	// Get 获取指定 key 的第一个值。
	//
	// 中文说明：
	// - key 不区分大小写；
	// - 如果 key 不存在，返回空字符串。
	Get(key string) string

	// Values 获取指定 key 的所有值。
	//
	// 中文说明：
	// - key 不区分大小写；
	// - 返回值列表，可能为空。
	Values(key string) []string

	// Set 设置指定 key 的值（覆盖）。
	//
	// 中文说明：
	// - key 不区分大小写；
	// - 覆盖已有值。
	Set(key, value string)

	// Add 添加指定 key 的值（追加）。
	//
	// 中文说明：
	// - key 不区分大小写；
	// - 追加到已有值列表。
	Add(key, value string)

	// Del 删除指定 key。
	Del(key string)

	// Range 遍历所有键值对。
	//
	// 中文说明：
	// - f 返回 false 时停止遍历。
	Range(f func(key string, values []string) bool)

	// Clone 深拷贝 metadata。
	Clone() Metadata

	// ToMap 转换为 map[string][]string。
	//
	// 中文说明：
	// - 用于 HTTP Header / gRPC Metadata 设置。
	ToMap() map[string][]string
}

// MetadataCarrier 定义 metadata 载体接口。
//
// 中文说明：
// - HTTP Header 实现 this interface；
// - gRPC Metadata 实现此接口；
// - 用于统一的 metadata 提取/注入。
type MetadataCarrier interface {
	// Get 获取指定 key 的值
	Get(key string) string

	// Set 设置指定 key 的值
	Set(key, value string)

	// Add 添加指定 key 的值
	Add(key, value string)

	// Keys 获取所有 key
	Keys() []string

	// Values 获取指定 key 的所有值
	Values(key string) []string
}

// MetadataPropagator 定义 metadata 传播器接口。
//
// 中文说明：
	// - 控制哪些 metadata 需要跨服务传递；
// - 支持前缀匹配（如 x-md- 开头的 header）；
// - 支持白名单过滤。
type MetadataPropagator interface {
	// Inject 从 context 提取 metadata 注入 carrier。
	//
	// 中文说明：
	// - ctx: 源 context（包含 metadata）；
	// - carrier: 目标载体（HTTP Header / gRPC Metadata）；
	// - 服务端 -> 客户端 调用时使用。
	Inject(ctx context.Context, carrier MetadataCarrier)

	// Extract 从 carrier 提取 metadata 存入 context。
	//
	// 中文说明：
	// - ctx: 源 context；
	// - carrier: 源载体（HTTP Header / gRPC Metadata）；
	// - 返回包含 metadata 的新 context；
	// - 客户端 -> 服务端 调用时使用。
	Extract(ctx context.Context, carrier MetadataCarrier) context.Context
}

// MetadataConfig 定义 metadata 配置。
//
// 中文说明：
// - PropagatePrefix: 需要传播的 key 前缀列表；
// - ConstantMetadata: 常量 metadata（每次请求都携带）；
// - MaxSize: metadata 最大大小（防止过大的 header）。
type MetadataConfig struct {
	// PropagatePrefix 需要传播的 key 前缀
	PropagatePrefix []string `mapstructure:"propagate_prefix"`

	// ConstantMetadata 常量 metadata
	ConstantMetadata map[string]string `mapstructure:"constant_metadata"`

	// MaxSize 最大 metadata 大小（字节）
	MaxSize int `mapstructure:"max_size"`
}

// contextKey 定义 context key 类型。
type contextKey struct{ name string }

var (
	// serverMetadataKey 服务端 metadata context key
	serverMetadataKey = &contextKey{name: "server-metadata"}

	// clientMetadataKey 客户端 metadata context key
	clientMetadataKey = &contextKey{name: "client-metadata"}
)

// NewServerContext 创建包含服务端 metadata 的 context。
//
// 中文说明：
// - 服务端接收到请求时使用；
// - 从 HTTP Header / gRPC Metadata 提取后存入 context。
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey, md)
}

// FromServerContext 从 context 获取服务端 metadata。
//
// 中文说明：
// - 服务端处理请求时使用；
// - 返回 metadata 和是否存在标志。
func FromServerContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(serverMetadataKey).(Metadata)
	return md, ok
}

// NewClientContext 创建包含客户端 metadata 的 context。
//
// 中文说明：
// - 客户端发起请求时使用；
// - metadata 会被注入到 HTTP Header / gRPC Metadata。
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, clientMetadataKey, md)
}

// FromClientContext 从 context 获取客户端 metadata。
//
// 中文说明：
// - 客户端发起请求时使用；
// - 返回 metadata 和是否存在标志。
func FromClientContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(clientMetadataKey).(Metadata)
	return md, ok
}

// AppendToClientContext 向 context 追加客户端 metadata。
//
// 中文说明：
// - kv 必须成对出现；
// - 追加到已有 metadata 中。
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	if len(kv)%2 != 0 {
		return ctx
	}
	md, _ := FromClientContext(ctx)
	if md == nil {
		md = NewMetadata()
	} else {
		md = md.Clone()
	}
	for i := 0; i < len(kv); i += 2 {
		md.Set(kv[i], kv[i+1])
	}
	return NewClientContext(ctx, md)
}

// NewMetadata 创建空的 Metadata 实例。
//
	// 中文说明：
// - 用于创建新的 metadata 对象；
// - 默认返回内存实现。
func NewMetadata() Metadata {
	return &mapMetadata{data: make(map[string][]string)}
}

// NewMetadataFromMap 从 map 创建 Metadata 实例。
//
// 中文说明：
// - 从 HTTP Header / gRPC Metadata 转换；
// - 复制数据，不影响原始 map。
func NewMetadataFromMap(m map[string][]string) Metadata {
	data := make(map[string][]string, len(m))
	for k, v := range m {
		data[lowerKey(k)] = copySlice(v)
	}
	return &mapMetadata{data: data}
}

// mapMetadata 是 Metadata 的 map 实现。
type mapMetadata struct {
	data map[string][]string
}

func (m *mapMetadata) Get(key string) string {
	key = lowerKey(key)
	if v, ok := m.data[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func (m *mapMetadata) Values(key string) []string {
	return m.data[lowerKey(key)]
}

func (m *mapMetadata) Set(key, value string) {
	if key == "" {
		return
	}
	m.data[lowerKey(key)] = []string{value}
}

func (m *mapMetadata) Add(key, value string) {
	if key == "" {
		return
	}
	k := lowerKey(key)
	m.data[k] = append(m.data[k], value)
}

func (m *mapMetadata) Del(key string) {
	delete(m.data, lowerKey(key))
}

func (m *mapMetadata) Range(f func(key string, values []string) bool) {
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}

func (m *mapMetadata) Clone() Metadata {
	data := make(map[string][]string, len(m.data))
	for k, v := range m.data {
		data[k] = copySlice(v)
	}
	return &mapMetadata{data: data}
}

func (m *mapMetadata) ToMap() map[string][]string {
	result := make(map[string][]string, len(m.data))
	for k, v := range m.data {
		result[k] = copySlice(v)
	}
	return result
}

// lowerKey 转换 key 为小写。
func lowerKey(key string) string {
	// 简单实现，实际项目可用 strings.ToLower
	b := make([]byte, len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// copySlice 复制字符串切片。
func copySlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	r := make([]string, len(s))
	copy(r, s)
	return r
}