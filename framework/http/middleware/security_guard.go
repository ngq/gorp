// Package middleware provides HTTP security guard middleware for gorp framework.
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

var (
	sqlInjectionRegex   = regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|drop\s+table|update\s+.*\s+set|or\s+1\s*=\s*1|'--|;\s*drop|;\s*shutdown)`)
	xssRegex            = regexp.MustCompile(`(?i)(<script[^>]*>|javascript:|onload\s*=|onerror\s*=|onclick\s*=|document\.cookie|<iframe[^>]*>)`)
	pathTraversalRegex = regexp.MustCompile(`(?i)(\.\./|\.\.\\|%2e%2e%2f|%2e%2e/|\.\.%2f)`)
)

// SecurityGuardOptions configures security guard rules.
type SecurityGuardOptions struct {
	BlockSQLi          bool
	BlockXSS           bool
	BlockPathTraversal bool
}

// DefaultSecurityGuardOptions returns production defaults.
func DefaultSecurityGuardOptions() SecurityGuardOptions {
	return SecurityGuardOptions{
		BlockSQLi:          true,
		BlockXSS:           true,
		BlockPathTraversal: true,
	}
}

// SecurityGuardOption configures SecurityGuardOptions.
type SecurityGuardOption func(*SecurityGuardOptions)

// SecurityGuardMiddleware protects against SQL Injection, XSS, and Path Traversal attacks.
func SecurityGuardMiddleware(opts ...SecurityGuardOption) transportcontract.Middleware {
	cfg := DefaultSecurityGuardOptions()
	for _, o := range opts {
		o(&cfg)
	}

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			req := c.Request()
			if req == nil {
				next(c)
				return
			}

			// 1. Check URL Path & Query Parameters (with URL unescaping)
			urlStr := req.URL.RequestURI()
			if unescaped, err := url.QueryUnescape(urlStr); err == nil {
				urlStr = unescaped
			}
			if isMaliciousPayload(urlStr, cfg) {
				c.JSON(http.StatusBadRequest, map[string]any{
					"error": "security violation: suspicious request payload detected in URL",
				})
				return
			}

			// 2. Check Headers
			for _, vals := range req.Header {
				for _, v := range vals {
					if isMaliciousPayload(v, cfg) {
						c.JSON(http.StatusBadRequest, map[string]any{
							"error": "security violation: suspicious payload detected in header",
						})
						return
					}
				}
			}

			// 3. Check Body if non-empty
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err == nil {
					_ = req.Body.Close()
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					if len(bodyBytes) > 0 && isMaliciousPayload(string(bodyBytes), cfg) {
						c.JSON(http.StatusBadRequest, map[string]any{
							"error": "security violation: malicious payload detected in request body",
						})
						return
					}
				}
			}

			next(c)
		}
	}
}

func isMaliciousPayload(input string, cfg SecurityGuardOptions) bool {
	if cfg.BlockSQLi && sqlInjectionRegex.MatchString(input) {
		return true
	}
	if cfg.BlockXSS && xssRegex.MatchString(input) {
		return true
	}
	if cfg.BlockPathTraversal && pathTraversalRegex.MatchString(input) {
		return true
	}
	return false
}
