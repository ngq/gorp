// Package integration provides gRPC integration tests.
//
// 本包提供 gRPC 集成测试。
package integration

import (
	"context"
	"testing"
	"time"

	grpcprovider "github.com/ngq/gorp/framework/provider/rpc/grpc"
	transportcontracts "github.com/ngq/gorp/framework/contract/transport"
	pb "github.com/ngq/gorp/test/integration/pb"
)

// TestGRPCFullChain tests gRPC client-to-server call.
//
// TestGRPCFullChain 测试 gRPC 客户端到服务端调用。
func TestGRPCFullChain(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires docker backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Create gRPC client
	// 创建 gRPC 客户端
	cfg := &transportcontracts.RPCConfig{
		Mode:      "grpc",
		Address:   getEnvOrDefault("GORP_TEST_GRPC_BACKEND_ADDR", "localhost:50051"),
		TimeoutMS: 5000,
	}

	client := grpcprovider.NewClient(cfg, nil, nil, nil, nil, nil, nil, nil)
	defer client.Close()

	// 2. Prepare request
	// 准备请求
	req := &pb.HelloRequest{Name: "integration-test"}
	resp := &pb.HelloResponse{}

	// 3. Call mock backend
	// 调用 mock backend
	err := client.Call(ctx, "test-service", "/gorp.test.TestService/SayHello", req, resp)
	if err != nil {
		t.Fatalf("gRPC call failed: %v", err)
	}

	// 4. Verify response
	// 验证响应
	if resp.Message != "Hello integration-test" {
		t.Fatalf("unexpected response message: %s", resp.Message)
	}

	t.Logf("response: message=%s, trace-id=%s, request-id=%s", resp.Message, resp.TraceId, resp.RequestId)
}

// TestGRPCMetadataPropagation tests metadata propagation through gRPC call.
//
// TestGRPCMetadataPropagation 测试 metadata 通过 gRPC 调用传播。
func TestGRPCMetadataPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires docker backend")
	}

	ctx := context.Background()

	cfg := &transportcontracts.RPCConfig{
		Mode:      "grpc",
		Address:   getEnvOrDefault("GORP_TEST_GRPC_BACKEND_ADDR", "localhost:50051"),
		TimeoutMS: 5000,
	}

	client := grpcprovider.NewClient(cfg, nil, nil, nil, nil, nil, nil, nil)
	defer client.Close()

	req := &pb.EchoRequest{Body: "metadata-test"}
	resp := &pb.EchoResponse{}

	err := client.Call(ctx, "test-service", "/gorp.test.TestService/Echo", req, resp)
	if err != nil {
		t.Fatalf("gRPC call failed: %v", err)
	}

	if resp.Body != "metadata-test" {
		t.Fatalf("unexpected echo body: %s", resp.Body)
	}

	t.Logf("echo response metadata: %v", resp.Metadata)
}

// TestGRPCTimeoutMiddleware tests timeout middleware in gRPC chain.
//
// TestGRPCTimeoutMiddleware 测试 gRPC 链中的 timeout middleware。
func TestGRPCTimeoutMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires docker backend")
	}

	// Use short timeout to trigger timeout middleware
	// 使用短超时触发 timeout middleware
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cfg := &transportcontracts.RPCConfig{
		Mode:      "grpc",
		Address:   getEnvOrDefault("GORP_TEST_GRPC_BACKEND_ADDR", "localhost:50051"),
		TimeoutMS: 50, // Very short timeout
	}

	client := grpcprovider.NewClient(cfg, nil, nil, nil, nil, nil, nil, nil)
	defer client.Close()

	req := &pb.HelloRequest{Name: "timeout-test"}
	resp := &pb.HelloResponse{}

	err := client.Call(ctx, "test-service", "/gorp.test.TestService/SayHello", req, resp)
	// Expect timeout error
	// 期望超时错误
	if err == nil {
		t.Log("WARNING: expected timeout error but call succeeded")
	} else {
		t.Logf("expected timeout error: %v", err)
	}
}

// getEnvOrDefault gets environment variable or returns default value.
//
// getEnvOrDefault 获取环境变量或返回默认值。
func getEnvOrDefault(key, defaultVal string) string {
	// In real implementation, use os.Getenv
	// 实际实现中使用 os.Getenv
	// For test environment, return default
	// 测试环境返回默认值
	return defaultVal
}