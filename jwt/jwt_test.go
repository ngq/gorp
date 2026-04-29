package jwt

import (
	"testing"

	"github.com/gin-gonic/gin"
	frameworkjwt "github.com/ngq/gorp/framework/provider/auth/jwt"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type exportJWTContainerStub struct {
	jwtSvc contract.JWTService
}

func (s *exportJWTContainerStub) Bind(string, contract.Factory, bool)                {}
func (s *exportJWTContainerStub) IsBind(string) bool                                 { return true }
func (s *exportJWTContainerStub) MustMake(key string) any                            { v, _ := s.Make(key); return v }
func (s *exportJWTContainerStub) RegisterProvider(contract.ServiceProvider) error     { return nil }
func (s *exportJWTContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }
func (s *exportJWTContainerStub) Make(key string) (any, error) {
	if key == contract.AuthJWTKey {
		return s.jwtSvc, nil
	}
	return nil, nil
}

func TestExportedJWTHelpers(t *testing.T) {
	jwtSvc := NewService("secret", "issuer", "aud")
	containerStub := &exportJWTContainerStub{jwtSvc: jwtSvc}

	made, err := Make(containerStub)
	require.NoError(t, err)
	require.Same(t, jwtSvc, made)
	require.Same(t, jwtSvc, MustMake(containerStub))

	claims := jwtSvc.NewClaims(1, "admin", "alice", []string{"owner"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)
	verified, err := made.Verify(token)
	require.NoError(t, err)
	require.Equal(t, int64(1), verified.SubjectID)

	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set(ContextSubjectIDKey, int64(9))
	subjectID, ok := SubjectIDFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(9), subjectID)

	require.Equal(t, frameworkjwt.ContextJWTClaimsKey, ContextJWTClaimsKey)
	require.Equal(t, frameworkjwt.ContextSubjectIDKey, ContextSubjectIDKey)
	require.Equal(t, frameworkjwt.ContextSubjectTypeKey, ContextSubjectTypeKey)
}
