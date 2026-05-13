// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 验证 proto-first 工作流中的 gen-client 命令级测试。
// - 验证类型化 RPC 客户端 wrapper 生成。
// - 验证边界失败场景（proto 文件不存在、无服务定义等）。
package proto

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// TestGenerator_GenClient 测试从 proto 文件生成类型化 RPC 客户端 wrapper。
func TestGenerator_GenClient(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试 proto 文件。
	protoContent := `syntax = "proto3";

package order.v1;

option go_package = "example.com/myproject/proto/order/v1;orderv1";

service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse) {}
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse) {}
}

message CreateOrderRequest {
  string item = 1;
}

message CreateOrderResponse {
  int64 id = 1;
}

message GetOrderRequest {
  int64 id = 1;
}

message GetOrderResponse {
  string item = 1;
}
`
	protoFile := filepath.Join(tmpDir, "order.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("failed to write proto file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "client", "order_client.go")
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	opts := integrationcontract.ClientGenOptions{
		ProtoFile:   protoFile,
		OutputFile:  outputFile,
		PackageName: "client",
	}

	if err := gen.GenClient(context.Background(), opts); err != nil {
		t.Fatalf("GenClient failed: %v", err)
	}

	// 验证文件生成。
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("client file not generated: %v", err)
	}

	clientText := string(content)
	// 验证客户端结构定义。
	if !strings.Contains(clientText, "type OrderClient struct") {
		t.Error("client missing OrderClient struct")
	}
	// 验证方法生成。
	if !strings.Contains(clientText, "func (c *OrderClient) CreateOrder") {
		t.Error("client missing CreateOrder method")
	}
	if !strings.Contains(clientText, "func (c *OrderClient) GetOrder") {
		t.Error("client missing GetOrder method")
	}
	// 验证构造函数。
	if !strings.Contains(clientText, "func NewOrderClient") {
		t.Error("client missing NewOrderClient constructor")
	}
	// 验证使用了 RPCClient。
	if !strings.Contains(clientText, "transportcontract.RPCClient") {
		t.Error("client missing RPCClient import")
	}
}

// TestGenerator_GenClient_FilterByServiceName 测试指定服务名生成客户端。
func TestGenerator_GenClient_FilterByServiceName(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建包含多个服务的 proto 文件。
	protoContent := `syntax = "proto3";
package multi.v1;
service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse) {}
}
service PaymentService {
  rpc Pay(PayRequest) returns (PayResponse) {}
}
message CreateOrderRequest {}
message CreateOrderResponse {}
message PayRequest {}
message PayResponse {}
`
	protoFile := filepath.Join(tmpDir, "multi.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("failed to write proto file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "client", "order_client.go")
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	opts := integrationcontract.ClientGenOptions{
		ProtoFile:   protoFile,
		OutputFile:  outputFile,
		PackageName: "client",
		ServiceName: "OrderService",
	}

	if err := gen.GenClient(context.Background(), opts); err != nil {
		t.Fatalf("GenClient failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("client file not generated: %v", err)
	}

	clientText := string(content)
	if strings.Contains(clientText, "PaymentClient") {
		t.Error("client should not contain PaymentClient when ServiceName is specified")
	}
	if !strings.Contains(clientText, "OrderClient") {
		t.Error("client missing OrderClient")
	}
}

// TestGenerator_GenClient_FailsOnMissingProtoFile 测试 proto 文件不存在时报错。
func TestGenerator_GenClient_FailsOnMissingProtoFile(t *testing.T) {
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	opts := integrationcontract.ClientGenOptions{
		ProtoFile:  "/nonexistent/path/service.proto",
		OutputFile: "/tmp/out.go",
	}

	err = gen.GenClient(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for missing proto file, got nil")
	}
}

// TestGenerator_GenClient_FailsOnNoServicesInProto 测试 proto 文件中没有 service 定义时报错。
func TestGenerator_GenClient_FailsOnNoServicesInProto(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax = "proto3";
package empty.v1;
message EmptyMessage {}
`
	protoFile := filepath.Join(tmpDir, "empty.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("failed to write proto file: %v", err)
	}

	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "client", "empty_client.go")
	opts := integrationcontract.ClientGenOptions{
		ProtoFile:   protoFile,
		OutputFile:  outputFile,
		PackageName: "client",
	}

	err = gen.GenClient(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for no services in proto, got nil")
	}
	if !strings.Contains(err.Error(), "no services found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGenerator_GenClient_FailsOnMissingRequiredArgs 测试缺少必填参数时报错。
func TestGenerator_GenClient_FailsOnMissingRequiredArgs(t *testing.T) {
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 缺少 proto file。
	err = gen.GenClient(context.Background(), integrationcontract.ClientGenOptions{
		OutputFile: "/tmp/out.go",
	})
	if err == nil {
		t.Fatal("expected error for missing proto file, got nil")
	}

	// 缺少 output file。
	err = gen.GenClient(context.Background(), integrationcontract.ClientGenOptions{
		ProtoFile: "/tmp/test.proto",
	})
	if err == nil {
		t.Fatal("expected error for missing output file, got nil")
	}
}
