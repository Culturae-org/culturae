// backend/internal/handler/profile.go

package handler

import (
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/identifier"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	Usecase        *usecase.UserUsecase
	GameUsecase    *usecase.GameUsecase
	SessionService service.SessionServiceInterface
	SessionConfig  *model.SessionConfig
	LoggingService service.LoggingServiceInterface
}

func NewProfileHandler(
	usecase *usecase.UserUsecase,
	gameUsecase *usecase.GameUsecase,
	sessionService service.SessionServiceInterface,
	sessionConfig *model.SessionConfig,
	loggingService service.LoggingServiceInterface,
) *ProfileHandler {
	return &ProfileHandler{
		Usecase:        usecase,
		GameUsecase:    gameUsecase,
		SessionService: sessionService,
		SessionConfig:  sessionConfig,
		LoggingService: loggingService,
	}
}

// -----------------------------------------------------
// Profile Handlers
//
// - UpdateProfile
// - GetCurrentUser
// - UserProfile
// - RegeneratePublicID
// - ChangePassword
// - DeleteAccount
// - GetUserStats
// -----------------------------------------------------

func (pc *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	currentUser, err := pc.Usecase.GetByID(userID.String())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	originalPassword := currentUser.Password

	var req struct {
		Username            string  `json:"username,omitempty"`
		Email               string  `json:"email,omitempty"`
		IsProfilePublic     *bool   `json:"is_profile_public,omitempty"`
		ShowOnlineStatus    *bool   `json:"show_online_status,omitempty"`
		AllowFriendRequests *bool   `json:"allow_friend_requests,omitempty"`
		AllowPartyInvites   *bool   `json:"allow_party_invites,omitempty"`
		Country             *string `json:"country,omitempty"`
		Language            string  `json:"language,omitempty"`
		Bio                 *string `json:"bio,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if req.Bio != nil && len(*req.Bio) > 50 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Bio must not exceed 50 characters")
		return
	}

	changes := make(map[string]map[string]interface{})

	if req.Username != "" && req.Username != currentUser.Username {
		changes["username"] = map[string]interface{}{"from": currentUser.Username, "to": req.Username}
		currentUser.Username = req.Username
	}
	if req.Email != "" && req.Email != currentUser.Email {
		changes["email"] = map[string]interface{}{"from": currentUser.Email, "to": req.Email}
		currentUser.Email = req.Email
	}
	if req.IsProfilePublic != nil && *req.IsProfilePublic != currentUser.IsProfilePublic {
		changes["is_profile_public"] = map[string]interface{}{"from": currentUser.IsProfilePublic, "to": *req.IsProfilePublic}
		currentUser.IsProfilePublic = *req.IsProfilePublic
	}
	if req.ShowOnlineStatus != nil && *req.ShowOnlineStatus != currentUser.ShowOnlineStatus {
		changes["show_online_status"] = map[string]interface{}{"from": currentUser.ShowOnlineStatus, "to": *req.ShowOnlineStatus}
		currentUser.ShowOnlineStatus = *req.ShowOnlineStatus
	}
	if req.AllowFriendRequests != nil && *req.AllowFriendRequests != currentUser.AllowFriendRequests {
		changes["allow_friend_requests"] = map[string]interface{}{"from": currentUser.AllowFriendRequests, "to": *req.AllowFriendRequests}
		currentUser.AllowFriendRequests = *req.AllowFriendRequests
	}
	if req.AllowPartyInvites != nil && *req.AllowPartyInvites != currentUser.AllowPartyInvites {
		changes["allow_party_invites"] = map[string]interface{}{"from": currentUser.AllowPartyInvites, "to": *req.AllowPartyInvites}
		currentUser.AllowPartyInvites = *req.AllowPartyInvites
	}
	if req.Language != "" && req.Language != currentUser.Language {
		changes["language"] = map[string]interface{}{"from": currentUser.Language, "to": req.Language}
		currentUser.Language = req.Language
	}
	if req.Bio != nil && (currentUser.Bio == nil || *req.Bio != *currentUser.Bio) {
		changes["bio"] = map[string]interface{}{"from": currentUser.Bio, "to": *req.Bio}
		currentUser.Bio = req.Bio
	}

	currentUser.Password = originalPassword

	if err := pc.Usecase.UpdateUser(currentUser); err != nil {
		go func() {
			if pc.LoggingService != nil {
				errorMsg := err.Error()
				_ = pc.LoggingService.LogUserAction(userID, "profile_update", httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
					"error":             "failed_to_update_user",
					"attempted_changes": changes,
				}, false, &errorMsg)
			}
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update user")
		return
	}

	go func() {
		if pc.LoggingService != nil && len(changes) > 0 {
			_ = pc.LoggingService.LogUserAction(userID, "profile_update", httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"changes": changes,
			}, true, nil)
		}
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Profile updated successfully", currentUser.ToUserView())
}

func (pc *ProfileHandler) GetCurrentUser(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	user, err := pc.Usecase.GetByID(userID.String())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	info := model.UserBasicInfo{
		PublicID:  user.PublicID,
		Username:  user.Username,
		HasAvatar: user.HasAvatar,
		Language:  user.Language,
		Role:      user.Role,
		Level:     user.Level,
		Rank:      user.Rank,
		EloRating: user.EloRating,
	}

	if pc.GameUsecase != nil {
		if activeGames, err := pc.GameUsecase.GetActiveGames(userID); err == nil && len(activeGames) > 0 {
			pid := activeGames[0].PublicID
			info.CurrentGamePublicID = &pid
		}
	}

	httputil.Success(c, http.StatusOK, info)
}

func (pc *ProfileHandler) UserProfile(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	user, err := pc.Usecase.GetByID(userID.String())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	httputil.Success(c, http.StatusOK, user.ToUserView())
}

func (pc *ProfileHandler) RegeneratePublicID(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid user ID")
		return
	}

	user, err := pc.Usecase.GetByID(userID.String())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	now := time.Now()
	if user.LastPublicIDRegeneration != nil && now.Sub(*user.LastPublicIDRegeneration) < 24*time.Hour {
		httputil.Error(c, http.StatusTooManyRequests, httputil.ErrCodeRateLimited, "You can only regenerate your Public ID once per day")
		return
	}

	if pc.GameUsecase != nil {
		if activeGames, err := pc.GameUsecase.GetActiveGames(userID); err == nil && len(activeGames) > 0 {
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeValidation, "Cannot regenerate public ID while in an active game")
			return
		}
	}

	oldPublicID := user.PublicID
	user.PublicID = identifier.GeneratePublicID()
	user.LastPublicIDRegeneration = &now

	if err := pc.Usecase.UpdateUser(user); err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update user")
		return
	}

	go func() {
		if pc.LoggingService != nil {
			_ = pc.LoggingService.LogUserAction(userID, "public_id_regeneration", httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"old_public_id": oldPublicID,
				"new_public_id": user.PublicID,
			}, true, nil)
		}
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Public ID regenerated successfully", map[string]string{
		"public_id": user.PublicID,
	})
}

func (pc *ProfileHandler) ChangePassword(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if err := pc.Usecase.ChangePassword(userID.String(), req.CurrentPassword, req.NewPassword); err != nil {
		if err.Error() == "current password is incorrect" {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodePasswordMismatch, err.Error())
			return
		}
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	if pc.SessionConfig != nil && pc.SessionConfig.RevokeOnPasswordChange && pc.SessionService != nil {
		currentTokenID := c.GetString("token_id")
		sessions, err := pc.SessionService.GetUserSessions(userID)
		if err == nil {
			for _, s := range sessions {
				if s.TokenID != currentTokenID && !s.IsRevoked {
					_ = pc.SessionService.RevokeSession(s.TokenID)
				}
			}
		}
	}

	go func() {
		if pc.LoggingService != nil {
			_ = pc.LoggingService.LogUserAction(userID, "change_password", httputil.GetRealIP(c), httputil.GetUserAgent(c), nil, true, nil)
		}
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Password changed successfully", nil)
}

func (pc *ProfileHandler) DeleteAccount(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	var req model.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if err := pc.Usecase.SoftDeleteAccount(userID.String(), req.Password); err != nil {
		if err.Error() == "password is incorrect" {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodePasswordMismatch, err.Error())
			return
		}
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	if pc.SessionService != nil {
		_ = pc.SessionService.RevokeAllUserSessions(userID)
	}

	go func() {
		if pc.LoggingService != nil {
			_ = pc.LoggingService.LogUserAction(userID, "delete_account", httputil.GetRealIP(c), httputil.GetUserAgent(c), nil, true, nil)
		}
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Account deleted successfully", nil)
}

func (pc *ProfileHandler) GetUserStats(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	period := c.Query("period")
	var stats interface{}
	var err error

	if period == "week" || period == "month" {
		stats, err = pc.GameUsecase.GetUserStatsByPeriod(userID, period)
	} else {
		stats, err = pc.GameUsecase.GetUserStats(userID)
	}

	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch stats")
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}
