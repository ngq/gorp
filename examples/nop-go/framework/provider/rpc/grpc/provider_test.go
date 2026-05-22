// Package grpc_test provides unit tests for gRPC server transport and interceptor registration.
//
// 适用场景：
// - 验证 gRPC server 的启动、interceptor 注册和 metadata 处理。
// - 确保 discovery、observability、resilience 等中间件的集成正确。
package grpc

import (
	"context"
	"net"
	"testing"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testCircuitBreaker struct {
	resource string
	called   bool
}

func (cb *testCircuitBreaker) Allow(ctx context.Context, resource string) error   { return nil }
func (cb *testCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {}
func (cb *testCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {
}
func (cb *testCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	cb.resource = resource
	cb.called = true
	return fn()
}
func (cb *testCircuitBreaker) State(ctx context.Context, resource string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
}

type testServiceTokenIssuer struct{}

func (testServiceTokenIssuer) GenerateToken(ctx context.Context, targetService string) (string, error) {
	return "svc-token", nil
}

type testServiceAuthenticator struct{}

func (testServiceAuthenticator) Authenticate(ctx context.Context) (*securitycontract.ServiceIdentity, error) {
	return &securitycontract.ServiceIdentity{ServiceName: "caller"}, nil
}

type tracerInjector struct{}

func (tracerInjector) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	return ctx, noopSpan{}
}
func (tracerInjector) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	return noopSpan{}
}
func (tracerInjector) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	carrier.Set("x-test-trace", "trace-from-tracer")
	return nil
}
func (tracerInjector) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

type noopSpan struct{}

func (noopSpan) End(options ...observabilitycontract.SpanEndOption)                      {}
func (noopSpan) AddEvent(name string, attributes map[string]interface{})                 {}
func (noopSpan) SetTag(key string, value interface{})                                    {}
func (noopSpan) SetAttributes(attributes map[string]interface{})                         {}
func (noopSpan) SetError(err error)                                                      {}
func (noopSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {}
func (noopSpan) SpanContext() observabilitycontract.SpanContext {
	return observabilitycontract.SpanContext{}
}
func (noopSpan) IsRecording() bool        { return false }
func (noopSpan) Context() context.Context { return context.Background() }

type selectorRegistry struct {
	instances []transportcontract.ServiceInstance
}

func (s selectorRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	return nil
}
func (s selectorRegistry) Deregister(ctx context.Context, name, addr string) error { return nil }
func (s selectorRegistry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	return s.instances, nil
}
func (s selectorRegistry) Close() error { return nil }

type captureSelector struct {
	doneInfo discoverycontract.DoneInfo
	called   bool
}

func (s *captureSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (transportcontract.ServiceInstance, discoverycontract.DoneFunc, error) {
	s.called = true
	return instances[0], func(ctx context.Context, info discoverycontract.DoneInfo) {
		s.doneInfo = info
	}, nil
}

// TestCircuitBreakerUnaryInterceptorUsesNormalizedResource 验证熔断拦截器使用标准化资源名。
//
// 中文说明：
// - 资源名格式为 "rpc.grpc.{service}.{package}.{service}.{method}"。
func TestCircuitBreakerUnaryInterceptorUsesNormalizedResource(t *testing.T) {
	cb := &testCircuitBreaker{}
	client := NewClient(&transportcontract.RPCConfig{Mode: "grpc", TimeoutMS: 1000}, nil, nil, nil, nil, nil, cb, nil)

	interceptor := client.circuitBreakerUnaryInterceptor("user-service")
	err := interceptor(
		context.Background(),
		"/user.v1.UserService/GetUser",
		nil, nil, nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *gogrpc.ClientConn, opts ...gogrpc.CallOption) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
	if !cb.called {
		t.Fatal("expected circuit breaker to wrap grpc invoke")
	}
	if cb.resource != "rpc.grpc.user-service.user.v1.UserService.GetUser" {
		t.Fatalf("unexpected resource %q", cb.resource)
	}
}

// TestServiceAuthUnaryClientInterceptorInjectsToken 验证客户端拦截器正确注入 service token。
//
// 中文说明：
// - serviceAuthUnaryClientInterceptor 在 outgoing metadata 中注入 x-service-token。
func TestServiceAuthUnaryClientInterceptorInjectsToken(t *testing.T) {
	interceptor := serviceAuthUnaryClientInterceptor(testServiceTokenIssuer{}, "billing-service")
	err := interceptor(
		context.Background(),
		"/billing.v1.BillingService/Pay",
		nil, nil, nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *gogrpc.ClientConn, opts ...gogrpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}
			if got := md.Get("x-service-token"); len(got) != 1 || got[0] != "svc-token" {
				t.Fatalf("expected service token in metadata, got %v", got)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
}

// TestServiceAuthUnaryServerInterceptorInjectsIdentity 验证服务端拦截器正确注入 service identity。
//
// 中文说明：
// - serviceAuthUnaryServerInterceptor 从 token 解析出 identity 并注入 context。
func TestServiceAuthUnaryServerInterceptorInjectsIdentity(t *testing.T) {
	interceptor := serviceAuthUnaryServerInterceptor(testServiceAuthenticator{})
	_, err := interceptor(
		metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-token", "svc-token")),
		nil,
		&gogrpc.UnaryServerInfo{FullMethod: "/billing.v1.BillingService/Pay"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, ok := securitycontract.FromServiceIdentityContext(ctx)
			if !ok || identity == nil || identity.ServiceName != "caller" {
				t.Fatalf("expected service identity in context, got %+v", identity)
			}
			if token, _ := ctx.Value("x-service-token").(string); token != "svc-token" {
				t.Fatalf("expected service token in context, got %q", token)
			}
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
}

// TestGRPCMetadataCarrierSupportsTracingInjection 验证 gRPC metadata carrier 支持 tracing injection。
//
// 中文说明：
// - tracerInjector 通过 gRPCMetadataCarrier 将 trace 信息注入 metadata。
func TestGRPCMetadataCarrierSupportsTracingInjection(t *testing.T) {
	md := metadata.New(nil)
	carrier := newGRPCMetadataCarrier(md)
	if err := (tracerInjector{}).Inject(context.Background(), carrier); err != nil {
		t.Fatalf("inject returned error: %v", err)
	}
	if values := carrier.Values("x-test-trace"); len(values) != 1 || values[0] != "trace-from-tracer" {
		t.Fatalf("expected tracer metadata, got %v", values)
	}
}

// TestClientCallUsesSelectorAndResourceAwareRetryOnFailure 验证客户端调用使用 selector 做服务发现。
//
// 中文说明：
// - client.Call 通过 selector 选取目标实例进行调用。
func TestClientCallUsesSelectorAndResourceAwareRetryOnFailure(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	server := gogrpc.NewServer()
	defer server.Stop()
	go func() {
		_ = server.Serve(lis)
	}()

	selector := &captureSelector{}
	client := NewClient(
		&transportcontract.RPCConfig{Mode: "grpc", TimeoutMS: 100},
		selectorRegistry{instances: []transportcontract.ServiceInstance{{ID: "1", Address: lis.Addr().String(), Healthy: true}}},
		selector,
		nil, nil, nil, nil, nil,
	)

	err = client.Call(context.Background(), "user-service", "/user.v1.UserService/GetUser", nil, nil)
	if err == nil {
		t.Fatal("expected grpc call to fail for unregistered method")
	}
	if !selector.called {
		t.Fatal("expected selector to participate in target selection")
	}
}
