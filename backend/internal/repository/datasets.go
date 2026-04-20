// backend/internal/repository/datasets.go

package repository

import (
	"errors"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestionDatasetRepositoryInterface interface {
	ListDatasets(activeOnly bool) ([]model.QuestionDataset, error)
	GetDatasetByID(id uuid.UUID) (*model.QuestionDataset, error)
	GetDatasetBySlug(slug string) (*model.QuestionDataset, error)
	GetDatasetByManifestURL(manifestURL string) (*model.QuestionDataset, error)
	CreateDataset(dataset *model.QuestionDataset) error
	UpdateDataset(id uuid.UUID, updates map[string]interface{}) error
	DeleteDataset(id uuid.UUID) error
	SetDefaultDataset(id uuid.UUID) error
	GetDefaultDataset() (*model.QuestionDataset, error)
	GetQuestionCountByDataset(datasetID uuid.UUID) (int64, error)
	GetThemeCountByDataset(datasetID uuid.UUID) (int64, error)
	CountOtherActiveDatasets(excludeID uuid.UUID) (int64, error)
	ExistsBySlug(slug string, excludeID *uuid.UUID) (bool, error)
	GetDB() *gorm.DB
}

type QuestionDatasetRepository struct {
	DB *gorm.DB
}

func NewQuestionDatasetRepository(db *gorm.DB) *QuestionDatasetRepository {
	return &QuestionDatasetRepository{
		DB: db,
	}
}

func (r *QuestionDatasetRepository) ListDatasets(activeOnly bool) ([]model.QuestionDataset, error) {
	var datasets []model.QuestionDataset
	query := r.DB.Preload("ImportJob")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("is_default DESC, created_at DESC").Find(&datasets).Error; err != nil {
		return nil, err
	}

	return datasets, nil
}

func (r *QuestionDatasetRepository) GetDatasetByID(id uuid.UUID) (*model.QuestionDataset, error) {
	var dataset model.QuestionDataset
	if err := r.DB.Preload("ImportJob").First(&dataset, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *QuestionDatasetRepository) GetDatasetBySlug(slug string) (*model.QuestionDataset, error) {
	var dataset model.QuestionDataset
	if err := r.DB.Preload("ImportJob").Where("slug = ?", slug).First(&dataset).Error; err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *QuestionDatasetRepository) GetDatasetByManifestURL(manifestURL string) (*model.QuestionDataset, error) {
	var dataset model.QuestionDataset
	if err := r.DB.Where("manifest_url = ?", manifestURL).First(&dataset).Error; err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *QuestionDatasetRepository) CreateDataset(dataset *model.QuestionDataset) error {
	return r.DB.Create(dataset).Error
}

func (r *QuestionDatasetRepository) UpdateDataset(id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&model.QuestionDataset{}).Where("id = ?", id).Updates(updates).Error
}

func (r *QuestionDatasetRepository) DeleteDataset(id uuid.UUID) error {
	return r.DB.Delete(&model.QuestionDataset{}, id).Error
}

func (r *QuestionDatasetRepository) SetDefaultDataset(id uuid.UUID) error {
	if err := r.DB.Model(&model.QuestionDataset{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return err
	}
	return r.DB.Model(&model.QuestionDataset{}).Where("id = ?", id).Update("is_default", true).Error
}

func (r *QuestionDatasetRepository) GetDefaultDataset() (*model.QuestionDataset, error) {
	var dataset model.QuestionDataset
	err := r.DB.Where("is_default = ? AND is_active = ?", true, true).First(&dataset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.DB.Where("is_active = ?", true).Order("created_at DESC").First(&dataset).Error; err != nil {
				return nil, errors.New("no active datasets found")
			}
			return &dataset, nil
		}
		return nil, err
	}
	return &dataset, nil
}

func (r *QuestionDatasetRepository) GetQuestionCountByDataset(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Question{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *QuestionDatasetRepository) GetThemeCountByDataset(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Table("questions").
		Select("COUNT(DISTINCT theme_id)").
		Where("dataset_id = ?", datasetID).
		Count(&count).Error
	return count, err
}

func (r *QuestionDatasetRepository) CountOtherActiveDatasets(excludeID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.QuestionDataset{}).
		Where("id != ? AND is_active = ?", excludeID, true).
		Count(&count).Error
	return count, err
}

func (r *QuestionDatasetRepository) ExistsBySlug(slug string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.DB.Model(&model.QuestionDataset{}).Where("slug = ?", slug)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *QuestionDatasetRepository) GetDB() *gorm.DB {
	return r.DB
}
