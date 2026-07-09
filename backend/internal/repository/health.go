// backend/internal/repository/monitoring.go

package repository

import (
	"context"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/infrastructure/storage"
	"github.com/Culturae-org/culturae/internal/model"
)

type MonitoringRepositoryInterface interface {
	GetSystemMetrics() (*model.SystemMetrics, error)
	GetAPIRequestStats(startDate, endDate *time.Time) (*model.APIRequestStats, error)
	GetAdminActionStats(startDate, endDate *time.Time) (*model.AdminActionStats, error)
	GetUserActionStats(startDate, endDate *time.Time) (*model.UserActionStats, error)
	GetAPIRequestTimestamps(method *string, statusCode *int, startDate, endDate *time.Time) ([]time.Time, error)
	GetDatabaseInfo() (map[string]interface{}, error)
	CheckDatabaseHealth() error
}

type CacheHealthRepositoryInterface interface {
	Ping(ctx context.Context) error
	GetInfo(ctx context.Context) (map[string]interface{}, error)
}

type CacheHealthAdapter struct {
	redisClient cache.RedisClientInterface
}

func NewCacheHealthAdapter(redisClient cache.RedisClientInterface) *CacheHealthAdapter {
	return &CacheHealthAdapter{redisClient: redisClient}
}

func (c *CacheHealthAdapter) Ping(ctx context.Context) error {
	return c.redisClient.Ping(ctx)
}

func (c *CacheHealthAdapter) GetInfo(ctx context.Context) (map[string]interface{}, error) {
	return c.redisClient.GetInfo(ctx)
}

type StorageHealthRepositoryInterface interface {
	CheckBucketExists() error
	GetBucketInfo() (map[string]interface{}, error)
}

type StorageHealthAdapter struct {
	s3Client storage.S3ClientInterface
}

func NewStorageHealthAdapter(s3Client storage.S3ClientInterface) *StorageHealthAdapter {
	return &StorageHealthAdapter{s3Client: s3Client}
}

func (s *StorageHealthAdapter) CheckBucketExists() error {
	return s.s3Client.CheckBucketExists()
}

func (s *StorageHealthAdapter) GetBucketInfo() (map[string]interface{}, error) {
	return s.s3Client.GetBucketInfo()
}
