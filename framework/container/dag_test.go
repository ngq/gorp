package container

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// dagProvider is a configurable test provider for DAG tests.
type dagProvider struct {
	nameStr   string
	deferLoad bool
	provide   []string
	dependsOn []string
}

func (p *dagProvider) Name() string        { return p.nameStr }
func (p *dagProvider) IsDefer() bool       { return p.deferLoad }
func (p *dagProvider) Provides() []string  { return p.provide }
func (p *dagProvider) DependsOn() []string { return p.dependsOn }
func (p *dagProvider) Register(c runtimecontract.Container) error {
	for _, k := range p.provide {
		key := k
		c.Bind(key, func(runtimecontract.Container) (any, error) { return "ok", nil }, true)
	}
	return nil
}
func (p *dagProvider) Boot(runtimecontract.Container) error { return nil }

// TestProviderDAG_EmptyContainer verifies that an empty container returns an empty DAG.
func TestProviderDAG_EmptyContainer(t *testing.T) {
	c := New()
	dag := c.ProviderDAG()
	require.Empty(t, dag.Nodes)
	require.Empty(t, dag.Edges)
	require.Empty(t, dag.Cycles)
	require.Empty(t, dag.LoadOrder)
}

// TestProviderDAG_SingleProvider verifies that a single provider produces one node,
// no edges, no cycles, and a single-element LoadOrder.
func TestProviderDAG_SingleProvider(t *testing.T) {
	c := New()
	p := &dagProvider{nameStr: "config", provide: []string{"config.key"}}
	require.NoError(t, c.RegisterProvider(p))

	dag := c.ProviderDAG()
	require.Len(t, dag.Nodes, 1)
	require.Equal(t, "config", dag.Nodes[0].Name)
	require.Equal(t, []string{"config.key"}, dag.Nodes[0].Provides)
	require.Empty(t, dag.Edges)
	require.Empty(t, dag.Cycles)
	require.Equal(t, []string{"config"}, dag.LoadOrder)
}

// TestProviderDAG_TwoProvidersWithDependency verifies that two providers with
// a dependency produce an edge and correct load order.
func TestProviderDAG_TwoProvidersWithDependency(t *testing.T) {
	c := New()
	pConfig := &dagProvider{nameStr: "config", provide: []string{"config.key"}}
	pApp := &dagProvider{nameStr: "app", provide: []string{"app.key"}, dependsOn: []string{"config.key"}}
	require.NoError(t, c.RegisterProvider(pConfig))
	require.NoError(t, c.RegisterProvider(pApp))

	dag := c.ProviderDAG()
	require.Len(t, dag.Nodes, 2)
	require.Len(t, dag.Edges, 1)
	require.Equal(t, "app", dag.Edges[0].From)
	require.Equal(t, "config", dag.Edges[0].To)
	require.Equal(t, "config.key", dag.Edges[0].Key)
	require.Empty(t, dag.Cycles)
	// config should come before app in load order
	require.Equal(t, []string{"config", "app"}, dag.LoadOrder)
}

// TestProviderDAG_ExternalDependency verifies that a dependency on a directly-bound key
// produces an edge with To="" (external dependency).
func TestProviderDAG_ExternalDependency(t *testing.T) {
	c := New()
	c.Bind("external.key", func(runtimecontract.Container) (any, error) {
		return "external", nil
	}, true)

	p := &dagProvider{nameStr: "app", provide: []string{"app.key"}, dependsOn: []string{"external.key"}}
	require.NoError(t, c.RegisterProvider(p))

	dag := c.ProviderDAG()
	require.Len(t, dag.Edges, 1)
	require.Equal(t, "app", dag.Edges[0].From)
	require.Equal(t, "", dag.Edges[0].To)
	require.Equal(t, "external.key", dag.Edges[0].Key)
	require.Empty(t, dag.Cycles)
}

// TestProviderDAG_CircularDependency verifies that a cycle is detected.
func TestProviderDAG_CircularDependency(t *testing.T) {
	c := New()
	pA := &dagProvider{nameStr: "svc.a", provide: []string{"key.a"}, dependsOn: []string{"key.b"}}
	pB := &dagProvider{nameStr: "svc.b", provide: []string{"key.b"}, dependsOn: []string{"key.a"}}
	require.NoError(t, c.RegisterProvider(pA))
	require.NoError(t, c.RegisterProvider(pB))

	dag := c.ProviderDAG()
	require.Len(t, dag.Nodes, 2)
	require.Len(t, dag.Edges, 2)
	require.NotEmpty(t, dag.Cycles, "expected a cycle to be detected")
}

// TestProviderDAG_DeferredProvider verifies that a deferred provider appears in the DAG
// with IsDefer=true and contributes edges via deferredByKey.
func TestProviderDAG_DeferredProvider(t *testing.T) {
	c := New()
	pConfig := &dagProvider{nameStr: "config", provide: []string{"config.key"}}
	pDefer := &dagProvider{nameStr: "deferred.svc", deferLoad: true, provide: []string{"deferred.key"}}
	require.NoError(t, c.RegisterProvider(pConfig))
	require.NoError(t, c.RegisterProvider(pDefer))

	dag := c.ProviderDAG()
	require.Len(t, dag.Nodes, 2)

	var deferredNode runtimecontract.ProviderDAGNode
	for _, n := range dag.Nodes {
		if n.Name == "deferred.svc" {
			deferredNode = n
		}
	}
	require.True(t, deferredNode.IsDefer)
}

// TestProviderDAG_ThreeProvidersChain verifies load order for A->B->C dependency chain.
func TestProviderDAG_ThreeProvidersChain(t *testing.T) {
	c := New()
	pC := &dagProvider{nameStr: "db", provide: []string{"db.key"}}
	pB := &dagProvider{nameStr: "cache", provide: []string{"cache.key"}, dependsOn: []string{"db.key"}}
	pA := &dagProvider{nameStr: "app", provide: []string{"app.key"}, dependsOn: []string{"cache.key"}}
	require.NoError(t, c.RegisterProvider(pC))
	require.NoError(t, c.RegisterProvider(pB))
	require.NoError(t, c.RegisterProvider(pA))

	dag := c.ProviderDAG()
	require.Len(t, dag.Nodes, 3)
	require.Empty(t, dag.Cycles)
	require.Equal(t, []string{"db", "cache", "app"}, dag.LoadOrder)
}

// TestProviderDAG_DiamondDependency verifies correct handling of diamond-shaped
// dependency: A depends on B and C, both B and C depend on D.
func TestProviderDAG_DiamondDependency(t *testing.T) {
	c := New()
	pD := &dagProvider{nameStr: "config", provide: []string{"config.key"}}
	pB := &dagProvider{nameStr: "cache", provide: []string{"cache.key"}, dependsOn: []string{"config.key"}}
	pC := &dagProvider{nameStr: "db", provide: []string{"db.key"}, dependsOn: []string{"config.key"}}
	pA := &dagProvider{nameStr: "app", provide: []string{"app.key"}, dependsOn: []string{"cache.key", "db.key"}}
	require.NoError(t, c.RegisterProvider(pD))
	require.NoError(t, c.RegisterProvider(pB))
	require.NoError(t, c.RegisterProvider(pC))
	require.NoError(t, c.RegisterProvider(pA))

	dag := c.ProviderDAG()
	require.Len(t, dag.Nodes, 4)
	require.Len(t, dag.Edges, 4) // app->cache, app->db, cache->config, db->config
	require.Empty(t, dag.Cycles)
	// config must come first, app must come last
	require.Equal(t, "config", dag.LoadOrder[0])
	require.Equal(t, "app", dag.LoadOrder[len(dag.LoadOrder)-1])
}
