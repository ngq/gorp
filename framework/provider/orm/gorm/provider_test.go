// Package gorm_test provides unit tests for the ORM GORM provider.
//
// 适用场景：
// - 验证 ORM GORM provider 的注册与数据库操作行为。
package gorm_test

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	testinghelper "github.com/ngq/gorp/framework/testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGormProvider_AppliesMaxOpenConns_AndLogger(t *testing.T) {
	c, cleanup := testinghelper.NewTestContainer(t)
	defer cleanup()

	anyDB, err := c.Make(datacontract.GormKey)
	require.NoError(t, err)
	db := anyDB.(*gorm.DB)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	// From config/app.testing.yaml: database.max_open_conns = 5
	require.Equal(t, 5, sqlDB.Stats().MaxOpenConnections)

	// Logger should be our bridge (type lives in gorm package)
	require.NotNil(t, db.Config.Logger)
}
