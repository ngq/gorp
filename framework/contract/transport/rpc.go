package transport

import "context"

const (
	RPCClientKey   = "framework.rpc.client"
	RPCServerKey   = "framework.rpc.server"
	RPCRegistryKey = "framework.rpc.registry"
)

type RPCConfig struct {
	Mode string `mapstructure:"mode"`

	Registry string `mapstructure:"registry"`
	Address  string `mapstructure:"address"`

	BaseURL string `mapstructure:"base_url"`

	Target    string `mapstructure:"target"`
	Insecure  bool   `mapstructure:"insecure"`
	TimeoutMS int    `mapstructure:"timeout_ms"`
}

type RPCClient interface {
	Call(ctx context.Context, service, method string, req, resp any) error
	CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error)
	Close() error
}

type RPCServer interface {
	Register(service string, handler any) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Addr() string
}

type ServiceRegistry interface {
	Register(ctx context.Context, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, name, addr string) error
	Discover(ctx context.Context, name string) ([]ServiceInstance, error)
	Close() error
}

type ServiceInstance struct {
	ID       string
	Name     string
	Address  string
	Metadata map[string]string
	Healthy  bool
}
