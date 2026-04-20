// backend/internal/handler/admin/imports.go

package admin

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminImportsHandler struct {
	AdminImportsUsecase *adminUsecase.AdminImportsUsecase
	logger              *zap.Logger
}

func NewAdminImportsHandler(
	adminImportsUsecase *adminUsecase.AdminImportsUsecase,
	logger *zap.Logger,
) *AdminImportsHandler {
	return &AdminImportsHandler{
		AdminImportsUsecase: adminImportsUsecase,
		logger:              logger,
	}
}

// -----------------------------------------------------
// Admin Imports Handlers
//
// - ListImportJobs
// - GetImportJob
// - GetImportJobLogs
// - GetImportStats
// -----------------------------------------------------

func (ic *AdminImportsHandler) ListImportJobs(c *gin.Context) {
	pagination := pagination.Parse(c, pagination.Config{
		DefaultLimit: 20,
		MaxLimit:     100,
	})

	datasetType := c.Query("type")
	var datasetTypePtr *string
	if datasetType != "" {
		datasetTypePtr = &datasetType
	}

	jobs, total, err := ic.AdminImportsUsecase.ListImportJobs(pagination.Limit, pagination.Offset, datasetTypePtr)
	if err != nil {
		ic.logger.Error("Failed to list import jobs", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list import jobs")
		return
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, jobs, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (ic *AdminImportsHandler) GetImportJob(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid job ID")
		return
	}

	job, err := ic.AdminImportsUsecase.GetImportJob(id)
	if err != nil {
		ic.logger.Error("Failed to get import job", zap.String("id", idParam), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to get import job")
		return
	}

	if job == nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Import job not found")
		return
	}

	httputil.Success(c, http.StatusOK, job)
}

func (ic *AdminImportsHandler) GetImportJobLogs(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid job ID")
		return
	}

	pagination := pagination.Parse(c, pagination.Config{
		DefaultLimit: 50,
		MaxLimit:     500,
	})
	action := c.Query("action")

	var logs []interface{}
	var total int64

	if action != "" {
		logResults, count, err := ic.AdminImportsUsecase.GetImportJobLogsByAction(id, action, pagination.Limit, pagination.Offset)
		if err != nil {
			ic.logger.Error("Failed to get import logs by action",
				zap.String("id", idParam),
				zap.String("action", action),
				zap.Error(err),
			)
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to get import logs")
			return
		}
		for _, log := range logResults {
			logs = append(logs, log)
		}
		total = count
	} else {
		logResults, count, err := ic.AdminImportsUsecase.GetImportJobLogs(id, pagination.Limit, pagination.Offset)
		if err != nil {
			ic.logger.Error("Failed to get import logs", zap.String("id", idParam), zap.Error(err))
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to get import logs")
			return
		}
		for _, log := range logResults {
			logs = append(logs, log)
		}
		total = count
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, logs, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (ic *AdminImportsHandler) GetImportStats(c *gin.Context) {
	stats, err := ic.AdminImportsUsecase.GetImportStats()
	if err != nil {
		ic.logger.Error("Failed to get import stats", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to get import statistics")
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}
