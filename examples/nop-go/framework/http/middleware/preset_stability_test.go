// Package middleware_test provides unit tests for HTTP middleware preset stability.
//
// 适用场景：
// - 验证默认中间件预设的稳定性。
// - 防止预设顺序与默认值决策发生意外漂移。
package middleware

import (
	"testing"
)

// TestDefaultMiddlewareSetStableSize verifies that the default middleware preset keeps a stable cardinality.
//
// TestDefaultMiddlewareSetStableSize 验证默认中间件预设保持稳定的数量。
func TestDefaultMiddlewareSetStableSize(t *testing.T) {
	set := DefaultMiddlewareSet(nil)
	if len(set) != 3 {
		t.Fatalf("expected 3 default middleware entries, got %d", len(set))
	}
}
