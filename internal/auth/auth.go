package auth

import (
	"context"
	"fmt"
	"net/http"

	"PingCodeMcp/internal/config"
)

var (
	authToken string
)

// authKey is a custom context key for storing the auth token.
type authKey struct{}

// WithAuthKey adds an auth key to the context.
func WithAuthKey(ctx context.Context, auth string) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

// AuthFromRequest extracts the auth token from the request headers.
func AuthFromRequest(ctx context.Context, r *http.Request) context.Context {
	authToken = getAuthToken()
	if authToken != "" {
		return WithAuthKey(ctx, authToken)
	}
	return WithAuthKey(ctx, r.Header.Get("Authorization"))
}

// TokenFromContext extracts the auth token from the context.
// This can be used by tools to extract the token regardless of the
// transport being used by the server.
func TokenFromContext(ctx context.Context) (string, error) {
	auth, ok := ctx.Value(authKey{}).(string)
	if !ok {
		return "", fmt.Errorf("missing auth")
	}
	return auth, nil
}

// 获取授权 Token 信息
func getAuthToken() string {
	if authToken == "" {
		authToken = config.GetPingCodeAuthToken()
	}
	return authToken
}
