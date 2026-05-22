// Package testing_test provides unit tests for test container and ORM capability bindings.
//
// 适用场景：
// - 验证 NewTestContainer 对 ORM capability keys 的绑定行为。
// - 确保测试容器正确注入数据库 mock 等依赖。
package testing

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

// TestNewTestContainer_BindsORMCapabilityKeys 验证测试容器正确绑定 ORM 相关 capability keys。
//
// 中文说明：
// - NewTestContainer 绑定 ORMBackendKey、DBRuntimeKey、MigratorKey、SQLExecutorKey。
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
