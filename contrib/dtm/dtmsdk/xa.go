// Package dtmsdk provides XA transaction pattern implementation for DTM.
// This file implements the XABuilder contract for XA-style distributed transactions.
//
// 本包提供 DTM XA 事务模式实现。
// 本文件实现 XABuilder 契约，用于 XA 风格分布式事务。
package dtmsdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ErrXANoSteps indicates XA transaction has no steps.
//
// ErrXANoSteps 表示 XA 事务没有步骤。
var ErrXANoSteps = errors.New("dtm: xa has no steps")

// ErrXAStepRequired indicates XA step url is required.
//
// ErrXAStepRequired 表示 XA 步骤需要 url。
var ErrXAStepRequired = errors.New("dtm: xa step url is required")

// xaBuilder implements integrationcontract.XABuilder.
// Builds and submits XA transactions to DTM server.
//
// xaBuilder 实现 integrationcontract.XABuilder。
// 构建并提交 XA 事务到 DTM 服务器。
type xaBuilder struct {
	client *DTMClient
	name   string
	gid    string
	steps  []xaStep
}

// xaStep represents one XA branch step.
//
// xaStep 表示一个 XA 分支步骤。
type xaStep struct {
	url     string
	payload any
}

// XATransaction represents a built XA transaction model.
//
// XATransaction 表示构建好的 XA 事务模型。
type XATransaction struct {
	GID   string
	Steps []XAStep
}

// XAStep represents one XA branch step in transaction model.
//
// XAStep 表示事务模型中的一个 XA 分支步骤。
type XAStep struct {
	URL     string
	Payload any
}

// Add adds a new XA step with url and payload.
// Implements integrationcontract.XABuilder.Add.
//
// Add 添加新的 XA 步骤，包含 url 和 payload。
// 实现 integrationcontract.XABuilder.Add。
func (b *xaBuilder) Add(url string, payload any) integrationcontract.XABuilder {
	b.steps = append(b.steps, xaStep{url: url, payload: payload})
	return b
}

// Submit builds and submits the XA transaction to DTM server.
// Implements integrationcontract.XABuilder.Submit.
//
// Submit 构建并提交 XA 事务到 DTM 服务器。
// 实现 integrationcontract.XABuilder.Submit。
func (b *xaBuilder) Submit(ctx context.Context) error {
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
		"trans_type":      "xa",
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

// Build constructs the XATransaction model from builder state.
//
// Build 从 builder 状态构造 XATransaction 模型。
func (b *xaBuilder) Build() (*XATransaction, error) {
	if len(b.steps) == 0 {
		return nil, ErrXANoSteps
	}
	steps := make([]XAStep, len(b.steps))
	for i, step := range b.steps {
		if strings.TrimSpace(step.url) == "" {
			return nil, ErrXAStepRequired
		}
		steps[i] = XAStep{URL: step.url, Payload: step.payload}
	}
	return &XATransaction{GID: b.gid, Steps: steps}, nil
}

// buildSteps converts xaStep slice to DTM API format.
//
// buildSteps 将 xaStep 列表转换为 DTM API 格式。
func (b *xaBuilder) buildSteps() ([]map[string]string, []string, error) {
	if len(b.steps) == 0 {
		return nil, nil, ErrXANoSteps
	}
	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.url) == "" {
			return nil, nil, ErrXAStepRequired
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return nil, nil, fmt.Errorf("dtm: marshal xa payload failed: %w", err)
		}
		steps = append(steps, map[string]string{"url": step.url})
		payloads = append(payloads, payload)
	}
	return steps, payloads, nil
}
