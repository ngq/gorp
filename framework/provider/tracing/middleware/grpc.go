package middleware

import (
	"context"
	"strings"

	"github.com/ngq/gorp/framework/contract"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor 创建 gRPC 服务端一元拦截器。
//
// 中文说明：
// - 自动为每个 RPC 调用创建 Span；
// - 从 gRPC metadata 提取追踪上下文；
// - 记录方法名、错误码等信息；
// - 支持与 OpenTelemetry 集成。
func UnaryServerInterceptor(tracer contract.Tracer, serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 从 metadata 提取追踪上下文
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			carrier := &grpcMetadataCarrier{md: md}
			var err error
			ctx, err = tracer.Extract(ctx, carrier)
			if err != nil {
				// 提取失败，使用新的追踪上下文
			}
		}

		// 创建 Span
		spanName := info.FullMethod
		ctx, span := tracer.StartSpan(ctx, spanName,
			WithSpanKind(contract.SpanKindServer),
			WithAttributes(map[string]any{
				"rpc.system":       "grpc",
				"rpc.method":       info.FullMethod,
				"rpc.service":      extractServiceName(info.FullMethod),
				"service.name":     serviceName,
			}),
		)
		defer span.End()

		// 执行 handler
		resp, err := handler(ctx, req)

		// 记录错误
		if err != nil {
			span.SetError(err)
			span.SetStatus(contract.SpanStatusCodeError, err.Error())
		} else {
			span.SetStatus(contract.SpanStatusCodeOk, "")
		}

		// 注入追踪信息到响应 metadata
		if traceID := span.SpanContext().TraceID; traceID != "" {
			md := metadata.New(map[string]string{
				"x-trace-id": traceID,
			})
			grpc.SetTrailer(ctx, md)
		}

		return resp, err
	}
}

// UnaryClientInterceptor 创建 gRPC 客户端一元拦截器。
//
// 中文说明：
// - 自动为每个 RPC 调用创建客户端 Span；
// - 将追踪上下文注入到 gRPC metadata；
// - 支持跨服务追踪传播。
func UnaryClientInterceptor(tracer contract.Tracer, serviceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 创建 Span
		spanName := method
		ctx, span := tracer.StartSpan(ctx, spanName,
			WithSpanKind(contract.SpanKindClient),
			WithAttributes(map[string]any{
				"rpc.system":       "grpc",
				"rpc.method":       method,
				"rpc.service":      extractServiceName(method),
				"service.name":     serviceName,
			}),
		)
		defer span.End()

		// 注入追踪上下文到 metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := &grpcMetadataCarrier{md: md}
		if err := tracer.Inject(ctx, carrier); err == nil {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		// 执行调用
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 记录错误
		if err != nil {
			span.SetError(err)
			span.SetStatus(contract.SpanStatusCodeError, err.Error())
		} else {
			span.SetStatus(contract.SpanStatusCodeOk, "")
		}

		return err
	}
}

// grpcMetadataCarrier 实现 TextMapCarrier 接口。
type grpcMetadataCarrier struct {
	md metadata.MD
}

func (c *grpcMetadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *grpcMetadataCarrier) Set(key string, value string) {
	c.md.Set(key, value)
}

func (c *grpcMetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

// extractServiceName 从完整方法名提取服务名。
//
// 中文说明：
// - 输入：/package.service/method；
// - 输出：package.service。
func extractServiceName(fullMethod string) string {
	// 格式：/package.service/method
	if len(fullMethod) < 2 {
		return fullMethod
	}
	if fullMethod[0] != '/' {
		return fullMethod
	}

	// 查找第二个 /
	idx := strings.Index(fullMethod[1:], "/")
	if idx < 0 {
		return fullMethod[1:]
	}
	return fullMethod[1 : idx+1]
}