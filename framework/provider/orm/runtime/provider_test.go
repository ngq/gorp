package runtime

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "orm.runtime", p.Name())
	require.False(t, p.IsDefer())
	require.ElementsMatch(t, []string{
		contract.ORMBackendKey,
		contract.DBRuntimeKey,
		contract.MigratorKey,
		contract.SQLExecutorKey,
	}, p.Provides())
}
