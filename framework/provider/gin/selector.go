package gin

import (
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type HTTPPredicate func(transportcontract.HTTPContext) bool

func When(predicate HTTPPredicate, middleware transportcontract.HTTPMiddleware) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		wrappedNext := next
		if middleware != nil {
			wrappedNext = middleware(next)
		}
		return func(c transportcontract.HTTPContext) {
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

func MatchPath(path string) HTTPPredicate {
	return func(c transportcontract.HTTPContext) bool {
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

func MatchPrefix(prefix string) HTTPPredicate {
	return func(c transportcontract.HTTPContext) bool {
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

func MatchMethod(method string) HTTPPredicate {
	return func(c transportcontract.HTTPContext) bool {
		if c == nil || c.Request() == nil {
			return false
		}
		return strings.EqualFold(c.Request().Method, method)
	}
}

func Any(predicates ...HTTPPredicate) HTTPPredicate {
	return func(c transportcontract.HTTPContext) bool {
		for _, predicate := range predicates {
			if predicate != nil && predicate(c) {
				return true
			}
		}
		return false
	}
}

func All(predicates ...HTTPPredicate) HTTPPredicate {
	return func(c transportcontract.HTTPContext) bool {
		for _, predicate := range predicates {
			if predicate != nil && !predicate(c) {
				return false
			}
		}
		return true
	}
}

func Not(predicate HTTPPredicate) HTTPPredicate {
	return func(c transportcontract.HTTPContext) bool {
		if predicate == nil {
			return true
		}
		return !predicate(c)
	}
}
