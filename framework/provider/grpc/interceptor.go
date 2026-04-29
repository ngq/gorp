package grpc

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// TraceIDKey 是 gRPC metadata 中 trace id 的 key
	TraceIDKey = "x-trace-id"
	// RequestIDKey 是 gRPC metadata 中 request id 的 key
	RequestIDKey = "x-request-id"
)

// contextKey 用于在 context 中存储 trace id
type contextKey string

const (
	traceIDContextKey   contextKey = "trace_id"
	requestIDContextKey contextKey = "request_id"
)

var (
	// grpcRequestsTotal gRPC 请求总数
	grpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_grpc_requests_total",
		Help: "Total number of gRPC requests.",
	}, []string{"method", "status"})

	// grpcRequestDuration gRPC 请求耗时
	grpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_grpc_request_duration_seconds",
		Help:    "gRPC request latency in seconds.",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "status"})

	// grpcRequestsInFlight 当前处理中的请求数
	grpcRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_grpc_requests_in_flight",
		Help: "Current number of gRPC requests being processed.",
	}, []string{"method"})
)

// UnaryServerInterceptor 创建一个 gRPC 服务端一元拦截器，用于从 metadata 中提取 trace id 并收集指标。
//
// 中文说明：
// - 从 incoming metadata 中读取 trace id 和 request id；
// - 如果不存在，则生成新的 id；
// - 将 id 存入 context，供后续处理使用；
// - 同时将 id 写入 outgoing metadata，实现跨服务透传；
// - 收集 Prometheus 指标：请求总数、耗时、当前处理数。
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 记录开始时间和当前处理请求数
		start := time.Now()
		method := info.FullMethod
		grpcRequestsInFlight.WithLabelValues(method).Inc()
		defer grpcRequestsInFlight.WithLabelValues(method).Dec()

		// 从 incoming metadata 提取 trace id
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if values := md.Get(TraceIDKey); len(values) > 0 {
				ctx = context.WithValue(ctx, traceIDContextKey, values[0])
			}
			if values := md.Get(RequestIDKey); len(values) > 0 {
				ctx = context.WithValue(ctx, requestIDContextKey, values[0])
			}
		}

		// 调用实际的处理函数
		resp, err := handler(ctx, req)

		// 记录指标
		st := status.Code(err).String()
		duration := time.Since(start).Seconds()
		grpcRequestsTotal.WithLabelValues(method, st).Inc()
		grpcRequestDuration.WithLabelValues(method, st).Observe(duration)

		return resp, err
	}
}

// StreamServerInterceptor 创建一个 gRPC 服务端流拦截器，用于从 metadata 中提取 trace id 并收集指标。
//
// 中文说明：
// - 功能与 UnaryServerInterceptor 相同，但用于流式 RPC。
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 记录开始时间和当前处理请求数
		start := time.Now()
		method := info.FullMethod
		grpcRequestsInFlight.WithLabelValues(method).Inc()
		defer grpcRequestsInFlight.WithLabelValues(method).Dec()

		ctx := ss.Context()

		// 从 incoming metadata 提取 trace id
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if values := md.Get(TraceIDKey); len(values) > 0 {
				ctx = context.WithValue(ctx, traceIDContextKey, values[0])
			}
			if values := md.Get(RequestIDKey); len(values) > 0 {
				ctx = context.WithValue(ctx, requestIDContextKey, values[0])
			}
		}

		// 包装 ServerStream 以传递更新后的 context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		err := handler(srv, wrapped)

		// 记录指标
		st := status.Code(err).String()
		duration := time.Since(start).Seconds()
		grpcRequestsTotal.WithLabelValues(method, st).Inc()
		grpcRequestDuration.WithLabelValues(method, st).Observe(duration)

		return err
	}
}

// wrappedServerStream 包装 grpc.ServerStream 以支持自定义 context。
//
// 中文说明：
// - 流式拦截器从 metadata 提取 trace/request id 后，需要把更新后的 context 继续传给业务 handler；
// - grpc.ServerStream 默认不会替换 Context，因此这里用一个轻量包装器把新 context 带下去。
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回包装后的上下文。
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetTraceID 从 gRPC context 中获取 Trace ID。
//
// 中文说明：
// - 供 gRPC 服务实现使用；
// - 如果不存在则返回空字符串。
func GetTraceID(ctx context.Context) string {
	if tid, ok := ctx.Value(traceIDContextKey).(string); ok {
		return tid
	}
	return ""
}

// GetRequestID 从 gRPC context 中获取 Request ID。
//
// 中文说明：
// - 供 gRPC 服务实现使用；
// - 如果不存在则返回空字符串。
func GetRequestID(ctx context.Context) string {
	if rid, ok := ctx.Value(requestIDContextKey).(string); ok {
		return rid
	}
	return ""
}

// UnaryClientInterceptor 创建一个 gRPC 客户端一元拦截器，用于向 metadata 中注入 trace id。
//
// 中文说明：
// - 用于 gRPC 客户端调用时自动注入 trace id；
// - 实现跨服务链路追踪；
// - 如果 context 中已有 trace id，则透传；否则生成新的。
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 尝试从 context 获取 trace id
		tid := GetTraceID(ctx)
		rid := GetRequestID(ctx)

		// 如果有 trace id，添加到 outgoing metadata
		if tid != "" || rid != "" {
			md := metadata.New(map[string]string{})
			if tid != "" {
				md.Set(TraceIDKey, tid)
			}
			if rid != "" {
				md.Set(RequestIDKey, rid)
			}
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor 创建一个 gRPC 客户端流拦截器，用于向 metadata 中注入 trace id。
//
// 中文说明：
// - 作用与 UnaryClientInterceptor 一致，但覆盖流式调用；
// - 这样无论是一元 RPC 还是流 RPC，都能复用同一套 trace/request id 透传约定。
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// 尝试从 context 获取 trace id
		tid := GetTraceID(ctx)
		rid := GetRequestID(ctx)

		// 如果有 trace id，添加到 outgoing metadata
		if tid != "" || rid != "" {
			md := metadata.New(map[string]string{})
			if tid != "" {
				md.Set(TraceIDKey, tid)
			}
			if rid != "" {
				md.Set(RequestIDKey, rid)
			}
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}
