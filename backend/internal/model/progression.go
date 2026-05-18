// backend/internal/model/progression.go

package model

import (
	"time"

	"github.com/google/uuid"
)

type UserProgressionSnapshot struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	GameID     *uuid.UUID `gorm:"type:uuid;index" json:"game_id"`
	GameMode   string     `gorm:"type:varchar(20)" json:"game_mode"`

	Elo            int    `json:"elo"`
	Experience     int64  `json:"experience"`
	Level          int    `json:"level"`
	Rank           string `gorm:"type:varchar(50)" json:"rank"`

	EloDelta        int   `json:"elo_delta"`
	ExperienceDelta int64 `json:"experience_delta"`
	LevelDelta      int   `json:"level_delta"`

	// Context
	Score    int  `json:"score"`
	IsWinner bool `json:"is_winner"`
	IsDrawn  bool `json:"is_drawn"`

	RecordedAt time.Time `gorm:"autoCreateTime" json:"recorded_at"`
}
