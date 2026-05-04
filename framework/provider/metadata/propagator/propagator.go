package propagator

import (
	"context"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type DefaultPropagator struct {
	propagatePrefix  []string
	constantMetadata map[string]string
}

func NewDefaultPropagator(prefix []string, constant map[string]string) *DefaultPropagator {
	if len(prefix) == 0 {
		prefix = []string{"x-md-"}
	}
	if constant == nil {
		constant = make(map[string]string)
	}
	return &DefaultPropagator{
		propagatePrefix:  prefix,
		constantMetadata: constant,
	}
}

func (p *DefaultPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
	for k, v := range p.constantMetadata {
		carrier.Set(k, v)
	}

	if clientMD, ok := transportcontract.FromClientContext(ctx); ok {
		clientMD.Range(func(key string, values []string) bool {
			for _, v := range values {
				carrier.Add(key, v)
			}
			return true
		})
	}

	if serverMD, ok := transportcontract.FromServerContext(ctx); ok {
		serverMD.Range(func(key string, values []string) bool {
			if p.matchPrefix(key) {
				for _, v := range values {
					carrier.Add(key, v)
				}
			}
			return true
		})
	}
}

func (p *DefaultPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	md := transportcontract.NewMetadata()

	for _, key := range carrier.Keys() {
		if p.matchPrefix(key) {
			values := carrier.Values(key)
			for _, v := range values {
				md.Add(key, v)
			}
		}
	}

	for k, v := range p.constantMetadata {
		md.Set(k, v)
	}

	return transportcontract.NewServerContext(ctx, md)
}

func (p *DefaultPropagator) matchPrefix(key string) bool {
	if len(p.propagatePrefix) == 0 {
		return true
	}
	lowerKey := strings.ToLower(key)
	for _, prefix := range p.propagatePrefix {
		if strings.HasPrefix(lowerKey, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

type NoopPropagator struct{}

func NewNoopPropagator() *NoopPropagator { return &NoopPropagator{} }

func (p *NoopPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
}

func (p *NoopPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	return ctx
}
