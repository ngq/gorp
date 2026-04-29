package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type exportRetryStub struct{}

func (s *exportRetryStub) Do(context.Context, func() error) error { return nil }
func (s *exportRetryStub) DoWithResult(context.Context, func() (any, error)) (any, error) {
	return "ok", nil
}
func (s *exportRetryStub) IsRetryable(err error) bool { return err != nil }

type exportRetryContainerStub struct {
	retry contract.Retry
}

func (s *exportRetryContainerStub) Bind(string, contract.Factory, bool)                {}
func (s *exportRetryContainerStub) IsBind(string) bool                                 { return true }
func (s *exportRetryContainerStub) MustMake(key string) any                            { v, _ := s.Make(key); return v }
func (s *exportRetryContainerStub) RegisterProvider(contract.ServiceProvider) error     { return nil }
func (s *exportRetryContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }
func (s *exportRetryContainerStub) Make(key string) (any, error) {
	if key == contract.RetryKey {
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
	require.Equal(t, contract.DefaultRetryPolicy(), DefaultRetryPolicy())
}
