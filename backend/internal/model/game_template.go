// backend/internal/model/game_template.go

package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type I18nStrings map[string]string

func (i I18nStrings) Value() (driver.Value, error) {
	if i == nil {
		return "{}", nil
	}
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (i *I18nStrings) Scan(value interface{}) error {
	if value == nil {
		*i = make(I18nStrings)
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, i)
	case string:
		return json.Unmarshal([]byte(v), i)
	}
	return fmt.Errorf("I18nStrings: cannot scan type %T", value)
}

func (i I18nStrings) Get(lang string) string {
	if s, ok := i[lang]; ok && s != "" {
		return s
	}
	if s, ok := i["en"]; ok && s != "" {
		return s
	}
	for _, v := range i {
		if v != "" {
			return v
		}
	}
	return ""
}

type GameTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	NameI18n I18nStrings `gorm:"type:jsonb;default:'{}'" json:"name_i18n"`
	DescI18n I18nStrings `gorm:"type:jsonb;default:'{}'" json:"description_i18n"`
	Slug     string      `gorm:"type:citext;uniqueIndex;not null" json:"slug"`

	MinPlayers int `gorm:"not null;default:1" json:"min_players"`
	MaxPlayers int `gorm:"not null;default:2" json:"max_players"`

	QuestionCount      int        `gorm:"not null;default:10" json:"question_count"`
	QuestionType       string     `gorm:"type:varchar(50)" json:"question_type"`
	DatasetID          *uuid.UUID `gorm:"type:uuid;index" json:"dataset_id,omitempty"`
	Category           string     `gorm:"type:varchar(50)" json:"category"`
	FlagVariant        string     `gorm:"type:varchar(50)" json:"flag_variant"`
	Continent          string     `gorm:"type:varchar(100)" json:"continent"`
	IncludeTerritories bool       `gorm:"default:false" json:"include_territories"`
	Language           string     `gorm:"type:varchar(10);default:'en'" json:"language"`

	PointsPerCorrect int    `gorm:"not null;default:100" json:"points_per_correct"`
	TimeBonus        bool   `gorm:"default:true" json:"time_bonus"`
	ScoreMode        string `gorm:"type:varchar(50);default:'time_bonus'" json:"score_mode"`
	XPMultiplier *float64 `gorm:"type:decimal(5,2)" json:"xp_multiplier,omitempty"`

	Mode string `gorm:"type:varchar(50)" json:"mode"`

	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}

type CreateGameTemplateRequest struct {
	Name               string      `json:"name" binding:"required,max=100"`
	Description        string      `json:"description" binding:"max=500"`
	NameI18n           I18nStrings `json:"name_i18n,omitempty"`
	DescI18n           I18nStrings `json:"description_i18n,omitempty"`
	Slug               string      `json:"slug" binding:"required"`
	Mode               string      `json:"mode"`
	MinPlayers         int         `json:"min_players" binding:"required,min=1"`
	MaxPlayers         int         `json:"max_players" binding:"required,min=1"`
	QuestionCount      int         `json:"question_count" binding:"required,min=1"`
	QuestionType       string      `json:"question_type" binding:"omitempty,oneof=mcq mcq_2 mcq_4 true_false single_choice_2 mcq_2_mix text_input ''"`
	DatasetID          *uuid.UUID  `json:"dataset_id,omitempty"`
	Category           string      `json:"category"`
	FlagVariant        string      `json:"flag_variant"`
	Continent          string      `json:"continent"`
	IncludeTerritories bool        `json:"include_territories"`
	Language           string      `json:"language"`
	PointsPerCorrect   int         `json:"points_per_correct" binding:"min=0"`
	TimeBonus          bool        `json:"time_bonus"`
	ScoreMode          string      `json:"score_mode" binding:"oneof=classic time_bonus fastest_wins ''"`
	XPMultiplier       *float64    `json:"xp_multiplier,omitempty" binding:"omitempty,min=0"`
	IsActive           bool        `json:"is_active"`
}

type UpdateGameTemplateRequest struct {
	Name               *string      `json:"name,omitempty" binding:"omitempty,max=100"`
	Description        *string      `json:"description,omitempty" binding:"omitempty,max=500"`
	NameI18n           *I18nStrings `json:"name_i18n,omitempty"`
	DescI18n           *I18nStrings `json:"description_i18n,omitempty"`
	Slug               *string      `json:"slug,omitempty"`
	Mode               *string      `json:"mode,omitempty"`
	MinPlayers         *int         `json:"min_players,omitempty" binding:"omitempty,min=1"`
	MaxPlayers         *int         `json:"max_players,omitempty" binding:"omitempty,min=1"`
	QuestionCount      *int         `json:"question_count,omitempty" binding:"omitempty,min=1"`
	QuestionType       *string      `json:"question_type,omitempty" binding:"omitempty,oneof=mcq mcq_2 mcq_4 true_false single_choice_2 mcq_2_mix text_input ''"`
	DatasetID          *uuid.UUID   `json:"dataset_id,omitempty"`
	Category           *string      `json:"category,omitempty"`
	FlagVariant        *string      `json:"flag_variant,omitempty"`
	Continent          *string      `json:"continent,omitempty"`
	IncludeTerritories *bool        `json:"include_territories,omitempty"`
	Language           *string      `json:"language,omitempty"`
	PointsPerCorrect   *int         `json:"points_per_correct,omitempty" binding:"omitempty,min=0"`
	TimeBonus          *bool        `json:"time_bonus,omitempty"`
	ScoreMode          *string      `json:"score_mode,omitempty" binding:"omitempty,oneof=classic time_bonus fastest_wins"`
	XPMultiplier       *float64     `json:"xp_multiplier,omitempty" binding:"omitempty,min=0"`
	IsActive           *bool        `json:"is_active,omitempty"`
}

type GameTemplateListParams struct {
	IsActive *bool
	Mode     string
	Category string
	Query    string
	Limit    int
	Offset   int
}
