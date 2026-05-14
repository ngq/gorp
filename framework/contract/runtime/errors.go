// Application scenarios:
// - Define sentinel errors for container operations.
// - Provide structured error types for circular dependency detection and container lifecycle.
//
// 适用场景：
// - 定义容器操作的哨兵错误。
// - 为循环依赖检测和容器生命周期提供结构化错误类型。
package runtime

import (
	"errors"
	"fmt"
	"strings"
)

// ErrContainerDestroyed is returned when Make/MakeNamed is called after Destroy.
//
// ErrContainerDestroyed 在 Destroy 之后调用 Make/MakeNamed 时返回。
var ErrContainerDestroyed = errors.New("container is destroyed")

// ErrCircularDependency is returned when a circular dependency is detected during service resolution.
//
// ErrCircularDependency 在服务解析期间检测到循环依赖时返回。
var ErrCircularDependency = errors.New("circular dependency detected")

// CircularDependencyError provides details about a circular dependency chain.
//
// CircularDependencyError 提供循环依赖链的详细信息。
type CircularDependencyError struct {
	Key   string   // the key that triggered the cycle
	Chain []string // the full resolution chain, e.g. ["A", "B", "A"]
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s", strings.Join(e.Chain, " -> "))
}

func (e *CircularDependencyError) Unwrap() error {
	return ErrCircularDependency
}

// Is allows errors.Is(err, ErrCircularDependency) to match CircularDependencyError.
//
// Is 使得 errors.Is(err, ErrCircularDependency) 可以匹配 CircularDependencyError。
func (e *CircularDependencyError) Is(target error) bool {
	return target == ErrCircularDependency
}
