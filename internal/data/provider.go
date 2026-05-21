package data

import (
	"fmt"
	"time"

	"artwork/internal/conf"
	pkgAuth "artwork/internal/pkg/auth"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func NewDB(data *Data) *gorm.DB {
	return data.db
}

func NewRedisClient(data *Data) *redis.Client {
	return data.rdb
}

// NewBucketName 提供 bucket 名称
func NewBucketName(c *conf.Data) string {
	if c.Minio != nil {
		return c.Minio.BucketName
	}
	return "artwork"
}

func NewTokenManager(c *conf.Data) (*pkgAuth.TokenManager, error) {
	secret := "artwork-dev-secret"
	accessTTL := 2 * time.Hour
	refreshTTL := 7 * 24 * time.Hour
	if c.Auth != nil {
		if c.Auth.JwtSecret != "" {
			secret = c.Auth.JwtSecret
		}
		if c.Auth.AccessTokenExpire != "" {
			parsed, err := time.ParseDuration(c.Auth.AccessTokenExpire)
			if err != nil {
				return nil, fmt.Errorf("解析 access_token_expire 失败: %w", err)
			}
			accessTTL = parsed
		}
		if c.Auth.RefreshTokenExpire != "" {
			parsed, err := time.ParseDuration(c.Auth.RefreshTokenExpire)
			if err != nil {
				return nil, fmt.Errorf("解析 refresh_token_expire 失败: %w", err)
			}
			refreshTTL = parsed
		}
	}
	return pkgAuth.NewTokenManager(secret, accessTTL, refreshTTL)
}
