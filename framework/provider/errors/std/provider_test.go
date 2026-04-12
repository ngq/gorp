package std

import (
	"testing"
)

func TestHTTPToGRPC(t *testing.T) {
	h := &errorHandler{}

	tests := []struct {
		httpCode int
		grpcCode int
	}{
		{200, 0},   // OK
		{400, 3},   // InvalidArgument
		{401, 16},  // Unauthenticated
		{403, 7},   // PermissionDenied
		{404, 5},   // NotFound
		{409, 6},   // AlreadyExists
		{429, 8},   // ResourceExhausted
		{500, 13},  // Internal
		{503, 14},  // Unavailable
		{504, 4},   // DeadlineExceeded
	}

	for _, tt := range tests {
		result := h.HTTPToGRPC(tt.httpCode)
		if result != tt.grpcCode {
			t.Errorf("HTTP %d -> expected gRPC %d, got %d", tt.httpCode, tt.grpcCode, result)
		}
	}
}

func TestGRPCToHTTP(t *testing.T) {
	h := &errorHandler{}

	tests := []struct {
		grpcCode int
		httpCode int
	}{
		{0, 200},   // OK
		{1, 499},   // Canceled
		{2, 500},   // Unknown
		{3, 400},   // InvalidArgument
		{4, 504},   // DeadlineExceeded
		{5, 404},   // NotFound
		{6, 409},   // AlreadyExists
		{7, 403},   // PermissionDenied
		{8, 429},   // ResourceExhausted
		{16, 401},  // Unauthenticated
	}

	for _, tt := range tests {
		result := h.GRPCToHTTP(tt.grpcCode)
		if result != tt.httpCode {
			t.Errorf("gRPC %d -> expected HTTP %d, got %d", tt.grpcCode, tt.httpCode, result)
		}
	}
}

func TestIsBusinessError(t *testing.T) {
	// 测试在 contract 层已覆盖
}

func TestIsSystemError(t *testing.T) {
	// 测试在 contract 层已覆盖
}

func TestProvider_Register(t *testing.T) {
	p := NewProvider()

	if p.Name() != "errors.default" {
		t.Errorf("unexpected name: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Error("expected IsDefer=true")
	}
}