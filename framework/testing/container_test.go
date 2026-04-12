package testing

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestNewTestContainer_BindsORMCapabilityKeys(t *testing.T) {
	c, cleanup := NewTestContainer(t)
	defer cleanup()

	backendAny, err := c.Make(contract.ORMBackendKey)
	require.NoError(t, err)
	require.Equal(t, string(contract.RuntimeBackendGorm), backendAny)

	runtimeAny, err := c.Make(contract.DBRuntimeKey)
	require.NoError(t, err)
	require.NotNil(t, runtimeAny)

	migratorAny, err := c.Make(contract.MigratorKey)
	require.NoError(t, err)
	require.NotNil(t, migratorAny)

	execAny, err := c.Make(contract.SQLExecutorKey)
	require.NoError(t, err)
	require.NotNil(t, execAny)
}
