// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 验证 Service-first 工作流：Go Service 接口 → Proto 文件的完整生成链路。
// - 验证同包多文件 DTO 自动发现、跨包 import-paths 递归闭包、嵌套类型推导。
// - 验证特殊类型（time.Time / time.Duration / []byte / any）的 proto 映射。
// - 验证 Proto-first 工作流：Proto 文件 → Go pb.go / gRPC service 代码生成。
// - 验证不可解析类型或不支持的 map 组合直接返回明确错误，而非产出 placeholder。
package proto

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
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
	cfg := &integrationcontract.ProtoGeneratorConfig{
		DefaultProtoDir: tmpDir,
	}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// 生成 Proto
	outputPath := filepath.Join(tmpDir, "user.proto")
	opts := integrationcontract.ServiceToProtoOptions{
		ServicePath: servicePath,
		ServiceName: "UserService",
		OutputPath:  outputPath,
		Package:     "api.user.v1",
		GoPackage:   "github.com/example/api/user/v1;v1",
		IncludeHTTP: true,
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
	if !strings.Contains(contentStr, "syntax = \"proto3\"") {
		t.Error("proto file missing syntax declaration")
	}
	if !strings.Contains(contentStr, "service UserService") {
		t.Error("proto file missing service declaration")
	}
	if !strings.Contains(contentStr, "rpc GetUser") {
		t.Error("proto file missing GetUser method")
	}
	if !strings.Contains(contentStr, "rpc CreateUser") {
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

	cfg := &integrationcontract.ProtoGeneratorConfig{}
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

	cfg := &integrationcontract.ProtoGeneratorConfig{
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
	cfg := &integrationcontract.ProtoGeneratorConfig{
		DefaultProtoDir:       "./proto",
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

// TestGenerator_GenFromService_MultiFilePackage 测试同包多文件 DTO 自动收集。
func TestGenerator_GenFromService_MultiFilePackage(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

type UserService interface {
	GetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error)
}
`
	requestContent := `package service

type GetProfileRequest struct {
	UserID int64 ` + "`json:\"user_id\"`" + `
}
`
	responseContent := `package service

type GetProfileResponse struct {
	Profile Profile ` + "`json:\"profile\"`" + `
}
`
	typesContent := `package service

// Profile 用户资料。
type Profile struct {
	Nickname string ` + "`json:\"nickname\" remark:\"昵称\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)
	writeTestFile(t, filepath.Join(tmpDir, "request.go"), requestContent)
	writeTestFile(t, filepath.Join(tmpDir, "response.go"), responseContent)
	writeTestFile(t, filepath.Join(tmpDir, "types.go"), typesContent)

	content := runGenFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "UserService",
		OutputPath:  filepath.Join(tmpDir, "user.proto"),
		Package:     "api.user.v1",
		GoPackage:   "github.com/example/api/user/v1;v1",
	})

	assertContainsAll(t, content,
		"message GetProfileResponse",
		"message Profile",
		"Profile profile = 1;",
		"string nickname = 1; // 昵称",
	)
	assertNotContainsAll(t, content, "placeholder", "TODO: Add fields")
}

// TestGenerator_GenFromService_ImportPathsRecursive 测试跨包 import-paths 递归闭包。
func TestGenerator_GenFromService_ImportPathsRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "service")
	sharedDir := filepath.Join(tmpDir, "shared", "dto")

	serviceContent := `package service

import (
	"context"
	"shared/dto"
)

type UserService interface {
	GetProfile(ctx context.Context, req *dto.GetProfileRequest) (*dto.GetProfileResponse, error)
}
`
	sharedContent := `package dto

type GetProfileRequest struct {
	UserID int64 ` + "`json:\"user_id\"`" + `
}

type GetProfileResponse struct {
	Profile Profile ` + "`json:\"profile\"`" + `
}

type Profile struct {
	Address Address ` + "`json:\"address\"`" + `
}

type Address struct {
	City string ` + "`json:\"city\"`" + `
}
`

	writeTestFile(t, filepath.Join(serviceDir, "service.go"), serviceContent)
	writeTestFile(t, filepath.Join(sharedDir, "types.go"), sharedContent)

	content := runGenFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(serviceDir, "service.go"),
		ServiceName: "UserService",
		OutputPath:  filepath.Join(tmpDir, "user.proto"),
		Package:     "api.user.v1",
		GoPackage:   "github.com/example/api/user/v1;v1",
		ImportPaths: []string{tmpDir},
	})

	assertContainsAll(t, content,
		"message GetProfileRequest",
		"message GetProfileResponse",
		"message Profile",
		"message Address",
		"Address address = 1;",
		"string city = 1;",
	)
	assertNotContainsAll(t, content, "placeholder", "TODO: Add fields")
}

// TestGenerator_GenFromService_SpecialTypesAndNested 测试特殊类型映射与嵌套递归。
func TestGenerator_GenFromService_SpecialTypesAndNested(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import (
	"context"
	"time"
)

type UserService interface {
	Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
}

type CreateRequest struct {
	Payload []byte ` + "`json:\"payload\"`" + `
	Meta any ` + "`json:\"meta\"`" + `
	Items []Item ` + "`json:\"items\"`" + `
	Labels map[string]Label ` + "`json:\"labels\"`" + `
	Profile *Profile ` + "`json:\"profile\"`" + `
}

type CreateResponse struct {
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	Timeout time.Duration ` + "`json:\"timeout\"`" + `
}

type Item struct {
	Name string ` + "`json:\"name\"`" + `
}

type Label struct {
	Value string ` + "`json:\"value\"`" + `
}

type Profile struct {
	Nickname string ` + "`json:\"nickname\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)

	content := runGenFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "UserService",
		OutputPath:  filepath.Join(tmpDir, "user.proto"),
		Package:     "api.user.v1",
		GoPackage:   "github.com/example/api/user/v1;v1",
	})

	assertContainsAll(t, content,
		"import \"google/protobuf/any.proto\";",
		"bytes payload = 1;",
		"google.protobuf.Any meta = 2;",
		"repeated Item items = 3;",
		"map<string, Label> labels = 4;",
		"Profile profile = 5;",
		"string created_at = 1;",
		"int64 timeout = 2;",
		"message Item",
		"message Label",
		"message Profile",
	)
	assertNotContainsAll(t, content, "repeated bytes payload", "placeholder")
}

// TestGenerator_GenFromService_UnresolvedType 测试无法解析的自定义类型直接报错。
func TestGenerator_GenFromService_UnresolvedType(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

type UserService interface {
	Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
}

type CreateRequest struct {
	Profile MissingProfile ` + "`json:\"profile\"`" + `
}

type CreateResponse struct {
	OK bool ` + "`json:\"ok\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)

	_, err := genFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "UserService",
		OutputPath:  filepath.Join(tmpDir, "user.proto"),
		Package:     "api.user.v1",
		GoPackage:   "github.com/example/api/user/v1;v1",
	})
	if err == nil {
		t.Fatal("expected unresolved type error, got nil")
	}
	if !strings.Contains(err.Error(), "references unresolved type MissingProfile") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGenerator_GenFromService_UnsupportedMapValue 测试 proto 不支持的 map value 组合直接报错。
func TestGenerator_GenFromService_UnsupportedMapValue(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

type UserService interface {
	Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
}

type CreateRequest struct {
	Attrs map[string][]string ` + "`json:\"attrs\"`" + `
}

type CreateResponse struct {
	OK bool ` + "`json:\"ok\"`" + `
}
`

	writeTestFile(t, filepath.Join(tmpDir, "service.go"), serviceContent)

	_, err := genFromService(t, integrationcontract.ServiceToProtoOptions{
		ServicePath: filepath.Join(tmpDir, "service.go"),
		ServiceName: "UserService",
		OutputPath:  filepath.Join(tmpDir, "user.proto"),
		Package:     "api.user.v1",
		GoPackage:   "github.com/example/api/user/v1;v1",
	})
	if err == nil {
		t.Fatal("expected unsupported map value error, got nil")
	}
	if !strings.Contains(err.Error(), "uses unsupported map value type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func runGenFromService(t *testing.T, opts integrationcontract.ServiceToProtoOptions) string {
	t.Helper()
	content, err := genFromService(t, opts)
	if err != nil {
		t.Fatalf("GenFromService failed: %v", err)
	}
	return content
}

func genFromService(t *testing.T, opts integrationcontract.ServiceToProtoOptions) (string, error) {
	t.Helper()
	cfg := &integrationcontract.ProtoGeneratorConfig{DefaultProtoDir: filepath.Dir(opts.OutputPath)}
	gen, err := NewGenerator(cfg)
	if err != nil {
		return "", err
	}
	if err := gen.GenFromService(context.Background(), opts); err != nil {
		return "", err
	}
	content, err := os.ReadFile(opts.OutputPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func assertContainsAll(t *testing.T, content string, want ...string) {
	t.Helper()
	for _, item := range want {
		if !strings.Contains(content, item) {
			t.Fatalf("generated content missing %q\n%s", item, content)
		}
	}
}

func assertNotContainsAll(t *testing.T, content string, unwanted ...string) {
	t.Helper()
	for _, item := range unwanted {
		if strings.Contains(content, item) {
			t.Fatalf("generated content should not contain %q\n%s", item, content)
		}
	}
}

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
		ProtoFile:   protoFile,
		OutputDir:   outputDir,
		Module:      "example.com/myproject",
		ServiceName: "OrderService",
		IncludeHTTP: true,
		IncludeGRPC: false,
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
