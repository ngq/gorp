// Package benchmark 提供 HTTP/gRPC 全链路性能基准测试。
//
// 本文件测试：
// - HTTP middleware chain（13 阶段全链路）
// - gRPC interceptor chain（9 阶段全链路）
// - Tracing span 创建开销
// - JSON 序列化/反序列化开销
//
// 运行方式：
//
//	go test ./benchmark/... -bench=. -benchmem
package benchmark

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	otelprovider "github.com/ngq/gorp/contrib/tracing/otel"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
)

// ============================================================
// HTTP Middleware Chain 性能测试
// 说明：13 阶段 middleware chain 是每请求必经路径。
// ============================================================

// mockMiddlewareChain 模拟完整 middleware chain（13 阶段）
func mockMiddlewareChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// 1. request_identity
		func(c *gin.Context) {
			c.Set("request_id", "test-123")
			c.Next()
		},
		// 2. logging (noop for benchmark)
		func(c *gin.Context) { c.Next() },
		// 3. recovery
		func(c *gin.Context) {
			defer func() {
				if r := recover(); r != nil {
					c.AbortWithStatus(500)
				}
			}()
			c.Next()
		},
		// 4. cors (skip for benchmark)
		func(c *gin.Context) { c.Next() },
		// 5. security_headers
		func(c *gin.Context) {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Next()
		},
		// 6. timeout (skip actual timeout for benchmark)
		func(c *gin.Context) { c.Next() },
		// 7. rate_limit (skip check for benchmark)
		func(c *gin.Context) { c.Next() },
		// 8. loadshedding (skip check for benchmark)
		func(c *gin.Context) { c.Next() },
		// 9. circuit_breaker (skip check for benchmark)
		func(c *gin.Context) { c.Next() },
		// 10. tracing (minimal overhead)
		func(c *gin.Context) {
			c.Set("trace_id", "trace-456")
			c.Next()
		},
		// 11. metadata
		func(c *gin.Context) {
			c.Set("metadata", map[string]string{"x-custom": "value"})
			c.Next()
		},
		// 12. auth (skip check for benchmark)
		func(c *gin.Context) { c.Next() },
		// 13. metrics
		func(c *gin.Context) {
			c.Next()
		},
	}
}

// BenchmarkMiddlewareChain_Full 测试完整 13 阶段 middleware chain
func BenchmarkMiddlewareChain_Full(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	chain := mockMiddlewareChain()
	for _, mw := range chain {
		router.Use(mw)
	}

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareChain_5Stages 测试 5 阶段 middleware chain（简化场景）
func BenchmarkMiddlewareChain_5Stages(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 只保留核心 5 个 middleware
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-123")
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		c.Set("trace_id", "trace-456")
		c.Next()
	})
	router.Use(func(c *gin.Context) { c.Next() }) // placeholder
	router.Use(func(c *gin.Context) { c.Next() }) // placeholder

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareChain_Short 测试短路场景（middleware 提前终止）
func BenchmarkMiddlewareChain_Short(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 在第 7 阶段模拟限流拒绝
	for i := 0; i < 7; i++ {
		router.Use(func(c *gin.Context) { c.Next() })
	}
	router.Use(func(c *gin.Context) {
		c.AbortWithStatus(429)
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareChain_NoMiddleware 测试无 middleware 的纯 handler
func BenchmarkMiddlewareChain_NoMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkMiddleware_PanicRecovery 测试 panic recovery 开销
func BenchmarkMiddleware_PanicRecovery(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkMiddleware_PanicRecovery_Active 测试 panic recovery 实际捕获开销
func BenchmarkMiddleware_PanicRecovery_Active(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ============================================================
// gRPC Interceptor Chain 性能测试（模拟）
// 说明：9 阶段 interceptor chain 是 gRPC 每请求必经路径。
// ============================================================

// mockGRPCUnaryInterceptorChain 模拟 9 阶段 gRPC interceptor chain
func mockGRPCUnaryInterceptorChain() []func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
	return []func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error){
		// 1. recovery
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			defer func() {
				if r := recover(); r != nil {
					// recovery logic
				}
			}()
			return handler(ctx, req)
		},
		// 2. logging (noop for benchmark)
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			return handler(ctx, req)
		},
		// 3. timeout
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return handler(ctx, req)
		},
		// 4. tracing
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			ctx = context.WithValue(ctx, "trace_id", "trace-789")
			return handler(ctx, req)
		},
		// 5. metadata
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			return handler(ctx, req)
		},
		// 6. serviceauth (skip for benchmark)
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			return handler(ctx, req)
		},
		// 7. rate_limit (skip for benchmark)
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			return handler(ctx, req)
		},
		// 8. loadshedding (skip for benchmark)
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			return handler(ctx, req)
		},
		// 9. metrics
		func(ctx context.Context, req any, info string, handler func(ctx context.Context, req any) (any, error)) (any, error) {
			return handler(ctx, req)
		},
	}
}

// BenchmarkGRPCInterceptorChain_Full 测试完整 9 阶段 interceptor chain
func BenchmarkGRPCInterceptorChain_Full(b *testing.B) {
	chain := mockGRPCUnaryInterceptorChain()

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		req := "test-request"

		// 模拟 interceptor chain 应用
		var finalHandler func(ctx context.Context, req any) (any, error) = handler
		for idx := len(chain) - 1; idx >= 0; idx-- {
			interceptor := chain[idx]
			nextHandler := finalHandler
			finalHandler = func(ctx context.Context, req any) (any, error) {
				return interceptor(ctx, req, "/test.Service/Method", nextHandler)
			}
		}

		_, _ = finalHandler(ctx, req)
	}
}

// BenchmarkGRPCInterceptorChain_5Stages 测试 5 阶段 interceptor chain
func BenchmarkGRPCInterceptorChain_5Stages(b *testing.B) {
	chain := mockGRPCUnaryInterceptorChain()[:5]

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		req := "test-request"

		var finalHandler func(ctx context.Context, req any) (any, error) = handler
		for idx := len(chain) - 1; idx >= 0; idx-- {
			interceptor := chain[idx]
			nextHandler := finalHandler
			finalHandler = func(ctx context.Context, req any) (any, error) {
				return interceptor(ctx, req, "/test.Service/Method", nextHandler)
			}
		}

		_, _ = finalHandler(ctx, req)
	}
}

// BenchmarkGRPCInterceptorChain_NoInterceptor 测试无 interceptor 的纯 handler
func BenchmarkGRPCInterceptorChain_NoInterceptor(b *testing.B) {
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		req := "test-request"
		_, _ = handler(ctx, req)
	}
}

// ============================================================
// Tracing Span 性能测试
// 说明：OTel span 创建有开销，需验证是否在可接受范围。
// ============================================================

// BenchmarkTracingSpanCreation 测试 OTel span 创建开销
func BenchmarkTracingSpanCreation(b *testing.B) {
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "benchmark-service",
		ExporterEndpoint: "",
		ExporterType:     "stdout",
		Enabled:          false,
		SamplingRate:     1.0,
	}

	tracerProvider, _ := otelprovider.NewTracerProvider(cfg)
	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, span := tracer.StartSpan(ctx, "benchmark-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindServer
		})
		span.End()
	}
}

// BenchmarkTracingSpanCreationWithAttributes 测试带 attributes 的 span 创建
func BenchmarkTracingSpanCreationWithAttributes(b *testing.B) {
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "benchmark-service",
		ExporterEndpoint: "",
		ExporterType:     "stdout",
		Enabled:          false,
		SamplingRate:     1.0,
	}

	tracerProvider, _ := otelprovider.NewTracerProvider(cfg)
	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, span := tracer.StartSpan(ctx, "benchmark-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindServer
			cfg.Attributes = map[string]interface{}{
				"user.id":    12345,
				"request.id": "req-123",
				"method":     "GET",
			}
		})
		span.SetAttributes(map[string]interface{}{
			"result": "success",
			"code":   200,
		})
		span.End()
	}
}

// BenchmarkTracingSpanCreation_Nested 测试嵌套 span 创建
func BenchmarkTracingSpanCreation_Nested(b *testing.B) {
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "benchmark-service",
		ExporterEndpoint: "",
		ExporterType:     "stdout",
		Enabled:          false,
		SamplingRate:     1.0,
	}

	tracerProvider, _ := otelprovider.NewTracerProvider(cfg)
	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()

		// parent span
		ctx, parentSpan := tracer.StartSpan(ctx, "parent-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindServer
		})

		// child span
		ctx, childSpan := tracer.StartSpan(ctx, "child-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindInternal
		})
		childSpan.End()

		parentSpan.End()
	}
}

// ============================================================
// JSON 序列化/反序列化性能测试
// 说明：每请求必经，直接影响吞吐量。
// ============================================================

type BenchmarkResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type BenchmarkUserData struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

// BenchmarkJSONSerialize 测试 JSON 序列化
func BenchmarkJSONSerialize(b *testing.B) {
	resp := BenchmarkResponse{
		Status:  200,
		Message: "success",
		Data: BenchmarkUserData{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Age:      25,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

// BenchmarkJSONDeserialize 测试 JSON 反序列化
func BenchmarkJSONDeserialize(b *testing.B) {
	data := []byte(`{"status":200,"message":"success","data":{"id":1,"username":"testuser","email":"test@example.com","age":25}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp BenchmarkResponse
		_ = json.Unmarshal(data, &resp)
	}
}

// BenchmarkJSONSerialize_Large 测试大型 JSON 序列化（100 个用户）
func BenchmarkJSONSerialize_Large(b *testing.B) {
	users := make([]BenchmarkUserData, 100)
	for i := 0; i < 100; i++ {
		users[i] = BenchmarkUserData{
			ID:       i,
			Username: "user" + string(rune('0'+i%10)),
			Email:    "user" + string(rune('0'+i%10)) + "@example.com",
			Age:      20 + i%50,
		}
	}

	resp := BenchmarkResponse{
		Status:  200,
		Message: "success",
		Data:    users,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

// ============================================================
// Context 操作性能测试
// 说明：middleware chain 中高频 context.Value 操作。
// ============================================================

// BenchmarkContextSet 测试 context.Set 开销（通过 gin.Context）
func BenchmarkContextSet(b *testing.B) {
	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("key", "value")
		_ = c
	}
}

// BenchmarkContextGet 测试 context.Get 开销（通过 gin.Context）
func BenchmarkContextGet(b *testing.B) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get("key")
	}
}

// BenchmarkContextWithValue 测试标准 context.WithValue 开销
func BenchmarkContextWithValue(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = context.WithValue(ctx, "key", "value")
	}
}

// BenchmarkContextValue 测试标准 context.Value 查询开销
func BenchmarkContextValue(b *testing.B) {
	ctx := context.WithValue(context.Background(), "key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.Value("key")
	}
}

// ============================================================
// 综合请求模拟测试
// ============================================================

// BenchmarkFullHTTPRequest 模拟完整 HTTP 请求生命周期
func BenchmarkFullHTTPRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 使用简化 middleware（模拟真实场景）
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-123")
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		c.Set("trace_id", "trace-456")
		c.Next()
	})

	router.GET("/api/users/:id", func(c *gin.Context) {
		// 模拟业务逻辑
		_ = c.Param("id")

		// 模拟数据库查询结果
		user := BenchmarkUserData{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Age:      25,
		}

		c.JSON(200, BenchmarkResponse{
			Status:  200,
			Message: "success",
			Data:    user,
		})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkConcurrentHTTPRequests 模拟并发请求场景
func BenchmarkConcurrentHTTPRequests(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}
