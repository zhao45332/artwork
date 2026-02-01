package data

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"artwork/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	minio "github.com/minio/minio-go/v7"
)

type imageRepo struct {
	data       *Data
	bucketName string
	log        *log.Helper
}

// NewImageRepo 创建图片仓储
func NewImageRepo(data *Data, bucketName string, logger log.Logger) biz.ImageRepo {
	return &imageRepo{
		data:       data,
		bucketName: bucketName,
		log:        log.NewHelper(logger),
	}
}

// UploadImage 上传图片到 MinIO
func (r *imageRepo) UploadImage(ctx context.Context, filename string, content []byte, contentType string, bucketName string) (string, error) {
	// 生成对象名称（使用时间戳避免冲突）
	objectName := fmt.Sprintf("artwork/%s/%d_%s",
		time.Now().Format("20060102"),
		time.Now().Unix(),
		filename)

	// 上传文件
	reader := bytes.NewReader(content)
	_, err := r.data.minio.PutObject(ctx, bucketName, objectName, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		r.log.Errorf("上传图片到 MinIO 失败: %v", err)
		return "", fmt.Errorf("上传图片到 MinIO 失败: %w", err)
	}

	// 返回对象名称（实际使用时需要配置 MinIO 的访问地址）
	// 如果需要公开访问，可以使用 PresignedGetObject 生成预签名URL
	url := fmt.Sprintf("/images/%s", objectName)
	r.log.Infof("图片上传成功: %s", objectName)
	return url, nil
}
