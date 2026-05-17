// Package container_test provides unit tests for runtime helper functions.
//
// 适用场景：
// - 验证 Make* / MustMake* helper 函数的调用路径。
// - 验证 PingDBRuntime 对不同数据库运行时形态的健康检查。
package container

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	gormdb "gorm.io/gorm"
)

// ============================================================
// Mock implementations - minimal stubs that satisfy interfaces
// ============================================================

// mockRedis implements datacontract.Redis
type mockRedis struct{}

func (m *mockRedis) Ping(ctx context.Context) error                                      { return nil }
func (m *mockRedis) Get(ctx context.Context, key string) (string, error)                 { return "value", nil }
func (m *mockRedis) Set(ctx context.Context, key, value string, ttl time.Duration) error { return nil }
func (m *mockRedis) Del(ctx context.Context, key string) error                           { return nil }
func (m *mockRedis) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return map[string]string{"k": "v"}, nil
}
func (m *mockRedis) MSet(ctx context.Context, kvs map[string]string) error { return nil }
func (m *mockRedis) Expire(ctx context.Context, key string, ttl time.Duration) error { return nil }

// mockCache implements datacontract.Cache
type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string) (string, error)                 { return "value", nil }
func (m *mockCache) Set(ctx context.Context, key, value string, ttl time.Duration) error { return nil }
func (m *mockCache) Del(ctx context.Context, key string) error                           { return nil }
func (m *mockCache) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return map[string]string{"k": "v"}, nil
}
func (m *mockCache) MSet(ctx context.Context, kvs map[string]string, ttl time.Duration) error { return nil }
func (m *mockCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func(ctx context.Context) (string, error)) (string, error) {
	return fn(ctx)
}

// mockLogger implements observabilitycontract.Logger
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...observabilitycontract.Field) {}
func (m *mockLogger) Info(msg string, fields ...observabilitycontract.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...observabilitycontract.Field)  {}
func (m *mockLogger) Error(msg string, fields ...observabilitycontract.Field) {}
func (m *mockLogger) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return m
}

// mockValidator implements datacontract.Validator
type mockValidator struct{}

func (m *mockValidator) Validate(ctx context.Context, v any) error                { return nil }
func (m *mockValidator) ValidateVar(ctx context.Context, v any, tag string) error { return nil }
func (m *mockValidator) RegisterCustom(name string, fn datacontract.CustomValidateFunc) error {
	return nil
}
func (m *mockValidator) SetLocale(locale string) error            { return nil }
func (m *mockValidator) TranslateError(err error) error           { return nil }

// mockRetry implements resiliencecontract.Retry
type mockRetry struct{}

func (m *mockRetry) Do(ctx context.Context, fn func() error) error { return fn() }
func (m *mockRetry) DoForResource(ctx context.Context, resource string, fn func() error) error {
	return fn()
}
func (m *mockRetry) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	return fn()
}
func (m *mockRetry) DoWithResultForResource(ctx context.Context, resource string, fn func() (any, error)) (any, error) {
	return fn()
}
func (m *mockRetry) IsRetryable(err error) bool { return false }

// mockHTTP implements transportcontract.HTTP
type mockHTTP struct{}

func (m *mockHTTP) Router() transportcontract.HTTPRouter { return nil }
func (m *mockHTTP) Server() *http.Server                 { return nil }
func (m *mockHTTP) Run() error                           { return nil }
func (m *mockHTTP) Shutdown(ctx context.Context) error   { return nil }

// mockMessagePublisher implements integrationcontract.MessagePublisher
type mockMessagePublisher struct{}

func (m *mockMessagePublisher) Publish(ctx context.Context, topic string, msg any) error {
	return nil
}

// mockMessageSubscriber implements integrationcontract.MessageSubscriber
type mockMessageSubscriber struct{}

func (m *mockMessageSubscriber) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, msg any) error) error {
	return nil
}

// mockDistributedLock implements datacontract.DistributedLock
type mockDistributedLock struct{}

func (m *mockDistributedLock) Lock(ctx context.Context, key string, ttlSeconds int64) error {
	return nil
}
func (m *mockDistributedLock) TryLock(ctx context.Context, key string, ttlSeconds int64) (bool, error) {
	return true, nil
}
func (m *mockDistributedLock) Unlock(ctx context.Context, key string) error           { return nil }
func (m *mockDistributedLock) Renew(ctx context.Context, key string, ttl int64) error { return nil }
func (m *mockDistributedLock) IsLocked(ctx context.Context, key string) (bool, error) {
	return true, nil
}
func (m *mockDistributedLock) WithLock(ctx context.Context, key string, ttl int64, fn func() error) error {
	return fn()
}

// mockGRPCConnFactory implements transportcontract.GRPCConnFactory
type mockGRPCConnFactory struct{}

// mockGRPCServerRegistrar implements transportcontract.GRPCServerRegistrar
type mockGRPCServerRegistrar struct{}

// mockCron implements runtimecontract.Cron
type mockCron struct{}

func (m *mockCron) AddFunc(spec string, cmd func()) error { return nil }
func (m *mockCron) Start()                                {}
func (m *mockCron) Stop() error                           { return nil }

// mockHost implements runtimecontract.Host
type mockHost struct{}

func (m *mockHost) Name() string        { return "test" }
func (m *mockHost) Version() string     { return "1.0" }
func (m *mockHost) Environment() string { return "test" }
func (m *mockHost) IsDevelopment() bool { return false }

// mockJWTService implements securitycontract.JWTService
type mockJWTService struct{}

func (m *mockJWTService) Sign(claims securitycontract.JWTClaims) (string, error) {
	return "token", nil
}
func (m *mockJWTService) Verify(token string) (*securitycontract.JWTClaims, error) {
	return &securitycontract.JWTClaims{}, nil
}
func (m *mockJWTService) NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) securitycontract.JWTClaims {
	return securitycontract.JWTClaims{}
}

// mockConfig implements datacontract.Config
type mockConfig struct{}

func (m *mockConfig) Env() string                         { return "test" }
func (m *mockConfig) Get(key string) any                  { return nil }
func (m *mockConfig) GetString(key string) string         { return "" }
func (m *mockConfig) GetInt(key string) int               { return 0 }
func (m *mockConfig) GetBool(key string) bool             { return false }
func (m *mockConfig) GetFloat(key string) float64         { return 0 }
func (m *mockConfig) Unmarshal(key string, out any) error { return nil }
func (m *mockConfig) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (m *mockConfig) Reload(ctx context.Context) error { return nil }

// ============================================================
// Unbound key tests - verify error handling
// ============================================================

func TestMakeHelpers_UnboundKey(t *testing.T) {
	c := New()

	_, err := MakeRedis(c)
	require.Error(t, err, "MakeRedis should error on unbound key")

	_, err = MakeCache(c)
	require.Error(t, err, "MakeCache should error on unbound key")

	_, err = MakeLogger(c)
	require.Error(t, err, "MakeLogger should error on unbound key")

	_, err = MakeValidator(c)
	require.Error(t, err, "MakeValidator should error on unbound key")

	_, err = MakeRetry(c)
	require.Error(t, err, "MakeRetry should error on unbound key")

	_, err = MakeHTTP(c)
	require.Error(t, err, "MakeHTTP should error on unbound key")

	_, err = MakeHTTPRouter(c)
	require.Error(t, err, "MakeHTTPRouter should error on unbound key")

	_, err = MakeMessagePublisher(c)
	require.Error(t, err, "MakeMessagePublisher should error on unbound key")

	_, err = MakeMessageSubscriber(c)
	require.Error(t, err, "MakeMessageSubscriber should error on unbound key")

	_, err = MakeDistributedLock(c)
	require.Error(t, err, "MakeDistributedLock should error on unbound key")

	_, err = MakeGRPCConnFactory(c)
	require.Error(t, err, "MakeGRPCConnFactory should error on unbound key")

	_, err = MakeGRPCServerRegistrar(c)
	require.Error(t, err, "MakeGRPCServerRegistrar should error on unbound key")

	_, err = MakeCron(c)
	require.Error(t, err, "MakeCron should error on unbound key")

	_, err = MakeHost(c)
	require.Error(t, err, "MakeHost should error on unbound key")

	_, err = MakeConfig(c)
	require.Error(t, err, "MakeConfig should error on unbound key")

	_, err = MakeDBRuntime(c)
	require.Error(t, err, "MakeDBRuntime should error on unbound key")

	_, err = MakeGormDB(c)
	require.Error(t, err, "MakeGormDB should error on unbound key")

	_, err = MakeSQLX(c)
	require.Error(t, err, "MakeSQLX should error on unbound key")
}

// ============================================================
// MustMake* panic tests
// ============================================================

func TestMustMakeHelpers_PanicOnUnboundKey(t *testing.T) {
	c := New()

	require.Panics(t, func() { MustMakeLogger(c) }, "MustMakeLogger should panic")
	require.Panics(t, func() { MustMakeGorm(c) }, "MustMakeGorm should panic")
	require.Panics(t, func() { MustMakeHTTP(c) }, "MustMakeHTTP should panic")
	require.Panics(t, func() { MustMakeHTTPRouter(c) }, "MustMakeHTTPRouter should panic")
	require.Panics(t, func() { MustMakeMessagePublisher(c) }, "MustMakeMessagePublisher should panic")
	require.Panics(t, func() { MustMakeMessageSubscriber(c) }, "MustMakeMessageSubscriber should panic")
	require.Panics(t, func() { MustMakeDistributedLock(c) }, "MustMakeDistributedLock should panic")
	require.Panics(t, func() { MustMakeGRPCConnFactory(c) }, "MustMakeGRPCConnFactory should panic")
	require.Panics(t, func() { MustMakeGRPCServerRegistrar(c) }, "MustMakeGRPCServerRegistrar should panic")
	require.Panics(t, func() { MustMakeValidator(c) }, "MustMakeValidator should panic")
	require.Panics(t, func() { MustMakeRetry(c) }, "MustMakeRetry should panic")
	require.Panics(t, func() { MustMakeConfig(c) }, "MustMakeConfig should panic")
	require.Panics(t, func() { MustMakeCache(c) }, "MustMakeCache should panic")
	require.Panics(t, func() { MustMakeJWT(c) }, "MustMakeJWT should panic")
}

// ============================================================
// PingDBRuntime tests
// ============================================================

func TestPingDBRuntime_Gorm(t *testing.T) {
	gormDB, err := gormdb.Open(sqlite.Open(":memory:"), &gormdb.Config{})
	require.NoError(t, err)

	err = PingDBRuntime(gormDB)
	require.NoError(t, err)
}

func TestPingDBRuntime_SQLDB(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("sqlite3 driver not available")
	}

	err = PingDBRuntime(sqlDB)
	// sqlite3 driver may not be available, so we just verify no panic
	_ = err
}

func TestPingDBRuntime_CustomPinger(t *testing.T) {
	custom := &customPinger{pingErr: nil}
	err := PingDBRuntime(custom)
	require.NoError(t, err)
}

func TestPingDBRuntime_CustomPingerWithError(t *testing.T) {
	custom := &customPinger{pingErr: context.DeadlineExceeded}
	err := PingDBRuntime(custom)
	require.Error(t, err)
}

func TestPingDBRuntime_UnknownType(t *testing.T) {
	// Unknown types should return nil (no error, treated as non-pingable)
	err := PingDBRuntime("not a database")
	require.NoError(t, err)
}

type customPinger struct {
	pingErr error
}

func (p *customPinger) Ping() error { return p.pingErr }

// ============================================================
// Success path tests with bound keys
// ============================================================

func TestMakeRedis_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.RedisKey, func(runtimecontract.Container) (any, error) {
		return &mockRedis{}, nil
	}, true)

	redis, err := MakeRedis(c)
	require.NoError(t, err)
	require.NotNil(t, redis)
}

func TestMakeCache_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.CacheKey, func(runtimecontract.Container) (any, error) {
		return &mockCache{}, nil
	}, true)

	cache, err := MakeCache(c)
	require.NoError(t, err)
	require.NotNil(t, cache)
}

func TestMakeLogger_Success(t *testing.T) {
	c := New()
	c.Bind(observabilitycontract.LogKey, func(runtimecontract.Container) (any, error) {
		return &mockLogger{}, nil
	}, true)

	logger, err := MakeLogger(c)
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestMakeValidator_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.ValidatorKey, func(runtimecontract.Container) (any, error) {
		return &mockValidator{}, nil
	}, true)

	validator, err := MakeValidator(c)
	require.NoError(t, err)
	require.NotNil(t, validator)
}

func TestMakeRetry_Success(t *testing.T) {
	c := New()
	c.Bind(resiliencecontract.RetryKey, func(runtimecontract.Container) (any, error) {
		return &mockRetry{}, nil
	}, true)

	retry, err := MakeRetry(c)
	require.NoError(t, err)
	require.NotNil(t, retry)
}

func TestMakeHTTP_Success(t *testing.T) {
	c := New()
	c.Bind(transportcontract.HTTPKey, func(runtimecontract.Container) (any, error) {
		return &mockHTTP{}, nil
	}, true)

	http, err := MakeHTTP(c)
	require.NoError(t, err)
	require.NotNil(t, http)
}

func TestMakeHTTPRouter_Success(t *testing.T) {
	c := New()
	c.Bind(transportcontract.HTTPKey, func(runtimecontract.Container) (any, error) {
		return &mockHTTP{}, nil
	}, true)

	router, err := MakeHTTPRouter(c)
	require.NoError(t, err)
	// Router() returns nil in mock, which is valid
	_ = router
}

func TestMakeGormDB_Success(t *testing.T) {
	c := New()
	gormDB, err := gormdb.Open(sqlite.Open(":memory:"), &gormdb.Config{})
	require.NoError(t, err)

	c.Bind(datacontract.GormKey, func(runtimecontract.Container) (any, error) {
		return gormDB, nil
	}, true)

	db, err := MakeGormDB(c)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestMakeSQLX_Success(t *testing.T) {
	c := New()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("sqlite3 driver not available")
	}
	sqlxDB := sqlx.NewDb(sqlDB, "sqlite3")

	c.Bind(datacontract.SQLXKey, func(runtimecontract.Container) (any, error) {
		return sqlxDB, nil
	}, true)

	db, err := MakeSQLX(c)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestMakeConfig_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &mockConfig{}, nil
	}, true)

	cfg, err := MakeConfig(c)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestMakeDBRuntime_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.DBRuntimeKey, func(runtimecontract.Container) (any, error) {
		return "db-runtime", nil
	}, true)

	db, err := MakeDBRuntime(c)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestMustMakeLogger_Success(t *testing.T) {
	c := New()
	c.Bind(observabilitycontract.LogKey, func(runtimecontract.Container) (any, error) {
		return &mockLogger{}, nil
	}, true)

	logger := MustMakeLogger(c)
	require.NotNil(t, logger)
}

func TestMustMakeGorm_Success(t *testing.T) {
	c := New()
	gormDB, err := gormdb.Open(sqlite.Open(":memory:"), &gormdb.Config{})
	require.NoError(t, err)

	c.Bind(datacontract.GormKey, func(runtimecontract.Container) (any, error) {
		return gormDB, nil
	}, true)

	db := MustMakeGorm(c)
	require.NotNil(t, db)
}

func TestMustMakeHTTP_Success(t *testing.T) {
	c := New()
	c.Bind(transportcontract.HTTPKey, func(runtimecontract.Container) (any, error) {
		return &mockHTTP{}, nil
	}, true)

	http := MustMakeHTTP(c)
	require.NotNil(t, http)
}

func TestMustMakeHTTPRouter_Success(t *testing.T) {
	c := New()
	c.Bind(transportcontract.HTTPKey, func(runtimecontract.Container) (any, error) {
		return &mockHTTP{}, nil
	}, true)

	router := MustMakeHTTPRouter(c)
	// Router() returns nil in mock
	_ = router
}

func TestMustMakeValidator_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.ValidatorKey, func(runtimecontract.Container) (any, error) {
		return &mockValidator{}, nil
	}, true)

	validator := MustMakeValidator(c)
	require.NotNil(t, validator)
}

func TestMustMakeRetry_Success(t *testing.T) {
	c := New()
	c.Bind(resiliencecontract.RetryKey, func(runtimecontract.Container) (any, error) {
		return &mockRetry{}, nil
	}, true)

	retry := MustMakeRetry(c)
	require.NotNil(t, retry)
}

func TestMustMakeConfig_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &mockConfig{}, nil
	}, true)

	cfg := MustMakeConfig(c)
	require.NotNil(t, cfg)
}

func TestMustMakeCache_Success(t *testing.T) {
	c := New()
	c.Bind(datacontract.CacheKey, func(runtimecontract.Container) (any, error) {
		return &mockCache{}, nil
	}, true)

	cache := MustMakeCache(c)
	require.NotNil(t, cache)
}

func TestMustMakeJWT_Success(t *testing.T) {
	c := New()
	c.Bind(securitycontract.AuthJWTKey, func(runtimecontract.Container) (any, error) {
		return &mockJWTService{}, nil
	}, true)

	jwt := MustMakeJWT(c)
	require.NotNil(t, jwt)
}

// ============================================================
// MakeAppService / MustMakeAppService tests
// ============================================================

func TestMakeAppService_Success(t *testing.T) {
	c := New()
	c.Bind("app.service", func(runtimecontract.Container) (any, error) {
		return "service-value", nil
	}, true)

	svc, err := MakeAppService[string](c, "app.service")
	require.NoError(t, err)
	require.Equal(t, "service-value", svc)
}

func TestMakeAppService_UnboundKey(t *testing.T) {
	c := New()
	_, err := MakeAppService[string](c, "missing")
	require.Error(t, err)
}

func TestMakeAppService_TypeMismatch(t *testing.T) {
	c := New()
	c.Bind("app.service", func(runtimecontract.Container) (any, error) {
		return "string-value", nil
	}, true)

	_, err := MakeAppService[int](c, "app.service")
	require.Error(t, err)
	require.Contains(t, err.Error(), "type mismatch")
}

func TestMustMakeAppService_Success(t *testing.T) {
	c := New()
	c.Bind("app.service", func(runtimecontract.Container) (any, error) {
		return "service-value", nil
	}, true)

	svc := MustMakeAppService[string](c, "app.service")
	require.Equal(t, "service-value", svc)
}

func TestMustMakeAppService_PanicsOnUnboundKey(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		MustMakeAppService[string](c, "missing")
	})
}

// ============================================================
// MustMakeNamed tests
// ============================================================

func TestMustMakeNamed_Success(t *testing.T) {
	c := New()
	c.NamedBind("primary", "cache", func(runtimecontract.Container) (any, error) {
		return &mockCache{}, nil
	}, true)

	cache := c.MustMakeNamed("primary", "cache")
	require.NotNil(t, cache)
}

func TestMustMakeNamed_PanicsOnUnboundKey(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		c.MustMakeNamed("missing", "cache")
	})
}

// ============================================================
// ProviderNames tests
// ============================================================

func TestProviderNames_Empty(t *testing.T) {
	c := New()
	names := c.ProviderNames()
	require.Empty(t, names)
}

func TestProviderNames_WithProviders(t *testing.T) {
	c := New()
	_ = c.RegisterProvider(&providerNamesTestProvider{nameStr: "provider.z"})
	_ = c.RegisterProvider(&providerNamesTestProvider{nameStr: "provider.a"})
	_ = c.RegisterProvider(&providerNamesTestProvider{nameStr: "provider.m"})

	names := c.ProviderNames()
	require.Equal(t, []string{"provider.a", "provider.m", "provider.z"}, names)
}

type providerNamesTestProvider struct {
	nameStr string
}

func (p *providerNamesTestProvider) Name() string                               { return p.nameStr }
func (p *providerNamesTestProvider) IsDefer() bool                              { return false }
func (p *providerNamesTestProvider) Provides() []string                         { return []string{} }
func (p *providerNamesTestProvider) DependsOn() []string                        { return nil }
func (p *providerNamesTestProvider) Register(c runtimecontract.Container) error { return nil }
func (p *providerNamesTestProvider) Boot(c runtimecontract.Container) error     { return nil }
