package server

import (
	"context"
	"strings"

	"artwork/internal/biz"
	pkgAuth "artwork/internal/pkg/auth"
	"artwork/internal/pkg/xctx"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func NewAuthMiddleware(tokenManager *pkgAuth.TokenManager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			authorization := strings.TrimSpace(tr.RequestHeader().Get("Authorization"))
			if authorization == "" {
				return handler(ctx, req)
			}
			parts := strings.SplitN(authorization, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return nil, kratosErrors.Unauthorized("INVALID_AUTHORIZATION", "Authorization 头格式无效")
			}
			claims, err := tokenManager.ParseAccessToken(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, kratosErrors.Unauthorized("INVALID_TOKEN", "访问令牌无效")
			}
			ctx = xctx.WithAuthUser(ctx, &xctx.AuthUser{UserID: claims.UserID, Role: claims.Role, SessionID: claims.SessionID})
			return handler(ctx, req)
		}
	}
}

func RequireAuth() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if _, ok := xctx.UserIDFromContext(ctx); !ok {
				return nil, biz.ErrUnauthenticated
			}
			return handler(ctx, req)
		}
	}
}
