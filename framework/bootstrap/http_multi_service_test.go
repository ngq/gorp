package bootstrap

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/ngq/gorp/framework/container"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/provider/host"
)

type multiServiceLogger struct{}

func (multiServiceLogger) Debug(string, ...observabilitycontract.Field) {}
func (multiServiceLogger) Info(string, ...observabilitycontract.Field)  {}
func (multiServiceLogger) Warn(string, ...observabilitycontract.Field)  {}
func (multiServiceLogger) Error(string, ...observabilitycontract.Field) {}
func (multiServiceLogger) With(...observabilitycontract.Field) observabilitycontract.Logger {
	return multiServiceLogger{}
}

type testRouteSurface interface{ transportcontract.Router }

type lifecycleHTTP struct {
	testRouteSurface
	started chan struct{}
	starts  atomic.Int32
	stops   atomic.Int32
}

func (s *lifecycleHTTP) Router() transportcontract.Router { return s.testRouteSurface }
func (s *lifecycleHTTP) Server() *http.Server             { return &http.Server{} }
func (s *lifecycleHTTP) Run() error                       { return nil }
func (s *lifecycleHTTP) Start(context.Context) error {
	s.starts.Add(1)
	if s.started != nil {
		close(s.started)
	}
	return nil
}
func (s *lifecycleHTTP) Stop(context.Context) error {
	s.stops.Add(1)
	return nil
}
func (s *lifecycleHTTP) Shutdown(ctx context.Context) error        { return s.Stop(ctx) }
func (s *lifecycleHTTP) UseGlobal(...transportcontract.Middleware) {}

type lifecycleRegistry struct {
	entries []transportcontract.HTTPServiceEntry
}

func (r lifecycleRegistry) Get(name string) (transportcontract.HTTP, bool) {
	for _, entry := range r.entries {
		if entry.Name == name {
			return entry.Service, true
		}
	}
	return nil, false
}
func (r lifecycleRegistry) Default() (transportcontract.HTTP, bool) {
	return r.Get(transportcontract.DefaultHTTPServiceName)
}
func (r lifecycleRegistry) Names() []string {
	names := make([]string, 0, len(r.entries))
	for _, entry := range r.entries {
		names = append(names, entry.Name)
	}
	return names
}
func (r lifecycleRegistry) Entries() []transportcontract.HTTPServiceEntry {
	return append([]transportcontract.HTTPServiceEntry(nil), r.entries...)
}
func (r lifecycleRegistry) Use(middleware ...transportcontract.Middleware) {
	for _, entry := range r.entries {
		if entry.Service != nil {
			entry.Service.Use(middleware...)
		}
	}
}

func TestRunHTTPContextSkipsDisabledServices(t *testing.T) {
	enabled := &lifecycleHTTP{started: make(chan struct{})}
	disabled := &lifecycleHTTP{}
	registry := lifecycleRegistry{entries: []transportcontract.HTTPServiceEntry{
		{Name: "admin", Service: enabled, Enabled: true},
		{Name: "wecom", Service: disabled, Enabled: false},
	}}
	c := container.New()
	c.Bind(transportcontract.HTTPRegistryKey, func(runtimecontract.Container) (any, error) { return registry, nil }, true)
	c.Bind(runtimecontract.HostKey, func(runtimecontract.Container) (any, error) { return host.NewDefaultHost(c), nil }, true)
	t.Cleanup(func() { _ = c.Destroy() })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- RunHTTPContext(ctx, c, multiServiceLogger{}) }()
	<-enabled.started
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if enabled.starts.Load() != 1 || enabled.stops.Load() != 1 {
		t.Fatalf("enabled service lifecycle mismatch: starts=%d stops=%d", enabled.starts.Load(), enabled.stops.Load())
	}
	if disabled.starts.Load() != 0 || disabled.stops.Load() != 0 {
		t.Fatalf("disabled service must not run: starts=%d stops=%d", disabled.starts.Load(), disabled.stops.Load())
	}
}
