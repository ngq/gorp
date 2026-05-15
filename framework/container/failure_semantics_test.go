// Package container_test provides unit tests for container failure semantics.
//
// 适用场景：
// - 验证容器在服务商注册、引导失败时的错误传播与处理行为。
package container

import (
	"errors"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type failingProvider struct {
	registerErr error
	bootErr     error
	loaded      *int
	booted      *int
}

func (p *failingProvider) Name() string       { return "failing" }
func (p *failingProvider) IsDefer() bool      { return false }
func (p *failingProvider) Provides() []string { return nil }
func (p *failingProvider) DependsOn() []string { return nil }
func (p *failingProvider) Register(runtimecontract.Container) error {
	*p.loaded++
	return p.registerErr
}
func (p *failingProvider) Boot(runtimecontract.Container) error {
	*p.booted++
	return p.bootErr
}

// TestContainer_LoadProviderRetriesAfterRegisterFailure verifies that after
// a provider's Register fails, subsequent load attempts still retry Register.
//
// TestContainer_LoadProviderRetriesAfterRegisterFailure 验证当服务提供商的
// Register 失败后，后续加载尝试仍会重试 Register。
func TestContainer_LoadProviderRetriesAfterRegisterFailure(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &failingProvider{
		registerErr: errors.New("register failed"),
		loaded:      &loaded,
		booted:      &booted,
	}

	// 先只登记 provider，不走 RegisterProvider 的整链路
	c.providersByName[p.Name()] = &providerState{p: p}

	err := c.loadProvider(p.Name())
	require.EqualError(t, err, "register failed")
	require.Equal(t, 1, loaded)
	require.Equal(t, 0, booted)

	// 再次触发 loadProvider，应该再次执行 Register
	err = c.loadProvider(p.Name())
	require.EqualError(t, err, "register failed")
	require.Equal(t, 2, loaded)
	require.Equal(t, 0, booted)
}

// TestContainer_BootProviderRetriesAfterBootFailure verifies that after a
// provider's Boot fails, subsequent boot attempts still retry Boot.
//
// TestContainer_BootProviderRetriesAfterBootFailure 验证当服务提供商的 Boot
// 失败后，后续引导尝试仍会重试 Boot。
func TestContainer_BootProviderRetriesAfterBootFailure(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &failingProvider{
		bootErr: errors.New("boot failed"),
		loaded:  &loaded,
		booted:  &booted,
	}

	c.providersByName[p.Name()] = &providerState{p: p, loaded: true}

	err := c.bootProvider(p.Name())
	require.EqualError(t, err, "boot failed")
	require.Equal(t, 0, loaded)
	require.Equal(t, 1, booted)

	// 再次触发 bootProvider，应该再次执行 Boot
	err = c.bootProvider(p.Name())
	require.EqualError(t, err, "boot failed")
	require.Equal(t, 0, loaded)
	require.Equal(t, 2, booted)
}
