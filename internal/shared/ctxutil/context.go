package ctxutil

import "context"

type contextKey string

const (
	SessionIDKey contextKey = "session_id"
	ProfileKey   contextKey = "profile"
)

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

func WithProfile(ctx context.Context, profile string) context.Context {
	return context.WithValue(ctx, ProfileKey, profile)
}

func GetSessionID(ctx context.Context) (string, bool) {
	val := ctx.Value(SessionIDKey)
	if val == nil {
		return "", false
	}
	sessionID, ok := val.(string)
	return sessionID, ok
}

func GetProfile(ctx context.Context) (string, bool) {
	val := ctx.Value(ProfileKey)
	if val == nil {
		return "", false
	}
	profile, ok := val.(string)
	return profile, ok
}
