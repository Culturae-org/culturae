// backend/internal/usecase/avatar.go

package usecase

import (
	"context"
	"mime/multipart"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"
	"github.com/Culturae-org/culturae/internal/service"
)

type AvatarUsecase struct {
	UserRepo      repository.UserRepositoryInterface
	AvatarStorage repository.AvatarStorageRepositoryInterface
	AvatarCache   repository.AvatarCacheRepositoryInterface
	LoggingSvc    service.LoggingServiceInterface
}

func NewAvatarUsecase(
	userRepo repository.UserRepositoryInterface,
	avatarStorage repository.AvatarStorageRepositoryInterface,
	avatarCache repository.AvatarCacheRepositoryInterface,
	loggingSvc service.LoggingServiceInterface,
) *AvatarUsecase {
	return &AvatarUsecase{
		UserRepo:      userRepo,
		AvatarStorage: avatarStorage,
		AvatarCache:   avatarCache,
		LoggingSvc:    loggingSvc,
	}
}

// -----------------------------------------------
// Avatar Usecase Methods
//
// - UpdateAvatar
// - GetAvatarBytes
// - UploadAvatar
// - GetAvatarConfig
// - DeleteAvatar
//
// -----------------------------------------------

func (uc *AvatarUsecase) UpdateAvatar(userID string, hasAvatar bool) error {
	return uc.UserRepo.UpdateAvatar(userID, hasAvatar)
}

func (uc *AvatarUsecase) GetAvatarBytes(userID string) (contentType string, data []byte, err error) {
	return uc.AvatarStorage.GetAvatarBytes(userID)
}

func (uc *AvatarUsecase) UploadAvatar(userID string, file *multipart.FileHeader) (string, error) {
	fileName, err := uc.AvatarStorage.UploadAvatar(userID, file)
	if err != nil {
		return "", err
	}
	if err := uc.UserRepo.UpdateAvatar(userID, true); err != nil {
		_ = uc.AvatarStorage.DeleteAvatar(userID)
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	_ = uc.AvatarCache.InvalidateUserCache(ctx, userID)
	return fileName, nil
}

func (uc *AvatarUsecase) GetAvatarConfig(ctx context.Context) model.AvatarConfig {
	cfg, err := uc.AvatarCache.GetAvatarConfig(ctx)
	if err != nil {
		return model.DefaultAvatarConfig()
	}
	return cfg
}

func (uc *AvatarUsecase) DeleteAvatar(userID string) error {
	if err := uc.AvatarStorage.DeleteAvatar(userID); err != nil {
		return err
	}
	if err := uc.UserRepo.UpdateAvatar(userID, false); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	_ = uc.AvatarCache.InvalidateUserCache(ctx, userID)
	return nil
}
