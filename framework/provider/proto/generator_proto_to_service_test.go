// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 验证 Proto-first 工作流：Proto 文件 → Go pb.go / gRPC service 代码生成。
package proto

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// TestGenerator_GenService 测试从 proto 文件生成 service skeleton。
func TestGenerator_GenService(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试 proto 文件。
	protoContent := `syntax = "proto3";

package user.v1;

option go_package = "example.com/myproject/proto/user/v1;userv1";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {}
}

message GetUserRequest {
  int64 id = 1;
}

message GetUserResponse {
  int64 id = 1;
  string username = 2;
}

message CreateUserRequest {
  string username = 1;
  string email = 2;
}

message CreateUserResponse {
  int64 id = 1;
}
`
	protoFile := filepath.Join(tmpDir, "user.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("failed to write proto file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	opts := integrationcontract.ServiceGenOptions{
		ProtoFile:      protoFile,
		OutputDir:      outputDir,
		Module:         "example.com/myproject",
		IncludeHTTP:    true,
		IncludeGRPC:    true,
		RegisterRoutes: true,
	}

	if err := gen.GenService(context.Background(), opts); err != nil {
		t.Fatalf("GenService failed: %v", err)
	}

	// 验证 HTTP handler 生成。
	handlerFile := filepath.Join(outputDir, "handler", "userservice.go")
	content, err := os.ReadFile(handlerFile)
	if err != nil {
		t.Fatalf("handler file not generated: %v", err)
	}
	handlerText := string(content)
	if !strings.Contains(handlerText, "type UserServiceHandler struct") {
		t.Error("handler missing UserServiceHandler struct")
	}
	if !strings.Contains(handlerText, "func (h *UserServiceHandler) GetUser") {
		t.Error("handler missing GetUser method")
	}
	if !strings.Contains(handlerText, "func (h *UserServiceHandler) CreateUser") {
		t.Error("handler missing CreateUser method")
	}

	// 验证 gRPC service 生成。
	serviceFile := filepath.Join(outputDir, "service", "userservice.go")
	content, err = os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("service file not generated: %v", err)
	}
	serviceText := string(content)
	if !strings.Contains(serviceText, "type UserServiceService struct") {
		t.Error("service missing UserServiceService struct")
	}
	if !strings.Contains(serviceText, "func (s *UserServiceService) GetUser") {
		t.Error("service missing GetUser method")
	}

	// 验证路由注册生成。
	routesFile := filepath.Join(outputDir, "routes", "userservice.go")
	content, err = os.ReadFile(routesFile)
	if err != nil {
		t.Fatalf("routes file not generated: %v", err)
	}
	routesText := string(content)
	if !strings.Contains(routesText, "func RegisterUserServiceRoutes") {
		t.Error("routes missing RegisterUserServiceRoutes function")
	}
}

// TestGenerator_GenService_SpecifyServiceName 测试指定服务名生成。
func TestGenerator_GenService_SpecifyServiceName(t *testing.T) {
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

	outputDir := filepath.Join(tmpDir, "output")
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 只生成 OrderService。
	opts := integrationcontract.ServiceGenOptions{
		ProtoFile:      protoFile,
		OutputDir:      outputDir,
		Module:         "example.com/myproject",
		ServiceName:    "OrderService",
		IncludeHTTP:    true,
		IncludeGRPC:    false,
		RegisterRoutes: false,
	}

	if err := gen.GenService(context.Background(), opts); err != nil {
		t.Fatalf("GenService failed: %v", err)
	}

	handlerFile := filepath.Join(outputDir, "handler", "orderservice.go")
	content, err := os.ReadFile(handlerFile)
	if err != nil {
		t.Fatalf("handler file not generated: %v", err)
	}
	handlerText := string(content)
	if strings.Contains(handlerText, "PaymentService") {
		t.Error("handler should not contain PaymentService when ServiceName is specified")
	}
	if !strings.Contains(handlerText, "OrderServiceHandler") {
		t.Error("handler missing OrderServiceHandler")
	}
}
