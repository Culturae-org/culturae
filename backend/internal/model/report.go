// backend/internal/model/report.go

package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAlreadyReported = errors.New("question already reported by this user")

type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusInProgress ReportStatus = "in_progress"
	ReportStatusResolved   ReportStatus = "resolved"
)

type QuestionReport struct {
	ID              uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID          uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:idx_user_question_report;uniqueIndex:idx_user_gq_report" json:"user_id"`
	QuestionID      *uuid.UUID   `gorm:"type:uuid;index;uniqueIndex:idx_user_question_report" json:"question_id,omitempty"`
	GameQuestionID  *uuid.UUID   `gorm:"type:uuid;index;uniqueIndex:idx_user_gq_report" json:"game_question_id,omitempty"`
	Reason          string       `gorm:"not null" json:"reason"`
	Message         string       `gorm:"type:text" json:"message"`
	ResolutionNotes string       `gorm:"type:text" json:"resolution_notes"`
	Status          ReportStatus `gorm:"type:citext;default:'pending';index" json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Question     *Question     `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	GameQuestion *GameQuestion `gorm:"foreignKey:GameQuestionID" json:"game_question,omitempty"`
}

type CreateReportRequest struct {
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	Reason     string    `json:"reason" binding:"required,oneof=wrong_answer typo offensive other"`
	Message    string    `json:"message" binding:"max=500"`
}

type CreateGameReportRequest struct {
	Reason  string `json:"reason" binding:"required,oneof=wrong_answer typo offensive other"`
	Message string `json:"message" binding:"max=500"`
}

type UpdateReportStatusRequest struct {
	Status          ReportStatus `json:"status" binding:"required,oneof=pending in_progress resolved"`
	ResolutionNotes string       `json:"resolution_notes"`
}
