// backend/internal/model/imports.go

package model

import (
	"time"

	"github.com/google/uuid"
)

type ImportJob struct {
	ID                uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	ManifestURL       string     `gorm:"type:text;not null" json:"manifest_url"`
	Dataset           string     `gorm:"type:text" json:"dataset"`
	Version           string     `gorm:"type:text" json:"version"`
	StartedAt         time.Time  `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
	Success           bool       `json:"success"`
	Added             int        `json:"added"`
	Updated           int        `json:"updated"`
	Skipped           int        `json:"skipped"`
	Errors            int        `json:"errors"`
	Message           string     `gorm:"type:text" json:"message,omitempty"`
	FlagsStartedAt    *time.Time `json:"flags_started_at,omitempty"`
	FlagsFinishedAt   *time.Time `json:"flags_finished_at,omitempty"`
	FlagsSVGCount     int        `json:"flags_svg_count"`
	FlagsPNG512Count  int        `json:"flags_png512_count"`
	FlagsPNG1024Count int        `json:"flags_png1024_count"`
}

type ImportQuestionLog struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	JobID   uuid.UUID `gorm:"type:uuid;index;not null" json:"job_id"`
	Line    int       `json:"line"`
	Slug    string    `gorm:"type:citext;index" json:"slug"`
	Action  string    `gorm:"type:text" json:"action"`
	Message string    `gorm:"type:text" json:"message,omitempty"`
}
