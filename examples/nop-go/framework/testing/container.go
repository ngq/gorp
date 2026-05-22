// Package testing provides testing utilities for gorp framework.
// This file provides container setup helpers for integration tests.
// Creates test container with config, logger, DB, Redis capabilities.
//
// 测试包提供 gorp 框架的测试工具能力。
// 本文件提供用于集成测试的容器设置 helper。
// 创建携带 config、logger、DB、Redis 能力的测试容器。
package testing

import (
	"context"
	"os"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/provider/config"
	"github.com/ngq/gorp/framework/provider/log"
	"github.com/ngq/gorp/framework/provider/orm/gorm"
	runtimeorm "github.com/ngq/gorp/framework/provider/orm/runtime"
	sqlxprovider "github.com/ngq/gorp/framework/provider/orm/sqlx"
	"github.com/ngq/gorp/framework/provider/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/ngq/gorp/framework/container"
	"github.com/stretchr/testify/require"
)

type cleanupFunc func()

// NewTestContainer builds a container configured for tests:
// - APP_ENV=testing
// - sqlite in-memory
// - miniredis
func NewTestContainer(t *testing.T) (runtimecontract.Container, cleanupFunc) {
	require.NoError(t, ChdirRepoRoot())

	// Set test env
	_ = os.Setenv("APP_ENV", "testing")

	// miniredis
	mr := miniredis.RunT(t)
	_ = os.Setenv("REDIS_ADDR", mr.Addr())

	c := container.New()
	require.NoError(t, c.RegisterProvider(config.NewProvider()))
	require.NoError(t, c.RegisterProvider(log.NewProvider()))
	require.NoError(t, c.RegisterProvider(gorm.NewProvider()))
	require.NoError(t, c.RegisterProvider(sqlxprovider.NewProvider()))
	require.NoError(t, c.RegisterProvider(runtimeorm.NewProvider()))
	require.NoError(t, c.RegisterProvider(redis.NewProvider()))

	// ensure db is up
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.Make(datacontract.GormKey)
	require.NoError(t, err)
	_, err = c.Make(datacontract.SQLXKey)
	require.NoError(t, err)
	_, err = c.Make(datacontract.ORMBackendKey)
	require.NoError(t, err)
	_, err = c.Make(datacontract.DBRuntimeKey)
	require.NoError(t, err)
	_, err = c.Make(datacontract.MigratorKey)
	require.NoError(t, err)
	_, err = c.Make(datacontract.SQLExecutorKey)
	require.NoError(t, err)

	rAny, err := c.Make(datacontract.RedisKey)
	require.NoError(t, err)
	redisSvc, ok := rAny.(datacontract.Redis)
	require.True(t, ok, "expected datacontract.Redis, got %T", rAny)
	require.NoError(t, redisSvc.Ping(ctx))

	return c, func() {
		mr.Close()
	}
}
