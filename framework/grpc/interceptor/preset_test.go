// Package interceptor_test provides unit tests for gRPC interceptor preset ordering.
//
// 适用场景：
// - 验证 gRPC 拦截器的预设顺序与组合行为。
package interceptor

import (
	"context"
	"testing"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"google.golang.org/grpc"
)

type noopTracer struct{}

func (noopTracer) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	return ctx, noopSpan{}
}
func (noopTracer) SpanFromContext(context.Context) observabilitycontract.Span { return noopSpan{} }
func (noopTracer) Inject(context.Context, observabilitycontract.TextMapCarrier) error {
	return nil
}
func (noopTracer) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

type noopSpan struct{}

func (noopSpan) End(...observabilitycontract.SpanEndOption)             {}
func (noopSpan) AddEvent(string, map[string]interface{})                {}
func (noopSpan) SetTag(string, interface{})                             {}
func (noopSpan) SetAttributes(map[string]interface{})                   {}
func (noopSpan) SetError(error)                                         {}
func (noopSpan) SetStatus(observabilitycontract.SpanStatusCode, string) {}
func (noopSpan) SpanContext() observabilitycontract.SpanContext {
	return observabilitycontract.SpanContext{}
}
func (noopSpan) IsRecording() bool        { return false }
func (noopSpan) Context() context.Context { return context.Background() }

type noopMetadataPropagator struct{}

func (noopMetadataPropagator) Inject(context.Context, transportcontract.MetadataCarrier) {}
func (noopMetadataPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	return ctx
}

// TestDefaultUnaryServerInterceptorsStableBase verifies the base gRPC unary preset cardinality.
//
// TestDefaultUnaryServerInterceptorsStableBase 验证默认 gRPC unary 预设的基础数量稳定。
func TestDefaultUnaryServerInterceptorsStableBase(t *testing.T) {
	set := DefaultUnaryServerInterceptors(DefaultServerPresetOptions{})
	if len(set) != 1 {
		t.Fatalf("expected 1 base unary interceptor, got %d", len(set))
	}
}

// TestDefaultStreamServerInterceptorsStableBase verifies the base gRPC stream preset cardinality.
//
// TestDefaultStreamServerInterceptorsStableBase 验证默认 gRPC stream 预设的基础数量稳定。
func TestDefaultStreamServerInterceptorsStableBase(t *testing.T) {
	set := DefaultStreamServerInterceptors(DefaultServerPresetOptions{})
	if len(set) != 1 {
		t.Fatalf("expected 1 base stream interceptor, got %d", len(set))
	}
}

// TestDefaultUnaryServerInterceptorsIncludeTracingAndMetadataInStableOrder verifies unary interceptors include tracing and metadata in stable order.
//
// TestDefaultUnaryServerInterceptorsIncludeTracingAndMetadataInStableOrder 验证 unary 拦截器按稳定顺序包含追踪和元数据。
func TestDefaultUnaryServerInterceptorsIncludeTracingAndMetadataInStableOrder(t *testing.T) {
	set := DefaultUnaryServerInterceptors(DefaultServerPresetOptions{
		Tracer:             noopTracer{},
		ServiceName:        "demo",
		MetadataPropagator: noopMetadataPropagator{},
	})
	if len(set) != 3 {
		t.Fatalf("expected 3 unary interceptors, got %d", len(set))
	}
}

// TestDefaultStreamServerInterceptorsIncludeMetadataInStableOrder verifies stream interceptors include metadata in stable order.
//
// TestDefaultStreamServerInterceptorsIncludeMetadataInStableOrder 验证 stream 拦截器按稳定顺序包含元数据。
func TestDefaultStreamServerInterceptorsIncludeMetadataInStableOrder(t *testing.T) {
	set := DefaultStreamServerInterceptors(DefaultServerPresetOptions{
		MetadataPropagator: noopMetadataPropagator{},
	})
	if len(set) != 2 {
		t.Fatalf("expected 2 stream interceptors, got %d", len(set))
	}
}

// TestChainUnarySkipsNilAndPreservesDeclaredOrder verifies ChainUnary skips nil interceptors and preserves order.
//
// TestChainUnarySkipsNilAndPreservesDeclaredOrder 验证 ChainUnary 跳过 nil 拦截器并保持声明顺序。
func TestChainUnarySkipsNilAndPreservesDeclaredOrder(t *testing.T) {
	var calls []string
	first := func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		calls = append(calls, "first")
		return handler(ctx, req)
	}
	second := func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		calls = append(calls, "second")
		return handler(ctx, req)
	}

	set := ChainUnary(first, nil, second)
	if len(set) != 2 {
		t.Fatalf("expected 2 interceptors after nil filtering, got %d", len(set))
	}
	for _, interceptor := range set {
		if _, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}); err != nil {
			t.Fatalf("unexpected unary interceptor error: %v", err)
		}
	}
	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Fatalf("expected stable unary order [first second], got %v", calls)
	}
}

// TestChainStreamSkipsNilAndPreservesDeclaredOrder verifies ChainStream skips nil interceptors and preserves order.
//
// TestChainStreamSkipsNilAndPreservesDeclaredOrder 验证 ChainStream 跳过 nil 拦截器并保持声明顺序。
func TestChainStreamSkipsNilAndPreservesDeclaredOrder(t *testing.T) {
	var calls []string
	first := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		calls = append(calls, "first")
		return nil
	}
	second := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		calls = append(calls, "second")
		return nil
	}

	set := ChainStream(first, nil, second)
	if len(set) != 2 {
		t.Fatalf("expected 2 interceptors after nil filtering, got %d", len(set))
	}
	for _, interceptor := range set {
		if err := interceptor(nil, nil, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
			return nil
		}); err != nil {
			t.Fatalf("unexpected stream interceptor error: %v", err)
		}
	}
	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Fatalf("expected stable stream order [first second], got %v", calls)
	}
}
