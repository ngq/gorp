// Package lifecycle provides service lifecycle management for gorp framework.
package lifecycle

import (
	"context"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// RegistryHookOptions configures the service registration lifecycle hook.
type RegistryHookOptions struct {
	ServiceName string
	Address     string
	Metadata    map[string]string
	DrainDelay  time.Duration // Time to wait after deregistration to allow in-flight traffic to drain.
}

// RegistryHook implements runtimecontract.Lifecycle to automate Post-Start registration
// and Pre-Stop deregistration with traffic draining.
type RegistryHook struct {
	registry transportcontract.ServiceRegistry
	opts     RegistryHookOptions
}

// NewRegistryHook creates a new service discovery lifecycle hook.
func NewRegistryHook(registry transportcontract.ServiceRegistry, opts RegistryHookOptions) *RegistryHook {
	if opts.DrainDelay <= 0 {
		opts.DrainDelay = 500 * time.Millisecond
	}
	return &RegistryHook{
		registry: registry,
		opts:     opts,
	}
}

// OnStarting is a no-op before service start.
func (h *RegistryHook) OnStarting(ctx context.Context) error { return nil }

// OnStarted registers the service to the discovery registry only after the server is listening.
func (h *RegistryHook) OnStarted(ctx context.Context) error {
	if h.registry == nil || h.opts.ServiceName == "" || h.opts.Address == "" {
		return nil
	}
	return h.registry.Register(ctx, h.opts.ServiceName, h.opts.Address, h.opts.Metadata)
}

// OnStopping deregisters the service and pauses for the drain window before shutting down the server.
func (h *RegistryHook) OnStopping(ctx context.Context) error {
	if h.registry == nil || h.opts.ServiceName == "" || h.opts.Address == "" {
		return nil
	}
	_ = h.registry.Deregister(ctx, h.opts.ServiceName, h.opts.Address)
	if h.opts.DrainDelay > 0 {
		select {
		case <-time.After(h.opts.DrainDelay):
		case <-ctx.Done():
		}
	}
	return nil
}

// OnStopped is a no-op after service stop.
func (h *RegistryHook) OnStopped(ctx context.Context) error { return nil }
