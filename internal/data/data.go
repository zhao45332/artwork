package data

import (
	"context"
	"time"

	"artwork/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewArtworkRepo,
	NewCategoryRepo,
	NewTagRepo,
	NewImageRepo,
	NewBucketName,
	NewDB,
	NewRedisClient,
	NewAuthRedisRepo,
	NewUserDataRepo,
	NewTokenManager,
)

// Data .
type Data struct {
	db        *gorm.DB
	rdb       *redis.Client
	minio     *minio.Client
	minioCore *minio.Core
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)

	// 初始化 MySQL
	db, err := initDB(c.Database, logger)
	if err != nil {
		return nil, nil, err
	}

	// 自动迁移数据库表
	if err := autoMigrate(db, logger); err != nil {
		logHelper.Errorf("数据库表迁移失败: %v", err)
		return nil, nil, err
	}

	// 初始化 Redis
	rdb, err := initRedis(c.Redis, logger)
	if err != nil {
		return nil, nil, err
	}

	// 初始化 MinIO
	minioClient, err := initMinIO(c.Minio, logger)
	if err != nil {
		return nil, nil, err
	}

	// 确保 MinIO bucket 存在
	if err := ensureBucket(context.Background(), minioClient, c.Minio.BucketName, logger); err != nil {
		logHelper.Warnf("确保 bucket 存在失败: %v", err)
	}

	cleanup := func() {
		logHelper.Info("正在关闭数据资源")
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		if rdb != nil {
			rdb.Close()
		}
	}

	return &Data{
		db:        db,
		rdb:       rdb,
		minio:     minioClient,
		minioCore: &minio.Core{Client: minioClient},
	}, cleanup, nil
}

// initDB 初始化 MySQL 数据库连接
func initDB(c *conf.Data_Database, logger log.Logger) (*gorm.DB, error) {
	logHelper := log.NewHelper(logger)

	db, err := gorm.Open(mysql.Open(c.Source), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent), // 使用 Silent 模式，通过 kratos logger 记录
	})
	if err != nil {
		logHelper.Errorf("连接 MySQL 失败: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		logHelper.Errorf("MySQL 连接测试失败: %v", err)
		return nil, err
	}

	logHelper.Info("MySQL 连接建立成功")
	return db, nil
}

// initRedis 初始化 Redis 连接
func initRedis(c *conf.Data_Redis, logger log.Logger) (*redis.Client, error) {
	logHelper := log.NewHelper(logger)

	opts := &redis.Options{
		Addr:         c.Addr,
		DB:           int(c.Db),
		ReadTimeout:  c.ReadTimeout.AsDuration(),
		WriteTimeout: c.WriteTimeout.AsDuration(),
	}

	if c.Password != "" {
		opts.Password = c.Password
	}

	rdb := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logHelper.Errorf("Redis 连接测试失败: %v", err)
		return nil, err
	}

	logHelper.Info("Redis 连接建立成功")
	return rdb, nil
}

// initMinIO 初始化 MinIO 客户端
func initMinIO(c *conf.Data_MinIO, logger log.Logger) (*minio.Client, error) {
	logHelper := log.NewHelper(logger)

	minioClient, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKeyId, c.SecretAccessKey, ""),
		Secure: c.UseSsl,
	})
	if err != nil {
		logHelper.Errorf("创建 MinIO 客户端失败: %v", err)
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		logHelper.Errorf("列出 MinIO buckets 失败: %v", err)
		return nil, err
	}

	logHelper.Info("MinIO 连接建立成功")
	return minioClient, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB, logger log.Logger) error {
	logHelper := log.NewHelper(logger)

	err := db.AutoMigrate(
		&Category{},
		&Tag{},
		&Artwork{},
		&ArtworkImage{},
		&User{},
		&UserStats{},
		&UserIdentity{},
	)
	if err != nil {
		logHelper.Errorf("数据库表迁移失败: %v", err)
		return err
	}

	logHelper.Info("数据库迁移完成")
	return nil
}

// ensureBucket 确保 bucket 存在，如果不存在则创建
func ensureBucket(ctx context.Context, client *minio.Client, bucketName string, logger log.Logger) error {
	logHelper := log.NewHelper(logger)

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		logHelper.Infof("创建 bucket 成功: %s", bucketName)
	} else {
		logHelper.Infof("Bucket 已存在: %s", bucketName)
	}

	return nil
}
