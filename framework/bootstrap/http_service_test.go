package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutoMigrateModelsNilRuntime(t *testing.T) {
	require.NoError(t, AutoMigrateModels(nil, struct{}{}))
}

func TestAutoMigrateModelsNilDB(t *testing.T) {
	rt := &HTTPServiceRuntime{}
	require.NoError(t, AutoMigrateModels(rt, struct{}{}))
}
