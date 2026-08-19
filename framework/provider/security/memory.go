// Package security provides an in-memory, zero-dependency Authorizer provider for gorp framework.
package security

import (
	"context"
	"path"
	"strings"
	"sync"
)

// MemoryAuthorizer is a lightweight in-memory Authorizer implementation.
// Maps role/subject names to allowed action:object patterns.
type MemoryAuthorizer struct {
	mu           sync.RWMutex
	rolePolicies map[string][]string // role -> ["*", "GET:/api/v1/users", "POST:/api/v1/*"]
}

// NewMemoryAuthorizer creates a MemoryAuthorizer with optional initial rules.
func NewMemoryAuthorizer(initialRules ...map[string][]string) *MemoryAuthorizer {
	m := &MemoryAuthorizer{
		rolePolicies: make(map[string][]string),
	}
	if len(initialRules) > 0 && initialRules[0] != nil {
		for k, v := range initialRules[0] {
			m.rolePolicies[strings.ToLower(k)] = v
		}
	}
	return m
}

// AddRule adds a permission pattern for a role/subject.
func (m *MemoryAuthorizer) AddRule(role, pattern string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := strings.ToLower(role)
	m.rolePolicies[r] = append(m.rolePolicies[r], pattern)
}

// Enforce checks if sub (subject/role) is permitted to access obj with act.
func (m *MemoryAuthorizer) Enforce(ctx context.Context, sub, obj, act string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	patterns, ok := m.rolePolicies[strings.ToLower(sub)]
	if !ok {
		return false, nil
	}

	target := strings.ToUpper(act) + ":" + obj
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "*" || p == target {
			return true, nil
		}
		// Match wildcards like "GET:/api/v1/*" or "/api/v1/*"
		if matched, _ := path.Match(p, target); matched {
			return true, nil
		}
		if matched, _ := path.Match(p, obj); matched {
			return true, nil
		}
	}

	return false, nil
}
