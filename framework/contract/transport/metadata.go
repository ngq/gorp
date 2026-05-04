package transport

import "context"

const (
	MetadataKey           = "framework.metadata"
	MetadataPropagatorKey = "framework.metadata.propagator"
)

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

type MetadataCarrier interface {
	Get(key string) string
	Set(key, value string)
	Add(key, value string)
	Keys() []string
	Values(key string) []string
}

type MetadataPropagator interface {
	Inject(ctx context.Context, carrier MetadataCarrier)
	Extract(ctx context.Context, carrier MetadataCarrier) context.Context
}

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

func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey, md)
}

func FromServerContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(serverMetadataKey).(Metadata)
	return md, ok
}

func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, clientMetadataKey, md)
}

func FromClientContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(clientMetadataKey).(Metadata)
	return md, ok
}

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

func NewMetadata() Metadata {
	return &mapMetadata{data: make(map[string][]string)}
}

func NewMetadataFromMap(m map[string][]string) Metadata {
	data := make(map[string][]string, len(m))
	for k, v := range m {
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
