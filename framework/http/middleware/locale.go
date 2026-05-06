package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// LocaleOptions controls request-level locale negotiation.
//
// LocaleOptions 控制请求级语言协商行为。
type LocaleOptions struct {
	Supported   []string
	Default     string
	QueryKeys   []string
	HeaderKeys  []string
	WriteHeader bool
}

// DefaultLocaleOptions returns the default locale negotiation behavior.
//
// DefaultLocaleOptions 返回默认语言协商配置。
func DefaultLocaleOptions() LocaleOptions {
	return LocaleOptions{
		Supported:   []string{"zh", "en"},
		Default:     "zh",
		QueryKeys:   []string{"lang", "locale"},
		HeaderKeys:  []string{"X-Locale", "Accept-Language"},
		WriteHeader: true,
	}
}

// Locale negotiates the request locale and stores it in the request context.
//
// Locale 协商请求语言，并将结果写入请求上下文。
func Locale(opts LocaleOptions) transportcontract.HTTPMiddleware {
	if len(opts.Supported) == 0 {
		opts = DefaultLocaleOptions()
	}
	if strings.TrimSpace(opts.Default) == "" {
		opts.Default = "zh"
	}
	if len(opts.QueryKeys) == 0 {
		opts.QueryKeys = []string{"lang", "locale"}
	}
	if len(opts.HeaderKeys) == 0 {
		opts.HeaderKeys = []string{"X-Locale", "Accept-Language"}
	}

	supported := normalizeSupportedLocales(opts.Supported)
	defaultLocale := normalizeLocaleTag(opts.Default)

	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if c == nil {
				if next != nil {
					next(c)
				}
				return
			}

			locale := resolveRequestLocale(c, opts, supported, defaultLocale)
			ctx := supportcontract.NewLocaleContext(c.Context(), locale)
			c.SetContext(ctx)
			if opts.WriteHeader {
				c.Header("Content-Language", locale)
			}

			if next != nil {
				next(c)
			}
		}
	}
}

// GetLocale reads the negotiated locale from a Gin request context.
//
// GetLocale 从 Gin 请求上下文中读取协商后的 locale。
func GetLocale(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if locale, ok := supportcontract.FromLocaleContext(c.Request.Context()); ok {
		return locale
	}
	return ""
}

func resolveRequestLocale(c transportcontract.HTTPContext, opts LocaleOptions, supported map[string]string, fallback string) string {
	for _, key := range opts.QueryKeys {
		if value := normalizeLocaleTag(c.Query(key)); value != "" {
			if locale, ok := matchSupportedLocale(value, supported); ok {
				return locale
			}
		}
	}

	for _, key := range opts.HeaderKeys {
		raw := strings.TrimSpace(c.GetHeader(key))
		if raw == "" {
			continue
		}
		if strings.EqualFold(key, "Accept-Language") {
			for _, candidate := range parseAcceptLanguage(raw) {
				if locale, ok := matchSupportedLocale(candidate, supported); ok {
					return locale
				}
			}
			continue
		}
		if locale, ok := matchSupportedLocale(normalizeLocaleTag(raw), supported); ok {
			return locale
		}
	}

	if locale, ok := matchSupportedLocale(fallback, supported); ok {
		return locale
	}
	return fallback
}

func normalizeSupportedLocales(values []string) map[string]string {
	result := make(map[string]string, len(values)*2)
	for _, value := range values {
		normalized := normalizeLocaleTag(value)
		if normalized == "" {
			continue
		}
		result[normalized] = normalized
		base := localeBase(normalized)
		if base != "" {
			if _, ok := result[base]; !ok {
				result[base] = normalized
			}
		}
	}
	return result
}

func normalizeLocaleTag(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func localeBase(value string) string {
	if idx := strings.Index(value, "-"); idx > 0 {
		return value[:idx]
	}
	return value
}

func matchSupportedLocale(value string, supported map[string]string) (string, bool) {
	if value == "" {
		return "", false
	}
	if locale, ok := supported[value]; ok {
		return locale, true
	}
	base := localeBase(value)
	locale, ok := supported[base]
	return locale, ok
}

func parseAcceptLanguage(header string) []string {
	parts := strings.Split(header, ",")
	locales := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.Index(part, ";"); idx >= 0 {
			part = part[:idx]
		}
		part = normalizeLocaleTag(part)
		if part == "" || part == "*" {
			continue
		}
		locales = append(locales, part)
	}
	return locales
}
