package data

import (
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
		if c.Auth.AccessTokenExpire != nil {
			accessTTL = c.Auth.AccessTokenExpire.AsDuration()
		}
		if c.Auth.RefreshTokenExpire != nil {
			refreshTTL = c.Auth.RefreshTokenExpire.AsDuration()
		}
	}
	return pkgAuth.NewTokenManager(secret, accessTTL, refreshTTL)
}
