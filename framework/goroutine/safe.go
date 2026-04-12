package goroutine

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ngq/gorp/framework/contract"
)

// SafeGo 在新的 goroutine 中执行 fn，并统一 recover panic 后记录日志。
//
// 中文说明：
// - 它适合替代直接 `go fn()` 的裸调用方式。
// - 当子 goroutine panic 时，错误不会把整个进程直接打崩，而是尽量记录到统一 logger。
func SafeGo(ctx context.Context, c contract.Container, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l := LoggerFromContainer(c)
				if l != nil {
					l.Error("panic in goroutine", contract.Field{Key: "recover", Value: r}, contract.Field{Key: "stack", Value: string(debug.Stack())})
				}
			}
		}()
		fn(ctx)
	}()
}

// SafeGoAndWait runs multiple functions concurrently.
//
// - Each function gets a derived context.
// - Panic is recovered and returned as an error.
// - The first returned error is returned.
func SafeGoAndWait(ctx context.Context, c contract.Container, fns ...func(context.Context) error) error {
	if len(fns) == 0 {
		return nil
	}

	errCh := make(chan error, len(fns))
	for _, fn := range fns {
		fn := fn
		SafeGo(ctx, c, func(ctx context.Context) {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic: %v", r)
				}
			}()
			errCh <- fn(ctx)
		})
	}

	var firstErr error
	for i := 0; i < len(fns); i++ {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
