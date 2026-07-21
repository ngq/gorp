package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestServerFailsClosedWhenServiceAuthResolutionFails(t *testing.T) {
	c := container.New()
	c.Bind(securitycontract.ServiceAuthKey, func(runtimecontract.Container) (any, error) {
		return nil, errors.New("secret missing")
	}, true)
	server := NewServer(&transportcontract.RPCConfig{Address: "127.0.0.1:0"}, c)
	err := server.Start(context.Background())
	require.ErrorContains(t, err, "resolve service authenticator")
	require.Empty(t, server.Addr())
}

func TestServerFailsClosedWhenServiceAuthHasWrongType(t *testing.T) {
	c := container.New()
	c.Bind(securitycontract.ServiceAuthKey, func(runtimecontract.Container) (any, error) {
		return "not-an-authenticator", nil
	}, true)
	server := NewServer(&transportcontract.RPCConfig{Address: "127.0.0.1:0"}, c)
	require.ErrorContains(t, server.Start(context.Background()), "invalid type")
}
