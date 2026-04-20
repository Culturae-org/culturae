// backend/internal/model/game_event_log.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type GameEventLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	GameID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"game_id"`
	EventType string         `gorm:"type:varchar(50);not null" json:"event_type"`
	Data      datatypes.JSON `gorm:"type:jsonb" json:"data"`
	OccurredAt time.Time     `gorm:"not null;index" json:"occurred_at"`
}
