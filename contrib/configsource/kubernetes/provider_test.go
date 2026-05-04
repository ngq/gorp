package kubernetes

import (
	"context"
	"errors"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.kubernetes", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}

func TestConfigSourceLoadAndGet(t *testing.T) {
	client := &fakeConfigMapClient{
		data: map[string]string{
			"app.yaml": "server:\n  port: 8080\nfeature:\n  enabled: true\n",
		},
	}

	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
		DataKey:       "app.yaml",
		AutoReload:    true,
	}, client)
	require.NoError(t, err)

	loaded, err := source.Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, 8080, loaded["server"].(map[string]any)["port"])

	value, err := source.Get(context.Background(), "feature.enabled")
	require.NoError(t, err)
	require.Equal(t, true, value)
}

func TestConfigSourceWatchDispatchesUpdatedValue(t *testing.T) {
	client := &fakeConfigMapClient{
		data: map[string]string{
			"app.yaml": "feature:\n  enabled: false\n",
		},
	}

	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
		DataKey:       "app.yaml",
		AutoReload:    true,
	}, client)
	require.NoError(t, err)
	_, err = source.Load(context.Background())
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)
	defer watcher.Stop()

	changed := make(chan any, 1)
	watcher.OnChange("feature.enabled", func(value any) {
		changed <- value
	})

	select {
	case value := <-changed:
		require.Equal(t, false, value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial config callback")
	}

	client.push(map[string]string{
		"app.yaml": "feature:\n  enabled: true\n",
	})

	select {
	case value := <-changed:
		require.Equal(t, true, value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected config change callback")
	}
}

func TestConfigSourceWatchDispatchesInitialCachedValue(t *testing.T) {
	client := &fakeConfigMapClient{
		data: map[string]string{
			"app.yaml": "feature:\n  enabled: true\n",
		},
	}

	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
		DataKey:       "app.yaml",
		AutoReload:    true,
	}, client)
	require.NoError(t, err)
	_, err = source.Load(context.Background())
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)
	defer watcher.Stop()

	changed := make(chan any, 1)
	watcher.OnChange("feature.enabled", func(value any) {
		changed <- value
	})

	select {
	case value := <-changed:
		require.Equal(t, true, value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial cached callback")
	}
}

func TestConfigSourceLoadReturnsNotFound(t *testing.T) {
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, &fakeConfigMapClient{loadErr: ErrConfigMapNotFound})
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, ErrConfigMapNotFound)
}

func TestConfigSourceLoadReturnsSourceError(t *testing.T) {
	expected := errors.New("kubernetes: api unreachable")
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, &fakeConfigMapClient{loadErr: expected})
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, expected)
}

func TestConfigSourceWatchAfterCloseFails(t *testing.T) {
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, &fakeConfigMapClient{})
	require.NoError(t, err)

	require.NoError(t, source.Close())

	_, err = source.Watch(context.Background(), "")
	require.EqualError(t, err, "kubernetes: config source closed")
}

func TestConfigSourceSetReturnsNotSupported(t *testing.T) {
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, &fakeConfigMapClient{})
	require.NoError(t, err)

	err = source.Set(context.Background(), "feature.enabled", true)
	require.ErrorIs(t, err, ErrSetNotSupported)
}

func TestConfigSourceCloseIsIdempotent(t *testing.T) {
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, &fakeConfigMapClient{})
	require.NoError(t, err)

	require.NoError(t, source.Close())
	require.NoError(t, source.Close())
}

func TestConfigSourceUnderlyingReturnsClient(t *testing.T) {
	client := &fakeConfigMapClient{}
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, client)
	require.NoError(t, err)

	require.Same(t, client, source.Underlying())
}

func TestConfigSourceAsProjectsNativeClient(t *testing.T) {
	native := fake.NewSimpleClientset()
	client := &fakeNativeConfigMapClient{native: native}
	source, err := NewConfigSourceWithClient(&KubernetesConfig{
		Namespace:     "default",
		ConfigMapName: "app-config",
	}, client)
	require.NoError(t, err)

	var projected *fake.Clientset
	require.True(t, source.As(&projected))
	require.Same(t, native, projected)
}

type fakeConfigMapClient struct {
	data      map[string]string
	watchOnce syncOnce
	updateCh  chan map[string]string
	loadErr   error
}

type fakeNativeConfigMapClient struct {
	fakeConfigMapClient
	native any
}

func (f *fakeNativeConfigMapClient) Underlying() any {
	return f.native
}

func (f *fakeConfigMapClient) LoadConfigMap(ctx context.Context, namespace, name string) (map[string]string, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return cloneStringMapForTest(f.data), nil
}

func (f *fakeConfigMapClient) WatchConfigMap(ctx context.Context, namespace, name string, onUpdate func(map[string]string)) error {
	f.watchOnce.Do(func() {
		if f.updateCh == nil {
			f.updateCh = make(chan map[string]string, 2)
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-f.updateCh:
			if errors.Is(f.loadErr, ErrConfigMapNotFound) {
				continue
			}
			f.data = cloneStringMapForTest(update)
			onUpdate(cloneStringMapForTest(update))
		}
	}
}

func (f *fakeConfigMapClient) push(update map[string]string) {
	f.watchOnce.Do(func() {
		f.updateCh = make(chan map[string]string, 2)
	})
	f.updateCh <- cloneStringMapForTest(update)
}

type syncOnce struct {
	done bool
}

func (o *syncOnce) Do(fn func()) {
	if o.done {
		return
	}
	o.done = true
	fn()
}

func cloneStringMapForTest(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
