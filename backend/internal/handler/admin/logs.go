// backend/internal/handler/admin/logs.go

package admin

import (
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminLogsHandler struct {
	Usecase *adminUsecase.AdminLogsUsecase
}

func NewAdminLogsHandler(
	uc *adminUsecase.AdminLogsUsecase,
) *AdminLogsHandler {
	return &AdminLogsHandler{
		Usecase: uc,
	}
}

// -----------------------------------------------------
// Admin Logs Handlers
//
// - GetAdminActionLogs
// - GetUserActionLogs
// - GetUserActionLogsByID
// - GetAllUserActionLogs
// - GetConnectionLogs
// - GetAPIRequestStats
// - GetAdminActionStats
// - GetUserActionStats
// - GetAPIRequestTimestamps
// -----------------------------------------------------

func (lc *AdminLogsHandler) GetAdminActionLogs(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	action := httputil.QueryString(c, "action")
	resource := httputil.QueryString(c, "resource")

	logs, total, err := lc.Usecase.GetAdminActionLogs(
		pag.Limit, pag.Offset,
		httputil.QueryUUID(c, "admin_id"),
		action, resource,
		httputil.QueryBool(c, "is_success"),
		httputil.QueryUUID(c, "resource_id"),
		startDate, endDate,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch admin action logs")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, logs, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (lc *AdminLogsHandler) GetUserActionLogs(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	logs, total, err := lc.Usecase.GetUserActionLogs(
		pag.Limit, pag.Offset,
		httputil.QueryUUID(c, "user_id"),
		httputil.QueryString(c, "action"),
		httputil.QueryString(c, "category"),
		startDate, endDate,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user action logs")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, logs, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (lc *AdminLogsHandler) GetUserActionLogsByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeMissingField, "User ID is required")
		return
	}

	uuid, err := uuid.Parse(userID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user ID")
		return
	}

	pag := pagination.Parse(c, pagination.AdminConfig())
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	logs, total, err := lc.Usecase.GetUserActionLogs(
		pag.Limit, pag.Offset,
		&uuid,
		httputil.QueryString(c, "action"),
		httputil.QueryString(c, "category"),
		startDate, endDate,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user action logs")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, logs, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (lc *AdminLogsHandler) GetAllUserActionLogs(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	logs, total, err := lc.Usecase.GetUserActionLogs(
		pag.Limit, pag.Offset,
		nil,
		httputil.QueryString(c, "action"),
		httputil.QueryString(c, "category"),
		startDate, endDate,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch all user action logs")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, logs, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (lc *AdminLogsHandler) GetConnectionLogs(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	logs, total, err := lc.Usecase.GetConnectionLogs(
		pag.Limit, pag.Offset,
		httputil.QueryUUID(c, "user_id"),
		httputil.QueryBool(c, "is_success"),
		startDate, endDate,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch connection logs")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, logs, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (lc *AdminLogsHandler) GetAPIRequestStats(c *gin.Context) {
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	stats, err := lc.Usecase.GetAPIRequestStats(startDate, endDate)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch API request stats")
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (lc *AdminLogsHandler) GetAdminActionStats(c *gin.Context) {
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	stats, err := lc.Usecase.GetAdminActionStats(startDate, endDate)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch admin action stats")
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (lc *AdminLogsHandler) GetUserActionStats(c *gin.Context) {
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	stats, err := lc.Usecase.GetUserActionStats(startDate, endDate)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch user action stats")
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (lc *AdminLogsHandler) GetAPIRequestTimestamps(c *gin.Context) {
	startDate, endDate := httputil.QueryDateRange(c, "start_date", "end_date")

	timestamps, err := lc.Usecase.GetAPIRequestTimestamps(
		httputil.QueryString(c, "method"),
		httputil.QueryInt(c, "status_code"),
		startDate, endDate,
	)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch API request timestamps")
		return
	}

	httputil.Success(c, http.StatusOK, timestamps)
}

