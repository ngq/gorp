// Application scenarios:
// - Restrict internal admin routes to office, bastion, or intranet IP ranges.
// - Block known abusive or disallowed client segments before business logic runs.
// - Protect callback and management endpoints with a simple IP policy.
//
// 适用场景：
// - 将内网管理路由限制在办公网、堡垒机或内网 IP 段内。
// - 在业务逻辑执行前拦截已知风险或不允许的客户端网段。
// - 通过简单 IP 策略保护回调接口和管理接口。
package middleware

import (
	"net"
	"net/http"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// IPAllowlist allows only requests whose client IP matches the configured entries.
//
// IPAllowlist 仅允许客户端 IP 命中配置项的请求继续执行。
func IPAllowlist(allowed ...string) transportcontract.HTTPMiddleware {
	rules := parseIPRules(allowed)
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if len(rules) == 0 {
				if next != nil {
					next(c)
				}
				return
			}

			clientIP := requestClientIP(c)
			if clientIP == "" || !matchesIPRules(clientIP, rules) {
				respondIPForbidden(c)
				return
			}

			if next != nil {
				next(c)
			}
		}
	}
}

// IPDenylist rejects requests whose client IP matches the configured entries.
//
// IPDenylist 拒绝客户端 IP 命中配置项的请求。
func IPDenylist(denied ...string) transportcontract.HTTPMiddleware {
	rules := parseIPRules(denied)
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if len(rules) == 0 {
				if next != nil {
					next(c)
				}
				return
			}

			clientIP := requestClientIP(c)
			if clientIP != "" && matchesIPRules(clientIP, rules) {
				respondIPForbidden(c)
				return
			}

			if next != nil {
				next(c)
			}
		}
	}
}

type ipRule struct {
	ip  net.IP
	net *net.IPNet
}

// parseIPRules parses IP and CIDR values into internal match rules.
//
// parseIPRules 将 IP 和 CIDR 配置解析为内部匹配规则。
func parseIPRules(values []string) []ipRule {
	rules := make([]ipRule, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if strings.Contains(value, "/") {
			if _, network, err := net.ParseCIDR(value); err == nil && network != nil {
				rules = append(rules, ipRule{net: network})
			}
			continue
		}
		if ip := net.ParseIP(value); ip != nil {
			rules = append(rules, ipRule{ip: ip})
		}
	}
	return rules
}

// requestClientIP extracts the client IP from headers or remote address.
//
// requestClientIP 从请求头或远端地址中提取客户端 IP。
func requestClientIP(c transportcontract.HTTPContext) string {
	if c == nil {
		return ""
	}
	if xff := strings.TrimSpace(c.GetHeader("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := strings.TrimSpace(c.GetHeader("X-Real-IP")); xri != "" {
		return xri
	}
	req := c.Request()
	if req == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(req.RemoteAddr)
}

// matchesIPRules reports whether the given IP matches any configured rule.
//
// matchesIPRules 判断给定 IP 是否命中任一配置规则。
func matchesIPRules(rawIP string, rules []ipRule) bool {
	ip := net.ParseIP(strings.TrimSpace(rawIP))
	if ip == nil {
		return false
	}
	for _, rule := range rules {
		if rule.ip != nil && rule.ip.Equal(ip) {
			return true
		}
		if rule.net != nil && rule.net.Contains(ip) {
			return true
		}
	}
	return false
}

// respondIPForbidden writes the unified forbidden response for IP policy failures.
//
// respondIPForbidden 为 IP 策略拒绝场景输出统一无权限响应。
func respondIPForbidden(c transportcontract.HTTPContext) {
	if gc, ok := unwrapGinContext(c); ok {
		writeGinResponseHeaders(gc)
		resp := Response{
			Code:    CodeForbidden,
			Message: "ip address is not allowed",
		}
		gc.JSON(http.StatusForbidden, resp)
		gc.Abort()
		return
	}

	c.JSON(http.StatusForbidden, map[string]any{
		"code":    CodeForbidden,
		"message": "ip address is not allowed",
	})
}
