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
