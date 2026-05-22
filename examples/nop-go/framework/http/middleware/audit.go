// Application scenarios:
// - Record operator, resource, action, and outcome for business-sensitive endpoints.
// - Provide a stable audit trail for compliance, troubleshooting, and risk control.
// - Reuse request identity, locale, and security context in a single audit record.
//
// 适用场景：
// - 为业务敏感接口记录操作者、资源、动作和结果。
// - 为审计合规、问题排查和风控分析提供稳定审计链路。
// - 在同一条审计记录中复用请求标识、语言环境和安全上下文。
package middleware

import (
	"net/http"
	"strings"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
)

// AuditOptions controls request-level audit logging behavior.
//
// AuditOptions 用于控制请求级审计日志的输出行为。
type AuditOptions struct {
	Event    string
	Action   func(transportcontract.Context) string
	Resource func(transportcontract.Context) string
	Skip     func(transportcontract.Context) bool
}

// AuditMiddleware writes a structured audit log for the current HTTP request.
//
// AuditMiddleware 为当前 HTTP 请求输出结构化审计日志。
//
// Example:
//
//	router.Use(httpmiddleware.AuditMiddleware(logger, httpmiddleware.AuditOptions{}))
func AuditMiddleware(base observabilitycontract.Logger, opts AuditOptions) transportcontract.Middleware {
	event := strings.TrimSpace(opts.Event)
	if event == "" {
		event = "http audit"
	}

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if c == nil {
				if next != nil {
					next(c)
				}
				return
			}
			if opts.Skip != nil && opts.Skip(c) {
				if next != nil {
					next(c)
				}
				return
			}

			start := time.Now()
			if next != nil {
				next(c)
			}

			logger := base
			if logger == nil {
				logger = frameworkbizlog.Ctx(c.Context())
			}

			action := auditAction(c, opts)
			resource := auditResource(c, opts)
			status := c.ResponseStatus()
			if status == 0 {
				status = http.StatusOK
			}
			outcome := "success"
			if status >= http.StatusBadRequest {
				outcome = "failure"
			}

			fields := []observabilitycontract.Field{
				{Key: "kind", Value: "audit"},
				{Key: "action", Value: action},
				{Key: "resource", Value: resource},
				{Key: "outcome", Value: outcome},
				{Key: "status", Value: status},
				{Key: "latency_ms", Value: time.Since(start).Milliseconds()},
			}

			if req := c.Request(); req != nil {
				fields = append(fields,
					observabilitycontract.Field{Key: "method", Value: req.Method},
				)
				if req.URL != nil {
					fields = append(fields, observabilitycontract.Field{Key: "path", Value: req.URL.Path})
				}
			}
			if route := c.RoutePath(); route != "" {
				fields = append(fields, observabilitycontract.Field{Key: "route", Value: route})
			}
			if rid, ok := supportcontract.FromRequestIDContext(c.Context()); ok && rid != "" {
				fields = append(fields, observabilitycontract.Field{Key: "request_id", Value: rid})
			}
			if tid, ok := supportcontract.FromTraceIDContext(c.Context()); ok && tid != "" {
				fields = append(fields, observabilitycontract.Field{Key: "trace_id", Value: tid})
			}
			if locale, ok := supportcontract.FromLocaleContext(c.Context()); ok && locale != "" {
				fields = append(fields, observabilitycontract.Field{Key: "locale", Value: locale})
			}
			fields = append(fields, auditActorFields(c.Context())...)

			logger.Info(event, fields...)
		}
	}
}

// auditAction resolves the audit action label for the current request.
//
// auditAction 解析当前请求对应的审计动作名称。
func auditAction(c transportcontract.Context, opts AuditOptions) string {
	if opts.Action != nil {
		if value := strings.TrimSpace(opts.Action(c)); value != "" {
			return value
		}
	}
	if c != nil && c.Request() != nil {
		return strings.ToUpper(strings.TrimSpace(c.Request().Method))
	}
	return "UNKNOWN"
}

// auditResource resolves the audit resource label for the current request.
//
// auditResource 解析当前请求对应的审计资源标识。
func auditResource(c transportcontract.Context, opts AuditOptions) string {
	if opts.Resource != nil {
		if value := strings.TrimSpace(opts.Resource(c)); value != "" {
			return value
		}
	}
	if c == nil {
		return ""
	}
	if route := strings.TrimSpace(c.RoutePath()); route != "" {
		return route
	}
	if req := c.Request(); req != nil && req.URL != nil {
		return req.URL.Path
	}
	return ""
}

// auditActorFields builds actor-related audit fields from request context.
//
// auditActorFields 从请求上下文中提取操作者相关审计字段。
func auditActorFields(ctx any) []observabilitycontract.Field {
	baseCtx, ok := ctx.(interface{ Done() <-chan struct{} })
	_ = baseCtx
	contextValue, ok := ctx.(interface{})
	if !ok {
		return nil
	}

	typedCtx, ok := contextValue.(interface {
		Value(any) any
	})
	_ = typedCtx

	realCtx, ok := ctx.(interface {
		Deadline() (time.Time, bool)
		Done() <-chan struct{}
		Err() error
		Value(any) any
	})
	if !ok {
		return nil
	}

	fields := make([]observabilitycontract.Field, 0, 5)
	if claims, ok := securitycontract.FromJWTClaimsContext(realCtx); ok && claims != nil {
		fields = append(fields,
			observabilitycontract.Field{Key: "actor_kind", Value: "user"},
			observabilitycontract.Field{Key: "actor_id", Value: claims.SubjectID},
			observabilitycontract.Field{Key: "actor_type", Value: claims.SubjectType},
		)
		if strings.TrimSpace(claims.SubjectName) != "" {
			fields = append(fields, observabilitycontract.Field{Key: "actor_name", Value: claims.SubjectName})
		}
		if len(claims.Roles) > 0 {
			fields = append(fields, observabilitycontract.Field{Key: "actor_roles", Value: strings.Join(claims.Roles, ",")})
		}
		return fields
	}
	if identity, ok := securitycontract.FromServiceIdentityContext(realCtx); ok && identity != nil {
		fields = append(fields,
			observabilitycontract.Field{Key: "actor_kind", Value: "service"},
			observabilitycontract.Field{Key: "actor_id", Value: identity.ServiceID},
			observabilitycontract.Field{Key: "actor_name", Value: identity.ServiceName},
		)
		if strings.TrimSpace(identity.Namespace) != "" {
			fields = append(fields, observabilitycontract.Field{Key: "actor_namespace", Value: identity.Namespace})
		}
		if strings.TrimSpace(identity.Environment) != "" {
			fields = append(fields, observabilitycontract.Field{Key: "actor_environment", Value: identity.Environment})
		}
	}
	return fields
}
