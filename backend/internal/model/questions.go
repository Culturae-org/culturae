// backend/internal/model/questions.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	QTypeTextInput = "text_input"
	QTypeMCQ       = "mcq"
	QTypeTrueFalse = "true-false"

	AnswerSlugTrue = "true"
)

type ThemeI18n struct {
	Label string `json:"label"`
}

type Theme struct {
	ID   uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Slug string         `gorm:"type:citext;unique;not null" json:"slug"`
	I18n datatypes.JSON `gorm:"type:jsonb" json:"i18n,omitempty"`
}

type Question struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Kind             string    `gorm:"not null" json:"kind"`
	Version          string    `gorm:"not null" json:"version"`
	Slug             string    `gorm:"type:citext;uniqueIndex:idx_slug_dataset;not null" json:"slug"`
	QType            string    `gorm:"not null" json:"qtype"`
	Difficulty       string    `gorm:"not null" json:"difficulty"`
	EstimatedSeconds int       `gorm:"not null" json:"estimated_seconds"`
	ShuffleAnswers   bool      `gorm:"default:true" json:"shuffle_answers"`

	DatasetID *uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_slug_dataset" json:"dataset_id,omitempty"`

	ThemeID uuid.UUID `gorm:"not null"`
	Theme   Theme     `gorm:"foreignKey:ThemeID" json:"theme"`

	Subthemes []Theme `gorm:"many2many:question_subthemes;constraint:OnDelete:CASCADE;" json:"subthemes"`
	Tags      []Theme `gorm:"many2many:question_tags;constraint:OnDelete:CASCADE;" json:"tags"`

	I18n datatypes.JSON `gorm:"type:jsonb;not null" json:"i18n"`

	Answers datatypes.JSON `gorm:"type:jsonb;not null" json:"answers"`

	Data datatypes.JSON `gorm:"type:jsonb" json:"data,omitempty"`

	Sources datatypes.JSON `gorm:"type:jsonb" json:"sources,omitempty"`

	CountryID   *uuid.UUID `gorm:"type:uuid;index" json:"country_id,omitempty"`
	CountrySlug string     `gorm:"type:varchar(100);index" json:"country_slug,omitempty"`
	Variant     string     `gorm:"type:varchar(50);index" json:"variant,omitempty"`

	TimesPlayed  int     `gorm:"default:0" json:"times_played"`
	TimesCorrect int     `gorm:"default:0" json:"times_correct"`
	SuccessRate  float64 `gorm:"default:0" json:"success_rate"`
	AvgTimeMs    float64 `gorm:"default:0" json:"avg_time_ms"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QuestionI18n struct {
	Title       string `json:"title"`
	Stem        string `json:"stem,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type Answer struct {
	Slug      string                `json:"slug"`
	IsCorrect bool                  `json:"is_correct"`
	I18n      map[string]AnswerI18n `json:"i18n"`
}

type AnswerI18n struct {
	Label string `json:"label"`
}

type QuestionCreateRequest struct {
	Kind             string                  `json:"kind" binding:"required"`
	Version          string                  `json:"version" binding:"required"`
	Slug             string                  `json:"slug" binding:"required"`
	Theme            Theme                   `json:"theme" binding:"required"`
	Subthemes        []Theme                 `json:"subthemes"`
	Tags             []Theme                 `json:"tags"`
	QType            string                  `json:"qtype" binding:"required"`
	Difficulty       string                  `json:"difficulty" binding:"required"`
	EstimatedSeconds int                     `json:"estimated_seconds" binding:"required,min=1"`
	ShuffleAnswers   bool                    `json:"shuffle_answers"`
	I18n             map[string]QuestionI18n `json:"i18n" binding:"required"`
	Answers          []Answer                `json:"answers" binding:"required,min=2,max=4"`
	Sources          []string                `json:"sources"`
	DatasetID        *uuid.UUID              `json:"dataset_id,omitempty"`
}

type QuestionUpdateRequest struct {
	Kind             *string                  `json:"kind,omitempty"`
	Version          *string                  `json:"version,omitempty"`
	Slug             *string                  `json:"slug,omitempty"`
	Theme            *Theme                   `json:"theme,omitempty"`
	Subthemes        *[]Theme                 `json:"subthemes,omitempty"`
	Tags             *[]Theme                 `json:"tags,omitempty"`
	QType            *string                  `json:"qtype,omitempty"`
	Difficulty       *string                  `json:"difficulty,omitempty"`
	EstimatedSeconds *int                     `json:"estimated_seconds,omitempty"`
	ShuffleAnswers   *bool                    `json:"shuffle_answers,omitempty"`
	I18n             *map[string]QuestionI18n `json:"i18n,omitempty"`
	Answers          *[]Answer                `json:"answers,omitempty"`
	Sources          *[]string                `json:"sources,omitempty"`
}

type QuestionFilters struct {
	DatasetID   *uuid.UUID `json:"dataset_id,omitempty"`
	Theme       *string    `json:"theme,omitempty"`
	Subtheme    *string    `json:"subtheme,omitempty"`
	Difficulty  *string    `json:"difficulty,omitempty"`
	QType       *string    `json:"qtype,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	SearchQuery *string    `json:"search_query,omitempty"`
}
