// Package middleware provides IP restriction and Geo-IP filtering middleware for gorp framework.
package middleware

import (
	"net"
	"net/http"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// IPRestrictionOptions configures IP filtering and Geo-IP access control.
type IPRestrictionOptions struct {
	AllowIPs       []string
	BlockIPs       []string
	AllowCountries []string
	BlockCountries []string
	GeoIPResolver  func(ip string) string // Resolves IP to Country Code (e.g. "CN")
	TrustedProxies []string               // CIDRs/IPs of trusted reverse proxies (default: 127.0.0.1, ::1)
}

// DefaultIPRestrictionOptions returns default configuration with localhost trusted proxy.
func DefaultIPRestrictionOptions() IPRestrictionOptions {
	return IPRestrictionOptions{
		TrustedProxies: []string{"127.0.0.1", "::1"},
	}
}

// IPRestrictionOption configures IPRestrictionOptions.
type IPRestrictionOption func(*IPRestrictionOptions)

// WithAllowIPs adds IP/CIDR to the whitelist.
func WithAllowIPs(ips ...string) IPRestrictionOption {
	return func(o *IPRestrictionOptions) {
		o.AllowIPs = append(o.AllowIPs, ips...)
	}
}

// WithBlockIPs adds IP/CIDR to the blacklist.
func WithBlockIPs(ips ...string) IPRestrictionOption {
	return func(o *IPRestrictionOptions) {
		o.BlockIPs = append(o.BlockIPs, ips...)
	}
}

// WithAllowCountries adds country codes to the whitelist.
func WithAllowCountries(countries ...string) IPRestrictionOption {
	return func(o *IPRestrictionOptions) {
		o.AllowCountries = append(o.AllowCountries, countries...)
	}
}

// WithGeoIPResolver configures a custom Geo-IP resolver function.
func WithGeoIPResolver(resolver func(ip string) string) IPRestrictionOption {
	return func(o *IPRestrictionOptions) {
		o.GeoIPResolver = resolver
	}
}

// WithTrustedProxies sets trusted reverse proxy IPs or CIDRs allowed to forward client IPs.
func WithTrustedProxies(proxies ...string) IPRestrictionOption {
	return func(o *IPRestrictionOptions) {
		o.TrustedProxies = append(o.TrustedProxies, proxies...)
	}
}

// IPRestrictionMiddleware provides IP CIDR whitelist/blacklist and Geo-IP region filtering with spoofing protection.
func IPRestrictionMiddleware(opts ...IPRestrictionOption) transportcontract.Middleware {
	cfg := DefaultIPRestrictionOptions()
	for _, o := range opts {
		o(&cfg)
	}

	allowNets := parseCIDRNets(cfg.AllowIPs)
	blockNets := parseCIDRNets(cfg.BlockIPs)
	trustedNets := parseCIDRNets(cfg.TrustedProxies)

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			clientIP := resolveClientIP(c, cfg.TrustedProxies, trustedNets)

			// 1. Blacklist CIDR check
			if matchIPOrCIDR(clientIP, cfg.BlockIPs, blockNets) {
				c.JSON(http.StatusForbidden, map[string]any{
					"error": "access denied: client IP is blacklisted",
				})
				return
			}

			// 2. Whitelist CIDR check (if specified)
			if len(cfg.AllowIPs) > 0 && !matchIPOrCIDR(clientIP, cfg.AllowIPs, allowNets) {
				c.JSON(http.StatusForbidden, map[string]any{
					"error": "access denied: client IP is not in whitelist",
				})
				return
			}

			// 3. Geo-IP Country check
			if cfg.GeoIPResolver != nil {
				country := cfg.GeoIPResolver(clientIP)

				if len(cfg.BlockCountries) > 0 {
					for _, bc := range cfg.BlockCountries {
						if strings.EqualFold(bc, country) {
							c.JSON(http.StatusForbidden, map[string]any{
								"error": "access denied: country/region is blocked",
							})
							return
						}
					}
				}

				if len(cfg.AllowCountries) > 0 {
					allowed := false
					for _, ac := range cfg.AllowCountries {
						if strings.EqualFold(ac, country) {
							allowed = true
							break
						}
					}
					if !allowed {
						c.JSON(http.StatusForbidden, map[string]any{
							"error": "access denied: country/region is not allowed",
						})
						return
					}
				}
			}

			next(c)
		}
	}
}

func parseCIDRNets(ipList []string) []*net.IPNet {
	var nets []*net.IPNet
	for _, item := range ipList {
		if strings.Contains(item, "/") {
			_, ipNet, err := net.ParseCIDR(item)
			if err == nil {
				nets = append(nets, ipNet)
			}
		}
	}
	return nets
}

func matchIPOrCIDR(clientIP string, strList []string, cidrNets []*net.IPNet) bool {
	parsedIP := net.ParseIP(clientIP)
	for _, str := range strList {
		if str == clientIP {
			return true
		}
	}
	if parsedIP != nil {
		for _, netRange := range cidrNets {
			if netRange.Contains(parsedIP) {
				return true
			}
		}
	}
	return false
}

func resolveClientIP(c transportcontract.Context, trustedProxies []string, trustedNets []*net.IPNet) string {
	req := c.Request()
	if req == nil {
		return "127.0.0.1"
	}

	remoteIP := req.RemoteAddr
	if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		remoteIP = host
	}

	// If remote IP is NOT a trusted proxy, never trust X-Forwarded-For or X-Real-IP headers (prevents IP spoofing)
	if !matchIPOrCIDR(remoteIP, trustedProxies, trustedNets) {
		return remoteIP
	}

	// Remote IP is trusted proxy: parse X-Forwarded-For or X-Real-IP
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		// Traverse XFF from right to left to find the first untrusted upstream IP
		for i := len(parts) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(parts[i])
			if net.ParseIP(ip) != nil {
				if !matchIPOrCIDR(ip, trustedProxies, trustedNets) {
					return ip
				}
			}
		}
		// If all IPs in XFF are trusted, return the leftmost client IP
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	return remoteIP
}
