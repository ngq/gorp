package dtmsdk

import (
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSAGABuilder_AddBranchKeepsBranchRetryMetadata(t *testing.T) {
	cfg := &integrationcontract.DTMConfig{Enabled: true}
	client, err := NewDTMClient(cfg)
	require.NoError(t, err)

	saga := client.SAGA("branch-options")
	saga.AddBranch("/action", "/compensate", nil, integrationcontract.BranchOptions{
		RetryCount:    3,
		RetryInterval: 5,
	})

	tx, err := saga.Build()
	require.NoError(t, err)
	require.Len(t, tx.Steps, 1)
	assert.EqualValues(t, 3, tx.Steps[0].RetryCount)
	assert.EqualValues(t, 5, tx.Steps[0].RetryInterval)
}
