// backend/internal/handler/admin/progression.go

package admin

import (
	"net/http"

	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminProgressionHandler struct {
	Repo adminRepo.ProgressionRepositoryInterface
}

func NewAdminProgressionHandler(repo adminRepo.ProgressionRepositoryInterface) *AdminProgressionHandler {
	return &AdminProgressionHandler{Repo: repo}
}

func (h *AdminProgressionHandler) GetUserProgression(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user ID")
		return
	}

	pag := pagination.Parse(c, pagination.AdminConfig())
	startDate, endDate := httputil.QueryDateRange(c, keyStartDate, keyEndDate)

	snaps, total, err := h.Repo.GetUserSnapshots(userID, pag.Limit, pag.Offset, startDate, endDate)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch progression history")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, snaps, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}
