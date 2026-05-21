package data

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	minio "github.com/minio/minio-go/v7"
)

// MultipartUploadInfo 分片上传信息
type MultipartUploadInfo struct {
	UploadID   string
	ObjectName string
	BucketName string
	Parts      []minio.CompletePart
}

// CreateMultipartUpload 创建分片上传
func (d *Data) CreateMultipartUpload(ctx context.Context, bucketName, objectName, contentType string, logger log.Logger) (string, error) {
	logHelper := log.NewHelper(logger)

	// MinIO Go SDK v7 使用 Core 类型的方法创建分片上传
	uploadID, err := d.minioCore.NewMultipartUpload(ctx, bucketName, objectName, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logHelper.Errorf("创建分片上传失败: %v", err)
		return "", fmt.Errorf("创建分片上传失败: %w", err)
	}

	logHelper.Infof("创建分片上传成功: uploadID=%s, objectName=%s", uploadID, objectName)
	return uploadID, nil
}

// UploadPart 上传分片
func (d *Data) UploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, partData []byte, logger log.Logger) (minio.ObjectPart, error) {
	logHelper := log.NewHelper(logger)

	reader := bytes.NewReader(partData)
	part, err := d.minioCore.PutObjectPart(ctx, bucketName, objectName, uploadID, partNumber, reader, int64(len(partData)), minio.PutObjectPartOptions{})
	if err != nil {
		logHelper.Errorf("上传分片 %d 失败: %v", partNumber, err)
		return minio.ObjectPart{}, fmt.Errorf("上传分片 %d 失败: %w", partNumber, err)
	}

	logHelper.Infof("上传分片 %d 成功: etag=%s", partNumber, part.ETag)
	return part, nil
}

// CompleteMultipartUpload 完成分片上传（合并）
func (d *Data) CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []minio.CompletePart, logger log.Logger) (minio.UploadInfo, error) {
	logHelper := log.NewHelper(logger)

	uploadInfo, err := d.minioCore.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, parts, minio.PutObjectOptions{})
	if err != nil {
		logHelper.Errorf("完成分片上传失败: %v", err)
		return minio.UploadInfo{}, fmt.Errorf("完成分片上传失败: %w", err)
	}

	logHelper.Infof("完成分片上传成功: objectName=%s, etag=%s", objectName, uploadInfo.ETag)
	return uploadInfo, nil
}

// AbortMultipartUpload 取消分片上传
func (d *Data) AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, logger log.Logger) error {
	logHelper := log.NewHelper(logger)

	err := d.minioCore.AbortMultipartUpload(ctx, bucketName, objectName, uploadID)
	if err != nil {
		logHelper.Errorf("取消分片上传失败: %v", err)
		return fmt.Errorf("取消分片上传失败: %w", err)
	}

	logHelper.Infof("取消分片上传成功: uploadID=%s", uploadID)
	return nil
}
