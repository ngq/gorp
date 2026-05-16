// Package dtmsdk provides barrier execution for DTM transactions.
// This file implements the BarrierHandler contract for idempotent transaction handling.
//
// 本包提供 DTM 事务的 barrier 执行。
// 本文件实现 BarrierHandler 契约，用于幂等事务处理。
package dtmsdk

import (
	"context"
	"errors"
	"strings"
)

// ErrBarrierTransType indicates barrier transType is required.
//
// ErrBarrierTransType 表示 barrier 需要事务类型。
var ErrBarrierTransType = errors.New("dtm: barrier transType is required")

// ErrBarrierUnsupportedType indicates barrier transType is unsupported.
//
// ErrBarrierUnsupportedType 表示不支持的事务类型。
var ErrBarrierUnsupportedType = errors.New("dtm: barrier transType is unsupported")

// ErrBarrierGID indicates barrier gid is required.
//
// ErrBarrierGID 表示 barrier 需要 gid。
var ErrBarrierGID = errors.New("dtm: barrier gid is required")

// ErrBarrierCallback indicates barrier callback is required.
//
// ErrBarrierCallback 表示 barrier 需要 callback。
var ErrBarrierCallback = errors.New("dtm: barrier callback is required")

// barrierHandler implements integrationcontract.BarrierHandler.
// Executes business callback WITHOUT database-level idempotent protection.
// The current implementation only validates parameters and delegates to the callback directly.
// For production use cases requiring true barrier protection (e.g., SAGA/TCC idempotency),
// integrate the official DTM SDK (github.com/dtm-labs/client) which provides
// database-backed barrier tables for preventing duplicate execution.
//
// ⚠ WARNING: This is a lightweight framework adapter — it does NOT create barrier tables
// or check for duplicate branch execution. Branch retries will cause the business callback
// to execute multiple times.
//
// barrierHandler 实现 integrationcontract.BarrierHandler。
// 当前实现不包含数据库层面的幂等保护，仅做参数校验后直接回调。
// 生产环境需要真正的 barrier 保护（如 SAGA/TCC 幂等），请集成 DTM 官方 SDK
// (github.com/dtm-labs/client)，它会创建 barrier 表防止重复执行。
//
// ⚠ 警告：这是轻量级框架适配器——不会创建 barrier 表，也不会检查分支重复执行。
// 分支重试会导致业务回调重复执行。
type barrierHandler struct {
	client    *DTMClient
	transType string
	gid       string
}

// BarrierContext carries barrier execution context.
//
// BarrierContext 携带 barrier 执行上下文。
type BarrierContext struct {
	TransType string
	GID       string
}

// Call executes the business callback WITHOUT database-level barrier protection.
// The current implementation only validates parameters and delegates to the callback.
// For true idempotent protection, integrate the official DTM SDK.
//
// ⚠ WARNING: Branch retries will cause fn to execute multiple times.
//
// Call 在无数据库 barrier 保护下执行业务回调。
// 当前实现仅做参数校验后直接回调。需要真正幂等保护请集成 DTM 官方 SDK。
//
// ⚠ 警告：分支重试会导致 fn 重复执行。
func (h *barrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	_ = ctx
	if strings.TrimSpace(h.transType) == "" {
		return ErrBarrierTransType
	}
	if !isSupportedBarrierType(h.transType) {
		return ErrBarrierUnsupportedType
	}
	if strings.TrimSpace(h.gid) == "" {
		return ErrBarrierGID
	}
	if fn == nil {
		return ErrBarrierCallback
	}
	return fn(&BarrierContext{TransType: h.transType, GID: h.gid})
}

// isSupportedBarrierType checks if the transaction type supports barrier.
//
// isSupportedBarrierType 检查事务类型是否支持 barrier。
func isSupportedBarrierType(transType string) bool {
	switch strings.ToLower(strings.TrimSpace(transType)) {
	case "saga", "tcc", "xa", "msg":
		return true
	default:
		return false
	}
}
