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

// TestGenerator_GenFromService_WithHTTPAnnotation 测试带 HTTP 注释的方法生成 HTTP 注解。
//
// 中文说明：
// - 验证方法注释中的 HTTP: 标记被正确解析为 google.api.http 注解；
// - 验证不带 HTTP: 标记的方法只生成纯 gRPC；
// - 验证自动导入 google/api/annotations.proto。
func TestGenerator_GenFromService_WithHTTPAnnotation(t *testing.T) {
	tmpDir := t.TempDir()

	serviceContent := `package service

import "context"

// UserService 用户服务接口。
type UserService interface {
	// GetUser 获取用户
	// HTTP: GET /v1/users/{id}
	GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)

	// CreateUser 创建用户
	// HTTP: POST /v1/users
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)

	// ValidateToken 验证 token - 只 gRPC，无 HTTP
	ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error)
}

type GetUserRequest struct {
	ID int64 ` + "`json:\"id\"`" + `
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

type ValidateTokenRequest struct {
	Token string ` + "`json:\"token\"`" + `
}

type ValidateTokenResponse struct {
	Valid bool ` + "`json:\"valid\"`" + `
}
`
	servicePath := filepath.Join(tmpDir, "user_service.go")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("failed to write service file: %v", err)
	}

	cfg := &integrationcontract.ProtoGeneratorConfig{
		DefaultProtoDir: tmpDir,
	}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "user.proto")
	opts := integrationcontract.ServiceToProtoOptions{
		ServicePath: servicePath,
		OutputPath:  outputPath,
		Package:     "user.v1",
		GoPackage:   "github.com/example/api/user/v1;userv1",
		// 不需要 IncludeHTTP=true，HTTP 注解从方法注释自动解析
	}

	err = gen.GenFromService(context.Background(), opts)
	if err != nil {
		t.Fatalf("GenFromService failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read proto file: %v", err)
	}
	contentStr := string(content)

	// 验证自动导入 HTTP 注解
	assertContainsAll(t, contentStr, `import "google/api/annotations.proto"`)

	// 验证 GetUser 有 HTTP 注解
	assertContainsAll(t, contentStr,
		`rpc GetUser(GetUserRequest) returns (GetUserResponse) {`,
		`option (google.api.http) = {`,
		`get: "/v1/users/{id}"`,
	)

	// 验证 CreateUser 有 HTTP 注解（POST 默认带 body）
	assertContainsAll(t, contentStr,
		`rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {`,
		`post: "/v1/users"`,
		`body: "*"`,
	)

	// 验证 ValidateToken 无 HTTP 注解（只有 gRPC）
	assertContainsAll(t, contentStr, `rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);`)
	assertNotContainsAll(t, contentStr, `rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse) {`)
}

// TestParseHTTPAnnotation 测试 HTTP 注释解析。
//
// 中文说明：
// - 验证各种 HTTP 注释格式的解析；
// - 验证格式不规范时的错误提示。
func TestParseHTTPAnnotation(t *testing.T) {
	cfg := &integrationcontract.ProtoGeneratorConfig{}
	gen, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	tests := []struct {
		name        string
		comments    []string
		wantRule    bool
		wantMethod  string
		wantPath    string
		wantBody    string
		wantErrPart string
	}{
		{
			name:       "valid GET",
			comments:   []string{"GetUser 获取用户", "HTTP: GET /v1/users/{id}"},
			wantRule:   true,
			wantMethod: "GET",
			wantPath:   "/v1/users/{id}",
		},
		{
			name:       "valid POST",
			comments:   []string{"CreateUser 创建用户", "HTTP: POST /v1/users"},
			wantRule:   true,
			wantMethod: "POST",
			wantPath:   "/v1/users",
			wantBody:   "*",
		},
		{
			name:       "valid PUT",
			comments:   []string{"UpdateUser 更新用户", "HTTP: PUT /v1/users/{id}"},
			wantRule:   true,
			wantMethod: "PUT",
			wantPath:   "/v1/users/{id}",
			wantBody:   "*",
		},
		{
			name:       "valid DELETE",
			comments:   []string{"DeleteUser 删除用户", "HTTP: DELETE /v1/users/{id}"},
			wantRule:   true,
			wantMethod: "DELETE",
			wantPath:   "/v1/users/{id}",
		},
		{
			name:       "valid PATCH",
			comments:   []string{"PatchUser 部分更新用户", "HTTP: PATCH /v1/users/{id}"},
			wantRule:   true,
			wantMethod: "PATCH",
			wantPath:   "/v1/users/{id}",
			wantBody:   "*",
		},
		{
			name:        "invalid - empty HTTP annotation",
			comments:    []string{"GetUser 获取用户", "HTTP:"},
			wantRule:    false,
			wantErrPart: "HTTP annotation is empty",
		},
		{
			name:        "invalid - missing path",
			comments:    []string{"GetUser 获取用户", "HTTP: GET"},
			wantRule:    false,
			wantErrPart: "invalid HTTP annotation format",
		},
		{
			name:        "invalid - path without leading slash",
			comments:    []string{"GetUser 获取用户", "HTTP: GET v1/users/{id}"},
			wantRule:    false,
			wantErrPart: "path must start with /",
		},
		{
			name:        "invalid - unsupported method",
			comments:    []string{"GetUser 获取用户", "HTTP: OPTIONS /v1/users"},
			wantRule:    false,
			wantErrPart: "invalid HTTP method",
		},
		{
			name:     "no HTTP annotation",
			comments: []string{"ValidateToken 验证 token", "这是一个内部方法"},
			wantRule: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := gen.parseHTTPAnnotation(tt.comments)

			if tt.wantErrPart != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrPart)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrPart)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.wantRule {
				if rule == nil {
					t.Error("expected HTTPRule, got nil")
					return
				}
				if rule.Method != tt.wantMethod {
					t.Errorf("method: got %q, want %q", rule.Method, tt.wantMethod)
				}
				if rule.Path != tt.wantPath {
					t.Errorf("path: got %q, want %q", rule.Path, tt.wantPath)
				}
				if tt.wantBody != "" && rule.Body != tt.wantBody {
					t.Errorf("body: got %q, want %q", rule.Body, tt.wantBody)
				}
			} else {
				if rule != nil {
					t.Errorf("expected no HTTPRule, got %+v", rule)
				}
			}
		})
	}
}
