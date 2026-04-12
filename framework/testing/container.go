package testing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
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
func NewTestContainer(t *testing.T) (contract.Container, cleanupFunc) {
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

	_, err := c.Make(contract.GormKey)
	require.NoError(t, err)
	_, err = c.Make(contract.SQLXKey)
	require.NoError(t, err)
	_, err = c.Make(contract.ORMBackendKey)
	require.NoError(t, err)
	_, err = c.Make(contract.DBRuntimeKey)
	require.NoError(t, err)
	_, err = c.Make(contract.MigratorKey)
	require.NoError(t, err)
	_, err = c.Make(contract.SQLExecutorKey)
	require.NoError(t, err)

	rAny, err := c.Make(contract.RedisKey)
	require.NoError(t, err)
	require.NoError(t, rAny.(contract.Redis).Ping(ctx))

	return c, func() {
		mr.Close()
	}
}
