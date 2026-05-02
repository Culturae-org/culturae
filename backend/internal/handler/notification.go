// backend/internal/handler/notification.go

package handler

import (
	"errors"
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	notifRepo repository.NotificationRepositoryInterface
}

func NewNotificationHandler(
	notifRepo repository.NotificationRepositoryInterface,
	) *NotificationHandler {
	return &NotificationHandler{
		notifRepo: notifRepo,
	}
}

// -----------------------------------------------------
// Notification Handlers
//
// - GetNotifications
// - MarkAsRead
// - MarkAllAsRead
// -----------------------------------------------------

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	limit, offset := httputil.ParsePagination(c, 20, 100)
	unreadOnly := c.Query("unread_only") == "true"

	notifs, err := h.notifRepo.GetByUserID(userID, limit, offset, unreadOnly)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch notifications")
		return
	}

	total, err := h.notifRepo.CountByUserID(userID, unreadOnly)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to count notifications")
		return
	}

	httputil.SuccessList(c, notifs, httputil.ParamsToPagination(total, limit, offset))
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	notifID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid notification ID")
		return
	}

	if err := h.notifRepo.MarkAsRead(notifID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Notification not found")
		} else {
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to mark notification as read")
		}
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{keyMessage: "marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	if err := h.notifRepo.MarkAllAsRead(userID); err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to mark notifications as read")
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{keyMessage: "all notifications marked as read"})
}
