// Package feature defines the feature flag and feature toggle contracts for gorp framework.
package feature

import (
	"context"
)

// FeatureKey is the container key for the feature flag manager capability.
const FeatureKey = "framework.feature"

// EvaluationContext contains contextual metadata used for feature flag evaluation.
type EvaluationContext struct {
	UserID   string         `json:"user_id,omitempty"`
	TenantID string         `json:"tenant_id,omitempty"`
	Groups   []string       `json:"groups,omitempty"`
	Attrs    map[string]any `json:"attrs,omitempty"`
}

// FeatureManager defines the contract for evaluating dynamic feature flags and canary rollouts.
type FeatureManager interface {
	// EvaluateBool evaluates a boolean feature flag.
	EvaluateBool(ctx context.Context, flagKey string, defaultVal bool, evalCtx EvaluationContext) bool

	// EvaluateString evaluates a string feature flag.
	EvaluateString(ctx context.Context, flagKey string, defaultVal string, evalCtx EvaluationContext) string

	// EvaluateInt evaluates an integer feature flag.
	EvaluateInt(ctx context.Context, flagKey string, defaultVal int, evalCtx EvaluationContext) int

	// EvaluateJSON evaluates a JSON object feature flag into the target out interface.
	EvaluateJSON(ctx context.Context, flagKey string, out any, evalCtx EvaluationContext) error
}
