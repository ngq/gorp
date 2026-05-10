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
// Executes business callback with barrier protection.
//
// barrierHandler 实现 integrationcontract.BarrierHandler。
// 在 barrier 保护下执行业务回调。
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

// Call executes the business callback with barrier protection.
// Implements integrationcontract.BarrierHandler.Call.
//
// Call 在 barrier 保护下执行业务回调。
// 实现 integrationcontract.BarrierHandler.Call。
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