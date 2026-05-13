package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestProtoFromServiceCommand_GeneratesProtoWithInferredDefaults(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	outputFile := filepath.Join(root, "proto", "customer.proto")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import "context"

type CustomerService interface {
	GetCustomer(ctx context.Context, req *GetCustomerRequest) (*GetCustomerResponse, error)
}

type GetCustomerRequest struct {
	CustomerID int64 `+"`json:\"customer_id\" remark:\"客户ID\"`"+`
}

type GetCustomerResponse struct {
	Name string `+"`json:\"name\"`"+`
}
`)

	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile

	stdout, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)
	require.Contains(t, stdout, "go-package 未指定")
	require.Contains(t, stdout, "success: Service→Proto")

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `package customer;`)
	require.Contains(t, content, `option go_package = "./`+filepath.ToSlash(filepath.Dir(outputFile))+`;customer";`)
	require.Contains(t, content, `service CustomerService`)
	require.Contains(t, content, `rpc GetCustomer(GetCustomerRequest) returns (GetCustomerResponse);`)
	require.Contains(t, content, `int64 customer_id = 1; // 客户ID`)
}

func TestProtoFromServiceCommand_GeneratesProtoAcrossPackageFiles(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	writeProtoFromServiceFixture(t, filepath.Join(root, "service.go"), `package service

import "context"

type CustomerService interface {
	GetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error)
}
`)
	writeProtoFromServiceFixture(t, filepath.Join(root, "request.go"), `package service

type GetProfileRequest struct {
	UserID int64 `+"`json:\"user_id\"`"+`
}
`)
	writeProtoFromServiceFixture(t, filepath.Join(root, "response.go"), `package service

type GetProfileResponse struct {
	Profile Profile `+"`json:\"profile\"`"+`
}
`)
	writeProtoFromServiceFixture(t, filepath.Join(root, "types.go"), `package service

type Profile struct {
	Nickname string `+"`json:\"nickname\" remark:\"昵称\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "customer.proto")
	resetProtoFromServiceFlags()
	servicePath = filepath.Join(root, "service.go")
	outputDir = outputFile
	protoPackage = "api.customer.v1"
	goPackage = "github.com/example/project/api/customer/v1;customerv1"

	_, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `package api.customer.v1;`)
	require.Contains(t, content, `option go_package = "github.com/example/project/api/customer/v1;customerv1";`)
	require.Contains(t, content, `message GetProfileResponse`)
	require.Contains(t, content, `message Profile`)
	require.Contains(t, content, `Profile profile = 1;`)
	require.Contains(t, content, `string nickname = 1; // 昵称`)
	require.NotContains(t, content, `placeholder`)
}

func TestProtoFromServiceCommand_ResolvesImportPathsRecursively(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceDir := filepath.Join(root, "service")
	sharedDir := filepath.Join(root, "shared", "dto")

	writeProtoFromServiceFixture(t, filepath.Join(serviceDir, "service.go"), `package service

import (
	"context"
	"shared/dto"
)

type CustomerService interface {
	GetProfile(ctx context.Context, req *dto.GetProfileRequest) (*dto.GetProfileResponse, error)
}
`)
	writeProtoFromServiceFixture(t, filepath.Join(sharedDir, "types.go"), `package dto

type GetProfileRequest struct {
	UserID int64 `+"`json:\"user_id\"`"+`
}

type GetProfileResponse struct {
	Profile Profile `+"`json:\"profile\"`"+`
}

type Profile struct {
	Address Address `+"`json:\"address\"`"+`
}

type Address struct {
	City string `+"`json:\"city\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "customer.proto")
	resetProtoFromServiceFlags()
	servicePath = filepath.Join(serviceDir, "service.go")
	outputDir = outputFile
	protoPackage = "api.customer.v1"
	goPackage = "github.com/example/project/api/customer/v1;customerv1"
	importPathsS = []string{root}

	_, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `message GetProfileRequest`)
	require.Contains(t, content, `message GetProfileResponse`)
	require.Contains(t, content, `message Profile`)
	require.Contains(t, content, `message Address`)
	require.Contains(t, content, `Address address = 1;`)
	require.Contains(t, content, `string city = 1;`)
	require.NotContains(t, content, `placeholder`)
}

// TestProtoFromServiceCommand_FailsOnMissingFile 验证不存在的 service 文件会返回明确错误。
func TestProtoFromServiceCommand_FailsOnMissingFile(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	resetProtoFromServiceFlags()
	servicePath = "/nonexistent/path/service.go"
	outputDir = filepath.Join(t.TempDir(), "proto", "out.proto")

	err := runProtoFromService(protoFromServiceCmd, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse service file")
}

// TestProtoFromServiceCommand_FailsOnNoServiceInterface 验证没有接口定义的文件会报错。
func TestProtoFromServiceCommand_FailsOnNoServiceInterface(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

type NotAnInterface struct {
	Name string
}
`)

	outputFile := filepath.Join(root, "proto", "customer.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile

	err := runProtoFromService(protoFromServiceCmd, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no service interface found")
}

// TestProtoFromServiceCommand_FailsOnUnresolvedType 验证不可解析的自定义类型会直接报错，而非生成 placeholder。
func TestProtoFromServiceCommand_FailsOnUnresolvedType(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import "context"

type OrderService interface {
	Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
}

type CreateRequest struct {
	Profile MissingProfile `+"`json:\"profile\"`"+`
}

type CreateResponse struct {
	OK bool `+"`json:\"ok\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "order.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile

	err := runProtoFromService(protoFromServiceCmd, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "references unresolved type MissingProfile")
}

// TestProtoFromServiceCommand_FailsOnUnsupportedMapValue 验证 proto 不支持的 map value 类型会直接报错。
func TestProtoFromServiceCommand_FailsOnUnsupportedMapValue(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import "context"

type OrderService interface {
	Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
}

type CreateRequest struct {
	Attrs map[string][]string `+"`json:\"attrs\"`"+`
}

type CreateResponse struct {
	OK bool `+"`json:\"ok\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "order.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile

	err := runProtoFromService(protoFromServiceCmd, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported map value type")
}

// TestProtoFromServiceCommand_GeneratesProtoWithMapTypes 验证 map 类型正确生成 proto map 语法。
func TestProtoFromServiceCommand_GeneratesProtoWithMapTypes(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import "context"

type ConfigService interface {
	GetConfig(ctx context.Context, req *GetConfigRequest) (*GetConfigResponse, error)
}

type GetConfigRequest struct {
	Key string `+"`json:\"key\" remark:\"配置键\"`"+`
}

type GetConfigResponse struct {
	Labels map[string]string `+"`json:\"labels\" remark:\"标签\"`"+`
	Meta   map[string]Meta   `+"`json:\"meta\"`"+`
}

type Meta struct {
	Value string `+"`json:\"value\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "config.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile
	protoPackage = "api.config.v1"
	goPackage = "github.com/example/project/api/config/v1;configv1"

	_, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `map<string, string> labels = 1; // 标签`)
	require.Contains(t, content, `map<string, Meta> meta = 2;`)
	require.Contains(t, content, `message Meta`)
	require.Contains(t, content, `string value = 1;`)
}

// TestProtoFromServiceCommand_GeneratesProtoWithSpecialTypes 验证特殊类型（time、bytes、any）映射。
func TestProtoFromServiceCommand_GeneratesProtoWithSpecialTypes(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import (
	"context"
	"time"
)

type EventService interface {
	Create(ctx context.Context, req *CreateEventRequest) (*CreateEventResponse, error)
}

type CreateEventRequest struct {
	Payload []byte         `+"`json:\"payload\"`"+`
	Meta    any            `+"`json:\"meta\"`"+`
	Items   []Item         `+"`json:\"items\"`"+`
}

type CreateEventResponse struct {
	CreatedAt time.Time     `+"`json:\"created_at\"`"+`
	Timeout   time.Duration `+"`json:\"timeout\"`"+`
}

type Item struct {
	Name string `+"`json:\"name\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "event.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile
	protoPackage = "api.event.v1"
	goPackage = "github.com/example/project/api/event/v1;eventv1"

	_, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `bytes payload = 1;`)
	require.Contains(t, content, `google.protobuf.Any meta = 2;`)
	require.Contains(t, content, `repeated Item items = 3;`)
	require.Contains(t, content, `string created_at = 1;`)
	require.Contains(t, content, `int64 timeout = 2;`)
	require.Contains(t, content, `import "google/protobuf/any.proto";`)
}

// TestProtoFromServiceCommand_SpecifyServiceName 验证 --service-name 只生成指定服务。
func TestProtoFromServiceCommand_SpecifyServiceName(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import "context"

type OrderService interface {
	Create(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
}

type PaymentService interface {
	Pay(ctx context.Context, req *PayRequest) (*PayResponse, error)
}

type CreateOrderRequest struct {
	Item string `+"`json:\"item\"`"+`
}

type CreateOrderResponse struct {
	OK bool `+"`json:\"ok\"`"+`
}

type PayRequest struct {
	Amount int64 `+"`json:\"amount\"`"+`
}

type PayResponse struct {
	OK bool `+"`json:\"ok\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "order.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile
	serviceName = "OrderService"
	protoPackage = "api.order.v1"
	goPackage = "github.com/example/project/api/order/v1;orderv1"

	_, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `service OrderService`)
	require.NotContains(t, content, `PaymentService`)
	require.Contains(t, content, `rpc Create`)
}

// TestProtoFromServiceCommand_IncludeHTTPAnnotation 验证 --include-http 生成 google.api.http 注解。
func TestProtoFromServiceCommand_IncludeHTTPAnnotation(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	serviceFile := filepath.Join(root, "service.go")
	writeProtoFromServiceFixture(t, serviceFile, `package service

import "context"

type UserService interface {
	GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)
}

type GetUserRequest struct {
	ID int64 `+"`json:\"id\"`"+`
}

type GetUserResponse struct {
	Name string `+"`json:\"name\"`"+`
}
`)

	outputFile := filepath.Join(root, "proto", "user.proto")
	resetProtoFromServiceFlags()
	servicePath = serviceFile
	outputDir = outputFile
	includeHTTPS = true
	protoPackage = "api.user.v1"
	goPackage = "github.com/example/project/api/user/v1;userv1"

	_, stderr := captureProcessOutput(t, func() error {
		return runProtoFromService(protoFromServiceCmd, nil)
	})
	require.Empty(t, stderr)

	content := readGeneratedProto(t, outputFile)
	require.Contains(t, content, `import "google/api/annotations.proto";`)
}

func resetProtoFromServiceFlags() {
	servicePath = ""
	outputDir = ""
	protoPackage = ""
	goPackage = ""
	serviceName = ""
	includeHTTPS = false
	importPathsS = nil
}

func writeProtoFromServiceFixture(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func readGeneratedProto(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return strings.ReplaceAll(string(content), "\\", "/")
}

func captureProcessOutput(t *testing.T, run func() error) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)
	stderrReader, stderrWriter, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	runErr := run()

	require.NoError(t, stdoutWriter.Close())
	require.NoError(t, stderrWriter.Close())
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	_, err = stdoutBuf.ReadFrom(stdoutReader)
	require.NoError(t, err)
	_, err = stderrBuf.ReadFrom(stderrReader)
	require.NoError(t, err)

	require.NoError(t, stdoutReader.Close())
	require.NoError(t, stderrReader.Close())
	require.NoError(t, runErr)

	return stdoutBuf.String(), stderrBuf.String()
}
