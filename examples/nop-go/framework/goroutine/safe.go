// Package goroutine provides goroutine safety utilities for gorp framework.
// This file provides SafeGo for panic recovery in goroutines.
// Prevents panic from crashing the entire process, logs errors instead.
//
// Goroutine 包提供 gorp 框架的 goroutine 安全工具能力。
// 本文件提供 SafeGo 用于 goroutine 中的 panic 恢复。
// 防止 panic 把整个进程直接打崩，而是尽量记录到统一 logger。
package goroutine

import (
	"context"
	"fmt"
	"runtime/debug"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// SafeGo 在新的 goroutine 中执行 fn，并统一 recover panic 后记录日志。
//
// 中文说明：
// - 它适合替代直接 `go fn()` 的裸调用方式。
// - 当子 goroutine panic 时，错误不会把整个进程直接打崩，而是尽量记录到统一 logger。
func SafeGo(ctx context.Context, c runtimecontract.Container, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l := LoggerFromContainer(c)
				if l != nil {
					l.Error("panic in goroutine", observabilitycontract.Field{Key: "recover", Value: r}, observabilitycontract.Field{Key: "stack", Value: string(debug.Stack())})
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
func SafeGoAndWait(ctx context.Context, c runtimecontract.Container, fns ...func(context.Context) error) error {
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
