// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 验证不可解析类型或不支持的 map 组合直接返回明确错误，而非产出 placeholder。
package proto

import (
	"path/filepath"
	"strings"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

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
