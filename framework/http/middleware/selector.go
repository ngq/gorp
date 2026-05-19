// Application scenarios:
// - Apply a middleware only to selected paths, methods, or route groups.
// - Build concise route-level governance rules without copying middleware code.
// - Compose middleware selection logic with Any / All / Not helpers.
//
// 适用场景：
// - 仅对指定路径、方法或路由组生效某个中间件。
// - 在不复制中间件代码的前提下构建精简的路由治理规则。
// - 用 Any / All / Not 组合复杂的中间件匹配逻辑。
package middleware

import (
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// HTTPPredicate decides whether the current middleware should apply to the request.
//
// HTTPPredicate 用于判断当前中间件是否应当作用于该请求。
type HTTPPredicate func(transportcontract.Context) bool

// When applies the wrapped middleware only when the predicate matches.
//
// When 仅在谓词命中时应用被包装的中间件。
func When(predicate HTTPPredicate, middleware transportcontract.Middleware) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		wrappedNext := next
		if middleware != nil {
			wrappedNext = middleware(next)
		}
		return func(c transportcontract.Context) {
			if predicate == nil || predicate(c) {
				if wrappedNext != nil {
					wrappedNext(c)
				}
				return
			}
			if next != nil {
				next(c)
			}
		}
	}
}

// MatchPath creates a predicate that matches an exact route or URL path.
//
// MatchPath 创建一个匹配精确路由或 URL 路径的谓词。
func MatchPath(path string) HTTPPredicate {
	return func(c transportcontract.Context) bool {
		if c == nil {
			return false
		}
		if c.RoutePath() == path {
			return true
		}
		if req := c.Request(); req != nil && req.URL != nil {
			return req.URL.Path == path
		}
		return false
	}
}

// MatchPrefix creates a predicate that matches a path prefix.
//
// MatchPrefix 创建一个匹配路径前缀的谓词。
func MatchPrefix(prefix string) HTTPPredicate {
	return func(c transportcontract.Context) bool {
		if c == nil {
			return false
		}
		path := c.RoutePath()
		if path == "" && c.Request() != nil && c.Request().URL != nil {
			path = c.Request().URL.Path
		}
		return strings.HasPrefix(path, prefix)
	}
}

// MatchMethod creates a predicate that matches an HTTP method.
//
// MatchMethod 创建一个匹配 HTTP 方法的谓词。
func MatchMethod(method string) HTTPPredicate {
	return func(c transportcontract.Context) bool {
		if c == nil || c.Request() == nil {
			return false
		}
		return strings.EqualFold(c.Request().Method, method)
	}
}

// Any returns a predicate that matches when any child predicate matches.
//
// Any 返回一个“任一子谓词命中即命中”的组合谓词。
func Any(predicates ...HTTPPredicate) HTTPPredicate {
	return func(c transportcontract.Context) bool {
		for _, predicate := range predicates {
			if predicate != nil && predicate(c) {
				return true
			}
		}
		return false
	}
}

// All returns a predicate that matches only when all child predicates match.
//
// All 返回一个“全部子谓词命中才命中”的组合谓词。
func All(predicates ...HTTPPredicate) HTTPPredicate {
	return func(c transportcontract.Context) bool {
		for _, predicate := range predicates {
			if predicate != nil && !predicate(c) {
				return false
			}
		}
		return true
	}
}

// Not returns the negated predicate.
//
// Not 返回一个取反后的谓词。
func Not(predicate HTTPPredicate) HTTPPredicate {
	return func(c transportcontract.Context) bool {
		if predicate == nil {
			return true
		}
		return !predicate(c)
	}
}
