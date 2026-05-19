// Package std_test provides unit tests for HTTP to gRPC error code mapping.
//
// 适用场景：
// - 验证 errorHandler 对 HTTP 状态码到 gRPC 错误码的正确转换。
// - 确保错误映射覆盖常见 HTTP 状态码。
package std

import (
	"testing"
)

// TestHTTPToGRPC 验证 HTTP 状态码到 gRPC 错误码的正确转换。
//
// 中文说明：
// - 覆盖常见 HTTP 状态码到 gRPC code 的映射关系。
func TestHTTPToGRPC(t *testing.T) {
	h := &errorHandler{}

	tests := []struct {
		httpCode int
		grpcCode int
	}{
		{200, 0},  // OK
		{400, 3},  // InvalidArgument
		{401, 16}, // Unauthenticated
		{403, 7},  // PermissionDenied
		{404, 5},  // NotFound
		{409, 6},  // AlreadyExists
		{429, 8},  // ResourceExhausted
		{500, 13}, // Internal
		{503, 14}, // Unavailable
		{504, 4},  // DeadlineExceeded
	}

	for _, tt := range tests {
		result := h.HTTPToGRPC(tt.httpCode)
		if result != tt.grpcCode {
			t.Errorf("HTTP %d -> expected gRPC %d, got %d", tt.httpCode, tt.grpcCode, result)
		}
	}
}

// TestGRPCToHTTP 验证 gRPC 错误码到 HTTP 状态码的反向映射。
//
// 中文说明：
// - 覆盖常见 gRPC code 到 HTTP 状态码的映射关系。
func TestGRPCToHTTP(t *testing.T) {
	h := &errorHandler{}

	tests := []struct {
		grpcCode int
		httpCode int
	}{
		{0, 200},  // OK
		{1, 499},  // Canceled
		{2, 500},  // Unknown
		{3, 400},  // InvalidArgument
		{4, 504},  // DeadlineExceeded
		{5, 404},  // NotFound
		{6, 409},  // AlreadyExists
		{7, 403},  // PermissionDenied
		{8, 429},  // ResourceExhausted
		{16, 401}, // Unauthenticated
	}

	for _, tt := range tests {
		result := h.GRPCToHTTP(tt.grpcCode)
		if result != tt.httpCode {
			t.Errorf("gRPC %d -> expected HTTP %d, got %d", tt.grpcCode, tt.httpCode, result)
		}
	}
}

// TestIsBusinessError 验证业务错误判断逻辑。
//
// 中文说明：
// - 业务错误在 contract 层已覆盖，此处做存根验证。
func TestIsBusinessError(t *testing.T) {
	// 测试在 contract 层已覆盖
}

// TestIsSystemError 验证系统错误判断逻辑。
//
// 中文说明：
// - 系统错误在 contract 层已覆盖，此处做存根验证。
func TestIsSystemError(t *testing.T) {
	// 测试在 contract 层已覆盖
}

// TestProvider_Register 验证 error provider 的注册信息。
//
// 中文说明：
// - Name 返回 "errors.default"。
// - IsDefer 返回 true。
func TestProvider_Register(t *testing.T) {
	p := NewProvider()

	if p.Name() != "errors.default" {
		t.Errorf("unexpected name: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Error("expected IsDefer=true")
	}
}
