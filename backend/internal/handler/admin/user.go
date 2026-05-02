// backend/internal/handler/admin/user.go

package admin

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/usecase"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminUserHandler struct {
	Usecase        *adminUsecase.AdminUserUsecase
	UserUsecase    *usecase.UserUsecase
	GameUsecase    *usecase.GameUsecase
	LoggingService service.LoggingServiceInterface
	wsService      service.WebSocketServiceInterface
}

func NewAdminUserHandler(
	uc *adminUsecase.AdminUserUsecase,
	userUsecase *usecase.UserUsecase,
	gameUsecase *usecase.GameUsecase,
	loggingService service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
) *AdminUserHandler {
	return &AdminUserHandler{
		Usecase:        uc,
		UserUsecase:    userUsecase,
		GameUsecase:    gameUsecase,
		LoggingService: loggingService,
		wsService:      wsService,
	}
}

// -----------------------------------------------------
// Admin User Handlers
//
// - GetAllUsers
// - GetUserCount
// - GetUserOnlineCount
// - GetWeeklyActiveUserCount
// - SearchUsers
// - GetCurrentUser
// - GetUserByID
// - CreateUser
// - UpdateUser
// - UpdateUserPassword
// - DeleteUser
// -----------------------------------------------------

func (ac *AdminUserHandler) GetAllUsers(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())

	var isOnlineFilter *bool
	if statusFilter := c.Query("status"); statusFilter != "" {
		switch statusFilter {
		case "online":
			t := true
			isOnlineFilter = &t
		case "offline":
			f := false
			isOnlineFilter = &f
		}
	}

	users, err := ac.Usecase.GetAllUsers(
		c.Query("role"),
		c.Query("rank"),
		c.Query("account_status"),
		isOnlineFilter,
		pag.Limit, pag.Offset,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch users")
		return
	}

	totalCount, err := ac.Usecase.GetUserCount(c.Query("role"), c.Query("rank"), c.Query("account_status"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user count")
		return
	}

	pag.WithTotal(int64(totalCount))
	httputil.SuccessList(c, users, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (ac *AdminUserHandler) GetUserCount(c *gin.Context) {
	count, err := ac.Usecase.GetUserCount(c.Query("role"), c.Query("level"), c.Query("account_status"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user count")
		return
	}
	httputil.Success(c, http.StatusOK, count)
}

func (ac *AdminUserHandler) GetUserOnlineCount(c *gin.Context) {
	count, err := ac.Usecase.GetUserOnlineCount(c.Query("role"), c.Query("level"), c.Query("account_status"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user count")
		return
	}
	httputil.Success(c, http.StatusOK, count)
}

func (ac *AdminUserHandler) GetWeeklyActiveUserCount(c *gin.Context) {
	count, err := ac.Usecase.GetWeeklyActiveUserCount(c.Query("role"), c.Query("level"), c.Query("account_status"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch weekly active user count")
		return
	}
	httputil.Success(c, http.StatusOK, count)
}

func (ac *AdminUserHandler) SearchUsers(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())
	query := c.Query("query")

	users, err := ac.Usecase.SearchUsers(query, pag.Limit, pag.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to search users")
		return
	}

	totalCount, err := ac.Usecase.SearchUserCount(query)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch search count")
		return
	}

	pag.WithTotal(int64(totalCount))
	httputil.SuccessList(c, users, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (ac *AdminUserHandler) GetCurrentUser(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}

	user, err := ac.Usecase.GetUserByID(userID.String())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}

	httputil.Success(c, http.StatusOK, user)
}

func (ac *AdminUserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	user, err := ac.Usecase.GetUserByID(userID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user")
		return
	}
	httputil.Success(c, http.StatusOK, user)
}

func (ac *AdminUserHandler) CreateUser(c *gin.Context) {
	var createUser model.AdminCreateUser
	if err := c.ShouldBindJSON(&createUser); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request payload")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_create"

	createdUser, createErr := ac.Usecase.CreateUser(createUser)
	if createErr != nil {
		errorMsg := createErr.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"email":          createUser.Email,
				"username":       createUser.Username,
				"role":           createUser.Role,
				"account_status": createUser.AccountStatus,
			}, false, &errorMsg)
		}()

		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to create user")
		return
	}

	resourceID := createdUser.ID
	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"email":          createUser.Email,
			"username":       createUser.Username,
			"role":           createUser.Role,
			"account_status": createUser.AccountStatus,
		}, true, nil)
	}()

	ac.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "user_created_admin",
		Data: map[string]interface{}{
			"username":   createUser.Username,
			keyAdminName: adminName,
		},
		EntityType: "user",
		EntityID:   createdUser.PublicID,
		ActionURL:  "/users/" + createdUser.PublicID,
	})

	httputil.Success(c, http.StatusCreated, createdUser)
}

func (ac *AdminUserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var userUpdate model.UserUpdate
	if err := c.ShouldBindJSON(&userUpdate); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request payload")
		return
	}

	if userUpdate.AccountStatus != "" {
		switch userUpdate.AccountStatus {
		case model.AccountStatusActive, model.AccountStatusSuspended, model.AccountStatusBanned, model.AccountStatusInactive:
		default:
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid account status")
			return
		}
	}

	if userUpdate.Language != "" {
		switch userUpdate.Language {
		case "en", "fr", "es":
		default:
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid language. Must be one of: en, fr, es")
			return
		}
	}

	if userUpdate.Bio != nil && len(*userUpdate.Bio) > 50 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Bio must not exceed 50 characters")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_update"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	changes := ac.buildUserChanges(currentUser, &userUpdate)

	updatedUser, err := ac.Usecase.UpdateUserByID(id, userUpdate)
	if err != nil {
		errorMsg := err.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keyUserID:       currentUser.ID,
				keyUserEmail:    currentUser.Email,
				keyUserUsername: currentUser.Username,
			}, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update user")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"changes":       changes,
			"summary":       httputil.GenerateChangeSummary(changes),
			keyUserID:       currentUser.ID,
			keyUserEmail:    currentUser.Email,
			keyUserUsername: currentUser.Username,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, updatedUser)
}

func (ac *AdminUserHandler) UpdateUserPassword(c *gin.Context) {
	id := c.Param("id")
	var updatePassword model.AdminUpdatePassword
	if err := c.ShouldBindJSON(&updatePassword); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request payload")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_password_update"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	details := map[string]interface{}{
		keyAction:        "password_update",
		keyUserID:       currentUser.ID,
		keyUserEmail:    currentUser.Email,
		keyUserUsername: currentUser.Username,
	}

	if updateErr := ac.Usecase.UpdateUserPassword(id, updatePassword.Password); updateErr != nil {
		errorMsg := updateErr.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update user password")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()
	httputil.SuccessWithMessage(c, http.StatusOK, "Password updated successfully", nil)
}

func (ac *AdminUserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	adminUUID := httputil.GetUserIDFromContext(c)
	if adminUUID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User not authenticated")
		return
	}
	adminName := c.GetString("username")
	action := "user_delete"
	resourceUUID := uuid.MustParse(id)

	currentUser, err := ac.Usecase.GetUserByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "User not found")
		return
	}

	details := map[string]interface{}{
		keyAction:        "delete",
		keyUserID:       currentUser.ID,
		keyUserEmail:    currentUser.Email,
		keyUserUsername: currentUser.Username,
	}

	if deleteErr := ac.Usecase.DeleteUserByID(id); deleteErr != nil {
		errorMsg := deleteErr.Error()
		go func() {
			httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMsg)
		}()
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to delete user")
		return
	}

	go func() {
		httputil.LogAdminAction(ac.LoggingService, adminUUID, adminName, action, "user", &resourceUUID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)
	}()
	httputil.SuccessWithMessage(c, http.StatusOK, "User deleted successfully", nil)
}

func (ac *AdminUserHandler) buildUserChanges(current *model.UserAdminView, update *model.UserUpdate) map[string]interface{} {
	changes := make(map[string]interface{})

	if update.Username != "" && update.Username != current.Username {
		changes["username"] = map[string]string{keyFrom: current.Username, keyTo: update.Username}
	}
	if update.Email != "" && update.Email != current.Email {
		changes["email"] = map[string]string{keyFrom: current.Email, keyTo: update.Email}
	}
	if update.Role != "" && update.Role != current.Role {
		changes["role"] = map[string]string{keyFrom: current.Role, keyTo: update.Role}
	}
	if update.AccountStatus != "" && update.AccountStatus != current.AccountStatus {
		changes["account_status"] = map[string]string{keyFrom: current.AccountStatus, keyTo: update.AccountStatus}
	}
	if update.IsProfilePublic != nil && *update.IsProfilePublic != current.IsProfilePublic {
		changes["is_profile_public"] = map[string]bool{keyFrom: current.IsProfilePublic, keyTo: *update.IsProfilePublic}
	}
	if update.ShowOnlineStatus != nil && *update.ShowOnlineStatus != current.ShowOnlineStatus {
		changes["show_online_status"] = map[string]bool{keyFrom: current.ShowOnlineStatus, keyTo: *update.ShowOnlineStatus}
	}
	if update.AllowFriendRequests != nil && *update.AllowFriendRequests != current.AllowFriendRequests {
		changes["allow_friend_requests"] = map[string]bool{keyFrom: current.AllowFriendRequests, keyTo: *update.AllowFriendRequests}
	}
	if update.AllowPartyInvites != nil && *update.AllowPartyInvites != current.AllowPartyInvites {
		changes["allow_party_invites"] = map[string]bool{keyFrom: current.AllowPartyInvites, keyTo: *update.AllowPartyInvites}
	}
	if update.Language != "" && update.Language != current.Language {
		changes["language"] = map[string]string{keyFrom: current.Language, keyTo: update.Language}
	}

	return changes
}
