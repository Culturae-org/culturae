// backend/internal/handler/friends.go

package handler

import (
	"errors"
	"net/http"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FriendsHandler struct {
	friendsUsecase *usecase.FriendsUsecase
}

func NewFriendsHandler(
	friendsUsecase *usecase.FriendsUsecase,
) *FriendsHandler {
	return &FriendsHandler{
		friendsUsecase: friendsUsecase,
	}
}

// -----------------------------------------------------
// Friends Handlers
//
// - SendFriendRequest
// - AcceptFriendRequest
// - RejectFriendRequest
// - CancelFriendRequest
// - BlockFriendRequest
// - ListFriendRequests
// - ListFriends
// - RemoveFriend
// - GetBlockedUsers
// - BlockUser
// - UnblockUser
// -----------------------------------------------------

func (ctrl *FriendsHandler) SendFriendRequest(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	toUserPublicID := c.Param("toUserPublicID")
	if toUserPublicID == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid user public ID")
		return
	}

	request, err := ctrl.friendsUsecase.SendFriendRequest(c, userID, toUserPublicID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.Success(c, http.StatusCreated, request)
}

func (ctrl *FriendsHandler) AcceptFriendRequest(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	requestID, err := uuid.Parse(c.Param("requestID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid request ID")
		return
	}

	err = ctrl.friendsUsecase.AcceptFriendRequest(c, requestID, userID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Friend request accepted", nil)
}

func (ctrl *FriendsHandler) RejectFriendRequest(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	requestID, err := uuid.Parse(c.Param("requestID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid request ID")
		return
	}

	err = ctrl.friendsUsecase.RejectFriendRequest(c, requestID, userID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Friend request rejected", nil)
}

func (ctrl *FriendsHandler) CancelFriendRequest(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	requestID, err := uuid.Parse(c.Param("requestID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid request ID")
		return
	}

	err = ctrl.friendsUsecase.CancelFriendRequest(c, requestID, userID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Friend request cancelled", nil)
}

func (ctrl *FriendsHandler) BlockFriendRequest(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	requestID, err := uuid.Parse(c.Param("requestID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid request ID")
		return
	}

	err = ctrl.friendsUsecase.BlockFriendRequest(c, requestID, userID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "User blocked", nil)
}

func (ctrl *FriendsHandler) ListFriendRequests(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	status := c.DefaultQuery("status", "pending")
	requestType := c.DefaultQuery("type", "incoming")
	pag := pagination.Parse(c)

	requests, total, err := ctrl.friendsUsecase.ListFriendRequests(c, userID, status, requestType, pag.Limit, pag.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list requests")
		return
	}

	if requests == nil {
		requests = []model.FriendRequestWithUser{}
	}

	httputil.SuccessList(c, requests, pag.WithTotal(total))
}

func (ctrl *FriendsHandler) ListFriends(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	pag := pagination.Parse(c)

	friends, total, err := ctrl.friendsUsecase.ListFriends(c, userID, pag.Limit, pag.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list friends")
		return
	}

	if friends == nil {
		friends = []model.FriendUserResponse{}
	}

	httputil.SuccessList(c, friends, pag.WithTotal(total))
}

func (ctrl *FriendsHandler) RemoveFriend(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	friendUserPublicID := c.Param("friendUserPublicID")
	if friendUserPublicID == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid user public ID")
		return
	}

	err := ctrl.friendsUsecase.RemoveFriend(c, userID, friendUserPublicID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Friend removed", nil)
}

func (ctrl *FriendsHandler) GetBlockedUsers(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	users, err := ctrl.friendsUsecase.GetBlockedUsers(userID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list blocked users")
		return
	}

	httputil.Success(c, http.StatusOK, users)
}

func (ctrl *FriendsHandler) BlockUser(c *gin.Context) {
	userPublicID := c.Param("userPublicID")
	if userPublicID == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid user public ID")
		return
	}

	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	if err := ctrl.friendsUsecase.BlockUser(c, userID, userPublicID); err != nil {
		if errors.Is(err, model.ErrSelfBlock) {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeSelfAction, err.Error())
			return
		}
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "User blocked", nil)
}

func (ctrl *FriendsHandler) UnblockUser(c *gin.Context) {
	userPublicID := c.Param("userPublicID")
	if userPublicID == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid user public ID")
		return
	}

	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	if err := ctrl.friendsUsecase.UnblockUser(c, userID, userPublicID); err != nil { //nolint:forcetypeassert
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "User unblocked", nil)
}
