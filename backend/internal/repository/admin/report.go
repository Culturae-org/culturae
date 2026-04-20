package admin

import (
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
)

type AdminReportRepositoryInterface interface {
	GetReportByID(id uuid.UUID) (*model.QuestionReport, error)
	ListReports(limit, offset int, status string) ([]model.QuestionReport, int64, error)
	UpdateReportStatus(id uuid.UUID, status model.ReportStatus, notes string) error
}
