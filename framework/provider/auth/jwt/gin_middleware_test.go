package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddlewareWritesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")
	claims := jwtSvc.NewClaims(7, "customer", "alice", []string{"user"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/me", func(c *gin.Context) {
		gotClaims, ok := securitycontract.FromJWTClaimsContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(7), gotClaims.SubjectID)

		subjectID, ok := securitycontract.FromSubjectIDContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(7), subjectID)

		subjectType, ok := securitycontract.FromSubjectTypeContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, "customer", subjectType)

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestRequestContextHelpers(t *testing.T) {
	claims := &securitycontract.JWTClaims{
		SubjectID:   9,
		SubjectType: "admin",
		SubjectName: "alice",
	}
	ctx := securitycontract.NewJWTClaimsContext(t.Context(), claims)
	ctx = securitycontract.NewSubjectIDContext(ctx, claims.SubjectID)
	ctx = securitycontract.NewSubjectTypeContext(ctx, claims.SubjectType)

	gotClaims, ok := ClaimsFromRequestContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(9), gotClaims.SubjectID)

	subjectID, ok := SubjectIDFromRequestContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(9), subjectID)

	subjectType, ok := SubjectTypeFromRequestContext(ctx)
	require.True(t, ok)
	require.Equal(t, "admin", subjectType)
}

func TestSubjectIDFromContextNilSafe(t *testing.T) {
	subjectID, ok := SubjectIDFromContext(nil)
	require.False(t, ok)
	require.Equal(t, int64(0), subjectID)
}
