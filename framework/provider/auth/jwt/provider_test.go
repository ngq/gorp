package jwt

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
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
func (s *stubConfig) Watch(_ context.Context, _ string) (contract.ConfigWatcher, error) {
	return nil, nil
}
func (s *stubConfig) Reload(_ context.Context) error { return nil }

func TestProviderMeta(t *testing.T) {
	p := NewProvider()
	if p.Name() != "auth.jwt" {
		t.Fatalf("unexpected provider name: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Fatal("auth.jwt provider should be defer")
	}
	if got := p.Provides(); len(got) != 1 || got[0] != contract.AuthJWTKey {
		t.Fatalf("unexpected provides: %#v", got)
	}
}

func TestProviderBindJWTService(t *testing.T) {
	c := container.New()
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return &stubConfig{values: map[string]string{
			"auth.jwt.secret": "s1",
			"auth.jwt.issuer": "issuer-1",
		}}, nil
	}, true)

	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}

	v, err := c.Make(contract.AuthJWTKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v.(contract.JWTService); !ok {
		t.Fatalf("expected contract.JWTService, got %T", v)
	}
}

func TestProviderCompatLegacySecretKey(t *testing.T) {
	c := container.New()
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return &stubConfig{values: map[string]string{
			"auth.jwt_secret": "legacy-secret",
		}}, nil
	}, true)

	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}

	v, err := c.Make(contract.AuthJWTKey)
	if err != nil {
		t.Fatal(err)
	}
	svc := v.(contract.JWTService)
	claims := svc.NewClaims(1, "user", "u1", nil, 60)
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Verify(token); err != nil {
		t.Fatalf("expected token verify pass with legacy key, got err: %v", err)
	}
}
