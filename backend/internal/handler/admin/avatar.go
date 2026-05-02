// backend/internal/handler/admin/avatar.go

package admin

import (
	"log"
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/fileutil"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminAvatarHandler struct {
	UserUsecase    *usecase.UserUsecase
	AvatarUsecase  *usecase.AvatarUsecase
	LoggingService service.LoggingServiceInterface
}

func NewAdminAvatarHandler(
	userUc *usecase.UserUsecase,
	avatarUc *usecase.AvatarUsecase,
	loggingService service.LoggingServiceInterface,
) *AdminAvatarHandler {
	return &AdminAvatarHandler{
		UserUsecase:    userUc,
		AvatarUsecase:  avatarUc,
		LoggingService: loggingService,
	}
}

// -----------------------------------------------------
// Admin Avatar Handlers
//
// - GetUserAvatar
// - UploadUserAvatar
// - DeleteUserAvatar
// -----------------------------------------------------

func (ac *AdminAvatarHandler) GetUserAvatar(c *gin.Context) {
	userID := c.Param("userID")

	user, err := ac.UserUsecase.GetByID(userID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	if !user.HasAvatar {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User has no avatar")
		return
	}

	contentType, data, err := ac.AvatarUsecase.GetAvatarBytes(userID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Avatar not found")
		return
	}
	c.Data(http.StatusOK, contentType, data)
}

func (ac *AdminAvatarHandler) UploadUserAvatar(c *gin.Context) {
	userID := c.Param("userID")

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		log.Printf("Invalid user_id in context")
		return
	}
	adminName := c.GetString("username")
	action := "avatar_upload"
	resourceUUID := uuid.MustParse(userID)

	file, err := c.FormFile("avatar")
	if err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError: "no_file_uploaded",
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "No file uploaded")
		return
	}

	avatarCfg := ac.AvatarUsecase.GetAvatarConfig(c.Request.Context())

	if !fileutil.IsValidFileTypeWithConfig(file, avatarCfg.AllowedMimeTypes, avatarCfg.AllowedExtensions) {
		errorMsg := "Invalid file type. Only .jpg, .jpeg, .png are allowed."
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError:     "invalid_file_type",
				keyFileName: file.Filename,
				keyFileSize: file.Size,
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid file type. Only .jpg, .jpeg, .png are allowed.")
		return
	}

	if file.Size > avatarCfg.MaxFileSizeBytes() {
		errorMsg := "File size exceeds the configured limit."
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError:     "file_too_large",
				keyFileName: file.Filename,
				keyFileSize: file.Size,
				"max_size":  avatarCfg.MaxFileSizeBytes(),
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "File size exceeds the configured limit.")
		return
	}

	if _, err := ac.UserUsecase.GetByID(userID); err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError: "failed_to_fetch_user",
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	_, err = ac.AvatarUsecase.UploadAvatar(userID, file)
	if err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError:     "failed_to_upload_avatar",
				keyFileName: file.Filename,
				keyFileSize: file.Size,
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to upload avatar")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			keyFileSize: file.Size,
			keyFileName: file.Filename,
		}, true, nil)
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Avatar uploaded successfully", gin.H{"has_avatar": true})
}

func (ac *AdminAvatarHandler) DeleteUserAvatar(c *gin.Context) {
	userID := c.Param("userID")

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		log.Printf("Invalid user_id in context")
		return
	}
	adminName := c.GetString("username")
	action := "avatar_delete"
	resourceUUID := uuid.MustParse(userID)

	user, err := ac.UserUsecase.GetByID(userID)
	if err != nil {
		go func() {
			errorMsg := err.Error()
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError: "failed_to_fetch_user",
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	if !user.HasAvatar {
		errorMsg := "No avatar to delete"
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError: "no_avatar_to_delete",
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "No avatar to delete")
		return
	}

	err = ac.AvatarUsecase.DeleteAvatar(userID)
	if err != nil {
		go func() {
			errorMsg := err.Error()
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyError:       "failed_to_delete_avatar",
				"avatar_path": fileutil.FormatAvatarURL(userID),
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to delete avatar")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "avatar", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"avatar_path": fileutil.FormatAvatarURL(userID),
		}, true, nil)
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Avatar deleted successfully", gin.H{"has_avatar": false})
}
