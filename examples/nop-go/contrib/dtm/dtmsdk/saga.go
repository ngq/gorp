// Package dtmsdk provides SAGA transaction pattern implementation for DTM.
// This file implements the SAGABuilder contract for orchestration-style transactions.
//
// 本包提供 DTM SAGA 事务模式实现。
// 本文件实现 SAGABuilder 契约，用于编排风格事务。
package dtmsdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ErrSagaNoSteps indicates SAGA transaction has no steps.
//
// ErrSagaNoSteps 表示 SAGA 事务没有步骤。
var ErrSagaNoSteps = errors.New("dtm: saga has no steps")

// ErrSagaStepRequired indicates SAGA step action and compensate are required.
//
// ErrSagaStepRequired 表示 SAGA 步骤需要 action 和 compensate。
var ErrSagaStepRequired = errors.New("dtm: saga step action and compensate are required")

// sagaSubmitRequest represents the SAGA submit request body for DTM API.
//
// sagaSubmitRequest 表示 DTM API 的 SAGA 提交请求体。
type sagaSubmitRequest struct {
	GID            string              `json:"gid"`
	TransType      string              `json:"trans_type"`
	Steps          []map[string]string `json:"steps,omitempty"`
	Payloads       []string            `json:"payloads,omitempty"`
	RetryInterval  int64               `json:"retry_interval,omitempty"`
	RetryCount     int64               `json:"retry_count,omitempty"`
	RequestTimeout int64               `json:"request_timeout,omitempty"`
}

// sagaBuilder implements integrationcontract.SAGABuilder.
// Builds and submits SAGA transactions to DTM server.
//
// sagaBuilder 实现 integrationcontract.SAGABuilder。
// 构建并提交 SAGA 事务到 DTM 服务器。
type sagaBuilder struct {
	client *DTMClient
	name   string
	gid    string
	steps  []sagaStep
}

// sagaStep represents one SAGA branch step.
//
// sagaStep 表示一个 SAGA 分支步骤。
type sagaStep struct {
	action        string
	compensate    string
	payload       any
	retryCount    int
	retryInterval int
	timeout       int
}

// Add adds a new SAGA step with action, compensate and payload.
// Implements integrationcontract.SAGABuilder.Add.
//
// Add 添加新的 SAGA 步骤，包含 action、compensate 和 payload。
// 实现 integrationcontract.SAGABuilder.Add。
func (b *sagaBuilder) Add(action string, compensate string, payload any) integrationcontract.SAGABuilder {
	b.steps = append(b.steps, sagaStep{action: action, compensate: compensate, payload: payload})
	return b
}

// AddBranch adds a new SAGA step with branch-level options.
// Implements integrationcontract.SAGABuilder.AddBranch.
//
// AddBranch 添加新的 SAGA 步骤，包含分支级选项。
// 实现 integrationcontract.SAGABuilder.AddBranch。
func (b *sagaBuilder) AddBranch(action string, compensate string, payload any, opts integrationcontract.BranchOptions) integrationcontract.SAGABuilder {
	b.steps = append(b.steps, sagaStep{
		action:        action,
		compensate:    compensate,
		payload:       payload,
		retryCount:    opts.RetryCount,
		retryInterval: opts.RetryInterval,
		timeout:       opts.Timeout,
	})
	return b
}

// Submit builds and submits the SAGA transaction to DTM server.
// Implements integrationcontract.SAGABuilder.Submit.
//
// Submit 构建并提交 SAGA 事务到 DTM 服务器。
// 实现 integrationcontract.SAGABuilder.Submit。
func (b *sagaBuilder) Submit(ctx context.Context) error {
	steps, payloads, err := b.buildSteps()
	if err != nil {
		return err
	}
	gid, err := b.client.newGID(ctx)
	if err != nil {
		return err
	}
	body := sagaSubmitRequest{
		GID:            gid,
		TransType:      "saga",
		Steps:          steps,
		Payloads:       payloads,
		RetryInterval:  int64(b.client.cfg.RetryInterval),
		RetryCount:     int64(b.client.cfg.RetryCount),
		RequestTimeout: int64(b.client.cfg.Timeout),
	}
	if err := b.client.submit(ctx, body); err != nil {
		return err
	}
	b.gid = gid
	return nil
}

// Build constructs the SAGATransaction model from builder state.
// Implements integrationcontract.SAGABuilder.Build.
//
// Build 从 builder 状态构造 SAGATransaction 模型。
// 实现 integrationcontract.SAGABuilder.Build。
func (b *sagaBuilder) Build() (*integrationcontract.SAGATransaction, error) {
	steps := make([]integrationcontract.SAGAStep, len(b.steps))
	for i, step := range b.steps {
		steps[i] = integrationcontract.SAGAStep{
			Action:        step.action,
			Compensate:    step.compensate,
			Payload:       step.payload,
			RetryCount:    step.retryCount,
			RetryInterval: step.retryInterval,
			Timeout:       step.timeout,
		}
	}
	return &integrationcontract.SAGATransaction{GID: b.gid, Steps: steps}, nil
}

// buildSteps converts sagaStep slice to DTM API format.
//
// buildSteps 将 sagaStep 列表转换为 DTM API 格式。
func (b *sagaBuilder) buildSteps() ([]map[string]string, []string, error) {
	if len(b.steps) == 0 {
		return nil, nil, ErrSagaNoSteps
	}
	steps := make([]map[string]string, 0, len(b.steps))
	payloads := make([]string, 0, len(b.steps))
	for _, step := range b.steps {
		if strings.TrimSpace(step.action) == "" || strings.TrimSpace(step.compensate) == "" {
			return nil, nil, ErrSagaStepRequired
		}
		payload, err := marshalPayload(step.payload)
		if err != nil {
			return nil, nil, fmt.Errorf("dtm: marshal saga payload failed: %w", err)
		}
		steps = append(steps, map[string]string{"action": step.action, "compensate": step.compensate})
		payloads = append(payloads, payload)
	}
	return steps, payloads, nil
}
