// Package jwt_test provides unit tests for JWT auth provider registration and config binding.
//
// 适用场景：
// - 验证 JWT auth provider 的注册、IsDefer、Provides 等接口契约。
// - 确保 Container 绑定和配置注入逻辑正确。
package jwt

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

type stubConfig struct {
	values map[string]string
}

func (s *stubConfig) Env() string                 { return "testing" }
func (s *stubConfig) Get(key string) any          { return s.values[key] }
func (s *stubConfig) GetString(key string) string { return s.values[key] }
func (s *stubConfig) GetInt(string) int           { return 0 }
func (s *stubConfig) GetBool(string) bool         { return false }
func (s *stubConfig) GetFloat(string) float64     { return 0 }
func (s *stubConfig) Unmarshal(string, any) error { return nil }
func (s *stubConfig) Watch(_ context.Context, _ string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (s *stubConfig) Reload(_ context.Context) error { return nil }

// TestProviderMeta verifies the JWT provider metadata including name, defer mode, and provided keys.
//
// TestProviderMeta 验证 JWT provider 元信息，包括名称、延迟模式及提供的 key 列表。
func TestProviderMeta(t *testing.T) {
	p := NewProvider()
	if p.Name() != "auth.jwt" {
		t.Fatalf("unexpected provider name: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Fatal("auth.jwt provider should be defer")
	}
	if got := p.Provides(); len(got) != 1 || got[0] != securitycontract.AuthJWTKey {
		t.Fatalf("unexpected provides: %#v", got)
	}
}

// TestProviderBindJWTService verifies the JWT provider can bind and create a JWTService from config.
//
// TestProviderBindJWTService 验证 JWT provider 能从配置绑定并创建 JWTService 实例。
func TestProviderBindJWTService(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &stubConfig{values: map[string]string{
			"auth.jwt.secret": "s1",
			"auth.jwt.issuer": "issuer-1",
		}}, nil
	}, true)

	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}

	v, err := c.Make(securitycontract.AuthJWTKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v.(securitycontract.JWTService); !ok {
		t.Fatalf("expected securitycontract.JWTService, got %T", v)
	}
}

// TestProviderCompatLegacySecretKey verifies the JWT provider supports legacy secret key "auth.jwt_secret".
//
// TestProviderCompatLegacySecretKey 验证 JWT provider 兼容旧版密钥 "auth.jwt_secret"。
func TestProviderCompatLegacySecretKey(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &stubConfig{values: map[string]string{
			"auth.jwt_secret": "legacy-secret",
		}}, nil
	}, true)

	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}

	v, err := c.Make(securitycontract.AuthJWTKey)
	if err != nil {
		t.Fatal(err)
	}
	svc := v.(securitycontract.JWTService)
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Verify(token); err != nil {
		t.Fatalf("expected token verify pass with legacy key, got err: %v", err)
	}
}

// TestJWTService_SignRejectsEmptySecret verifies that signing fails when secret is empty.
//
// TestJWTService_SignRejectsEmptySecret 验证空 secret 时签发会失败。
func TestJWTService_SignRejectsEmptySecret(t *testing.T) {
	svc := NewJWTService("", "issuer", "aud")
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	_, err := svc.Sign(claims)
	if err == nil {
		t.Fatal("expected error when signing with empty secret")
	}
}

// TestJWTService_VerifyRejectsEmptySecret verifies that verification fails when secret is empty.
//
// TestJWTService_VerifyRejectsEmptySecret 验证空 secret 时校验会失败。
func TestJWTService_VerifyRejectsEmptySecret(t *testing.T) {
	svc := NewJWTService("", "issuer", "aud")
	_, err := svc.Verify("some.token.value")
	if err == nil {
		t.Fatal("expected error when verifying with empty secret")
	}
}

// TestJWTService_VerifyRejectsInvalidToken verifies that verification fails for malformed tokens.
//
// TestJWTService_VerifyRejectsInvalidToken 验证畸形 token 校验会失败。
func TestJWTService_VerifyRejectsInvalidToken(t *testing.T) {
	svc := NewJWTService("secret", "issuer", "aud")
	cases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "abc"},
		{"two parts", "abc.def"},
		{"four parts", "a.b.c.d"},
		{"invalid base64 payload", "abc.!!!.def"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Verify(tc.token)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

// TestJWTService_VerifyRejectsWrongSecret verifies that a token signed with a different secret is rejected.
//
// TestJWTService_VerifyRejectsWrongSecret 验证使用不同 secret 签发的 token 会被拒绝。
func TestJWTService_VerifyRejectsWrongSecret(t *testing.T) {
	signer := NewJWTService("secret-a", "issuer", "aud")
	verifier := NewJWTService("secret-b", "issuer", "aud")
	claims := signer.NewClaims(1, "user", "u1", nil, 60)
	token, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	_, err = verifier.Verify(token)
	if err == nil {
		t.Fatal("expected error when verifying token signed with different secret")
	}
}

// TestJWTService_VerifyRejectsExpiredToken verifies that expired tokens are rejected.
//
// TestJWTService_VerifyRejectsExpiredToken 验证过期 token 会被拒绝。
func TestJWTService_VerifyRejectsExpiredToken(t *testing.T) {
	svc := NewJWTService("secret", "issuer", "aud")
	// 构建一个已过期的 claims，ExpiresAt 设为过去的时间
	claims := securitycontract.JWTClaims{
		SubjectID:   1,
		SubjectType: "user",
		SubjectName: "u1",
		Roles:       nil,
		ExpiresAt:   1, // 1970-01-01，已经过期
		IssuedAt:    1,
		Issuer:      "issuer",
		Audience:    "aud",
	}
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Verify(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

// TestJWTService_VerifyRejectsZeroSubjectID verifies that tokens with zero SubjectID are rejected.
//
// TestJWTService_VerifyRejectsZeroSubjectID 验证 SubjectID 为零值的 token 会被拒绝。
func TestJWTService_VerifyRejectsZeroSubjectID(t *testing.T) {
	svc := NewJWTService("secret", "issuer", "aud")
	claims := securitycontract.JWTClaims{
		SubjectID:   0, // 零值
		SubjectType: "user",
		SubjectName: "u1",
		ExpiresAt:   9999999999,
		IssuedAt:    1,
		Issuer:      "issuer",
		Audience:    "aud",
	}
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Verify(token)
	if err == nil {
		t.Fatal("expected error for zero SubjectID token")
	}
}

// TestProviderDefaultFallbackSecret verifies that when no secret is configured, the provider uses the default fallback.
//
// TestProviderDefaultFallbackSecret 验证未配置 secret 时，provider 使用默认回退 secret。
func TestProviderDefaultFallbackSecret(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &stubConfig{values: map[string]string{}}, nil
	}, true)

	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}

	v, err := c.Make(securitycontract.AuthJWTKey)
	if err != nil {
		t.Fatal(err)
	}
	svc := v.(securitycontract.JWTService)
	// 用默认 secret 也能正常签发和校验
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Verify(token); err != nil {
		t.Fatalf("expected token verify pass with default secret, got err: %v", err)
	}
}

// TestJWTService_VerifyRejectsIssuerMismatch verifies that tokens with mismatched issuer are rejected.
//
// TestJWTService_VerifyRejectsIssuerMismatch 验证 issuer 不匹配的 token 会被拒绝。
func TestJWTService_VerifyRejectsIssuerMismatch(t *testing.T) {
	// 创建一个配置了 issuer 的服务
	svc := NewJWTService("secret", "expected-issuer", "")

	// 用正确的 issuer 签发 token
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}

	// 用配置了不同 issuer 的验证器校验
	wrongIssuerSvc := NewJWTService("secret", "wrong-issuer", "")
	_, err = wrongIssuerSvc.Verify(token)
	if err == nil {
		t.Fatal("expected error when issuer mismatch")
	}
}

// TestJWTService_VerifyAcceptsMatchingIssuer verifies that tokens with matching issuer are accepted.
//
// TestJWTService_VerifyAcceptsMatchingIssuer 验证 issuer 匹配的 token 会被接受。
func TestJWTService_VerifyAcceptsMatchingIssuer(t *testing.T) {
	svc := NewJWTService("secret", "my-issuer", "")
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	// 同一个服务验证，issuer 匹配
	verified, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("expected token verify pass, got err: %v", err)
	}
	if verified.Issuer != "my-issuer" {
		t.Fatalf("expected issuer 'my-issuer', got '%s'", verified.Issuer)
	}
}

// TestJWTService_VerifyRejectsAudienceMismatch verifies that tokens with mismatched audience are rejected.
//
// TestJWTService_VerifyRejectsAudienceMismatch 验证 audience 不匹配的 token 会被拒绝。
func TestJWTService_VerifyRejectsAudienceMismatch(t *testing.T) {
	// 创建一个配置了 audience 的服务
	svc := NewJWTService("secret", "", "expected-audience")

	// 用正确的 audience 签发 token
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}

	// 用配置了不同 audience 的验证器校验
	wrongAudSvc := NewJWTService("secret", "", "wrong-audience")
	_, err = wrongAudSvc.Verify(token)
	if err == nil {
		t.Fatal("expected error when audience mismatch")
	}
}

// TestJWTService_VerifyAcceptsMatchingAudience verifies that tokens with matching audience are accepted.
//
// TestJWTService_VerifyAcceptsMatchingAudience 验证 audience 匹配的 token 会被接受。
func TestJWTService_VerifyAcceptsMatchingAudience(t *testing.T) {
	svc := NewJWTService("secret", "", "my-audience")
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	// 同一个服务验证，audience 匹配
	verified, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("expected token verify pass, got err: %v", err)
	}
	if verified.Audience != "my-audience" {
		t.Fatalf("expected audience 'my-audience', got '%s'", verified.Audience)
	}
}

// TestJWTService_VerifySkipsIssuerCheckWhenEmpty verifies that issuer check is skipped when service has no issuer configured.
//
// TestJWTService_VerifySkipsIssuerCheckWhenEmpty 验证服务未配置 issuer 时跳过 issuer 校验。
func TestJWTService_VerifySkipsIssuerCheckWhenEmpty(t *testing.T) {
	// 服务未配置 issuer
	svc := NewJWTService("secret", "", "")

	// 手动构建一个带任意 issuer 的 claims
	claims := securitycontract.JWTClaims{
		SubjectID:   1,
		SubjectType: "user",
		SubjectName: "u1",
		ExpiresAt:   9999999999,
		IssuedAt:    1,
		Issuer:      "any-issuer", // 任意 issuer
	}
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	// 服务未配置 issuer，应该接受任意 issuer
	_, err = svc.Verify(token)
	if err != nil {
		t.Fatalf("expected token verify pass when no issuer configured, got err: %v", err)
	}
}

// TestJWTService_VerifySkipsAudienceCheckWhenEmpty verifies that audience check is skipped when service has no audience configured.
//
// TestJWTService_VerifySkipsAudienceCheckWhenEmpty 验证服务未配置 audience 时跳过 audience 校验。
func TestJWTService_VerifySkipsAudienceCheckWhenEmpty(t *testing.T) {
	// 服务未配置 audience
	svc := NewJWTService("secret", "", "")

	// 手动构建一个带任意 audience 的 claims
	claims := securitycontract.JWTClaims{
		SubjectID:   1,
		SubjectType: "user",
		SubjectName: "u1",
		ExpiresAt:   9999999999,
		IssuedAt:    1,
		Audience:    "any-audience", // 任意 audience
	}
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	// 服务未配置 audience，应该接受任意 audience
	_, err = svc.Verify(token)
	if err != nil {
		t.Fatalf("expected token verify pass when no audience configured, got err: %v", err)
	}
}
