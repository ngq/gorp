// Package main provides mock gRPC backend for integration tests.
// This server implements TestService with tracing/metadata recording.
//
// 本包提供集成测试用的 mock gRPC backend。
// 此 server 实现 TestService，记录 tracing/metadata 信息。
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/ngq/gorp/test/integration/pb"
)

// TestServiceServer implements pb.TestServiceServer for integration testing.
//
// TestServiceServer 实现 pb.TestServiceServer 用于集成测试。
type TestServiceServer struct {
	pb.UnimplementedTestServiceServer

	mu      sync.Mutex
	calls   []CallRecord
	traceID string
}

// CallRecord 记录一次调用收到的信息
type CallRecord struct {
	Method    string
	TraceID   string
	RequestID string
	Metadata  map[string]string
}

// SayHello handles simple request-response test.
// Records received trace-id and request-id for verification.
//
// SayHello 处理简单请求-响应测试。
// 记录收到的 trace-id 和 request-id 用于验证。
func (s *TestServiceServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	traceID := ""
	requestID := ""
	if values := md.Get("x-trace-id"); len(values) > 0 {
		traceID = values[0]
	}
	if values := md.Get("x-request-id"); len(values) > 0 {
		requestID = values[0]
	}

	s.mu.Lock()
	s.calls = append(s.calls, CallRecord{
		Method:    "SayHello",
		TraceID:   traceID,
		RequestID: requestID,
		Metadata:  extractMetadataMap(md),
	})
	s.traceID = traceID
	s.mu.Unlock()

	log.Printf("SayHello: name=%s, trace-id=%s, request-id=%s", req.Name, traceID, requestID)

	return &pb.HelloResponse{
		Message:   "Hello " + req.Name,
		TraceId:   traceID,
		RequestId: requestID,
	}, nil
}

// Echo returns request content with received metadata for propagation testing.
//
// Echo 返回请求内容与收到的 metadata，用于传播测试。
func (s *TestServiceServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	s.mu.Lock()
	s.calls = append(s.calls, CallRecord{
		Method:   "Echo",
		Metadata: extractMetadataMap(md),
	})
	s.mu.Unlock()

	log.Printf("Echo: body=%s, metadata=%v", req.Body, extractMetadataMap(md))

	return &pb.EchoResponse{
		Body:     req.Body,
		Metadata: extractMetadataMap(md),
	}, nil
}

// StreamTest handles streaming test.
//
// StreamTest 处理流式测试。
func (s *TestServiceServer) StreamTest(req *pb.StreamRequest, stream pb.TestService_StreamTestServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())

	s.mu.Lock()
	s.calls = append(s.calls, CallRecord{
		Method:   "StreamTest",
		Metadata: extractMetadataMap(md),
	})
	s.mu.Unlock()

	for i := 0; i < int(req.Count); i++ {
		if err := stream.Send(&pb.StreamResponse{
			Index:   int32(i),
			Message: fmt.Sprintf("stream message %d", i),
		}); err != nil {
			return err
		}
	}
	return nil
}

// GetCalls returns all recorded calls for test verification.
//
// GetCalls 返回所有记录的调用，用于测试验证。
func (s *TestServiceServer) GetCalls() []CallRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]CallRecord(nil), s.calls...)
}

// GetTraceID returns the last received trace-id.
//
// GetTraceID 返回最后收到的 trace-id。
func (s *TestServiceServer) GetTraceID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.traceID
}

// extractMetadataMap converts grpc metadata to map[string]string.
//
// extractMetadataMap 将 grpc metadata 转换为 map[string]string。
func extractMetadataMap(md metadata.MD) map[string]string {
	result := make(map[string]string)
	for k, values := range md {
		if len(values) > 0 {
			result[k] = values[0]
		}
	}
	return result
}

var globalServer = &TestServiceServer{}

func main() {
	port := os.Getenv("GORP_TEST_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterTestServiceServer(srv, globalServer)

	log.Printf("Mock gRPC backend started on port %s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
