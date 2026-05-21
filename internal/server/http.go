package server

import (
	"fmt"
	"io"
	"strconv"
	"time"

	artworkv1 "artwork/api/artwork/v1"
	authv1 "artwork/api/auth/v1"
	userv1 "artwork/api/user/v1"
	"artwork/internal/conf"
	"artwork/internal/data"
	"artwork/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	minio "github.com/minio/minio-go/v7"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, artwork *service.ArtworkService, auth *service.AuthService, user *service.UserService, data *data.Data, bucketName string, authMiddleware middleware.Middleware, logger log.Logger) *kratosHttp.Server {
	var opts = []kratosHttp.ServerOption{
		kratosHttp.Middleware(
			recovery.Recovery(),
			authMiddleware,
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratosHttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratosHttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratosHttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratosHttp.NewServer(opts...)
	authv1.RegisterAuthServiceHTTPServer(srv, auth)
	artworkv1.RegisterArtworkServiceHTTPServer(srv, artwork)
	userv1.RegisterUserServiceHTTPServer(srv, user)

	registerMultipartUploadRoutes(srv, data, bucketName, logger)

	return srv
}

// 分片上传请求结构
type ChunkUploadRequest struct {
	UploadID   string `json:"upload_id"`
	PartNumber int    `json:"part_number"`
	Filename   string `json:"filename"`
}

// 分片上传响应结构
type ChunkUploadResponse struct {
	UploadID   string `json:"upload_id"`
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// 合并上传请求结构
type CompleteUploadRequest struct {
	UploadID    string               `json:"upload_id"`
	Filename    string               `json:"filename"`
	Parts       []minio.CompletePart `json:"parts"`
	ContentType string               `json:"content_type"`
}

// 合并上传响应结构
type CompleteUploadResponse struct {
	URL        string `json:"url"`
	ObjectName string `json:"object_name"`
	ETag       string `json:"etag"`
}

// registerMultipartUploadRoutes 注册分片上传路由
func registerMultipartUploadRoutes(srv *kratosHttp.Server, data *data.Data, bucketName string, logger log.Logger) {
	logHelper := log.NewHelper(logger)
	r := srv.Route("/")

	// 创建分片上传
	r.POST("/v1/artwork/upload/init", func(ctx kratosHttp.Context) error {
		var req struct {
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
		}

		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": fmt.Sprintf("请求参数无效: %v", err),
			})
		}

		// 生成对象名称
		objectName := fmt.Sprintf("artwork/%s/%d_%s",
			time.Now().Format("20060102"),
			time.Now().Unix(),
			req.Filename)

		contentType := req.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		uploadID, err := data.CreateMultipartUpload(ctx.Request().Context(), bucketName, objectName, contentType, logger)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"upload_id":   uploadID,
			"object_name": objectName,
		})
	})

	// 上传分片
	r.POST("/v1/artwork/upload/chunk", func(ctx kratosHttp.Context) error {
		// 解析表单数据
		uploadID := ctx.Request().FormValue("upload_id")
		objectName := ctx.Request().FormValue("object_name")
		partNumberStr := ctx.Request().FormValue("part_number")

		if uploadID == "" || objectName == "" || partNumberStr == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "缺少必填字段: upload_id, object_name, part_number",
			})
		}

		partNumber, err := strconv.Atoi(partNumberStr)
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": fmt.Sprintf("分片序号无效: %v", err),
			})
		}

		// 读取分片数据
		file, _, err := ctx.Request().FormFile("chunk")
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": fmt.Sprintf("读取分片文件失败: %v", err),
			})
		}
		defer file.Close()

		partData, err := io.ReadAll(file)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": fmt.Sprintf("读取分片数据失败: %v", err),
			})
		}

		// 上传分片
		part, err := data.UploadPart(ctx.Request().Context(), bucketName, objectName, uploadID, partNumber, partData, logger)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, ChunkUploadResponse{
			UploadID:   uploadID,
			PartNumber: partNumber,
			ETag:       part.ETag,
		})
	})

	// 合并分片
	r.POST("/v1/artwork/upload/complete", func(ctx kratosHttp.Context) error {
		var req CompleteUploadRequest

		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": fmt.Sprintf("请求参数无效: %v", err),
			})
		}

		if req.UploadID == "" || len(req.Parts) == 0 || req.Filename == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "缺少必填字段: upload_id, parts, filename",
			})
		}

		// 从请求中获取 object_name，如果没有则根据 filename 生成（兼容旧版本）
		objectName := ctx.Request().FormValue("object_name")
		if objectName == "" {
			// 如果没有传递 object_name，则根据 filename 生成（兼容性处理）
			objectName = fmt.Sprintf("artwork/%s/%d_%s",
				time.Now().Format("20060102"),
				time.Now().Unix(),
				req.Filename)
		}

		// 合并分片
		uploadInfo, err := data.CompleteMultipartUpload(ctx.Request().Context(), bucketName, objectName, req.UploadID, req.Parts, logger)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		url := fmt.Sprintf("/images/%s", objectName)

		return ctx.JSON(200, CompleteUploadResponse{
			URL:        url,
			ObjectName: objectName,
			ETag:       uploadInfo.ETag,
		})
	})

	logHelper.Info("分片上传路由注册成功")
}
