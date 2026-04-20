// backend/internal/repository/avatar.go

package repository

import (
	"context"
	"io"
	"mime/multipart"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/infrastructure/storage"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/fileutil"
)

type AvatarStorageRepositoryInterface interface {
	UploadAvatar(userID string, file *multipart.FileHeader) (string, error)
	DeleteAvatar(userID string) error
	GetAvatarBytes(userID string) (contentType string, data []byte, err error)
}

type AvatarStorageAdapter struct {
	minioClient storage.MinIOClientInterface
}

func NewAvatarStorageAdapter(minioClient storage.MinIOClientInterface) *AvatarStorageAdapter {
	return &AvatarStorageAdapter{minioClient: minioClient}
}

func (a *AvatarStorageAdapter) UploadAvatar(userID string, file *multipart.FileHeader) (string, error) {
	return a.minioClient.UploadAvatar(userID, file)
}

func (a *AvatarStorageAdapter) DeleteAvatar(userID string) error {
	avatarPath := fileutil.FormatAvatarURL(userID)
	return a.minioClient.DeleteAvatar(avatarPath)
}

func (a *AvatarStorageAdapter) GetAvatarBytes(userID string) (contentType string, data []byte, err error) {
	avatarPath := fileutil.FormatAvatarURL(userID)
	objInfo, err := a.minioClient.GetAvatarObjectInfo(avatarPath)
	if err != nil {
		return "", nil, err
	}
	rc, err := a.minioClient.GetAvatarFile(avatarPath)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		_ = rc.Close()
	}()
	data, err = io.ReadAll(rc)
	if err != nil {
		return "", nil, err
	}
	return objInfo.ContentType, data, nil
}

type AvatarCacheRepositoryInterface interface {
	GetAvatarConfig(ctx context.Context) (model.AvatarConfig, error)
	SetAvatarConfig(ctx context.Context, cfg model.AvatarConfig, ttl time.Duration) error
	InvalidateUserCache(ctx context.Context, userID string) error
}

type AvatarCacheAdapter struct {
	redisClient cache.RedisClientInterface
	defaultTTL  time.Duration
}

const avatarConfigKey = "system:avatar:config"

func NewAvatarCacheAdapter(redisClient cache.RedisClientInterface) *AvatarCacheAdapter {
	return &AvatarCacheAdapter{
		redisClient: redisClient,
		defaultTTL:  10 * time.Minute,
	}
}

func (a *AvatarCacheAdapter) GetAvatarConfig(ctx context.Context) (model.AvatarConfig, error) {
	var cfg model.AvatarConfig
	if err := a.redisClient.GetJSON(ctx, avatarConfigKey, &cfg); err != nil {
		return model.DefaultAvatarConfig(), nil
	}
	if cfg.MaxFileSizeMB <= 0 || len(cfg.AllowedMimeTypes) == 0 || len(cfg.AllowedExtensions) == 0 {
		return model.DefaultAvatarConfig(), nil
	}
	return cfg, nil
}

func (a *AvatarCacheAdapter) SetAvatarConfig(ctx context.Context, cfg model.AvatarConfig, ttl time.Duration) error {
	if ttl == 0 {
		ttl = a.defaultTTL
	}
	return a.redisClient.SetJSON(ctx, avatarConfigKey, cfg, ttl)
}

func (a *AvatarCacheAdapter) InvalidateUserCache(ctx context.Context, userID string) error {
	cacheKey := "user:" + userID
	return a.redisClient.Delete(ctx, cacheKey)
}
