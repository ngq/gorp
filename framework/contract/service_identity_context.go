package contract

import "context"

type serviceIdentityContextKey struct{}

// NewServiceIdentityContext stores service identity in context.
func NewServiceIdentityContext(ctx context.Context, identity *ServiceIdentity) context.Context {
	return context.WithValue(ctx, serviceIdentityContextKey{}, identity)
}

// FromServiceIdentityContext retrieves service identity from context.
func FromServiceIdentityContext(ctx context.Context) (*ServiceIdentity, bool) {
	if ctx == nil {
		return nil, false
	}
	identity, ok := ctx.Value(serviceIdentityContextKey{}).(*ServiceIdentity)
	return identity, ok && identity != nil
}
