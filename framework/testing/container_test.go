package testing

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestNewTestContainer_BindsORMCapabilityKeys(t *testing.T) {
	c, cleanup := NewTestContainer(t)
	defer cleanup()

	backendAny, err := c.Make(datacontract.ORMBackendKey)
	require.NoError(t, err)
	require.Equal(t, string(datacontract.RuntimeBackendGorm), backendAny)

	runtimeAny, err := c.Make(datacontract.DBRuntimeKey)
	require.NoError(t, err)
	require.NotNil(t, runtimeAny)

	migratorAny, err := c.Make(datacontract.MigratorKey)
	require.NoError(t, err)
	require.NotNil(t, migratorAny)

	execAny, err := c.Make(datacontract.SQLExecutorKey)
	require.NoError(t, err)
	require.NotNil(t, execAny)
}
