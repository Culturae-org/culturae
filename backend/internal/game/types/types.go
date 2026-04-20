// backend/internal/game/types/types.go

package types

import (
	"time"

	"github.com/google/uuid"
)

type Answer struct {
	QuestionID uuid.UUID          `json:"question_id"`
	AnswerType string             `json:"answer_type"`
	AnswerData interface{}        `json:"answer_data"`
	IsCorrect  bool               `json:"is_correct"`
	TimeSpent  time.Duration      `json:"time_spent"`
	Points     int                `json:"points"`
	AnsweredAt time.Time          `json:"answered_at"`

	ServerTimeSpent time.Duration `json:"server_time_spent"`
	ReceivedAt      time.Time     `json:"received_at"`

	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (a *Answer) GetAnswerSlug() string {
	if a.AnswerType == "mcq" {
		if slug, ok := a.AnswerData.(string); ok {
			return slug
		}
	}
	return ""
}
