// Package middleware_test provides unit tests for cache and compression middleware.
//
// 适用场景：
// - Cache-Control 头控制
// - ETag 生成与条件请求处理
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestCacheControlWritesHeader verifies fixed Cache-Control output.
//
// TestCacheControlWritesHeader 验证固定 Cache-Control 头输出。
func TestCacheControlWritesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, CacheControl("public, max-age=60"))
	router.GET("/cache", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/cache", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Cache-Control") != "public, max-age=60" {
		t.Fatalf("unexpected Cache-Control header %q", recorder.Header().Get("Cache-Control"))
	}
}

// TestETagWritesHeaderAndReturnsNotModifiedOnMatch verifies ETag generation and conditional 304 responses.
//
// TestETagWritesHeaderAndReturnsNotModifiedOnMatch 验证 ETag 生成与条件 304 响应。
func TestETagFlushEnablesStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, ETag())
	flushed := make(chan struct{}, 1)
	router.GET("/events", func(c *gin.Context) {
		c.Writer.Write([]byte("first"))
		c.Writer.Flush()
		flushed <- struct{}{}
		c.Writer.Write([]byte("second"))
	})

	recorder := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/events", nil))
		close(done)
	}()
	select {
	case <-flushed:
	case <-time.After(time.Second):
		t.Fatal("handler did not reach flush")
	}
	<-done
	if recorder.Body.String() != "firstsecond" {
		t.Fatalf("expected streamed body, got %q", recorder.Body.String())
	}
	if recorder.Header().Get("ETag") != "" {
		t.Fatal("streaming response must not receive an ETag")
	}
}

func TestETagOversizedResponseSpillsToWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, ETag())
	router.GET("/large", func(c *gin.Context) {
		_, _ = c.Writer.Write(make([]byte, maxETagResponseBytes+1))
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/large", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Body.Len() != maxETagResponseBytes+1 {
		t.Fatalf("expected full streamed body, got %d bytes", recorder.Body.Len())
	}
	if recorder.Header().Get("ETag") != "" {
		t.Fatal("oversized response must bypass ETag generation")
	}
}

func TestETagWritesHeaderAndReturnsNotModifiedOnMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, ETag())
	router.GET("/etag", func(c *gin.Context) {
		c.String(http.StatusOK, "payload")
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/etag", nil)
	firstRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstRecorder, firstReq)

	etag := firstRecorder.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header")
	}
	if firstRecorder.Body.String() != "payload" {
		t.Fatalf("expected payload body, got %q", firstRecorder.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/etag", nil)
	secondReq.Header.Set("If-None-Match", etag)
	secondRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondRecorder, secondReq)

	if secondRecorder.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", secondRecorder.Code)
	}
	if secondRecorder.Body.Len() != 0 {
		t.Fatalf("expected empty 304 body, got %q", secondRecorder.Body.String())
	}
	if secondRecorder.Header().Get("ETag") != etag {
		t.Fatalf("expected same ETag header, got %q", secondRecorder.Header().Get("ETag"))
	}
}
