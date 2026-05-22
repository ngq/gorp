// Package middleware_test provides unit tests for load shedding and circuit breaker middleware.
//
// 适用场景：
// - 过载保护与限流
// - 并发槽位管理与等待
// - 熔断器状态管理与资源记录
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestLoadSheddingRejectsWhenConcurrentSlotsAreFull verifies fast failure when no concurrency slot is available.
//
// TestLoadSheddingRejectsWhenConcurrentSlotsAreFull 验证无可用并发槽位时会快速失败。
func TestLoadSheddingRejectsWhenConcurrentSlotsAreFull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, LoadShedding(1))

	started := make(chan struct{})
	release := make(chan struct{})
	router.GET("/busy", func(c *gin.Context) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-release
		c.Status(http.StatusNoContent)
	})

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/busy", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		firstDone <- recorder
	}()

	<-started

	req := httptest.NewRequest(http.MethodGet, "/busy", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}

	close(release)
	<-firstDone
}

// TestConcurrencyLimitWaitsForSlotWithinTimeout verifies queued waiting within the configured timeout budget.
//
// TestConcurrencyLimitWaitsForSlotWithinTimeout 验证会在配置的超时预算内等待空闲槽位。
func TestConcurrencyLimitWaitsForSlotWithinTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, ConcurrencyLimit(1, 100*time.Millisecond))

	started := make(chan struct{})
	release := make(chan struct{})
	router.GET("/wait", func(c *gin.Context) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-release
		c.Status(http.StatusNoContent)
	})

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/wait", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		firstDone <- recorder
	}()

	<-started

	secondDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/wait", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		secondDone <- recorder
	}()

	time.Sleep(20 * time.Millisecond)
	close(release)

	firstRecorder := <-firstDone
	secondRecorder := <-secondDone

	if firstRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected first request 204, got %d", firstRecorder.Code)
	}
	if secondRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected second request to wait and succeed with 204, got %d", secondRecorder.Code)
	}
}

// TestCircuitBreakerRejectsWhenOpen verifies that open breakers return the unified service-unavailable response.
//
// TestCircuitBreakerRejectsWhenOpen 验证熔断器打开时会返回统一的服务不可用响应。
func TestCircuitBreakerRejectsWhenOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	breaker := &stubCircuitBreaker{allowErr: errors.New("open")}
	applyTransportMiddleware(router, CircuitBreaker(breaker, ""))
	router.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

// TestCircuitBreakerRecordsSuccessAndFailureWithRouteResource verifies breaker success and failure accounting by route resource.
//
// TestCircuitBreakerRecordsSuccessAndFailureWithRouteResource 验证熔断器按路由资源记录成功与失败。
func TestCircuitBreakerRecordsSuccessAndFailureWithRouteResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	successBreaker := &stubCircuitBreaker{}
	successRouter := NewTestEngine()
	applyTransportMiddleware(successRouter, CircuitBreaker(successBreaker, ""))
	successRouter.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	successReq := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	successRecorder := httptest.NewRecorder()
	successRouter.ServeHTTP(successRecorder, successReq)

	if len(successBreaker.successResources) != 1 {
		t.Fatalf("expected one success record, got %d", len(successBreaker.successResources))
	}
	if successBreaker.successResources[0] != "http.get.orders.id" {
		t.Fatalf("unexpected success resource %q", successBreaker.successResources[0])
	}

	failureBreaker := &stubCircuitBreaker{}
	failureRouter := NewTestEngine()
	applyTransportMiddleware(failureRouter, CircuitBreaker(failureBreaker, ""))
	failureRouter.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	failureReq := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	failureRecorder := httptest.NewRecorder()
	failureRouter.ServeHTTP(failureRecorder, failureReq)

	if len(failureBreaker.failureResources) != 1 {
		t.Fatalf("expected one failure record, got %d", len(failureBreaker.failureResources))
	}
	if failureBreaker.failureResources[0] != "http.get.orders.id" {
		t.Fatalf("unexpected failure resource %q", failureBreaker.failureResources[0])
	}
}
