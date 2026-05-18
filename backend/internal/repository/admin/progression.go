// backend/internal/repository/admin/progression.go

package admin

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProgressionRepositoryInterface interface {
	SaveSnapshot(snap *model.UserProgressionSnapshot) error
	UpdateSnapshotElo(userID uuid.UUID, gameID uuid.UUID, newElo, delta int) error
	GetUserSnapshots(userID uuid.UUID, limit, offset int, startDate, endDate *time.Time) ([]model.UserProgressionSnapshot, int64, error)
}

type ProgressionRepository struct {
	DB *gorm.DB
}

func NewProgressionRepository(db *gorm.DB) *ProgressionRepository {
	return &ProgressionRepository{DB: db}
}

func (r *ProgressionRepository) SaveSnapshot(snap *model.UserProgressionSnapshot) error {
	return r.DB.Create(snap).Error
}

func (r *ProgressionRepository) UpdateSnapshotElo(userID uuid.UUID, gameID uuid.UUID, newElo, delta int) error {
	return r.DB.Model(&model.UserProgressionSnapshot{}).
		Where("user_id = ? AND game_id = ?", userID, gameID).
		Updates(map[string]interface{}{
			"elo":       newElo,
			"elo_delta": delta,
		}).Error
}

func (r *ProgressionRepository) GetUserSnapshots(userID uuid.UUID, limit, offset int, startDate, endDate *time.Time) ([]model.UserProgressionSnapshot, int64, error) {
	var snaps []model.UserProgressionSnapshot
	var total int64

	query := r.DB.Model(&model.UserProgressionSnapshot{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("recorded_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("recorded_at <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("recorded_at ASC").Limit(limit).Offset(offset).Find(&snaps).Error; err != nil {
		return nil, 0, err
	}

	return snaps, total, nil
}
