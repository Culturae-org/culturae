package admin

import (
	"encoding/json"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameEventLogRepositoryInterface interface {
	SaveEvent(gameID uuid.UUID, eventType string, data map[string]interface{}, occurredAt time.Time) error
	GetEventsByGameID(gameID uuid.UUID) ([]model.GameEventLog, error)
}

type GameEventLogRepository struct {
	DB *gorm.DB
}

func NewGameEventLogRepository(db *gorm.DB) *GameEventLogRepository {
	return &GameEventLogRepository{DB: db}
}

func (r *GameEventLogRepository) SaveEvent(gameID uuid.UUID, eventType string, data map[string]interface{}, occurredAt time.Time) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	log := model.GameEventLog{
		GameID:     gameID,
		EventType:  eventType,
		Data:       raw,
		OccurredAt: occurredAt,
	}

	return r.DB.Create(&log).Error
}

func (r *GameEventLogRepository) GetEventsByGameID(gameID uuid.UUID) ([]model.GameEventLog, error) {
	var events []model.GameEventLog
	err := r.DB.Where("game_id = ?", gameID).
		Order("occurred_at ASC").
		Find(&events).Error
	return events, err
}
