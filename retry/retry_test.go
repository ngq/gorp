package retry

import (
	"context"
	"errors"
	"io"
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type exportRetryStub struct{}

func (s *exportRetryStub) Do(context.Context, func() error) error { return nil }
func (s *exportRetryStub) DoForResource(_ context.Context, _ string, fn func() error) error {
	return fn()
}
func (s *exportRetryStub) DoWithResult(context.Context, func() (any, error)) (any, error) {
	return "ok", nil
}
func (s *exportRetryStub) IsRetryable(err error) bool { return err != nil }

type exportRetryContainerStub struct {
	retry resiliencecontract.Retry
}

func (s *exportRetryContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (s *exportRetryContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (s *exportRetryContainerStub) IsBind(string) bool                                      { return true }
func (s *exportRetryContainerStub) IsBindNamed(string, string) bool                         { return false }
func (s *exportRetryContainerStub) MustMake(key string) any                                 { v, _ := s.Make(key); return v }
func (s *exportRetryContainerStub) MustMakeNamed(string, string) any                        { return nil }
func (s *exportRetryContainerStub) RegisterCloser(string, io.Closer)                        {}
func (s *exportRetryContainerStub) Destroy() error                                          { return nil }
func (s *exportRetryContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo     { return nil }
func (s *exportRetryContainerStub) DebugPrint() string                                      { return "" }
func (s *exportRetryContainerStub) ProviderDAG() runtimecontract.ProviderDAG {
	return runtimecontract.ProviderDAG{}
}
func (s *exportRetryContainerStub) MakeNamed(string, string) (any, error) { return nil, nil }
func (s *exportRetryContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportRetryContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportRetryContainerStub) Make(key string) (any, error) {
	if key == resiliencecontract.RetryKey {
		return s.retry, nil
	}
	return nil, context.DeadlineExceeded
}

func TestExportedRetryHelpers(t *testing.T) {
	stub := &exportRetryStub{}
	containerStub := &exportRetryContainerStub{retry: stub}

	retrySvc, err := Make(containerStub)
	require.NoError(t, err)
	require.Same(t, stub, retrySvc)
	require.Same(t, stub, MustMake(containerStub))

	err = Do(context.Background(), containerStub, func() error { return nil })
	require.NoError(t, err)

	result, err := DoWithResult(context.Background(), containerStub, func() (any, error) {
		return 1, errors.New("ignored")
	})
	require.NoError(t, err)
	require.Equal(t, "ok", result)

	ok, err := IsRetryable(containerStub, errors.New("boom"))
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, resiliencecontract.DefaultRetryPolicy(), DefaultRetryPolicy())
}
