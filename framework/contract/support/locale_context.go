package support

import "context"

type localeContextKey struct{}

func NewLocaleContext(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey{}, locale)
}

func FromLocaleContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	locale, ok := ctx.Value(localeContextKey{}).(string)
	return locale, ok && locale != ""
}
