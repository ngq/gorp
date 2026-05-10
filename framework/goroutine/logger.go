// Package goroutine provides goroutine safety utilities for gorp framework.
// This file provides logger extraction helper from container.
// Used by SafeGo for unified panic recovery logging.
//
// Goroutine 包提供 gorp 框架的 goroutine 安全工具能力。
// 本文件提供从容器获取 logger 的 helper。
// 用于 SafeGo 统一 panic 恢复日志记录。
package goroutine

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// LoggerFromContainer tries to get framework logger.
func LoggerFromContainer(c runtimecontract.Container) observabilitycontract.Logger {
	v, err := c.Make(observabilitycontract.LogKey)
	if err != nil {
		return nil
	}
	l, _ := v.(observabilitycontract.Logger)
	return l
}
