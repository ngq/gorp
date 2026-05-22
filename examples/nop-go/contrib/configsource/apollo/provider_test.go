package apollo

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apolloconfig/agollo/v4"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.apollo", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}

func TestConfigSourceLoadUsesClient(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{
		getContent: "database:\n  host: 127.0.0.1\n",
	})
	require.NoError(t, err)

	loaded, err := source.Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", loaded["database"].(map[string]any)["host"])

	value, err := source.Get(context.Background(), "database.host")
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", value)
}

func TestConfigWatcherEmitsInitialAndUpdatedValues(t *testing.T) {
	client := &fakeApolloClient{
		watchUpdates: make(chan apolloConfigSnapshot, 1),
	}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	initial := make(chan any, 1)
	watcher.OnChange("database.host", func(value any) {
		initial <- value
	})

	select {
	case value := <-initial:
		require.Equal(t, "127.0.0.1", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	client.push("database:\n  host: 10.0.0.2\n")

	select {
	case value := <-initial:
		require.Equal(t, "10.0.0.2", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected updated callback")
	}
}

func TestWatchLoadsInitialSnapshotWithoutExplicitLoad(t *testing.T) {
	client := &fakeApolloClient{watchUpdates: make(chan apolloConfigSnapshot, 1)}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 1)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case value := <-values:
		require.Equal(t, "127.0.0.1", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback after implicit load")
	}
}

func TestLoadReturnsNotFound(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{
		getErr: ErrConfigNotFound,
	})
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, ErrConfigNotFound)
}

func TestLoadReturnsSourceError(t *testing.T) {
	expected := errors.New("apollo: server unavailable")
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{
		getErr: expected,
	})
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, expected)
}

func TestLoadReturnsAuthFailed(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{
		getErr: ErrAuthFailed,
	})
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, ErrAuthFailed)
}

func TestConfigSourceUnderlyingReturnsClient(t *testing.T) {
	client := &fakeApolloClient{}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	require.Same(t, client, source.Underlying())
}

func TestConfigSourceAsProjectsClient(t *testing.T) {
	client := &fakeApolloClient{}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	var projected *fakeApolloClient
	require.True(t, source.As(&projected))
	require.Same(t, client, projected)
}

func TestConfigSourceAsRejectsInvalidTarget(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{})
	require.NoError(t, err)

	require.False(t, source.As(nil))
	require.False(t, source.As(fakeApolloClient{}))
	var projected agollo.Client
	require.False(t, source.As(&projected))
}

func TestDecodeContentFallsBackToNamespaceRawValue(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{
		getContent: "raw-value",
	})
	require.NoError(t, err)

	loaded, err := source.Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "raw-value", loaded["application"])
}

func TestSetNotSupported(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{})
	require.NoError(t, err)

	err = source.Set(context.Background(), "database.host", "127.0.0.1")
	require.ErrorIs(t, err, ErrSetNotSupported)
}

func TestWatchAfterCloseFails(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{})
	require.NoError(t, err)
	require.NoError(t, source.Close())

	_, err = source.Watch(context.Background(), "")
	require.ErrorIs(t, err, ErrConfigSourceClosed)
}

func TestNewConfigSourceRequiresAppID(t *testing.T) {
	cfg := testApolloConfig()
	cfg.AppID = ""

	source, err := NewConfigSourceWithClient(cfg, &fakeApolloClient{})
	require.Nil(t, source)
	require.ErrorIs(t, err, ErrAppIDRequired)
}

func TestNewConfigSourceRequiresMetaServer(t *testing.T) {
	cfg := testApolloConfig()
	cfg.MetaServer = ""

	source, err := NewConfigSourceWithClient(cfg, &fakeApolloClient{})
	require.Nil(t, source)
	require.ErrorIs(t, err, ErrMetaRequired)
}

func TestWatcherStopPreventsFurtherDispatch(t *testing.T) {
	client := &fakeApolloClient{
		watchUpdates: make(chan apolloConfigSnapshot, 2),
	}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	_, err = source.Load(context.Background())
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	changes := make(chan any, 2)
	watcher.OnChange("database.host", func(value any) {
		changes <- value
	})

	select {
	case <-changes:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	require.NoError(t, watcher.Stop())
	client.push("database:\n  host: 10.0.0.3\n")

	select {
	case value := <-changes:
		t.Fatalf("unexpected callback after stop: %v", value)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloClient{})
	require.NoError(t, err)
	require.NoError(t, source.Close())
	require.NoError(t, source.Close())
}

func TestWatchRetriesAfterSourceUnavailable(t *testing.T) {
	client := &fakeApolloClient{
		watchUpdates: make(chan apolloConfigSnapshot, 1),
		watchErrs:    []error{ErrSourceUnavailable},
	}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 2)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case <-values:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	client.push("database:\n  host: 10.0.0.9\n")

	select {
	case value := <-values:
		require.Equal(t, "10.0.0.9", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected callback after retry")
	}
	require.GreaterOrEqual(t, client.watchCalls.Load(), int32(2))
}

type fakeApolloClient struct {
	getContent   string
	getRevision  string
	getSnapshots []apolloConfigSnapshot
	getErr       error
	watchErr     error
	watchErrs    []error
	watchUpdates chan apolloConfigSnapshot
	watchCalls   atomic.Int32
}

type fakeApolloNativeClient struct {
	fakeApolloClient
	native any
}

func (f *fakeApolloNativeClient) Underlying() any {
	return f.native
}

func (f *fakeApolloClient) GetConfig(ctx context.Context, cfg *ApolloConfig) (apolloConfigSnapshot, error) {
	if f.getErr != nil {
		return apolloConfigSnapshot{}, f.getErr
	}
	if len(f.getSnapshots) > 0 {
		snapshot := f.getSnapshots[0]
		if len(f.getSnapshots) > 1 {
			f.getSnapshots = f.getSnapshots[1:]
		}
		if snapshot.Revision == "" {
			snapshot.Revision = normalizeApolloRevision(snapshot)
		}
		return snapshot, nil
	}
	content := "database:\n  host: 127.0.0.1\n"
	if f.getContent != "" {
		content = f.getContent
	}
	revision := f.getRevision
	if revision == "" {
		revision = "rev-initial"
	}
	return apolloConfigSnapshot{Content: content, Revision: revision}, nil
}

func (f *fakeApolloClient) WatchConfig(ctx context.Context, cfg *ApolloConfig, lastRevision string, onUpdate func(snapshot apolloConfigSnapshot)) error {
	f.watchCalls.Add(1)
	if len(f.watchErrs) > 0 {
		err := f.watchErrs[0]
		f.watchErrs = f.watchErrs[1:]
		return err
	}
	if f.watchUpdates == nil {
		<-ctx.Done()
		return f.watchErr
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-f.watchUpdates:
			onUpdate(update)
		}
	}
}

func (f *fakeApolloClient) push(content string) {
	f.pushSnapshot(apolloConfigSnapshot{Content: content})
}

func (f *fakeApolloClient) pushSnapshot(snapshot apolloConfigSnapshot) {
	if f.watchUpdates == nil {
		f.watchUpdates = make(chan apolloConfigSnapshot, 1)
	}
	f.watchUpdates <- snapshot
}

func testApolloConfig() *ApolloConfig {
	return &ApolloConfig{
		AppID:              "demo",
		Cluster:            "default",
		Namespace:          "application",
		MetaServer:         "http://apollo.local",
		PollInterval:       10 * time.Millisecond,
		WatchRetryInterval: 10 * time.Millisecond,
	}
}

func TestConfigSourceUnderlyingPrefersNativeClient(t *testing.T) {
	native := struct{ Name string }{Name: "apollo-native"}
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloNativeClient{native: native})
	require.NoError(t, err)

	require.Equal(t, native, source.Underlying())
}

func TestConfigSourceAsProjectsNativeClientWhenAvailable(t *testing.T) {
	native := &struct{ Name string }{Name: "apollo-native"}
	source, err := NewConfigSourceWithClient(testApolloConfig(), &fakeApolloNativeClient{native: native})
	require.NoError(t, err)

	var projected *struct{ Name string }
	require.True(t, source.As(&projected))
	require.Same(t, native, projected)
}

func TestWatchSkipsDuplicateRevisionAfterRetry(t *testing.T) {
	client := &fakeApolloClient{
		watchUpdates: make(chan apolloConfigSnapshot, 2),
		watchErrs:    []error{ErrSourceUnavailable},
		getRevision:  "rev-1",
	}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 3)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case value := <-values:
		require.Equal(t, "127.0.0.1", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	client.pushSnapshot(apolloConfigSnapshot{
		Content:  "database:\n  host: 127.0.0.1\n",
		Revision: "rev-1",
	})

	select {
	case value := <-values:
		t.Fatalf("unexpected duplicate callback: %v", value)
	case <-time.After(100 * time.Millisecond):
	}

	client.pushSnapshot(apolloConfigSnapshot{
		Content:  "database:\n  host: 10.0.0.8\n",
		Revision: "rev-2",
	})

	select {
	case value := <-values:
		require.Equal(t, "10.0.0.8", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected callback for new revision")
	}
}

func TestWatchStopsAfterAuthFailed(t *testing.T) {
	client := &fakeApolloClient{
		watchErrs:   []error{ErrAuthFailed},
		getRevision: "rev-1",
	}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 1)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case <-values:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	require.Eventually(t, func() bool {
		return client.watchCalls.Load() == 1
	}, time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, int32(1), client.watchCalls.Load())
}

func TestWatchStopsAfterConfigNotFound(t *testing.T) {
	client := &fakeApolloClient{
		watchErrs:   []error{ErrConfigNotFound},
		getRevision: "rev-1",
	}
	source, err := NewConfigSourceWithClient(testApolloConfig(), client)
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 1)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case <-values:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	require.Eventually(t, func() bool {
		return client.watchCalls.Load() == 1
	}, time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, int32(1), client.watchCalls.Load())
}

func TestTranslateApolloSDKErrorClassifiesFailures(t *testing.T) {
	require.ErrorIs(t, translateApolloSDKError(errors.New("401 unauthorized")), ErrAuthFailed)
	require.ErrorIs(t, translateApolloSDKError(errors.New("404 not found")), ErrConfigNotFound)
	require.ErrorIs(t, translateApolloSDKError(errors.New("dial tcp: connection refused")), ErrSourceUnavailable)
}

func TestWatchPollFallbackRefreshesSnapshot(t *testing.T) {
	client := &fakeApolloClient{
		getSnapshots: []apolloConfigSnapshot{
			{Content: "database:\n  host: 127.0.0.1\n", Revision: "rev-1"},
			{Content: "database:\n  host: 10.0.0.7\n", Revision: "rev-2"},
		},
	}
	cfg := testApolloConfig()
	cfg.PollInterval = 20 * time.Millisecond
	source, err := NewConfigSourceWithClient(cfg, client)
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 2)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case value := <-values:
		require.Equal(t, "127.0.0.1", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	select {
	case value := <-values:
		require.Equal(t, "10.0.0.7", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected callback from poll fallback")
	}
}
