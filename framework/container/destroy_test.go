// Package container_test provides unit tests for the Destroy lifecycle.
//
// 适用场景：
// - 验证 RegisterCloser / Destroy 语义。
// - 验证销毁后 Make 返回 ErrContainerDestroyed。
// - 验证 Closer 调用顺序（逆序）。
package container

import (
	"errors"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestDestroy_ReverseOrder verifies that closers are called in reverse registration order.
//
// TestDestroy_ReverseOrder 验证 Closer 按注册逆序调用。
func TestDestroy_ReverseOrder(t *testing.T) {
	c := New()
	var order []string

	c.RegisterCloser("first", &closeFunc{fn: func() error {
		order = append(order, "first")
		return nil
	}})
	c.RegisterCloser("second", &closeFunc{fn: func() error {
		order = append(order, "second")
		return nil
	}})
	c.RegisterCloser("third", &closeFunc{fn: func() error {
		order = append(order, "third")
		return nil
	}})

	err := c.Destroy()
	require.NoError(t, err)
	require.Equal(t, []string{"third", "second", "first"}, order)
}

// TestDestroy_ReturnsPartialErrors verifies that Destroy collects all close errors.
//
// TestDestroy_ReturnsPartialErrors 验证 Destroy 收集所有关闭错误。
func TestDestroy_ReturnsPartialErrors(t *testing.T) {
	c := New()

	c.RegisterCloser("ok", &closeFunc{fn: func() error { return nil }})
	c.RegisterCloser("fail1", &closeFunc{fn: func() error { return errors.New("close fail1") }})
	c.RegisterCloser("ok2", &closeFunc{fn: func() error { return nil }})
	c.RegisterCloser("fail2", &closeFunc{fn: func() error { return errors.New("close fail2") }})

	err := c.Destroy()
	require.Error(t, err)
	require.Contains(t, err.Error(), "close fail1")
	require.Contains(t, err.Error(), "close fail2")
}

// TestDestroy_DoubleDestroyReturnsError verifies that calling Destroy twice returns an error.
//
// TestDestroy_DoubleDestroyReturnsError 验证调用 Destroy 两次返回错误。
func TestDestroy_DoubleDestroyReturnsError(t *testing.T) {
	c := New()
	require.NoError(t, c.Destroy())
	require.Error(t, c.Destroy())
}

// TestDestroy_MakeAfterDestroyReturnsError verifies that Make returns
// ErrContainerDestroyed after Destroy is called.
//
// TestDestroy_MakeAfterDestroyReturnsError 验证 Destroy 后 Make 返回 ErrContainerDestroyed。
func TestDestroy_MakeAfterDestroyReturnsError(t *testing.T) {
	c := New()
	c.Bind("svc", func(runtimecontract.Container) (any, error) {
		return "value", nil
	}, true)

	require.NoError(t, c.Destroy())

	_, err := c.Make("svc")
	require.Error(t, err)
	require.True(t, errors.Is(err, runtimecontract.ErrContainerDestroyed))
}

// TestDestroy_MakeNamedAfterDestroyReturnsError verifies that MakeNamed returns
// ErrContainerDestroyed after Destroy.
//
// TestDestroy_MakeNamedAfterDestroyReturnsError 验证 Destroy 后 MakeNamed 返回 ErrContainerDestroyed。
func TestDestroy_MakeNamedAfterDestroyReturnsError(t *testing.T) {
	c := New()
	require.NoError(t, c.Destroy())

	_, err := c.MakeNamed("name", "key")
	require.Error(t, err)
	require.True(t, errors.Is(err, runtimecontract.ErrContainerDestroyed))
}

// closeFunc is a simple io.Closer implementation for testing.
type closeFunc struct {
	fn func() error
}

func (c *closeFunc) Close() error {
	if c.fn != nil {
		return c.fn()
	}
	return nil
}
