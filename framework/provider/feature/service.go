// Package feature implements the config-driven feature flag and canary rollout provider.
package feature

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"strings"
	"sync"

	featurecontract "github.com/ngq/gorp/framework/contract/feature"
)

// FlagRule defines the configuration rule for a single feature flag.
type FlagRule struct {
	Enabled         bool           `mapstructure:"enabled" json:"enabled"`
	DefaultValue    any            `mapstructure:"default_value" json:"default_value"`
	UserWhitelist   []string       `mapstructure:"user_whitelist" json:"user_whitelist"`
	TenantWhitelist []string       `mapstructure:"tenant_whitelist" json:"tenant_whitelist"`
	Percentage      int            `mapstructure:"percentage" json:"percentage"` // 0 - 100
	ValueMap        map[string]any `mapstructure:"value_map" json:"value_map"`
}

// Service is the in-memory/config-backed FeatureManager implementation.
type Service struct {
	mu    sync.RWMutex
	flags map[string]FlagRule
}

// NewService creates a new feature flag Service with initial flag rules.
func NewService(initialFlags ...map[string]FlagRule) *Service {
	s := &Service{
		flags: make(map[string]FlagRule),
	}
	if len(initialFlags) > 0 && initialFlags[0] != nil {
		for k, v := range initialFlags[0] {
			s.flags[k] = v
		}
	}
	return s
}

// SetRule sets or updates a feature flag rule dynamically.
func (s *Service) SetRule(flagKey string, rule FlagRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags[flagKey] = rule
}

// EvaluateBool evaluates a boolean feature flag.
func (s *Service) EvaluateBool(ctx context.Context, flagKey string, defaultVal bool, evalCtx featurecontract.EvaluationContext) bool {
	s.mu.RLock()
	rule, exists := s.flags[flagKey]
	s.mu.RUnlock()

	if !exists {
		return defaultVal
	}
	if !rule.Enabled {
		return false
	}

	// 1. User Whitelist Check
	if len(rule.UserWhitelist) > 0 && evalCtx.UserID != "" {
		for _, u := range rule.UserWhitelist {
			if u == evalCtx.UserID {
				return true
			}
		}
	}

	// 2. Tenant Whitelist Check
	if len(rule.TenantWhitelist) > 0 && evalCtx.TenantID != "" {
		for _, t := range rule.TenantWhitelist {
			if t == evalCtx.TenantID {
				return true
			}
		}
	}

	// 3. Percentage Rollout Check (Hash UserID or TenantID % 100)
	if rule.Percentage > 0 {
		keyToHash := evalCtx.UserID
		if keyToHash == "" {
			keyToHash = evalCtx.TenantID
		}
		if keyToHash != "" {
			h := fnv.New32a()
			_, _ = h.Write([]byte(keyToHash + ":" + flagKey))
			bucket := int(h.Sum32() % 100)
			return bucket < rule.Percentage
		}
	}

	// If default value is provided in rule, use it
	if boolVal, ok := rule.DefaultValue.(bool); ok {
		return boolVal
	}
	return rule.Enabled
}

// EvaluateString evaluates a string feature flag.
func (s *Service) EvaluateString(ctx context.Context, flagKey string, defaultVal string, evalCtx featurecontract.EvaluationContext) string {
	s.mu.RLock()
	rule, exists := s.flags[flagKey]
	s.mu.RUnlock()

	if !exists || !rule.Enabled {
		return defaultVal
	}
	if strVal, ok := rule.DefaultValue.(string); ok {
		return strVal
	}
	return defaultVal
}

// EvaluateInt evaluates an integer feature flag.
func (s *Service) EvaluateInt(ctx context.Context, flagKey string, defaultVal int, evalCtx featurecontract.EvaluationContext) int {
	s.mu.RLock()
	rule, exists := s.flags[flagKey]
	s.mu.RUnlock()

	if !exists || !rule.Enabled {
		return defaultVal
	}
	switch v := rule.DefaultValue.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return defaultVal
}

// EvaluateJSON evaluates a JSON object feature flag into target out.
func (s *Service) EvaluateJSON(ctx context.Context, flagKey string, out any, evalCtx featurecontract.EvaluationContext) error {
	s.mu.RLock()
	rule, exists := s.flags[flagKey]
	s.mu.RUnlock()

	if !exists || !rule.Enabled || rule.DefaultValue == nil {
		return nil
	}

	b, err := json.Marshal(rule.DefaultValue)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// ParseEvalContextFromHeader constructs EvaluationContext from HTTP headers (X-User-ID, X-Tenant-ID).
func ParseEvalContextFromHeader(getHeader func(string) string) featurecontract.EvaluationContext {
	if getHeader == nil {
		return featurecontract.EvaluationContext{}
	}
	userID := getHeader("X-User-ID")
	tenantID := getHeader("X-Tenant-ID")
	groupsStr := getHeader("X-User-Groups")
	var groups []string
	if groupsStr != "" {
		groups = strings.Split(groupsStr, ",")
	}
	return featurecontract.EvaluationContext{
		UserID:   userID,
		TenantID: tenantID,
		Groups:   groups,
	}
}
