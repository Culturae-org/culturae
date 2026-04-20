// backend/internal/handler/admin/reports.go

package admin

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/service"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminReportsHandler struct {
	reportUsecase  *adminUsecase.AdminReportUsecase
	LoggingService service.LoggingServiceInterface
	logger         *zap.Logger
}

func NewAdminReportsHandler(
	reportUsecase *adminUsecase.AdminReportUsecase,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AdminReportsHandler {
	return &AdminReportsHandler{
		reportUsecase:  reportUsecase,
		LoggingService: loggingService,
		logger:         logger,
	}
}

// -----------------------------------------------------
// Admin Reports Handlers
//
// - ListReports
// - UpdateReportStatus
// - GetReport
// -----------------------------------------------------

func (arc *AdminReportsHandler) ListReports(c *gin.Context) {
	pagination := pagination.Parse(c, pagination.Config{
		DefaultLimit: 50,
		MaxLimit:     100,
	})

	status := c.Query("status")

	reports, total, err := arc.reportUsecase.ListReports(pagination.Limit, pagination.Offset, status)
	if err != nil {
		arc.logger.Error("Failed to list reports", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list reports")
		return
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, reports, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (arc *AdminReportsHandler) UpdateReportStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid report ID")
		return
	}

	var req model.UpdateReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	adminUUID, adminName := httputil.GetUserIDFromContext(c), c.GetString("username")
	if err := arc.reportUsecase.UpdateReportStatus(id, req.Status, req.ResolutionNotes); err != nil {
		arc.logger.Error("Failed to update report status", zap.Error(err))
		errMsg := err.Error()
		httputil.LogAdminAction(arc.LoggingService, adminUUID, adminName, "report_update_status", "report", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"report_id": id, "status": req.Status}, false, &errMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update status")
		return
	}

	httputil.LogAdminAction(arc.LoggingService, adminUUID, adminName, "report_update_status", "report", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"report_id": id, "status": req.Status}, true, nil)
	httputil.SuccessWithMessage(c, http.StatusOK, "Report status updated", nil)
}

func (arc *AdminReportsHandler) GetReport(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid report ID")
		return
	}

	report, err := arc.reportUsecase.GetReport(id)
	if err != nil {
		arc.logger.Error("Failed to get report", zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Report not found")
		return
	}

	httputil.Success(c, http.StatusOK, report)
}
