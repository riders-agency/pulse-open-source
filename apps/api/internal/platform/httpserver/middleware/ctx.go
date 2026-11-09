package middleware

import (
	"context"

	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/platform/authz"
)

type ridKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ridKey{}, id)
}

func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(ridKey{}).(string)
	return v
}

func NewRequestID() string {
	return uuid.NewString()
}

type SessionClaims = authz.SessionClaims

func WithSessionClaims(ctx context.Context, s SessionClaims) context.Context {
	return authz.WithSessionClaims(ctx, s)
}

func SessionClaimsFrom(ctx context.Context) (SessionClaims, bool) {
	return authz.SessionClaimsFrom(ctx)
}
