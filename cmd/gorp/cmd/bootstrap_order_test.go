package cmd

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

type bootstrapRecordingProvider struct {
	name     string
	provides []string
	calls    *[]string
}

func (p *bootstrapRecordingProvider) Name() string        { return p.name }
func (p *bootstrapRecordingProvider) IsDefer() bool       { return false }
func (p *bootstrapRecordingProvider) Provides() []string  { return p.provides }
func (p *bootstrapRecordingProvider) DependsOn() []string { return nil }
func (p *bootstrapRecordingProvider) Register(c runtimecontract.Container) error {
	*p.calls = append(*p.calls, p.name+":register")
	for _, key := range p.provides {
		value := p.name
		bindKey := key
		c.Bind(bindKey, func(runtimecontract.Container) (any, error) { return value, nil }, true)
	}
	return nil
}
func (p *bootstrapRecordingProvider) Boot(runtimecontract.Container) error {
	*p.calls = append(*p.calls, p.name+":boot")
	return nil
}

func TestBootstrapRegistersRuntimeThenExtras(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	old := readBootstrapHooks()
	defer RegisterBootstrapProviders(old.runtimeProvider, old.extraProviders...)

	calls := []string{}
	runtimeProvider := &bootstrapRecordingProvider{name: "runtime-a", provides: []string{"runtime.key"}, calls: &calls}
	extraA := &bootstrapRecordingProvider{name: "extra-a", provides: []string{"extra.a"}, calls: &calls}
	extraB := &bootstrapRecordingProvider{name: "extra-b", provides: []string{"extra.b"}, calls: &calls}

	RegisterBootstrapProviders(runtimeProvider, extraA, extraB)

	_, c, err := bootstrap()
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Equal(t, []string{
		"runtime-a:register",
		"runtime-a:boot",
		"extra-a:register",
		"extra-a:boot",
		"extra-b:register",
		"extra-b:boot",
	}, calls)

	v, err := c.Make("runtime.key")
	require.NoError(t, err)
	require.Equal(t, "runtime-a", v)
	v, err = c.Make("extra.a")
	require.NoError(t, err)
	require.Equal(t, "extra-a", v)
	v, err = c.Make("extra.b")
	require.NoError(t, err)
	require.Equal(t, "extra-b", v)
}
