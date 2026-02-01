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
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	db    *gorm.DB
	rdb   *redis.Client
	minio *minio.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)

	// 初始化 MySQL
	db, err := initDB(c.Database, logger)
	if err != nil {
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
		logHelper.Warnf("Failed to ensure bucket exists: %v", err)
	}

	cleanup := func() {
		logHelper.Info("closing the data resources")
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		if rdb != nil {
			rdb.Close()
		}
	}

	return &Data{
		db:    db,
		rdb:   rdb,
		minio: minioClient,
	}, cleanup, nil
}

// initDB 初始化 MySQL 数据库连接
func initDB(c *conf.Data_Database, logger log.Logger) (*gorm.DB, error) {
	logHelper := log.NewHelper(logger)

	db, err := gorm.Open(mysql.Open(c.Source), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent), // 使用 Silent 模式，通过 kratos logger 记录
	})
	if err != nil {
		logHelper.Errorf("failed opening connection to mysql: %v", err)
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
		logHelper.Errorf("failed to ping mysql: %v", err)
		return nil, err
	}

	logHelper.Info("MySQL connection established")
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
		logHelper.Errorf("failed to ping redis: %v", err)
		return nil, err
	}

	logHelper.Info("Redis connection established")
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
		logHelper.Errorf("failed to create minio client: %v", err)
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		logHelper.Errorf("failed to list buckets: %v", err)
		return nil, err
	}

	logHelper.Info("MinIO connection established")
	return minioClient, nil
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
		logHelper.Infof("Created bucket: %s", bucketName)
	} else {
		logHelper.Infof("Bucket already exists: %s", bucketName)
	}

	return nil
}
