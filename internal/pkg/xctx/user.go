package xctx

import "context"

type userContextKey struct{}

type AuthUser struct {
	UserID    int64
	Role      int32
	SessionID string
}

func WithAuthUser(ctx context.Context, user *AuthUser) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func AuthUserFromContext(ctx context.Context) (*AuthUser, bool) {
	user, ok := ctx.Value(userContextKey{}).(*AuthUser)
	return user, ok && user != nil
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	user, ok := AuthUserFromContext(ctx)
	if !ok {
		return 0, false
	}
	return user.UserID, true
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	user, ok := AuthUserFromContext(ctx)
	if !ok || user.SessionID == "" {
		return "", false
	}
	return user.SessionID, true
}
