// Package middleware_test provides unit tests for audit middleware.
//
// 适用场景：
// - 审计日志记录
// - Actor、resource、outcome 字段记录
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestAuditMiddlewareWritesActorAndOutcomeFields verifies audit fields such as actor, locale, resource, and outcome.
//
// TestAuditMiddlewareWritesActorAndOutcomeFields 验证 actor、locale、resource 与 outcome 等审计字段。
func TestAuditMiddlewareWritesActorAndOutcomeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := newStubLogger()
	router := gin.New()
	applyTransportMiddleware(router,
		RequestIdentity(),
		Locale(DefaultLocaleOptions()),
		func(next transportcontract.Handler) transportcontract.Handler {
			return func(c transportcontract.Context) {
				claims := &securitycontract.JWTClaims{
					SubjectID:   7,
					SubjectType: "user",
					SubjectName: "alice",
					Roles:       []string{"admin", "writer"},
				}
				c.Set("jwt_claims", claims)
				c.Set("subject_id", int64(7))
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
		},
		AuditMiddleware(logger, AuditOptions{Event: "admin audit"}),
	)
	router.POST("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/orders/42?lang=en", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected audit log entry")
	}
	entry := entries[len(entries)-1]
	if entry.msg != "admin audit" {
		t.Fatalf("expected admin audit event, got %q", entry.msg)
	}
	if fieldValue(entry.fields, "kind") != "audit" {
		t.Fatalf("expected audit kind, got %v", fieldValue(entry.fields, "kind"))
	}
	if fieldValue(entry.fields, "action") != "POST" {
		t.Fatalf("expected action POST, got %v", fieldValue(entry.fields, "action"))
	}
	if fieldValue(entry.fields, "resource") != "/orders/:id" {
		t.Fatalf("expected resource /orders/:id, got %v", fieldValue(entry.fields, "resource"))
	}
	if fieldValue(entry.fields, "outcome") != "success" {
		t.Fatalf("expected outcome success, got %v", fieldValue(entry.fields, "outcome"))
	}
	if fieldValue(entry.fields, "actor_kind") != "user" {
		t.Fatalf("expected actor_kind user, got %v", fieldValue(entry.fields, "actor_kind"))
	}
	if fieldValue(entry.fields, "actor_id") != int64(7) {
		t.Fatalf("expected actor_id 7, got %v", fieldValue(entry.fields, "actor_id"))
	}
	if fieldValue(entry.fields, "actor_roles") != "admin,writer" {
		t.Fatalf("expected actor_roles admin,writer, got %v", fieldValue(entry.fields, "actor_roles"))
	}
	if fieldValue(entry.fields, "locale") != "en" {
		t.Fatalf("expected locale en, got %v", fieldValue(entry.fields, "locale"))
	}
}
