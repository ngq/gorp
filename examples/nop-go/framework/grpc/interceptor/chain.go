// Application scenarios:
// - Provide one provider-neutral home for the framework's gRPC governance interceptor presets.
// - Keep interceptor ordering explicit while reusing concrete tracing/metadata/runtime helpers.
// - Let bootstrap and top-level helpers assemble gRPC governance without reaching into provider internals directly.
//
// 适用场景：
// - 为框架的 gRPC 治理拦截器预设提供统一的 provider-neutral 主线目录。
// - 在复用 tracing、metadata、runtime helper 的同时，保持拦截器顺序显式可控。
// - 让 bootstrap 和顶层 helper 在不直接下沉 provider 内部细节的前提下装配 gRPC 治理能力。
package interceptor

import "google.golang.org/grpc"

// ChainUnary returns the unary interceptors unchanged in the declared order.
//
// ChainUnary 按声明顺序返回 unary 拦截器集合。
func ChainUnary(interceptors ...grpc.UnaryServerInterceptor) []grpc.UnaryServerInterceptor {
	result := make([]grpc.UnaryServerInterceptor, 0, len(interceptors))
	for _, interceptor := range interceptors {
		if interceptor != nil {
			result = append(result, interceptor)
		}
	}
	return result
}

// ChainStream returns the stream interceptors unchanged in the declared order.
//
// ChainStream 按声明顺序返回 stream 拦截器集合。
func ChainStream(interceptors ...grpc.StreamServerInterceptor) []grpc.StreamServerInterceptor {
	result := make([]grpc.StreamServerInterceptor, 0, len(interceptors))
	for _, interceptor := range interceptors {
		if interceptor != nil {
			result = append(result, interceptor)
		}
	}
	return result
}
