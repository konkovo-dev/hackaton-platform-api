package auth

import "context"

type contextKey string

const (
	claimsContextKey      contextKey = "auth:claims"
	serviceCallContextKey contextKey = "auth:service_call"
)

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func GetUserID(ctx context.Context) (string, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.UserID, true
}

func WithServiceCall(ctx context.Context) context.Context {
	return context.WithValue(ctx, serviceCallContextKey, true)
}

func IsServiceCall(ctx context.Context) bool {
	val, ok := ctx.Value(serviceCallContextKey).(bool)
	return ok && val
}
