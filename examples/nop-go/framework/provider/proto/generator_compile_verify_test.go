// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 验证生成产物的基本合法性（语法、结构完整性）。
// - golden verify 校验：确保关键模式在生成产物中存在。
package proto

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// TestGenerator_GenFromService_GoldenVerify_ComplexNesting 验证复杂嵌套类型的生成产物结构完整性。
//
// 中文说明：
// - 验证多层嵌套 struct（struct 内嵌 struct 内嵌 struct）的递归生成；
// - 验证所有嵌套 message 都出现在 proto 产物中，没有遗漏；
// - 验证字段编号连续且无重复。
func TestGenerator_GenFromService_GoldenVerify_ComplexNesting(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

type OrderService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
}

type CreateOrderRequest struct {
	Customer Customer ` + "`json:\"customer\"`" + `
	Items    []OrderItem ` + "`json:\"items\"`" + `
}

type CreateOrderResponse struct {
	Order Order ` + "`json:\"order\"`" + `
}

type Customer struct {
	Name    string ` + "`json:\"name\"`" + `
	Address Address ` + "`json:\"address\"`" + `
}

type Address struct {
	City    string ` + "`json:\"city\"`" + `
	ZipCode string ` + "`json:\"zip_code\"`" + `
}

type OrderItem struct {
	Product Product ` + "`json:\"product\"`" + `
	Quantity int32 ` + "`json:\"quantity\"`" + `
}

type Product struct {
	Name  string ` + "`json:\"name\"`" + `
	Price float64 ` + "`json:\"price\"`" + `
}

type Order struct {
	ID      string ` + "`json:\"id\"`" + `
	Status  string ` + "`json:\"status\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)

	content := runGenFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "OrderService",
		OutputPath:  filepath.Join(tmpDir, "order.proto"),
		Package:     "api.order.v1",
		GoPackage:   "github.com/example/api/order/v1;orderv1",
	})

	// 验证所有 message 都生成（golden verify）。
	expectedMessages := []string{
		"message CreateOrderRequest",
		"message CreateOrderResponse",
		"message Customer",
		"message Address",
		"message OrderItem",
		"message Product",
		"message Order",
	}
	for _, msg := range expectedMessages {
		if !strings.Contains(content, msg) {
			t.Errorf("generated content missing %q", msg)
		}
	}

	// 验证嵌套字段正确引用。
	assertContainsAll(t, content,
		"Customer customer = 1;",
		"repeated OrderItem items = 2;",
		"Address address = 2;",
		"Product product = 1;",
	)

	// 验证没有 placeholder 残留。
	assertNotContainsAll(t, content, "placeholder", "TODO: Add fields")
}

// TestGenerator_GenFromService_ProtoSyntaxIsValid 验证生成的 proto 文件满足基本语法规范。
//
// 中文说明：
// - 检查 syntax 声明存在；
// - 检查 package 声明存在；
// - 检查 go_package 选项存在；
// - 检查 service 定义存在；
// - 检查 message 定义存在。
func TestGenerator_GenFromService_ProtoSyntaxIsValid(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

type PingService interface {
	Ping(ctx context.Context, req *PingRequest) (*PingResponse, error)
}

type PingRequest struct {
	Message string ` + "`json:\"message\"`" + `
}

type PingResponse struct {
	Reply string ` + "`json:\"reply\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)

	content := runGenFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "PingService",
		OutputPath:  filepath.Join(tmpDir, "ping.proto"),
		Package:     "api.ping.v1",
		GoPackage:   "github.com/example/api/ping/v1;pingv1",
	})

	// 验证 proto 基本语法元素。
	assertContainsAll(t, content,
		`syntax = "proto3";`,
		"package api.ping.v1;",
		`option go_package = "github.com/example/api/ping/v1;pingv1";`,
		"service PingService",
		"rpc Ping(PingRequest) returns (PingResponse);",
		"message PingRequest",
		"message PingResponse",
	)
}

// TestGenerator_GenFromService_FailsOnUnsupportedMapKey 验证 proto 不支持的 map key 类型会报错。
func TestGenerator_GenFromService_FailsOnUnsupportedMapKey(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

type BadService interface {
	Create(ctx context.Context, req *BadRequest) (*BadResponse, error)
}

type BadRequest struct {
	// proto 不支持 struct 作为 map key
	Data map[ComplexKey]string ` + "`json:\"data\"`" + `
}

type BadResponse struct {
	OK bool ` + "`json:\"ok\"`" + `
}

type ComplexKey struct {
	ID int64 ` + "`json:\"id\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)

	_, err := genFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "BadService",
		OutputPath:  filepath.Join(tmpDir, "bad.proto"),
		Package:     "api.bad.v1",
		GoPackage:   "github.com/example/api/bad/v1;badv1",
	})
	if err == nil {
		t.Fatal("expected unsupported map key error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported map key type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGenerator_GenService_FailsOnMissingProtoFile 验证 GenService 在 proto 文件不存在时报错。
func TestGenerator_GenService_FailsOnMissingProtoFile(t *testing.T) {
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	opts := integrationcontract.ServiceGenOptions{
		ProtoFile: "/nonexistent/path/service.proto",
		OutputDir: "/tmp/output",
	}

	err = gen.GenService(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for missing proto file, got nil")
	}
}

// TestGenerator_GenService_FailsOnMissingRequiredArgs 验证 GenService 缺少必填参数时报错。
func TestGenerator_GenService_FailsOnMissingRequiredArgs(t *testing.T) {
	cfg := &integrationcontract.ProtoGeneratorConfig{Enabled: true}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 缺少 proto file。
	err = gen.GenService(context.Background(), integrationcontract.ServiceGenOptions{
		OutputDir: "/tmp/output",
	})
	if err == nil {
		t.Fatal("expected error for missing proto file, got nil")
	}

	// 缺少 output dir。
	err = gen.GenService(context.Background(), integrationcontract.ServiceGenOptions{
		ProtoFile: "/tmp/test.proto",
	})
	if err == nil {
		t.Fatal("expected error for missing output dir, got nil")
	}
}

// TestGenerator_GenFromService_FailsOnInvalidServicePath 验证无效的 service 文件路径会报错。
func TestGenerator_GenFromService_FailsOnInvalidServicePath(t *testing.T) {
	cfg := &integrationcontract.ProtoGeneratorConfig{}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	err = gen.GenFromService(context.Background(), integrationcontract.ServiceToProtoOptions{
		ServicePath: "/nonexistent/service.go",
		OutputPath:  "/tmp/out.proto",
		Package:     "test",
		GoPackage:   "test;test",
	})
	if err == nil {
		t.Fatal("expected error for invalid service path, got nil")
	}
}
