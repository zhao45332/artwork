package data

import (
	"context"
	"strings"
	"time"

	pkgAuth "artwork/internal/pkg/auth"

	"github.com/redis/go-redis/v9"
)

type AuthRedisRepo struct {
	rdb *redis.Client
}

func NewAuthRedisRepo(rdb *redis.Client) *AuthRedisRepo {
	return &AuthRedisRepo{rdb: rdb}
}

func (r *AuthRedisRepo) SaveRefreshSession(ctx context.Context, refreshToken string, session *pkgAuth.RefreshSession, ttl time.Duration) error {
	key := r.refreshTokenKey(refreshToken)
	data, err := pkgAuth.EncodeRefreshSession(session)
	if err != nil {
		return err
	}
	if err := r.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return err
	}
	return r.rdb.Set(ctx, r.sessionTokenKey(session.SessionID), strings.TrimSpace(refreshToken), ttl).Err()
}

func (r *AuthRedisRepo) GetRefreshSession(ctx context.Context, refreshToken string) (*pkgAuth.RefreshSession, error) {
	key := r.refreshTokenKey(refreshToken)
	data, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return pkgAuth.DecodeRefreshSession(data)
}

func (r *AuthRedisRepo) DeleteRefreshSession(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}
	session, err := r.GetRefreshSession(ctx, refreshToken)
	if err == nil && session != nil && session.SessionID != "" {
		if err := r.rdb.Del(ctx, r.sessionTokenKey(session.SessionID)).Err(); err != nil {
			return err
		}
	}
	return r.rdb.Del(ctx, r.refreshTokenKey(refreshToken)).Err()
}

func (r *AuthRedisRepo) DeleteRefreshSessionBySessionID(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	refreshToken, err := r.rdb.Get(ctx, r.sessionTokenKey(sessionID)).Result()
	if err != nil {
		return err
	}
	if err := r.rdb.Del(ctx, r.sessionTokenKey(sessionID)).Err(); err != nil {
		return err
	}
	return r.rdb.Del(ctx, r.refreshTokenKey(refreshToken)).Err()
}

func (r *AuthRedisRepo) refreshTokenKey(token string) string {
	return "auth:refresh:" + strings.TrimSpace(token)
}

func (r *AuthRedisRepo) sessionTokenKey(sessionID string) string {
	return "auth:session:" + strings.TrimSpace(sessionID)
}
