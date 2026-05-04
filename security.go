package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/contract/security"
	"github.com/ngq/gorp/framework/facade"
)

type ServiceIdentity = security.ServiceIdentity

func WithServiceIdentity(ctx context.Context, identity *ServiceIdentity) context.Context {
	return facade.WithServiceIdentity(ctx, identity)
}

func FromServiceIdentity(ctx context.Context) (*ServiceIdentity, bool) {
	return facade.FromServiceIdentity(ctx)
}

func FromJWTClaimsContext(ctx context.Context) (*security.JWTClaims, bool) {
	return security.FromJWTClaimsContext(ctx)
}

func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	return security.FromSubjectIDContext(ctx)
}

func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	return security.FromSubjectTypeContext(ctx)
}
