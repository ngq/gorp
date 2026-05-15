// Application scenarios:
// - Define transport-level metadata propagation contracts for HTTP and RPC flows.
// - Provide a reusable in-memory metadata implementation with cloning and context propagation helpers.
// - Keep metadata access, injection, and extraction semantics uniform across providers.
//
// 适用场景：
// - 定义 HTTP 和 RPC 流程共享的 transport 层 metadata 透传契约。
// - 提供一个可复用的内存 metadata 实现，以及克隆和 context 透传助手。
// - 在不同 provider 之间统一 metadata 的访问、注入和提取语义。
package transport

import "context"

const (
	MetadataKey           = "framework.metadata"
	MetadataPropagatorKey = "framework.metadata.propagator"
)

// Metadata defines the transport metadata abstraction.
//
// Metadata 定义 transport metadata 抽象。
type Metadata interface {
	Get(key string) string
	Values(key string) []string
	Set(key, value string)
	Add(key, value string)
	Del(key string)
	Range(f func(key string, values []string) bool)
	Clone() Metadata
	ToMap() map[string][]string
}

// MetadataCarrier defines the read/write carrier abstraction used during propagation.
//
// MetadataCarrier 定义透传过程中使用的读写载体抽象。
type MetadataCarrier interface {
	Get(key string) string
	Set(key, value string)
	Add(key, value string)
	Keys() []string
	Values(key string) []string
}

// MetadataPropagator defines how metadata should be injected and extracted.
//
// MetadataPropagator 定义 metadata 的注入与提取方式。
type MetadataPropagator interface {
	Inject(ctx context.Context, carrier MetadataCarrier)
	Extract(ctx context.Context, carrier MetadataCarrier) context.Context
}

// MetadataConfig describes metadata propagation behavior.
//
// MetadataConfig 描述 metadata 透传配置。
type MetadataConfig struct {
	PropagatePrefix  []string          `mapstructure:"propagate_prefix"`
	ConstantMetadata map[string]string `mapstructure:"constant_metadata"`
	MaxSize          int               `mapstructure:"max_size"`
}

type contextKey struct{ name string }

var (
	serverMetadataKey = &contextKey{name: "server-metadata"}
	clientMetadataKey = &contextKey{name: "client-metadata"}
)

// NewServerContext attaches server-side metadata into a context.
//
// NewServerContext 将服务端 metadata 绑定到 context。
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey, md)
}

// FromServerContext reads server-side metadata from a context.
//
// FromServerContext 从 context 中读取服务端 metadata。
func FromServerContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(serverMetadataKey).(Metadata)
	return md, ok
}

// NewClientContext attaches client-side metadata into a context.
//
// NewClientContext 将客户端 metadata 绑定到 context。
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, clientMetadataKey, md)
}

// FromClientContext reads client-side metadata from a context.
//
// FromClientContext 从 context 中读取客户端 metadata。
func FromClientContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(clientMetadataKey).(Metadata)
	return md, ok
}

// AppendToClientContext appends metadata key-value pairs into a client context.
//
// AppendToClientContext 向客户端 context 追加 metadata 键值对。
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	if len(kv)%2 != 0 {
		// Ignore malformed input pairs so callers do not accidentally poison metadata state.
		// 忽略不成对的输入，避免调用方意外破坏 metadata 状态。
		return ctx
	}
	md, _ := FromClientContext(ctx)
	if md == nil {
		md = NewMetadata()
	} else {
		// Clone before mutation so upstream callers do not observe in-place metadata side effects.
		// 写入前先克隆，避免上游调用方观察到原地修改带来的副作用。
		md = md.Clone()
	}
	for i := 0; i < len(kv); i += 2 {
		md.Set(kv[i], kv[i+1])
	}
	return NewClientContext(ctx, md)
}

// NewMetadata creates an empty in-memory metadata implementation.
//
// NewMetadata 创建一个空的内存 metadata 实现。
func NewMetadata() Metadata {
	return &mapMetadata{data: make(map[string][]string)}
}

// NewMetadataFromMap creates an in-memory metadata object from a map snapshot.
//
// NewMetadataFromMap 从 map 快照创建内存 metadata 对象。
func NewMetadataFromMap(m map[string][]string) Metadata {
	data := make(map[string][]string, len(m))
	for k, v := range m {
		// Normalize keys to lower case so metadata lookup stays case-insensitive across carriers.
		// 将 key 统一规整为小写，保证不同 carrier 之间的 metadata 查找保持大小写不敏感。
		data[lowerKey(k)] = copySlice(v)
	}
	return &mapMetadata{data: data}
}

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
		data[k] = cloneValues(v)
	}
	return &mapMetadata{data: data}
}

func (m *mapMetadata) ToMap() map[string][]string {
	result := make(map[string][]string, len(m.data))
	for k, v := range m.data {
		result[k] = cloneValues(v)
	}
	return result
}

func lowerKey(key string) string {
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

func copySlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	r := make([]string, len(s))
	copy(r, s)
	return r
}

// cloneValues clones a value slice. For single-element slices (the common case
// from Set), it avoids the overhead of copy() and allocates directly.
//
// cloneValues 克隆值切片。对于单元素切片（Set 的常见场景），
// 跳过 copy() 开销直接分配，减少内存分配次数。
func cloneValues(v []string) []string {
	if len(v) == 0 {
		return nil
	}
	if len(v) == 1 {
		return []string{v[0]}
	}
	r := make([]string, len(v))
	copy(r, v)
	return r
}
