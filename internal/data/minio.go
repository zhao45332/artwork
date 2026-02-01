package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	minio "github.com/minio/minio-go/v7"
)

// UploadImage 上传图片到 MinIO
func (d *Data) UploadImage(ctx context.Context, filename string, content []byte, contentType string, bucketName string) (string, error) {
	// 生成对象名称（使用时间戳避免冲突）
	objectName := fmt.Sprintf("artwork/%s/%d_%s",
		time.Now().Format("20060102"),
		time.Now().Unix(),
		filename)

	// 上传文件
	reader := bytes.NewReader(content)
	_, err := d.minio.PutObject(ctx, bucketName, objectName, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("上传图片到 MinIO 失败: %w", err)
	}

	// 返回对象URL（这里返回对象名称，实际使用时需要配置 MinIO 的访问地址）
	// 如果需要公开访问，可以使用 PresignedGetObject 生成预签名URL
	url := fmt.Sprintf("/images/%s", objectName)
	return url, nil
}

// GetImageURL 获取图片访问URL
func (d *Data) GetImageURL(ctx context.Context, objectName string, bucketName string, expiry time.Duration) (string, error) {
	// 生成预签名URL（用于临时访问）
	url, err := d.minio.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}
	return url.String(), nil
}

// DeleteImage 删除图片
func (d *Data) DeleteImage(ctx context.Context, objectName string, bucketName string) error {
	err := d.minio.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("从 MinIO 删除图片失败: %w", err)
	}
	return nil
}

// GetImage 获取图片内容
func (d *Data) GetImage(ctx context.Context, objectName string, bucketName string) (io.ReadCloser, error) {
	obj, err := d.minio.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("从 MinIO 获取图片失败: %w", err)
	}
	return obj, nil
}
