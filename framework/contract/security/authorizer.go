// Package security provides the minimal authorization contract for gorp framework.
package security

import (
	"context"
)

// AuthorizerKey is the container key for the authorizer capability.
const AuthorizerKey = "framework.security.authorizer"

// Authorizer defines the minimal authorization contract.
type Authorizer interface {
	// Enforce checks whether the subject (sub) is permitted to access resource (obj) with action (act).
	Enforce(ctx context.Context, sub, obj, act string) (bool, error)
}
