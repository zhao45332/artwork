package data

import (
	"artwork/internal/conf"
)

// NewBucketName 提供 bucket 名称
func NewBucketName(c *conf.Data) string {
	if c.Minio != nil {
		return c.Minio.BucketName
	}
	return "artwork"
}
