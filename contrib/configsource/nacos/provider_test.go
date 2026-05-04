package nacos

import (
	"context"
	"errors"
	"testing"
	"time"

	configclient "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.nacos", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}

func TestConfigSourceLoadAndSet(t *testing.T) {
	client := &fakeNacosClient{
		content: "server:\n  port: 8081\nfeature:\n  enabled: true\n",
	}

	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, client)
	require.NoError(t, err)

	loaded, err := source.Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, 8081, loaded["server"].(map[string]any)["port"])

	require.NoError(t, source.Set(context.Background(), "app.yaml", map[string]any{
		"feature": map[string]any{"enabled": false},
	}))
	require.Contains(t, client.content, "enabled: false")
}

func TestConfigSourceWatchDispatchesUpdatedValue(t *testing.T) {
	client := &fakeNacosClient{content: "feature:\n  enabled: false\n"}
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
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

	client.push("feature:\n  enabled: true\n")

	select {
	case value := <-changed:
		require.Equal(t, true, value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected config update")
	}
}

func TestConfigSourceWatchDispatchesInitialCachedValue(t *testing.T) {
	client := &fakeNacosClient{content: "feature:\n  enabled: true\n"}
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
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

func TestDecodeContentFallsBackToString(t *testing.T) {
	loaded, err := decodeContent("not: [valid", "raw.txt")
	require.NoError(t, err)
	require.Equal(t, "not: [valid", loaded["raw.txt"])
}

func TestConfigSourceLoadReturnsNotFound(t *testing.T) {
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, &fakeNacosClient{getErr: ErrConfigNotFound})
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, ErrConfigNotFound)
}

func TestConfigSourceSetReturnsPublishError(t *testing.T) {
	expected := errors.New("nacos: publish failed")
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, &fakeNacosClient{publishErr: expected})
	require.NoError(t, err)

	err = source.Set(context.Background(), "app.yaml", map[string]any{"enabled": true})
	require.ErrorIs(t, err, expected)
}

func TestConfigSourceWatchAfterCloseFails(t *testing.T) {
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, &fakeNacosClient{})
	require.NoError(t, err)

	require.NoError(t, source.Close())

	_, err = source.Watch(context.Background(), "")
	require.EqualError(t, err, "nacos: config source closed")
}

func TestConfigSourceSetRejectsDifferentDataID(t *testing.T) {
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, &fakeNacosClient{})
	require.NoError(t, err)

	err = source.Set(context.Background(), "other.yaml", map[string]any{"enabled": true})
	require.EqualError(t, err, "nacos: set only supports data_id app.yaml")
}

func TestConfigSourceCloseIsIdempotent(t *testing.T) {
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, &fakeNacosClient{})
	require.NoError(t, err)

	require.NoError(t, source.Close())
	require.NoError(t, source.Close())
}

func TestConfigSourceUnderlyingReturnsInjectedClient(t *testing.T) {
	client := &fakeNacosClient{}
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, client)
	require.NoError(t, err)

	require.Same(t, client, source.Underlying())
}

func TestConfigSourceAsProjectsInjectedClient(t *testing.T) {
	client := &fakeNacosClient{}
	source, err := NewConfigSourceWithClient(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	}, client)
	require.NoError(t, err)

	var projected *fakeNacosClient
	require.True(t, source.As(&projected))
	require.Same(t, client, projected)
}

func TestConfigSourceAsProjectsOfficialConfigClientOnDefaultClient(t *testing.T) {
	source, err := NewConfigSource(&NacosConfig{
		ServerAddr: "127.0.0.1",
		Port:       8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "app.yaml",
	})
	require.NoError(t, err)

	var projected configclient.IConfigClient
	require.True(t, source.As(&projected))
	require.NotNil(t, projected)
	require.NotNil(t, source.Underlying())
}

type fakeNacosClient struct {
	content    string
	updateCh   chan string
	getErr     error
	publishErr error
}

func (f *fakeNacosClient) GetConfig(ctx context.Context, cfg *NacosConfig) (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	return f.content, nil
}

func (f *fakeNacosClient) PublishConfig(ctx context.Context, cfg *NacosConfig, content string) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	f.content = content
	return nil
}

func (f *fakeNacosClient) WatchConfig(ctx context.Context, cfg *NacosConfig, onUpdate func(string)) error {
	if f.updateCh == nil {
		f.updateCh = make(chan string, 2)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-f.updateCh:
			f.content = update
			onUpdate(update)
		}
	}
}

func (f *fakeNacosClient) push(update string) {
	if f.updateCh == nil {
		f.updateCh = make(chan string, 2)
	}
	f.updateCh <- update
}
