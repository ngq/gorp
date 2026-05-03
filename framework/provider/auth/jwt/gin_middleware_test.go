package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
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
		httpCtx := contract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, c.Status, func() int { return c.Writer.Status() })
		AuthMiddleware(jwtSvc, "customer")(httpCtx, func() {
			c.Request = httpCtx.Request()
			c.Next()
		})
	})
	r.GET("/me", func(c *gin.Context) {
		gotClaims, ok := contract.FromJWTClaimsContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(7), gotClaims.SubjectID)

		subjectID, ok := contract.FromSubjectIDContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(7), subjectID)

		subjectType, ok := contract.FromSubjectTypeContext(c.Request.Context())
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
	claims := &contract.JWTClaims{
		SubjectID:   9,
		SubjectType: "admin",
		SubjectName: "alice",
	}
	ctx := contract.NewJWTClaimsContext(t.Context(), claims)
	ctx = contract.NewSubjectIDContext(ctx, claims.SubjectID)
	ctx = contract.NewSubjectTypeContext(ctx, claims.SubjectType)

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
