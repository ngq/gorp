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

// trustedProxies holds the set of IP networks that are considered trusted reverse proxies.
// Only when the request originates from a trusted proxy will X-Forwarded-For and
// X-Real-IP headers be used to determine the client IP.
// By default this is empty, meaning those headers are never trusted — this is the
// safest default because X-Forwarded-For can be spoofed by any client.
//
// trustedProxies 保存被认为是可信反向代理的 IP 网络集合。
// 仅当请求来自可信代理时，才会使用 X-Forwarded-For 和 X-Real-IP 头来确定客户端 IP。
// 默认为空，表示这些头永远不被信任——这是最安全的默认值，
// 因为 X-Forwarded-For 可被任何客户端伪造。
var trustedProxies []*net.IPNet

// SetTrustedProxies configures the networks that are considered trusted reverse proxies.
// Only requests from these networks will have their X-Forwarded-For / X-Real-IP
// headers honoured by requestClientIP and IPKeyFunc.
// Pass nil or an empty slice to disable header trust entirely (safest).
//
// SetTrustedProxies 配置被认为是可信反向代理的网络。
// 仅来自这些网络的请求才会让其 X-Forwarded-For / X-Real-IP 头被
// requestClientIP 和 IPKeyFunc 使用。
// 传入 nil 或空切片可完全禁用头信任（最安全）。
func SetTrustedProxies(cidrs []string) {
	trustedProxies = parseTrustedProxies(cidrs)
}

func parseTrustedProxies(cidrs []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		// Support plain IPs like "127.0.0.1" by converting to CIDR
		if !strings.Contains(cidr, "/") {
			ip := net.ParseIP(cidr)
			if ip == nil {
				continue
			}
			if ipv4 := ip.To4(); ipv4 != nil {
				cidr += "/32"
			} else {
				cidr += "/128"
			}
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil || network == nil {
			continue
		}
		nets = append(nets, network)
	}
	return nets
}

// isFromTrustedProxy reports whether the remote address belongs to a trusted proxy network.
//
// isFromTrustedProxy 判断远端地址是否来自可信代理网络。
func isFromTrustedProxy(remoteAddr string) bool {
	if len(trustedProxies) == 0 {
		return false
	}
	host := remoteAddr
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil && h != "" {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range trustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// IPAllowlist allows only requests whose client IP matches the configured entries.
//
// IPAllowlist 仅允许客户端 IP 命中配置项的请求继续执行。
func IPAllowlist(allowed ...string) transportcontract.Middleware {
	rules := parseIPRules(allowed)
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
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
func IPDenylist(denied ...string) transportcontract.Middleware {
	rules := parseIPRules(denied)
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
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
// X-Forwarded-For and X-Real-IP headers are ONLY used when the request
// originates from a trusted reverse proxy (configured via SetTrustedProxies).
// Otherwise, these headers are ignored because they can be spoofed by any client.
//
// requestClientIP 从请求头或远端地址中提取客户端 IP。
// X-Forwarded-For 和 X-Real-IP 头仅在请求来自可信反向代理时
// （通过 SetTrustedProxies 配置）才会被使用。
// 否则这些头将被忽略，因为它们可被任何客户端伪造。
func requestClientIP(c transportcontract.Context) string {
	if c == nil {
		return ""
	}

	req := c.Request()
	remoteAddr := ""
	if req != nil {
		remoteAddr = strings.TrimSpace(req.RemoteAddr)
	}

	// Only trust forwarding headers when the request comes from a trusted proxy.
	// 仅当请求来自可信代理时才信任转发头。
	if isFromTrustedProxy(remoteAddr) {
		if xff := strings.TrimSpace(c.GetHeader("X-Forwarded-For")); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
		if xri := strings.TrimSpace(c.GetHeader("X-Real-IP")); xri != "" {
			return xri
		}
	}

	// Fallback: use RemoteAddr directly.
	// 回退：直接使用 RemoteAddr。
	if req == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}
	return remoteAddr
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
func respondIPForbidden(c transportcontract.Context) {
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
