package gin

import (
	"testing"

	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

type providerResponderStub struct{}

func (providerResponderStub) Success(transportcontract.HTTPContext, any)                    {}
func (providerResponderStub) SuccessWithMessage(transportcontract.HTTPContext, string, any) {}
func (providerResponderStub) SuccessWithStatus(transportcontract.HTTPContext, int, any)     {}
func (providerResponderStub) Error(transportcontract.HTTPContext, error)                    {}
func (providerResponderStub) BadRequest(transportcontract.HTTPContext, string)              {}
func (providerResponderStub) InternalError(transportcontract.HTTPContext, string)           {}

func TestProviderRegistersDefaultResponderWhenMissing(t *testing.T) {
	c := container.New()

	err := NewProvider().Register(c)
	require.NoError(t, err)

	v, err := c.Make(transportcontract.HTTPResponderKey)
	require.NoError(t, err)
	_, ok := v.(transportcontract.HTTPResponder)
	require.True(t, ok)
	require.IsType(t, DefaultResponder{}, v)
}

func TestProviderDoesNotOverrideBusinessResponder(t *testing.T) {
	c := container.New()
	custom := providerResponderStub{}
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return custom, nil
	}, true)

	err := NewProvider().Register(c)
	require.NoError(t, err)

	v, err := c.Make(transportcontract.HTTPResponderKey)
	require.NoError(t, err)
	require.IsType(t, providerResponderStub{}, v)
}
