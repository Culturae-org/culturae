// backend/internal/handler/avatar.go

package handler

import (
	"github.com/Culturae-org/culturae/internal/pkg/fileutil"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AvatarHandler struct {
	UserUsecase    *usecase.UserUsecase
	AvatarUsecase  *usecase.AvatarUsecase
	FriendsUsecase *usecase.FriendsUsecase
	LoggingService service.LoggingServiceInterface
	logger         *zap.Logger
}

func NewAvatarHandler(
	userUc *usecase.UserUsecase,
	avatarUc *usecase.AvatarUsecase,
	friendsUc *usecase.FriendsUsecase,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AvatarHandler {
	return &AvatarHandler{
		UserUsecase:    userUc,
		AvatarUsecase:  avatarUc,
		FriendsUsecase: friendsUc,
		LoggingService: loggingService,
		logger:         logger,
	}
}

// -----------------------------------------------------
// Avatar Handlers
//
// - UploadAvatar
// - GetAvatar
// - DeleteAvatar
// -----------------------------------------------------

func (ac *AvatarHandler) UploadAvatar(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		ac.logger.Warn("Invalid or missing user_id in context")
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		errorMsg := err.Error()
		go func() {
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_upload",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error": "no_file_uploaded",
				},
				false,
				&errorMsg)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "No file uploaded")
		return
	}

	avatarCfg := ac.AvatarUsecase.GetAvatarConfig(c.Request.Context())

	if !fileutil.IsValidFileTypeWithConfig(file, avatarCfg.AllowedMimeTypes, avatarCfg.AllowedExtensions) {
		go func() {
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_upload",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error":     "invalid_file_type",
					"file_name": file.Filename,
					"file_size": file.Size,
				},
				false,
				nil)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid file type. Only .jpg, .jpeg, .png are allowed.")
		return
	}

	if file.Size > avatarCfg.MaxFileSizeBytes() {
		go func() {
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_upload",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error":     "file_too_large",
					"file_name": file.Filename,
					"file_size": file.Size,
					"max_size":  avatarCfg.MaxFileSizeBytes(),
				},
				false,
				nil)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "File size exceeds the configured limit.")
		return
	}

	if _, err := ac.UserUsecase.GetByID(userID.String()); err != nil {
		errorMsg := err.Error()
		go func() {
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_upload",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error": "failed_to_fetch_user",
				},
				false,
				&errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	_, err = ac.AvatarUsecase.UploadAvatar(userID.String(), file)
	if err != nil {
		errorMsg := err.Error()
		go func() {
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_upload",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error":     "failed_to_upload_avatar",
					"file_name": file.Filename,
					"file_size": file.Size,
				},
				false,
				&errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to upload avatar")
		return
	}

	go func() {
		_ = ac.LoggingService.LogUserAction(
			userID,
			"avatar_upload",
			httputil.GetRealIP(c),
			httputil.GetUserAgent(c),
			map[string]interface{}{
				"file_size": file.Size,
				"file_name": file.Filename,
			},
			true,
			nil)
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Avatar uploaded successfully", gin.H{
		"success":    true,
		"has_avatar": true,
	})
}

func (ac *AvatarHandler) GetAvatar(c *gin.Context) {
	publicID := c.Param("publicID")

	user, err := ac.UserUsecase.GetByPublicID(publicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	currentUserID := httputil.GetUserIDFromContext(c)
	if !user.IsProfilePublic {
		if currentUserID == uuid.Nil {
			httputil.Error(c, http.StatusForbidden, httputil.ErrCodeForbidden, "Profile is private")
			return
		}
		isFriend, err := ac.FriendsUsecase.IsFriend(currentUserID, user.ID)
		if err != nil || !isFriend {
			httputil.Error(c, http.StatusForbidden, httputil.ErrCodeForbidden, "Profile is private")
			return
		}
	}

	if !user.HasAvatar {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "No avatar found")
		return
	}

	contentType, data, err := ac.AvatarUsecase.GetAvatarBytes(user.ID.String())
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Avatar not found")
		return
	}
	c.Data(http.StatusOK, contentType, data)
}

func (ac *AvatarHandler) DeleteAvatar(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		ac.logger.Warn("Invalid or missing user_id in context")
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	user, err := ac.UserUsecase.GetByID(userID.String())
	if err != nil {
		go func() {
			errorMsg := err.Error()
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_delete",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error": "failed_to_fetch_user",
				},
				false,
				&errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	if !user.HasAvatar {
		go func() {
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_delete",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error": "no_avatar_to_delete",
				},
				false,
				nil)
		}()
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "No avatar to delete")
		return
	}

	if err := ac.AvatarUsecase.DeleteAvatar(userID.String()); err != nil {
		go func() {
			errorMsg := err.Error()
			_ = ac.LoggingService.LogUserAction(
				userID,
				"avatar_delete",
				httputil.GetRealIP(c),
				httputil.GetUserAgent(c),
				map[string]interface{}{
					"error":       "failed_to_delete_avatar",
					"avatar_path": fileutil.FormatAvatarURL(userID.String()),
				},
				false,
				&errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to delete avatar")
		return
	}

	go func() {
		_ = ac.LoggingService.LogUserAction(
			userID,
			"avatar_delete",
			httputil.GetRealIP(c),
			httputil.GetUserAgent(c),
			map[string]interface{}{
				"avatar_path": fileutil.FormatAvatarURL(userID.String()),
			},
			true,
			nil)
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Avatar deleted successfully", gin.H{
		"success":    true,
		"has_avatar": false,
	})
}
