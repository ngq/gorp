// Package middleware_test provides unit tests for authorization and logging middleware.
//
// 适用场景：
// - 认证与授权检查
// - 访问日志记录
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestAuthorizationMiddleware verifies unauthorized rejection, successful role checks, and forbidden rejection.
//
// TestAuthorizationMiddleware 验证未认证拒绝、角色校验成功以及无权限拒绝。
func TestAuthorizationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	unauthorizedRouter := gin.New()
	applyTransportMiddleware(unauthorizedRouter, RequireAuthorization())
	unauthorizedRouter.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	unauthorizedRecorder := httptest.NewRecorder()
	unauthorizedRouter.ServeHTTP(unauthorizedRecorder, unauthorizedReq)

	if unauthorizedRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthorizedRecorder.Code)
	}

	roleRouter := gin.New()
	applyTransportMiddleware(roleRouter, func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			claims := &securitycontract.JWTClaims{
				SubjectID:   1,
				SubjectType: "user",
				Roles:       []string{"admin", "writer"},
			}
			c.Set("jwt_claims", claims)
			c.Set("subject_id", int64(1))
			c.Set("subject_type", "user")
			// Also update gin.Request.Context for context.Context value propagation
			if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
				ctx := gc.Request.Context()
				ctx = securitycontract.NewJWTClaimsContext(ctx, claims)
				ctx = securitycontract.NewSubjectIDContext(ctx, claims.SubjectID)
				ctx = securitycontract.NewSubjectTypeContext(ctx, claims.SubjectType)
				gc.Request = gc.Request.WithContext(ctx)
			}
			next(c)
		}
	}, RequireAnyRole("admin"), RequireAllRoles("admin", "writer"), RequireSubjectType("user"))
	roleRouter.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	roleReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	roleRecorder := httptest.NewRecorder()
	roleRouter.ServeHTTP(roleRecorder, roleReq)

	if roleRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected authorized request 204, got %d", roleRecorder.Code)
	}

	forbiddenRouter := gin.New()
	applyTransportMiddleware(forbiddenRouter, func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			claims := &securitycontract.JWTClaims{
				SubjectID:   1,
				SubjectType: "service",
				Roles:       []string{"reader"},
			}
			c.Set("jwt_claims", claims)
			c.Set("subject_id", int64(1))
			c.Set("subject_type", "service")
			// Also update gin.Request.Context for context.Context value propagation
			if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
				ctx := gc.Request.Context()
				ctx = securitycontract.NewJWTClaimsContext(ctx, claims)
				ctx = securitycontract.NewSubjectIDContext(ctx, claims.SubjectID)
				ctx = securitycontract.NewSubjectTypeContext(ctx, claims.SubjectType)
				gc.Request = gc.Request.WithContext(ctx)
			}
			next(c)
		}
	}, RequireAnyRole("admin"))
	forbiddenRouter.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	forbiddenReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	forbiddenRecorder := httptest.NewRecorder()
	forbiddenRouter.ServeHTTP(forbiddenRecorder, forbiddenReq)

	if forbiddenRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", forbiddenRecorder.Code)
	}
}

// TestLoggingMiddlewareWritesAccessLogWithRequestIdentity verifies access log fields and propagated request identity.
//
// TestLoggingMiddlewareWritesAccessLogWithRequestIdentity 验证访问日志字段与传播的请求标识。
func TestLoggingMiddlewareWritesAccessLogWithRequestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := newStubLogger()
	router := gin.New()
	applyTransportMiddleware(router, RequestIdentity(), LoggingMiddleware(logger))
	router.GET("/logs/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/logs/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected log entries")
	}
	entry := entries[len(entries)-1]
	if entry.msg != "http request" {
		t.Fatalf("expected http request log, got %q", entry.msg)
	}
	if fieldValue(entry.fields, "method") != http.MethodGet {
		t.Fatalf("expected method GET, got %v", fieldValue(entry.fields, "method"))
	}
	if fieldValue(entry.fields, "path") != "/logs/42" {
		t.Fatalf("expected path /logs/42, got %v", fieldValue(entry.fields, "path"))
	}
	if fieldValue(entry.fields, "route") != "/logs/:id" {
		t.Fatalf("expected route /logs/:id, got %v", fieldValue(entry.fields, "route"))
	}
	if fieldValue(entry.fields, "status") != http.StatusNoContent {
		t.Fatalf("expected status 204, got %v", fieldValue(entry.fields, "status"))
	}
	if fieldValue(entry.fields, "request_id") == nil {
		t.Fatal("expected request_id field")
	}
	if fieldValue(entry.fields, "trace_id") == nil {
		t.Fatal("expected trace_id field")
	}
}
