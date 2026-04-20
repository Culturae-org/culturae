// backend/internal/service/user_cache.go

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
)

const (
	UserCachePrefix = "user:"
	SystemConfigKey = "system:config"
)

type UserCacheServiceInterface interface {
	GetUser(ctx context.Context, userID string) (*model.User, error)
	SetUser(ctx context.Context, user *model.User) error
	InvalidateUser(ctx context.Context, userID string) error
	UpdateUserCache(ctx context.Context, user *model.User) error
	IsUserCached(ctx context.Context, userID string) (bool, error)
	GetOrSetUser(ctx context.Context, userID string, fetchFunc func() (*model.User, error)) (*model.User, error)
}

type UserCacheService struct {
	RedisClient *cache.RedisClient
}

func NewUserCacheService(redisClient *cache.RedisClient) *UserCacheService {
	return &UserCacheService{
		RedisClient: redisClient,
	}
}

func (ucs *UserCacheService) getUserCacheTTL(ctx context.Context) time.Duration {
	var cfg model.SystemConfig
	if err := ucs.RedisClient.GetJSON(ctx, SystemConfigKey, &cfg); err != nil || cfg.UserCacheTTLMinutes <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(cfg.UserCacheTTLMinutes) * time.Minute
}

func (ucs *UserCacheService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := ucs.RedisClient.GetJSON(ctx, UserCachePrefix+userID, &user)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(userID)
	if err == nil {
		user.ID = id
	}

	return &user, nil
}

func (ucs *UserCacheService) SetUser(ctx context.Context, user *model.User) error {
	ttl := ucs.getUserCacheTTL(ctx)
	err := ucs.RedisClient.SetJSON(ctx, UserCachePrefix+user.ID.String(), user, ttl)
	if err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}

	return nil
}

func (ucs *UserCacheService) InvalidateUser(ctx context.Context, userID string) error {
	cacheKey := UserCachePrefix + userID
	return ucs.RedisClient.Delete(ctx, cacheKey)
}

func (ucs *UserCacheService) UpdateUserCache(ctx context.Context, user *model.User) error {
	if err := ucs.InvalidateUser(ctx, user.ID.String()); err != nil {
		return err
	}

	if err := ucs.SetUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (ucs *UserCacheService) IsUserCached(ctx context.Context, userID string) (bool, error) {
	return ucs.RedisClient.Exists(ctx, UserCachePrefix+userID)
}

func (ucs *UserCacheService) GetOrSetUser(ctx context.Context, userID string, fetchFunc func() (*model.User, error)) (*model.User, error) {
	if user, err := ucs.GetUser(ctx, userID); err == nil {
		return user, nil
	}

	user, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	if err := ucs.SetUser(ctx, user); err != nil {
		_ = err
		return user, nil
	}

	return user, nil
}
