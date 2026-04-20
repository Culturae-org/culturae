// backend/internal/handler/admin/friends.go

package admin

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminFriendsHandler struct {
	Usecase *adminUsecase.AdminFriendsUsecase
}

func NewAdminFriendsHandler(
	usecase *adminUsecase.AdminFriendsUsecase,
) *AdminFriendsHandler {
	return &AdminFriendsHandler{
		Usecase: usecase,
	}
}

// -----------------------------------------------------
// Admin Friends Handlers
//
// - ListFriendRequestsForUser
// - ListFriendsForUser
// -----------------------------------------------------

func (afc *AdminFriendsHandler) ListFriendRequestsForUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user ID")
		return
	}

	pag := pagination.Parse(c, pagination.AdminConfig())
	statusFilter := httputil.QueryString(c, "status")
	direction := httputil.QueryString(c, "direction")

	requests, total, err := afc.Usecase.GetFriendRequestsForUser(userID, pag.Limit, pag.Offset, statusFilter, direction)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch friend requests")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, requests, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (afc *AdminFriendsHandler) ListFriendsForUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user ID")
		return
	}

	pag := pagination.Parse(c, pagination.AdminConfig())

	friendships, total, err := afc.Usecase.GetFriendsForUser(userID, pag.Limit, pag.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch friends")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, friendships, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}
