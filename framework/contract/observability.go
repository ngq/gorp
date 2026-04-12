package contract

import (
	"context"
	"time"
)

const (
	ObservabilityKey = "framework.observability"
	TracerKey        = "framework.tracer"
	TracerProviderKey = "framework.tracer.provider"
)

// Observability 统一观测接口。
//
// 中文说明：
// - 整合 Metrics、Tracing、Logging 三方面的能力；
// - 提供统一的观测入口，简化业务代码使用；
// - 后续可对接 OpenTelemetry 等标准。
type Observability interface {
	// Metrics 返回指标收集器
	Metrics() Metrics
	// Tracer 返回链路追踪器
	Tracer() Tracer
	// Logger 返回日志记录器
	Logger() Logger
	// ErrorReporter 返回错误上报器
	ErrorReporter() ErrorReporter
}

// Metrics 指标收集接口。
//
// 中文说明：
// - 定义统一的指标收集接口；
// - 支持 Counter、Gauge、Histogram 三种类型；
// - 可对接 Prometheus、OpenTelemetry 等。
type Metrics interface {
	// Counter 增加计数器
	Counter(name string, labels map[string]string, delta float64)
	// Gauge 设置仪表值
	Gauge(name string, labels map[string]string, value float64)
	// Histogram 记录直方图值
	Histogram(name string, labels map[string]string, value float64)
	// Timing 记录耗时
	Timing(name string, labels map[string]string, duration time.Duration)
}

// TracerProvider 追踪器提供者接口。
//
// 中文说明：
// - 创建和管理 Tracer 实例；
// - 支持全局注册和关闭；
// - noop 实现返回空追踪器，单体项目零依赖。
type TracerProvider interface {
	// Tracer 创建或获取指定名称的追踪器。
	//
	// 中文说明：
	// - name: 追踪器名称，通常为服务名或模块名；
	// - options: 可选配置（如 schema URL、版本等）。
	Tracer(name string, options ...TracerOption) Tracer

	// Shutdown 关闭追踪器提供者。
	//
	// 中文说明：
	// - 刷新所有未导出的 Span；
	// - 关闭与后端的连接。
	Shutdown(ctx context.Context) error

	// ForceFlush 强制刷新所有 Span。
	ForceFlush(ctx context.Context) error
}

// TracerOption 追踪器配置选项。
type TracerOption func(*TracerConfig)

// TracerConfig 追踪器配置。
type TracerConfig struct {
	// SchemaURL OpenTelemetry schema URL
	SchemaURL string
	// Version 追踪器版本
	Version string
	// Attributes 默认属性
	Attributes map[string]interface{}
}

// Tracer 链路追踪接口。
//
// 中文说明：
// - 定义统一的链路追踪接口；
// - 支持 Span 的创建和上下文传递；
// - 可对接 OpenTelemetry、Jaeger 等。
type Tracer interface {
	// StartSpan 开始一个新的 Span。
	//
	// 中文说明：
	// - ctx: 父 Span 上下文；
	// - name: Span 名称（如 "HTTP GET /api/users"）；
	// - opts: Span 选项（如 SpanKind、Attributes）。
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// SpanFromContext 从上下文中获取当前 Span。
	SpanFromContext(ctx context.Context) Span

	// Inject 将追踪信息注入到载体中。
	//
	// 中文说明：
	// - 用于跨进程传递追踪上下文；
	// - carrier 类型：HTTP header、gRPC metadata 等。
	Inject(ctx context.Context, carrier TextMapCarrier) error

	// Extract 从载体中提取追踪信息。
	//
	// 中文说明：
	// - 用于接收跨进程传递的追踪上下文；
	// - 返回带有父 Span 上下文的 context。
	Extract(ctx context.Context, carrier TextMapCarrier) (context.Context, error)
}

// SpanOption Span 配置选项。
type SpanOption func(*SpanConfig)

// SpanConfig Span 配置。
type SpanConfig struct {
	// Kind Span 类型
	Kind SpanKind
	// Attributes Span 属性
	Attributes map[string]interface{}
	// StartTime Span 开始时间（默认为当前时间）
	StartTime time.Time
	// Links 关联的其他 Span
	Links []SpanLink
}

// SpanKind Span 类型。
type SpanKind int

const (
	SpanKindUnspecified SpanKind = iota
	SpanKindInternal
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// SpanLink 关联的 Span。
type SpanLink struct {
	SpanContext SpanContext
	Attributes  map[string]interface{}
}

// SpanContext Span 上下文信息。
type SpanContext struct {
	// TraceID 链路追踪 ID
	TraceID string
	// SpanID Span ID
	SpanID string
	// TraceFlags 追踪标志（如 sampled）
	TraceFlags TraceFlags
	// Remote 是否来自远程
	Remote bool
}

// TraceFlags 追踪标志。
type TraceFlags byte

const (
	TraceFlagsSampled TraceFlags = 0x01
)

// TextMapCarrier 文本映射载体接口。
//
// 中文说明：
// - 用于在 HTTP header、gRPC metadata 等载体中传递追踪信息；
// - Get/Set 方法用于读写键值对。
type TextMapCarrier interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

// Span 追踪 Span 接口。
type Span interface {
	// End 结束 Span。
	//
	// 中文说明：
	// - 记录 Span 结束时间；
	// - 将 Span 导出到后端。
	End(options ...SpanEndOption)

	// AddEvent 添加事件。
	//
	// 中文说明：
	// - name: 事件名称；
	// - attributes: 事件属性。
	AddEvent(name string, attributes map[string]interface{})

	// SetTag 设置标签。
	SetTag(key string, value interface{})

	// SetAttributes 设置多个属性。
	SetAttributes(attributes map[string]interface{})

	// SetError 设置错误。
	//
	// 中文说明：
	// - 自动设置 error=true 标签；
	// - 记录错误类型和消息。
	SetError(err error)

	// SetStatus 设置 Span 状态。
	//
	// 中文说明：
	// - code: 状态码（Ok、Error、Unset）；
	// - description: 状态描述。
	SetStatus(code SpanStatusCode, description string)

	// SpanContext 返回 Span 上下文。
	SpanContext() SpanContext

	// IsRecording 是否正在记录。
	//
	// 中文说明：
	// - 返回 false 表示 Span 被采样丢弃；
	// - 用于避免不必要的属性计算开销。
	IsRecording() bool

	// Context 返回包含此 Span 的 context。
	Context() context.Context
}

// SpanEndOption Span 结束选项。
type SpanEndOption func(*SpanEndConfig)

// SpanEndConfig Span 结束配置。
type SpanEndConfig struct {
	// EndTime Span 结束时间（默认为当前时间）
	EndTime time.Time
}

// SpanStatusCode Span 状态码。
type SpanStatusCode int

const (
	SpanStatusCodeUnset SpanStatusCode = iota
	SpanStatusCodeOk
	SpanStatusCodeError
)

// TracingConfig 追踪配置。
type TracingConfig struct {
	// Enabled 是否启用追踪
	Enabled bool

	// ServiceName 服务名称
	ServiceName string

	// Environment 环境标识
	Environment string

	// Version 版本号
	Version string

	// ExporterType 导出器类型：noop/jaeger/zipkin/otlp/stdout
	ExporterType string

	// ExporterEndpoint 导出器地址
	ExporterEndpoint string

	// SamplingRate 采样率（0-1）
	SamplingRate float64

	// Propagators 传播器：tracecontext/baggage/b3
	Propagators []string

	// ResourceAttributes 资源属性
	ResourceAttributes map[string]string

	// BatchTimeout 批量导出超时（秒）
	BatchTimeout int

	// MaxQueueSize 最大队列大小
	MaxQueueSize int

	// MaxExportBatchSize 最大导出批次大小
	MaxExportBatchSize int
}

// ObservabilityConfig 观测配置。
type ObservabilityConfig struct {
	// MetricsEnabled 是否启用指标收集
	MetricsEnabled bool
	// TracingEnabled 是否启用链路追踪
	TracingEnabled bool
	// ErrorReportingEnabled 是否启用错误上报
	ErrorReportingEnabled bool
	// ServiceName 服务名称
	ServiceName string
	// Environment 环境标识
	Environment string
	// Version 版本号
	Version string
	// SamplingRate 采样率（0-1）
	SamplingRate float64
}

// RequestContext 请求上下文，用于统一观测。
//
// 中文说明：
// - 包含请求的完整观测信息；
// - 用于在中间件、业务代码中传递；
// - 自动关联 TraceID、RequestID 等。
type RequestContext struct {
	// TraceID 链路追踪 ID
	TraceID string
	// RequestID 请求 ID
	RequestID string
	// Span 当前 Span
	Span Span
	// StartTime 请求开始时间
	StartTime time.Time
	// Labels 默认标签
	Labels map[string]string
}