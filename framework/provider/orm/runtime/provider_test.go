// Package runtime_test provides unit tests for the ORM runtime provider.
//
// 适用场景：
// - 验证 ORM runtime provider 的注册与迁移行为。
package runtime

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "orm.runtime", p.Name())
	require.False(t, p.IsDefer())
	require.ElementsMatch(t, []string{
		datacontract.ORMBackendKey,
		datacontract.DBRuntimeKey,
		datacontract.MigratorKey,
		datacontract.SQLExecutorKey,
	}, p.Provides())
}
