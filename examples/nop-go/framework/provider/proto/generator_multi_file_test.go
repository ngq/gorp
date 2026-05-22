// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 验证同包多文件 DTO 自动发现、跨包 import-paths 递归闭包、嵌套类型推导。
// - 验证特殊类型（time.Time / time.Duration / []byte / any）的 proto 映射。
package proto

import (
	"path/filepath"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

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
