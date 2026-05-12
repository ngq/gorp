// Package sqlx_test provides unit tests for the ORM sqlx provider.
//
// 适用场景：
// - 验证 ORM sqlx provider 的注册与数据库操作行为。
package sqlx

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "orm.sqlx", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{datacontract.SQLXKey}, p.Provides())
}
