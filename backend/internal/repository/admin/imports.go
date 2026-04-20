// backend/internal/repository/admin/imports.go

package admin

import (
	"errors"
	"fmt"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImportJobRepositoryInterface interface {
	ListImportJobs(limit int, offset int, datasetType *string) ([]model.ImportJob, int64, error)
	GetImportStats() (map[string]interface{}, error)
	GetImportJob(jobID uuid.UUID) (*model.ImportJob, error)
	GetImportJobLogs(jobID uuid.UUID, limit, offset int) ([]model.ImportQuestionLog, int64, error)
	GetImportJobLogsByAction(jobID uuid.UUID, action string, limit, offset int) ([]model.ImportQuestionLog, int64, error)
	SaveImportLog(log *model.ImportQuestionLog) error
	CreateImportJob(job *model.ImportJob) error
	UpdateImportJobStatus(jobID uuid.UUID, updates map[string]interface{}) error
}

type ImportJobRepository struct {
	DB *gorm.DB
}

func NewImportJobRepository(
	db *gorm.DB,
) *ImportJobRepository {
	return &ImportJobRepository{
		DB: db,
	}
}

func (r *ImportJobRepository) ListImportJobs(limit int, offset int, datasetType *string) ([]model.ImportJob, int64, error) {
	var jobs []model.ImportJob
	query := r.DB.Model(&model.ImportJob{})

	if datasetType != nil && *datasetType != "" {
		query = query.Where("dataset LIKE ?", fmt.Sprintf("%%%s%%", *datasetType))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("started_at DESC").Limit(limit).Offset(offset).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

func (r *ImportJobRepository) GetImportStats() (map[string]interface{}, error) {
	var totalJobs int64
	var successfulJobs int64
	var failedJobs int64

	if err := r.DB.Model(&model.ImportJob{}).Count(&totalJobs).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&model.ImportJob{}).Where("success = ?", true).Count(&successfulJobs).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&model.ImportJob{}).Where("success = ?", false).Count(&failedJobs).Error; err != nil {
		return nil, err
	}

	var lastJob model.ImportJob
	if err := r.DB.Model(&model.ImportJob{}).Order("started_at DESC").First(&lastJob).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return map[string]interface{}{
		"total_jobs":      totalJobs,
		"successful_jobs": successfulJobs,
		"failed_jobs":     failedJobs,
		"last_import":     lastJob,
	}, nil
}

func (r *ImportJobRepository) GetImportJob(jobID uuid.UUID) (*model.ImportJob, error) {
	var job model.ImportJob
	if err := r.DB.First(&job, "id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ImportJobRepository) GetImportJobLogs(jobID uuid.UUID, limit, offset int) ([]model.ImportQuestionLog, int64, error) {
	var logs []model.ImportQuestionLog
	var total int64

	query := r.DB.Model(&model.ImportQuestionLog{}).Where("job_id = ?", jobID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("line ASC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *ImportJobRepository) GetImportJobLogsByAction(jobID uuid.UUID, action string, limit, offset int) ([]model.ImportQuestionLog, int64, error) {
	var logs []model.ImportQuestionLog
	var total int64

	query := r.DB.Model(&model.ImportQuestionLog{}).Where("job_id = ? AND action = ?", jobID, action)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("line ASC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *ImportJobRepository) SaveImportLog(log *model.ImportQuestionLog) error {
	return r.DB.Create(log).Error
}

func (r *ImportJobRepository) CreateImportJob(job *model.ImportJob) error {
	return r.DB.Create(job).Error
}

func (r *ImportJobRepository) UpdateImportJobStatus(jobID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&model.ImportJob{}).Where("id = ?", jobID).Updates(updates).Error
}
