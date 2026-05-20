package jwt

import (
	"context"
	"io"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	frameworkjwt "github.com/ngq/gorp/framework/provider/auth/jwt"
	"github.com/stretchr/testify/require"
)

type exportJWTContainerStub struct {
	jwtSvc securitycontract.JWTService
}

func (s *exportJWTContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (s *exportJWTContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (s *exportJWTContainerStub) IsBind(string) bool                                      { return true }
func (s *exportJWTContainerStub) IsBindNamed(string, string) bool                         { return false }
func (s *exportJWTContainerStub) MustMake(key string) any                                 { v, _ := s.Make(key); return v }
func (s *exportJWTContainerStub) MustMakeNamed(string, string) any                        { return nil }
func (s *exportJWTContainerStub) RegisterCloser(string, io.Closer)                        {}
func (s *exportJWTContainerStub) Destroy() error                                          { return nil }
func (s *exportJWTContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo     { return nil }
func (s *exportJWTContainerStub) DebugPrint() string                                      { return "" }
func (s *exportJWTContainerStub) ProviderDAG() runtimecontract.ProviderDAG {
	return runtimecontract.ProviderDAG{}
}
func (s *exportJWTContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportJWTContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportJWTContainerStub) Make(key string) (any, error) {
	if key == securitycontract.AuthJWTKey {
		return s.jwtSvc, nil
	}
	return nil, nil
}
func (s *exportJWTContainerStub) MakeNamed(string, string) (any, error) { return nil, nil }

func TestExportedJWTHelpers(t *testing.T) {
	jwtSvc := NewService("secret", "issuer", "aud")
	containerStub := &exportJWTContainerStub{jwtSvc: jwtSvc}

	made, err := Get(containerStub)
	require.NoError(t, err)
	require.Same(t, jwtSvc, made)
	require.Same(t, jwtSvc, GetOrPanic(containerStub))

	claims := jwtSvc.NewClaims(1, "admin", "alice", []string{"owner"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)
	verified, err := made.Verify(token)
	require.NoError(t, err)
	require.Equal(t, int64(1), verified.SubjectID)

	requestCtx := securitycontract.NewJWTClaimsContext(context.Background(), verified)
	requestCtx = securitycontract.NewSubjectIDContext(requestCtx, int64(9))
	subjectID, ok := SubjectIDFromContext(requestCtx)
	require.True(t, ok)
	require.Equal(t, int64(9), subjectID)

	requestCtx = securitycontract.NewSubjectIDContext(requestCtx, verified.SubjectID)
	requestCtx = securitycontract.NewSubjectTypeContext(requestCtx, verified.SubjectType)

	gotClaims, ok := ClaimsFromRequestContext(requestCtx)
	require.True(t, ok)
	require.Equal(t, verified.SubjectID, gotClaims.SubjectID)

	gotSubjectID, ok := SubjectIDFromRequestContext(requestCtx)
	require.True(t, ok)
	require.Equal(t, verified.SubjectID, gotSubjectID)

	gotSubjectType, ok := SubjectTypeFromRequestContext(requestCtx)
	require.True(t, ok)
	require.Equal(t, verified.SubjectType, gotSubjectType)

	require.Equal(t, frameworkjwt.ContextJWTClaimsKey, ContextJWTClaimsKey)
	require.Equal(t, frameworkjwt.ContextSubjectIDKey, ContextSubjectIDKey)
	require.Equal(t, frameworkjwt.ContextSubjectTypeKey, ContextSubjectTypeKey)
}
