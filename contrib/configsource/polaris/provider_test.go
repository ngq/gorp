package polaris

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.polaris", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.ConfigSourceKey}, p.Provides())
}

func TestConfigSourceLoadUsesClient(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{
		getContent: "database:\n  host: 127.0.0.1\n",
	})
	require.NoError(t, err)

	loaded, err := source.Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", loaded["database"].(map[string]any)["host"])
}

func TestWatcherEmitsInitialAndUpdatedValues(t *testing.T) {
	client := &fakePolarisConfigClient{watchUpdates: make(chan polarisConfigSnapshot, 1)}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
	require.NoError(t, err)
	_, err = source.Load(context.Background())
	require.NoError(t, err)

	watcher, err := source.Watch(context.Background(), "")
	require.NoError(t, err)

	values := make(chan any, 1)
	watcher.OnChange("database.host", func(value any) { values <- value })

	select {
	case value := <-values:
		require.Equal(t, "127.0.0.1", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial callback")
	}

	client.push("database:\n  host: 10.0.0.2\n")
	select {
	case value := <-values:
		require.Equal(t, "10.0.0.2", value)
	case <-time.After(2 * time.Second):
		t.Fatal("expected updated callback")
	}
}

func TestWatchLoadsInitialSnapshotWithoutExplicitLoad(t *testing.T) {
	client := &fakePolarisConfigClient{watchUpdates: make(chan polarisConfigSnapshot, 1)}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
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
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{getErr: ErrConfigNotFound})
	require.NoError(t, err)
	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, ErrConfigNotFound)
}

func TestLoadReturnsSourceError(t *testing.T) {
	expected := errors.New("polaris unavailable")
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{getErr: expected})
	require.NoError(t, err)
	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, expected)
}

func TestLoadReturnsAuthFailed(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{getErr: ErrAuthFailed})
	require.NoError(t, err)
	_, err = source.Load(context.Background())
	require.ErrorIs(t, err, ErrAuthFailed)
}

func TestConfigSourceUnderlyingReturnsClient(t *testing.T) {
	client := &fakePolarisConfigClient{}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
	require.NoError(t, err)

	require.Same(t, client, source.Underlying())
}

func TestConfigSourceAsProjectsClient(t *testing.T) {
	client := &fakePolarisConfigClient{}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
	require.NoError(t, err)

	var projected *fakePolarisConfigClient
	require.True(t, source.As(&projected))
	require.Same(t, client, projected)
}

func TestConfigSourceAsRejectsInvalidTarget(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{})
	require.NoError(t, err)

	require.False(t, source.As(nil))
	require.False(t, source.As(fakePolarisConfigClient{}))
}

func TestDecodeContentFallsBackToRawValue(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{getContent: "raw-value"})
	require.NoError(t, err)
	loaded, err := source.Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "raw-value", loaded["config.yaml"])
}

func TestSetNotSupported(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{})
	require.NoError(t, err)
	err = source.Set(context.Background(), "a", "b")
	require.ErrorIs(t, err, ErrSetNotSupported)
}

func TestWatchAfterCloseFails(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{})
	require.NoError(t, err)
	require.NoError(t, source.Close())
	_, err = source.Watch(context.Background(), "")
	require.ErrorIs(t, err, ErrConfigSourceClosed)
}

func TestNewConfigSourceRequiresServerAddress(t *testing.T) {
	cfg := testPolarisConfig()
	cfg.ServerAddress = ""

	source, err := NewConfigSourceWithClient(cfg, &fakePolarisConfigClient{})
	require.Nil(t, source)
	require.ErrorIs(t, err, ErrServerAddressRequired)
}

func TestNewConfigSourceRequiresFileGroup(t *testing.T) {
	cfg := testPolarisConfig()
	cfg.FileGroup = ""

	source, err := NewConfigSourceWithClient(cfg, &fakePolarisConfigClient{})
	require.Nil(t, source)
	require.ErrorIs(t, err, ErrFileGroupRequired)
}

func TestNewConfigSourceRequiresFileName(t *testing.T) {
	cfg := testPolarisConfig()
	cfg.FileName = ""

	source, err := NewConfigSourceWithClient(cfg, &fakePolarisConfigClient{})
	require.Nil(t, source)
	require.ErrorIs(t, err, ErrFileNameRequired)
}

func TestWatcherStopPreventsFurtherDispatch(t *testing.T) {
	client := &fakePolarisConfigClient{watchUpdates: make(chan polarisConfigSnapshot, 2)}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
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

	require.NoError(t, watcher.Stop())
	client.push("database:\n  host: 10.0.0.3\n")

	select {
	case value := <-values:
		t.Fatalf("unexpected callback after stop: %v", value)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisConfigClient{})
	require.NoError(t, err)
	require.NoError(t, source.Close())
	require.NoError(t, source.Close())
}

func TestWatchRetriesAfterSourceUnavailable(t *testing.T) {
	client := &fakePolarisConfigClient{
		watchUpdates: make(chan polarisConfigSnapshot, 1),
		watchErrs:    []error{ErrSourceUnavailable},
	}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
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
	require.GreaterOrEqual(t, client.watchCalls, 2)
}

type fakePolarisConfigClient struct {
	getContent   string
	getRevision  string
	getErr       error
	getSnapshots []polarisConfigSnapshot
	watchErrs    []error
	watchUpdates chan polarisConfigSnapshot
	watchCalls   int
}

type fakePolarisNativeClient struct {
	fakePolarisConfigClient
	native any
}

func (f *fakePolarisNativeClient) Underlying() any {
	return f.native
}

func (f *fakePolarisConfigClient) GetConfig(ctx context.Context, cfg *PolarisConfig) (polarisConfigSnapshot, error) {
	if f.getErr != nil {
		return polarisConfigSnapshot{}, f.getErr
	}
	if len(f.getSnapshots) > 0 {
		snapshot := f.getSnapshots[0]
		if len(f.getSnapshots) > 1 {
			f.getSnapshots = f.getSnapshots[1:]
		}
		if snapshot.Revision == "" {
			snapshot.Revision = normalizePolarisRevision(snapshot)
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
	return polarisConfigSnapshot{Content: content, Revision: revision}, nil
}

func (f *fakePolarisConfigClient) WatchConfig(ctx context.Context, cfg *PolarisConfig, lastRevision string, onUpdate func(snapshot polarisConfigSnapshot)) error {
	f.watchCalls++
	if len(f.watchErrs) > 0 {
		err := f.watchErrs[0]
		f.watchErrs = f.watchErrs[1:]
		return err
	}
	if f.watchUpdates == nil {
		<-ctx.Done()
		return nil
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

func (f *fakePolarisConfigClient) push(content string) {
	f.pushSnapshot(polarisConfigSnapshot{Content: content})
}

func (f *fakePolarisConfigClient) pushSnapshot(snapshot polarisConfigSnapshot) {
	if f.watchUpdates == nil {
		f.watchUpdates = make(chan polarisConfigSnapshot, 1)
	}
	f.watchUpdates <- snapshot
}

func testPolarisConfig() *PolarisConfig {
	return &PolarisConfig{
		ServerAddress:      "http://polaris.local",
		Namespace:          "default",
		FileGroup:          "app",
		FileName:           "config.yaml",
		PollInterval:       10 * time.Millisecond,
		WatchRetryInterval: 10 * time.Millisecond,
	}
}

func TestConfigSourceUnderlyingPrefersNativeClient(t *testing.T) {
	native := &struct{ Name string }{Name: "polaris-native"}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), &fakePolarisNativeClient{native: native})
	require.NoError(t, err)

	require.Same(t, native, source.Underlying())
}

func TestWatchSkipsDuplicateRevisionAfterRetry(t *testing.T) {
	client := &fakePolarisConfigClient{
		watchUpdates: make(chan polarisConfigSnapshot, 2),
		watchErrs:    []error{ErrSourceUnavailable},
		getRevision:  "rev-1",
	}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
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

	client.pushSnapshot(polarisConfigSnapshot{
		Content:  "database:\n  host: 127.0.0.1\n",
		Revision: "rev-1",
	})

	select {
	case value := <-values:
		t.Fatalf("unexpected duplicate callback: %v", value)
	case <-time.After(100 * time.Millisecond):
	}

	client.pushSnapshot(polarisConfigSnapshot{
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
	client := &fakePolarisConfigClient{
		watchErrs:   []error{ErrAuthFailed},
		getRevision: "rev-1",
	}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
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
		return client.watchCalls == 1
	}, time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, 1, client.watchCalls)
}

func TestWatchStopsAfterConfigNotFound(t *testing.T) {
	client := &fakePolarisConfigClient{
		watchErrs:   []error{ErrConfigNotFound},
		getRevision: "rev-1",
	}
	source, err := NewConfigSourceWithClient(testPolarisConfig(), client)
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
		return client.watchCalls == 1
	}, time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, 1, client.watchCalls)
}

func TestTranslatePolarisSDKErrorClassifiesFailures(t *testing.T) {
	require.ErrorIs(t, translatePolarisSDKError(errors.New("401 unauthorized")), ErrAuthFailed)
	require.ErrorIs(t, translatePolarisSDKError(errors.New("404 not found")), ErrConfigNotFound)
	require.ErrorIs(t, translatePolarisSDKError(errors.New("dial tcp: connection refused")), ErrSourceUnavailable)
}

func TestWatchPollFallbackRefreshesSnapshot(t *testing.T) {
	client := &fakePolarisConfigClient{
		getSnapshots: []polarisConfigSnapshot{
			{Content: "database:\n  host: 127.0.0.1\n", Revision: "rev-1"},
			{Content: "database:\n  host: 10.0.0.7\n", Revision: "rev-2"},
		},
	}
	cfg := testPolarisConfig()
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
