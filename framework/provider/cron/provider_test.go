package cron

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "cron", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{contract.CronKey}, p.Provides())
}
