// backend/internal/handler/admin/user_actions.go

package admin

import (
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// -----------------------------------------------------
// Admin User Handlers
//
// - DeactivateUser
// - UpdateUserStatus
// - BanUser
// - UnbanUser
// - RegeneratePublicID
// - GetUserLevelStats
// - GetUserRoleStats
// - GetUserCreationDates
// - GetUserConnectionLogs
// - GetUserActiveSessions
// -----------------------------------------------------

func (ac *AdminUserHandler) DeactivateUser(c *gin.Context) {
	id := c.Param("id")

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_deactivate"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	details := map[string]interface{}{
		keyAction:          "deactivate",
		"new_status":      "inactive",
		keyPreviousStatus: currentUser.AccountStatus,
		keyUserID:         currentUser.ID,
		keyUserEmail:      currentUser.Email,
		keyUserUsername:   currentUser.Username,
	}

	if err := ac.Usecase.DeactivateUserByID(id); err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to deactivate user")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()

	ac.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "user_deactivated",
		Data: map[string]interface{}{
			keyAction:          "deactivate",
			keyAdminName:      adminName,
			keyTargetUsername: currentUser.Username,
		},
		EntityType: entityUser,
		EntityID:   currentUser.PublicID,
		ActionURL:  "/users/" + currentUser.PublicID,
	})

	httputil.SuccessWithMessage(c, http.StatusOK, "User deactivated successfully", nil)
}

func (ac *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var statusUpdate struct {
		AccountStatus string `json:"account_status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request payload")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_status_update"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	details := map[string]interface{}{
		"new_status":      statusUpdate.AccountStatus,
		keyPreviousStatus: currentUser.AccountStatus,
		keyUserID:         currentUser.ID,
		keyUserEmail:      currentUser.Email,
		keyUserUsername:   currentUser.Username,
	}

	if updateStatusErr := ac.Usecase.UpdateUserStatusByID(id, statusUpdate.AccountStatus); updateStatusErr != nil {
		errorMsg := updateStatusErr.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update user status")
		return
	}

	action = "user_status_" + statusUpdate.AccountStatus
	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()

	ac.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "user_status_updated",
		Data: map[string]interface{}{
			keyAction:          "status_" + statusUpdate.AccountStatus,
			keyAdminName:      adminName,
			keyTargetUsername: currentUser.Username,
		},
		EntityType: entityUser,
		EntityID:   currentUser.PublicID,
		ActionURL:  "/users/" + currentUser.PublicID,
	})

	updatedUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch updated user")
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "User status updated successfully", updatedUser)
}

func (ac *AdminUserHandler) BanUser(c *gin.Context) {
	id := c.Param("id")
	var banReq model.AdminBanUser
	if err := c.ShouldBindJSON(&banReq); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request payload")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_ban"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	details := map[string]interface{}{
		"duration":        banReq.Duration,
		"reason":          banReq.Reason,
		keyPreviousStatus: currentUser.AccountStatus,
		keyUserID:         currentUser.ID,
		keyUserEmail:      currentUser.Email,
		keyUserUsername:   currentUser.Username,
	}

	updatedUser, err := ac.Usecase.BanUser(id, banReq.Duration, banReq.Reason)
	if err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()

	ac.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "user_banned",
		Data: map[string]interface{}{
			keyAction:          "ban",
			keyAdminName:      adminName,
			keyTargetUsername: currentUser.Username,
		},
		EntityType: entityUser,
		EntityID:   currentUser.PublicID,
		ActionURL:  "/users/" + currentUser.PublicID,
	})

	httputil.SuccessWithMessage(c, http.StatusOK, "User banned successfully", updatedUser)
}

func (ac *AdminUserHandler) UnbanUser(c *gin.Context) {
	id := c.Param("id")

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_unban"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	details := map[string]interface{}{
		keyPreviousStatus: currentUser.AccountStatus,
		keyUserID:         currentUser.ID,
		keyUserEmail:      currentUser.Email,
		keyUserUsername:   currentUser.Username,
	}

	updatedUser, err := ac.Usecase.UnbanUser(id)
	if err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to unban user")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()

	ac.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "user_unbanned",
		Data: map[string]interface{}{
			keyAction:          "unban",
			keyAdminName:      adminName,
			keyTargetUsername: currentUser.Username,
		},
		EntityType: entityUser,
		EntityID:   currentUser.PublicID,
		ActionURL:  "/users/" + currentUser.PublicID,
	})

	httputil.SuccessWithMessage(c, http.StatusOK, "User unbanned successfully", updatedUser)
}

func (ac *AdminUserHandler) RegeneratePublicID(c *gin.Context) {
	id := c.Param("id")

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_regenerate_public_id"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	if activeGames, err := ac.GameUsecase.GetActiveGames(currentUser.ID); err == nil && len(activeGames) > 0 {
		httputil.Error(c, http.StatusConflict, httputil.ErrCodeValidation, "Cannot regenerate public ID while user is in an active game")
		return
	}

	oldPublicID := currentUser.PublicID
	details := map[string]interface{}{
		keyUserID:       currentUser.ID,
		keyUserEmail:    currentUser.Email,
		keyUserUsername: currentUser.Username,
		"old_public_id": oldPublicID,
	}

	if err := ac.Usecase.RegeneratePublicID(id); err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to regenerate public ID")
		return
	}

	updatedUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch updated user")
		return
	}

	details["new_public_id"] = updatedUser.PublicID
	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, entityUser, &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Public ID regenerated successfully", updatedUser)
}

func (ac *AdminUserHandler) GetUserLevelStats(c *gin.Context) {
	stats, err := ac.Usecase.GetUserLevelStats()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user level stats")
		return
	}
	httputil.Success(c, http.StatusOK, stats)
}

func (ac *AdminUserHandler) GetUserRoleStats(c *gin.Context) {
	stats, err := ac.Usecase.GetUserRoleStats()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user role stats")
		return
	}
	httputil.Success(c, http.StatusOK, stats)
}

func (ac *AdminUserHandler) GetUserCreationDates(c *gin.Context) {
	var startDate, endDate *time.Time

	if start := c.Query(keyStartDate); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = &t
		}
	}
	if end := c.Query(keyEndDate); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = &t
		}
	}

	dates, err := ac.Usecase.GetUserCreationDates(startDate, endDate)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user creation dates")
		return
	}
	httputil.Success(c, http.StatusOK, dates)
}

func (ac *AdminUserHandler) GetUserConnectionLogs(c *gin.Context) {
	id := c.Param("id")
	logs, err := ac.Usecase.GetUserConnectionLogs(id, httputil.QueryBool(c, "success"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user connection logs")
		return
	}
	httputil.Success(c, http.StatusOK, logs)
}

func (ac *AdminUserHandler) GetUserActiveSessions(c *gin.Context) {
	id := c.Param("id")
	sessions, err := ac.Usecase.GetUserActiveSessions(id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user active sessions")
		return
	}
	httputil.Success(c, http.StatusOK, sessions)
}
