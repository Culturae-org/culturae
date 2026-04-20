// backend/internal/repository/report.go

package repository

import (
	"errors"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepositoryInterface interface {
	CreateReport(report *model.QuestionReport) error
}

type ReportRepository struct {
	DB *gorm.DB
}

func NewReportRepository(
	db *gorm.DB,
) *ReportRepository {
	return &ReportRepository{
		DB: db,
	}
}

func (r *ReportRepository) CreateReport(report *model.QuestionReport) error {
	if err := r.DB.Create(report).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrAlreadyReported
		}
		return err
	}
	return nil
}

func (r *ReportRepository) GetReportByID(id uuid.UUID) (*model.QuestionReport, error) {
	var report model.QuestionReport
	err := r.DB.Preload("User").Preload("Question").Preload("GameQuestion").First(&report, "id = ?", id).Error
	return &report, err
}

func (r *ReportRepository) ListReports(limit, offset int, status string) ([]model.QuestionReport, int64, error) {
	var reports []model.QuestionReport
	var total int64

	query := r.DB.Model(&model.QuestionReport{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Preload("Question").Preload("GameQuestion").
		Order("created_at desc").
		Limit(limit).Offset(offset).
		Find(&reports).Error

	return reports, total, err
}

func (r *ReportRepository) UpdateReportStatus(id uuid.UUID, status model.ReportStatus, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": gorm.Expr("NOW()"),
	}
	if notes != "" {
		updates["resolution_notes"] = notes
	}
	return r.DB.Model(&model.QuestionReport{}).Where("id = ?", id).Updates(updates).Error
}
