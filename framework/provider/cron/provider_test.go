package cron

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "cron", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{runtimecontract.CronKey}, p.Provides())
}
