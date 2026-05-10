// Package dtmsdk provides TCC transaction pattern implementation for DTM.
// This file implements the TCCBuilder contract for try-confirm-cancel transactions.
//
// 本包提供 DTM TCC 事务模式实现。
// 本文件实现 TCCBuilder 契约，用于 try-confirm-cancel 事务。
package dtmsdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ErrTCCNoSteps indicates TCC transaction has no steps.
//
// ErrTCCNoSteps 表示 TCC 事务没有步骤。
var ErrTCCNoSteps = errors.New("dtm: tcc has no steps")

// ErrTCCStepRequired indicates TCC step try/confirm/cancel are required.
//
// ErrTCCStepRequired 表示 TCC 步骤需要 try/confirm/cancel。
var ErrTCCStepRequired = errors.New("dtm: tcc step try/confirm/cancel are required")

// tccBuilder implements integrationcontract.TCCBuilder.
// Builds and submits TCC transactions to DTM server.
//
// tccBuilder 实现 integrationcontract.TCCBuilder。
// 构建并提交 TCC 事务到 DTM 服务器。
type tccBuilder struct {
	client *DTMClient
	name   string
	gid    string
	steps  []tccStep
}

// tccStep represents one TCC branch step.
//
// tccStep 表示一个 TCC 分支步骤。
type tccStep struct {
	try     string
	confirm string
	cancel  string
	payload any
}

// TCCTransaction represents a built TCC transaction model.
//
// TCCTransaction 表示构建好的 TCC 事务模型。
type TCCTransaction struct {
	GID   string
	Steps []TCCStep
}

// TCCStep represents one TCC branch step in transaction model.
//
// TCCStep 表示事务模型中的一个 TCC 分支步骤。
type TCCStep struct {
	Try     string
	Confirm string
	Cancel  string
	Payload any
}

// Add adds a new TCC step with try, confirm, cancel and payload.
// Implements integrationcontract.TCCBuilder.Add.
//
// Add 添加新的 TCC 步骤，包含 try、confirm、cancel 和 payload。
// 实现 integrationcontract.TCCBuilder.Add。
func (b *tccBuilder) Add(try string, confirm string, cancel string, payload any) integrationcontract.TCCBuilder {
	b.steps = append(b.steps, tccStep{try: try, confirm: confirm, cancel: cancel, payload: payload})
	return b
}

// Submit builds and submits the TCC transaction to DTM server.
// Implements integrationcontract.TCCBuilder.Submit.
//
// Submit 构建并提交 TCC 事务到 DTM 服务器。
// 实现 integrationcontract.TCCBuilder.Submit。
func (b *tccBuilder) Submit(ctx context.Context) error {
	steps, payloads, err := b.buildSteps()
	if err != nil {
		return err
	}
	gid, err := b.client.newGID(ctx)
	if err != nil {
		return err
	}
	body := map[string]any{
		"gid":             gid,
		"trans_type":      "tcc",
		"steps":           steps,
		"payloads":        payloads,
		"retry_interval":  int64(b.client.cfg.RetryInterval),
		"retry_count":     int64(b.client.cfg.RetryCount),
		"request_timeout": int64(b.client.cfg.Timeout),
	}
	if err := b.client.submit(ctx, body); err != nil {
		return err
	}
	b.gid = gid
	return nil
}

// Build constructs the TCCTransaction model from builder state.
//
// Build 从 builder 状态构造 TCCTransaction 模型。
func (b *tccBuilder) Build() (*TCCTransaction, error) {
	if len(b.steps) == 0 {
		return nil, ErrTCCNoSteps
	}
	steps := make([]TCCStep, len(b.steps))
	for i, step := range b.steps {
		if strings.TrimSpace(step.try) == "" || strings.TrimSpace(step.confirm) == "" || strings.TrimSpace(step.cancel) == "" {
			return nil, ErrTCCStepRequired
		}
		steps[i] = TCCStep{Try: step.try, Confirm: step.confirm, Cancel: step.cancel, Payload: step.payload}
	}
	return &TCCTransaction{GID: b.gid, Steps: steps}, nil
}

// buildSteps converts tccStep slice to DTM API format.
//
// buildSteps 将 tccStep 列表转换为 DTM API 格式。
func (b *tccBuilder) buildSteps() ([]map[string]string, []string, error) {
	if len(b.steps) == 0 {
		return nil, nil, ErrTCCNoSteps
	}
	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.try) == "" || strings.TrimSpace(step.confirm) == "" || strings.TrimSpace(step.cancel) == "" {
			return nil, nil, ErrTCCStepRequired
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return nil, nil, fmt.Errorf("dtm: marshal tcc payload failed: %w", err)
		}
		steps = append(steps, map[string]string{"try": step.try, "confirm": step.confirm, "cancel": step.cancel})
		payloads = append(payloads, payload)
	}
	return steps, payloads, nil
}