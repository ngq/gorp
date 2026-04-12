package proto

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

// TestGenerator_GenFromService 测试从 Service 生成 Proto。
//
// 中文说明：
// - 验证 Service-first 工作流。
func TestGenerator_GenFromService(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建测试 Service 文件
	serviceContent := `package service

import "context"

// UserService 用户服务接口。
type UserService interface {
	// GetUser 获取用户。
	GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)

	// CreateUser 创建用户。
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
}

type GetUserRequest struct {
	UserID int64 ` + "`json:\"user_id\"`" + `
}

type GetUserResponse struct {
	ID       int64  ` + "`json:\"id\"`" + `
	Username string ` + "`json:\"username\"`" + `
}

type CreateUserRequest struct {
	Username string ` + "`json:\"username\"`" + `
	Email    string ` + "`json:\"email\"`" + `
}

type CreateUserResponse struct {
	ID int64 ` + "`json:\"id\"`" + `
}
`

	servicePath := filepath.Join(tmpDir, "user_service.go")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("failed to write service file: %v", err)
	}

	// 创建 Generator
	cfg := &contract.ProtoGeneratorConfig{
		DefaultProtoDir: tmpDir,
	}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 生成 Proto
	outputPath := filepath.Join(tmpDir, "user.proto")
	opts := contract.ServiceToProtoOptions{
		ServicePath:  servicePath,
		ServiceName:  "UserService",
		OutputPath:   outputPath,
		Package:      "api.user.v1",
		GoPackage:    "github.com/example/api/user/v1;v1",
		IncludeHTTP:  true,
	}

	err = gen.GenFromService(context.Background(), opts)
	if err != nil {
		t.Errorf("failed to generate proto: %v", err)
	}

	// 验证文件生成
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("proto file was not created")
	}

	// 读取生成的内容
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read proto file: %v", err)
	}

	// 验证基本结构
	contentStr := string(content)
	if !contains(contentStr, "syntax = \"proto3\"") {
		t.Error("proto file missing syntax declaration")
	}
	if !contains(contentStr, "service UserService") {
		t.Error("proto file missing service declaration")
	}
	if !contains(contentStr, "rpc GetUser") {
		t.Error("proto file missing GetUser method")
	}
	if !contains(contentStr, "rpc CreateUser") {
		t.Error("proto file missing CreateUser method")
	}
}

// TestGenerator_ExtractServices 测试服务提取。
//
// 中文说明：
// - 验证从 Go AST 正确提取服务定义。
func TestGenerator_ExtractServices(t *testing.T) {
	// 创建测试 Service 文件内容
	serviceContent := `package service

import "context"

type OrderService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error)
	ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error)
}
`

	tmpDir := t.TempDir()
	servicePath := filepath.Join(tmpDir, "order_service.go")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("failed to write service file: %v", err)
	}

	cfg := &contract.ProtoGeneratorConfig{}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 解析文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, servicePath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse service file: %v", err)
	}

	// 提取服务
	services := gen.extractServices(f)

	if len(services) == 0 {
		t.Error("no services extracted")
	}

	svc := services[0]
	if svc.Name != "OrderService" {
		t.Errorf("expected OrderService, got: %s", svc.Name)
	}

	if len(svc.Methods) != 3 {
		t.Errorf("expected 3 methods, got: %d", len(svc.Methods))
	}

	// 验证方法名
	methodNames := make(map[string]bool)
	for _, m := range svc.Methods {
		methodNames[m.Name] = true
	}
	if !methodNames["CreateOrder"] || !methodNames["GetOrder"] || !methodNames["ListOrders"] {
		t.Error("missing expected methods")
	}
}

// TestGenerator_ScanProtoFiles 测试 Proto 文件扫描。
//
// 中文说明：
// - 验证扫描目录下的 proto 文件。
func TestGenerator_ScanProtoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个 proto 文件
	protoFiles := []string{
		"user.proto",
		"order.proto",
		"subdir/product.proto",
	}

	for _, f := range protoFiles {
		path := filepath.Join(tmpDir, f)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("syntax = \"proto3\";"), 0644); err != nil {
			t.Fatalf("failed to write proto file: %v", err)
		}
	}

	cfg := &contract.ProtoGeneratorConfig{
		DefaultProtoDir: tmpDir,
	}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 扫描文件
	files, err := gen.scanProtoFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan proto files: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 proto files, got: %d", len(files))
	}
}

// TestNewGenerator 测试创建 Generator。
//
// 中文说明：
// - 验证 Generator 正确初始化。
func TestNewGenerator(t *testing.T) {
	cfg := &contract.ProtoGeneratorConfig{
		DefaultProtoDir:      "./proto",
		IncludeHTTPAnnotation: true,
		ThirdPartyPaths:       []string{"/usr/local/include"},
	}

	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	if gen == nil {
		t.Error("generator is nil")
	}

	// 验证配置
	retrievedCfg := gen.GetConfig()
	if retrievedCfg.DefaultProtoDir != "./proto" {
		t.Errorf("unexpected DefaultProtoDir: %s", retrievedCfg.DefaultProtoDir)
	}
}

// contains 辅助函数，检查字符串包含。
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}