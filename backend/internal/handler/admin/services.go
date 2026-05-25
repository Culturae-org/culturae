// backend/internal/handler/admin/services.go

package admin

import (
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminServicesHandler struct {
	Usecase *adminUsecase.AdminLogsUsecase
}

func NewAdminServicesHandler(
	usecase *adminUsecase.AdminLogsUsecase,
) *AdminServicesHandler {
	return &AdminServicesHandler{
		Usecase: usecase,
	}
}

// -----------------------------------------------------
// Services Handlers
//
// - GetSystemMetrics
// - GetServiceStatus
// -----------------------------------------------------

func (lc *AdminServicesHandler) GetSystemMetrics(c *gin.Context) {
	metrics, err := lc.Usecase.GetSystemMetrics()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch system metrics")
		return
	}

	httputil.Success(c, http.StatusOK, metrics)
}

func (lc *AdminServicesHandler) GetServiceStatus(c *gin.Context) {
	statuses, err := lc.Usecase.CheckServiceStatus()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to check service status")
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"services": statuses})
}
