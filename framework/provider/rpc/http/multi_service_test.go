package http

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type captureRouter struct {
	transportcontract.Router
	method string
	path   string
}

func (r *captureRouter) POST(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.method = "POST"
	r.path = path
}

type registryHTTP struct {
	transportcontract.HTTP
	router transportcontract.Router
}

func (s *registryHTTP) Router() transportcontract.Router { return s.router }

type rpcHTTPRegistry struct {
	services map[string]transportcontract.HTTP
}

func (r rpcHTTPRegistry) Get(name string) (transportcontract.HTTP, bool) {
	service, ok := r.services[name]
	return service, ok
}
func (r rpcHTTPRegistry) Default() (transportcontract.HTTP, bool) {
	return r.Get(transportcontract.DefaultHTTPServiceName)
}
func (r rpcHTTPRegistry) Names() []string {
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
func (r rpcHTTPRegistry) Entries() []transportcontract.HTTPServiceEntry { return nil }
func (r rpcHTTPRegistry) Use(...transportcontract.Middleware)           {}

func TestServerRegistersRoutesOnConfiguredHTTPService(t *testing.T) {
	adminRouter := &captureRouter{}
	wecomRouter := &captureRouter{}
	registry := rpcHTTPRegistry{services: map[string]transportcontract.HTTP{
		"admin": &registryHTTP{router: adminRouter},
		"wecom": &registryHTTP{router: wecomRouter},
	}}
	c := container.New()
	c.Bind(transportcontract.HTTPRegistryKey, func(runtimecontract.Container) (any, error) { return registry, nil }, true)
	t.Cleanup(func() { _ = c.Destroy() })

	server := NewServer(&transportcontract.RPCConfig{HTTPService: "wecom"}, c)
	_ = server.Register("events", transportcontract.Handler(func(transportcontract.Context) {}))
	if err := server.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if adminRouter.path != "" || wecomRouter.method != "POST" || wecomRouter.path != "/rpc/events" {
		t.Fatalf("RPC route registered on wrong service: admin=%q wecom=%q", adminRouter.path, wecomRouter.path)
	}
}

func TestServerRejectsMissingHTTPServiceTarget(t *testing.T) {
	c := container.New()
	registry := rpcHTTPRegistry{services: map[string]transportcontract.HTTP{}}
	c.Bind(transportcontract.HTTPRegistryKey, func(runtimecontract.Container) (any, error) { return registry, nil }, true)
	t.Cleanup(func() { _ = c.Destroy() })

	server := NewServer(&transportcontract.RPCConfig{HTTPService: "missing"}, c)
	if err := server.Start(context.Background()); err == nil {
		t.Fatal("expected missing HTTP service target to fail")
	}
}
