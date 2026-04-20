// backend/internal/handler/reports.go

package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReportsHandler struct {
	reportUsecase *usecase.ReportUsecase
	logger        *zap.Logger
}

func NewReportsHandler(
	reportUsecase *usecase.ReportUsecase,
	logger *zap.Logger,
) *ReportsHandler {
	return &ReportsHandler{
		reportUsecase: reportUsecase,
		logger:        logger,
	}
}

// -----------------------------------------------------
// Reports Handlers
//
// - CreateReport
// -----------------------------------------------------

func (rc *ReportsHandler) CreateReportFromGame(c *gin.Context) {
	gamePublicID := c.Param("gameID")
	questionNumberStr := c.Param("number")

	questionNumber, err := strconv.Atoi(questionNumberStr)
	if err != nil || questionNumber < 1 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid question number")
		return
	}

	userID := httputil.GetUserIDFromContext(c)

	var req model.CreateGameReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	report, err := rc.reportUsecase.CreateReportFromGame(userID, gamePublicID, questionNumber, req)
	if err != nil {
		if errors.Is(err, model.ErrAlreadyReported) {
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, "You have already reported this question")
			return
		}
		rc.logger.Error("Failed to create game question report", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to create report")
		return
	}

	httputil.Success(c, http.StatusCreated, report)
}
