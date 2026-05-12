// Package goroutine_test provides unit tests for goroutine-safe utilities and container integration.
//
// 适用场景：
// - 验证 goroutine 安全工具在并发场景下的行为正确。
// - 确保 SafeGo 等工具在 container 集成中的安全性和资源管理。
package goroutine

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type noopLogger struct{}

func (n *noopLogger) Debug(string, ...observabilitycontract.Field) {}
func (n *noopLogger) Info(string, ...observabilitycontract.Field)  {}
func (n *noopLogger) Warn(string, ...observabilitycontract.Field)  {}
func (n *noopLogger) Error(string, ...observabilitycontract.Field) {}
func (n *noopLogger) With(...observabilitycontract.Field) observabilitycontract.Logger {
	return n
}

func TestSafeGoAndWait_ReturnsFirstError(t *testing.T) {
	c := container.New()
	c.Bind(observabilitycontract.LogKey, func(runtimecontract.Container) (any, error) { return &noopLogger{}, nil }, true)

	err := SafeGoAndWait(context.Background(), c,
		func(context.Context) error { return nil },
		func(context.Context) error { return context.Canceled },
		func(context.Context) error { return nil },
	)
	require.Error(t, err)
}

func TestSafeGoAndWait_RecoversPanic(t *testing.T) {
	c := container.New()
	c.Bind(observabilitycontract.LogKey, func(runtimecontract.Container) (any, error) { return &noopLogger{}, nil }, true)

	err := SafeGoAndWait(context.Background(), c,
		func(context.Context) error { panic("boom") },
	)
	require.Error(t, err)
}
