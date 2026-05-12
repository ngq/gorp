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
