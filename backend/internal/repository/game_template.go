// backend/internal/repository/game_template.go

package repository

import (
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameTemplateRepositoryInterface interface {
	GetByID(id uuid.UUID) (*model.GameTemplate, error)
	GetBySlug(slug string) (*model.GameTemplate, error)
	List(params model.GameTemplateListParams) ([]model.GameTemplate, error)
	Count(params model.GameTemplateListParams) (int64, error)
	Create(t *model.GameTemplate) error
	Update(t *model.GameTemplate) error
	Delete(id uuid.UUID) error
	DeleteAll() error
}

type GameTemplateRepository struct {
	db *gorm.DB
}

func NewGameTemplateRepository(
	db *gorm.DB,
) *GameTemplateRepository {
	return &GameTemplateRepository{
		db: db,
	}
}

func (r *GameTemplateRepository) GetByID(id uuid.UUID) (*model.GameTemplate, error) {
	var t model.GameTemplate
	if err := r.db.Where("id = ?", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *GameTemplateRepository) GetBySlug(slug string) (*model.GameTemplate, error) {
	var t model.GameTemplate
	if err := r.db.Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *GameTemplateRepository) filterQuery(params model.GameTemplateListParams) *gorm.DB {
	q := r.db.Model(&model.GameTemplate{})
	if params.IsActive != nil {
		q = q.Where("is_active = ?", *params.IsActive)
	}
	if params.Mode != "" {
		q = q.Where("mode = ?", params.Mode)
	}
	if params.Category != "" {
		q = q.Where("category = ?", params.Category)
	}
	if params.Query != "" {
		like := "%" + params.Query + "%"
		q = q.Where("name ILIKE ? OR slug ILIKE ? OR description ILIKE ?", like, like, like)
	}
	return q
}

func (r *GameTemplateRepository) List(params model.GameTemplateListParams) ([]model.GameTemplate, error) {
	var templates []model.GameTemplate
	q := r.filterQuery(params).Order("mode ASC, name ASC")
	if params.Limit > 0 {
		q = q.Limit(params.Limit).Offset(params.Offset)
	}
	if err := q.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *GameTemplateRepository) Count(params model.GameTemplateListParams) (int64, error) {
	var count int64
	return count, r.filterQuery(params).Count(&count).Error
}

func (r *GameTemplateRepository) Create(t *model.GameTemplate) error {
	return r.db.Create(t).Error
}

func (r *GameTemplateRepository) Update(t *model.GameTemplate) error {
	return r.db.Save(t).Error
}

func (r *GameTemplateRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.GameTemplate{}, "id = ?", id).Error
}

func (r *GameTemplateRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&model.GameTemplate{}, "id = ?", id).Error
}

func (r *GameTemplateRepository) DeleteAll() error {
	return r.db.Where("1 = 1").Delete(&model.GameTemplate{}).Error
}
