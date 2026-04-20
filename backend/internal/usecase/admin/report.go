// backend/internal/usecase/admin/report.go

package admin

import (
	"github.com/Culturae-org/culturae/internal/model"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/google/uuid"
)

type AdminReportUsecase struct {
	reportRepo adminRepo.AdminReportRepositoryInterface
	wsService  service.WebSocketServiceInterface
}

func NewAdminReportUsecase(
	reportRepo adminRepo.AdminReportRepositoryInterface,
	wsService service.WebSocketServiceInterface,
) *AdminReportUsecase {
	return &AdminReportUsecase{
		reportRepo: reportRepo,
		wsService:  wsService,
	}
}

// -----------------------------------------------
// Admin Report Usecase Methods
//
// - ListReports
// - UpdateReportStatus
// - GetReport
//
// -----------------------------------------------

func (u *AdminReportUsecase) ListReports(limit, offset int, status string) ([]model.QuestionReport, int64, error) {
	return u.reportRepo.ListReports(limit, offset, status)
}

func (u *AdminReportUsecase) UpdateReportStatus(id uuid.UUID, status model.ReportStatus, notes string) error {
	if err := u.reportRepo.UpdateReportStatus(id, status, notes); err != nil {
		return err
	}

	u.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "report_resolved",
		Data: map[string]interface{}{
			"new_status": string(status),
		},
		EntityType: "report",
		EntityID:   id.String(),
		ActionURL:  "/reports/" + id.String(),
	})

	return nil
}

func (u *AdminReportUsecase) GetReport(id uuid.UUID) (*model.QuestionReport, error) {
	return u.reportRepo.GetReportByID(id)
}
