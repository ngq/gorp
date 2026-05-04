// Package gorp exposes the stable public entrypoints of the framework.
//
// The root package stays intentionally thin:
//   - application startup and mainline assembly entrypoints live here
//   - a small set of high-frequency convenience helpers live here
//   - default HTTP response helpers live here as fallback behavior
//   - concrete implementations and advanced capabilities stay in framework/* and contrib/*
//
// For business code, prefer the narrowest public surface that solves the task:
//   - use package gorp for startup, common entry helpers, and default HTTP response fallback
//   - use focused top-level helper packages such as cache/log/retry/jwt/dlock/validate for capability access
//   - drop into framework/contract/* or provider-specific APIs only when lower-level control is really needed
package gorp
