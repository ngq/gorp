// Package main provides mock gRPC backend server implementation.
// This file implements the TestService gRPC server.
//
// 本包提供 mock gRPC backend server 实现。
// 本文件实现 TestService gRPC server。
package main

import (
	"context"

	pb "github.com/ngq/gorp/test/integration/pb"
)

// grpc_server.go provides additional gRPC server utilities.
// The main server implementation is in main.go.
//
// grpc_server.go 提供额外的 gRPC server 工具。
// 主要 server 实现在 main.go 中。

// RecordingServer wraps TestServiceServer with call recording.
//
// RecordingServer 包装 TestServiceServer，提供调用记录。
type RecordingServer struct {
	*TestServiceServer
}

// NewRecordingServer creates a new RecordingServer.
//
// NewRecordingServer 创建新的 RecordingServer。
func NewRecordingServer() *RecordingServer {
	return &RecordingServer{
		TestServiceServer: &TestServiceServer{},
	}
}

// VerifyTracingPropagation verifies that trace-id was propagated correctly.
//
// VerifyTracingPropagation 验证 trace-id 正确传播。
func (s *RecordingServer) VerifyTracingPropagation(expectedTraceID string) bool {
	return s.GetTraceID() == expectedTraceID
}

// VerifyMetadataPropagation verifies that metadata was propagated correctly.
// Checks against the metadata recorded on actual calls (the previous version
// verified an empty metadata.MD{}, which always returned false for any
// non-empty expectedKeys and was useless).
//
// VerifyMetadataPropagation 验证 metadata 正确传播。
// 基于实际调用记录中的 metadata 校验（旧版对空 metadata.MD{} 校验，
// expectedKeys 非空时恒 false，完全失效）。
func (s *RecordingServer) VerifyMetadataPropagation(expectedKeys []string) bool {
	calls := s.GetCalls()
	if len(calls) == 0 {
		return false
	}
	// 任意一次调用包含全部 expected key 即视为传播成功。
	for _, call := range calls {
		all := true
		for _, key := range expectedKeys {
			if _, ok := call.Metadata[key]; !ok {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

// ClearCalls clears all recorded calls.
//
// ClearCalls 清除所有记录的调用。
func (s *RecordingServer) ClearCalls() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = nil
}

// SayHelloWithContext records call with custom context.
//
// SayHelloWithContext 记录带自定义 context 的调用。
func (s *RecordingServer) SayHelloWithContext(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return s.TestServiceServer.SayHello(ctx, req)
}
