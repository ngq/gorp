// Package container_test provides unit tests for service provider registration and boot order.
//
// 适用场景：
// - 验证服务商注册、引导顺序与 boot 调用的正确性。
package container

import (
	"errors"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type orderedProvider struct {
	name     string
	calls    *[]string
	failBoot error
}

func (p *orderedProvider) Name() string        { return p.name }
func (p *orderedProvider) IsDefer() bool       { return false }
func (p *orderedProvider) Provides() []string  { return []string{p.name} }
func (p *orderedProvider) DependsOn() []string { return nil }
func (p *orderedProvider) Register(c runtimecontract.Container) error {
	*p.calls = append(*p.calls, p.name+":register")
	c.Bind(p.name, func(runtimecontract.Container) (any, error) { return p.name, nil }, true)
	return nil
}
func (p *orderedProvider) Boot(runtimecontract.Container) error {
	*p.calls = append(*p.calls, p.name+":boot")
	return p.failBoot
}

func TestRegisterProvidersKeepsInputOrder(t *testing.T) {
	c := New()
	calls := []string{}

	p1 := &orderedProvider{name: "p1", calls: &calls}
	p2 := &orderedProvider{name: "p2", calls: &calls}
	p3 := &orderedProvider{name: "p3", calls: &calls}

	require.NoError(t, c.RegisterProviders(p1, p2, p3))
	require.Equal(t, []string{
		"p1:register", "p1:boot",
		"p2:register", "p2:boot",
		"p3:register", "p3:boot",
	}, calls)
}

func TestRegisterProvidersStopsAtFirstFailure(t *testing.T) {
	c := New()
	calls := []string{}

	p1 := &orderedProvider{name: "p1", calls: &calls}
	p2 := &orderedProvider{name: "p2", calls: &calls, failBoot: errors.New("boot failed")}
	p3 := &orderedProvider{name: "p3", calls: &calls}

	err := c.RegisterProviders(p1, p2, p3)
	require.EqualError(t, err, "register provider p2: boot failed")
	require.Equal(t, []string{
		"p1:register", "p1:boot",
		"p2:register", "p2:boot",
	}, calls)
}
