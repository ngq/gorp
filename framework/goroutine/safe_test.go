package goroutine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

type noopLogger struct{}

func (n *noopLogger) Debug(string, ...contract.Field) {}
func (n *noopLogger) Info(string, ...contract.Field)  {}
func (n *noopLogger) Warn(string, ...contract.Field)  {}
func (n *noopLogger) Error(string, ...contract.Field) {}
func (n *noopLogger) With(...contract.Field) contract.Logger {
	return n
}

func TestSafeGoAndWait_ReturnsFirstError(t *testing.T) {
	c := container.New()
	c.Bind(contract.LogKey, func(contract.Container) (any, error) { return &noopLogger{}, nil }, true)

	err := SafeGoAndWait(context.Background(), c,
		func(context.Context) error { return nil },
		func(context.Context) error { return context.Canceled },
		func(context.Context) error { return nil },
	)
	require.Error(t, err)
}

func TestSafeGoAndWait_RecoversPanic(t *testing.T) {
	c := container.New()
	c.Bind(contract.LogKey, func(contract.Container) (any, error) { return &noopLogger{}, nil }, true)

	err := SafeGoAndWait(context.Background(), c,
		func(context.Context) error { panic("boom") },
	)
	require.Error(t, err)
}
