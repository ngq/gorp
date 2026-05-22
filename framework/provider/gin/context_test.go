// Package gin_test provides unit tests for ginContext.Value BUG-001 fix.
//
// BUG-001: ginContext.Value 栈溢出问题修复验证
// 问题根因：当 ginContext 被包装在 context.WithValue 链中时，
// 调用 Value() 方法会导致无限递归，最终栈溢出崩溃。
package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGinContextValue_NoStackOverflow 验证 ginContext.Value 不会栈溢出。
//
// 测试场景：
// 1. 创建嵌套的 context.WithValue 链
// 2. 将 ginContext 包装在链中
// 3. 调用 Value() 方法不应导致栈溢出
func TestGinContextValue_NoStackOverflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var valueResult any
	var panicOccurred bool

	engine.GET("/test", func(gc *gin.Context) {
		// 模拟框架中间件创建嵌套 context
		ctx := context.WithValue(gc.Request.Context(), "key1", "value1")
		ctx = context.WithValue(ctx, "key2", "value2")
		ctx = context.WithValue(ctx, "key3", "value3")

		// 更新请求的 context
		gc.Request = gc.Request.WithContext(ctx)

		// 创建 ginContext 包装
		c := newContext(gc)

		// 这个调用在修复前会导致栈溢出
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
				t.Errorf("Stack overflow or panic occurred: %v", r)
			}
		}()

		// 测试获取值
		valueResult = c.Context().Value("key1")
		gc.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if panicOccurred {
		t.Fatal("panic occurred during Value() call")
	}

	if valueResult != "value1" {
		t.Errorf("expected value1, got %v", valueResult)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestGinContextValue_NestedContextValues 验证嵌套 context 值的正确获取。
//
// 测试场景：
// 1. 多层 context.WithValue 嵌套
// 2. 验证每一层的值都能正确获取
func TestGinContextValue_NestedContextValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	results := make(map[string]any)

	engine.GET("/test", func(gc *gin.Context) {
		// 创建多层嵌套 context
		ctx := context.WithValue(gc.Request.Context(), "level1", "value1")
		ctx = context.WithValue(ctx, "level2", "value2")
		ctx = context.WithValue(ctx, "level3", "value3")

		gc.Request = gc.Request.WithContext(ctx)

		c := newContext(gc)

		// 验证每一层的值
		results["level1"] = c.Context().Value("level1")
		results["level2"] = c.Context().Value("level2")
		results["level3"] = c.Context().Value("level3")
		results["notexist"] = c.Context().Value("notexist")

		gc.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if results["level1"] != "value1" {
		t.Errorf("level1: expected value1, got %v", results["level1"])
	}
	if results["level2"] != "value2" {
		t.Errorf("level2: expected value2, got %v", results["level2"])
	}
	if results["level3"] != "value3" {
		t.Errorf("level3: expected value3, got %v", results["level3"])
	}
	if results["notexist"] != nil {
		t.Errorf("notexist: expected nil, got %v", results["notexist"])
	}
}

// TestGinContextValue_GinInternalStorage 验证 gin.Context 内部存储优先级。
//
// 测试场景：
// 1. gin.Context 内部存储设置值
// 2. Request.Context() 也设置相同 key
// TestGinContextValue_GinInternalStorage 验证重构后行为：
// 重构后 Context() 直接返回 Request.Context()，不再混合 gin 内部存储。
// 业务想读 gin 内部存储应使用 c.Get(key)。
func TestGinContextValue_GinInternalStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var ctxValue any
	var ginValue any

	engine.GET("/test", func(gc *gin.Context) {
		// 在 gin.Context 内部存储设置值
		gc.Set("mykey", "gin-value")

		// 在 Request.Context() 设置相同 key
		ctx := context.WithValue(gc.Request.Context(), "mykey", "context-value")
		gc.Request = gc.Request.WithContext(ctx)

		c := newContext(gc)

		// c.Context().Value 走标准 context 链，返回 context 中的值
		ctxValue = c.Context().Value("mykey")
		// c.Get 走 gin 内部存储
		ginValue = c.Get("mykey")

		gc.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if ctxValue != "context-value" {
		t.Errorf("c.Context().Value should return context value, got %v", ctxValue)
	}
	if ginValue != "gin-value" {
		t.Errorf("c.Get should return gin internal value, got %v", ginValue)
	}
}

// TestGinContextValue_NonStringKey 验证非字符串 key 的处理。
//
// 测试场景：
// 1. 使用非字符串 key（如 int）
// 2. 验证正确回退到 Request.Context()
func TestGinContextValue_NonStringKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	type customKey int
	const myKey customKey = 1

	var result any

	engine.GET("/test", func(gc *gin.Context) {
		// 使用自定义类型 key
		ctx := context.WithValue(gc.Request.Context(), myKey, "custom-value")
		gc.Request = gc.Request.WithContext(ctx)

		c := newContext(gc)

		result = c.Context().Value(myKey)

		gc.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if result != "custom-value" {
		t.Errorf("expected custom-value, got %v", result)
	}
}

// TestGinContextValue_DeepNesting 验证深度嵌套不会栈溢出。
//
// 测试场景：
// 1. 创建 100 层嵌套的 context.WithValue
// 2. 验证不会栈溢出
func TestGinContextValue_DeepNesting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var result any
	var panicOccurred bool

	engine.GET("/test", func(gc *gin.Context) {
		// 创建深度嵌套的 context
		ctx := gc.Request.Context()
		for i := 0; i < 100; i++ {
			ctx = context.WithValue(ctx, i, i*10)
		}
		gc.Request = gc.Request.WithContext(ctx)

		c := newContext(gc)

		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
				t.Errorf("Panic with deep nesting: %v", r)
			}
		}()

		// 获取中间层的值
		result = c.Context().Value(50)

		gc.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if panicOccurred {
		t.Fatal("panic occurred with deep nesting")
	}

	if result != 500 {
		t.Errorf("expected 500, got %v", result)
	}
}
