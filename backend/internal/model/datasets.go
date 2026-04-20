// backend/internal/model/datasets.go

package model

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

var (
	ErrDatasetNotFound = errors.New("dataset not found")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func ValidateSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

type DatasetType string

const (
	DatasetTypeQuestions DatasetType = "questions"
	DatasetTypeGeography DatasetType = "geography"
)

type QuestionDataset struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`

	Version string `gorm:"not null" json:"version"`

	ManifestURL  string         `gorm:"type:text" json:"manifest_url"`
	ManifestData datatypes.JSON `gorm:"type:jsonb" json:"manifest_data,omitempty"`
	Source       string         `gorm:"default:'custom'" json:"source"`

	ImportJobID *uuid.UUID `gorm:"type:uuid" json:"import_job_id,omitempty"`
	ImportJob   *ImportJob `gorm:"foreignKey:ImportJobID" json:"import_job,omitempty"`
	ImportedAt  time.Time  `json:"imported_at"`

	QuestionCount int `gorm:"default:0" json:"question_count"`
	ThemeCount    int `gorm:"default:0" json:"theme_count"`

	LatestAvailableVersion string     `gorm:"type:varchar(50)" json:"latest_available_version,omitempty"`
	UpdateCheckedAt        *time.Time `json:"update_checked_at,omitempty"`

	IsActive  bool `gorm:"default:true" json:"is_active"`
	IsDefault bool `gorm:"default:false" json:"is_default"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Questions []Question `gorm:"foreignKey:DatasetID;constraint:OnDelete:SET NULL" json:"questions,omitempty"`
}

func (QuestionDataset) TableName() string {
	return "question_datasets"
}

type GeographyDataset struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`

	Version string `gorm:"not null" json:"version"`

	ManifestURL  string         `gorm:"type:text" json:"manifest_url"`
	ManifestData datatypes.JSON `gorm:"type:jsonb" json:"manifest_data,omitempty"`
	Source       string         `gorm:"default:'custom'" json:"source"`

	ImportJobID *uuid.UUID `gorm:"type:uuid" json:"import_job_id,omitempty"`
	ImportJob   *ImportJob `gorm:"foreignKey:ImportJobID" json:"import_job,omitempty"`
	ImportedAt  time.Time  `json:"imported_at"`

	CountryCount     int `gorm:"default:0" json:"country_count"`
	ContinentCount   int `gorm:"default:0" json:"continent_count"`
	RegionCount      int `gorm:"default:0" json:"region_count"`
	FlagCount        int `gorm:"default:0" json:"flag_count"`
	FlagPNG512Count  int `gorm:"default:0" json:"flag_png512_count"`
	FlagPNG1024Count int `gorm:"default:0" json:"flag_png1024_count"`

	LatestAvailableVersion string     `gorm:"type:varchar(50)" json:"latest_available_version,omitempty"`
	UpdateCheckedAt        *time.Time `json:"update_checked_at,omitempty"`

	IsActive  bool `gorm:"default:true" json:"is_active"`
	IsDefault bool `gorm:"default:false" json:"is_default"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Countries  []Country   `gorm:"foreignKey:DatasetID;constraint:OnDelete:CASCADE" json:"countries,omitempty"`
	Continents []Continent `gorm:"foreignKey:DatasetID;constraint:OnDelete:CASCADE" json:"continents,omitempty"`
	Regions    []Region    `gorm:"foreignKey:DatasetID;constraint:OnDelete:CASCADE" json:"regions,omitempty"`
}

func (GeographyDataset) TableName() string {
	return "geography_datasets"
}

type DatasetUpdateInfo struct {
	HasUpdate      bool             `json:"has_update"`
	CurrentVersion string           `json:"current_version"`
	LatestVersion  string           `json:"latest_version"`
	UpdatedAt      string           `json:"updated_at,omitempty"`
	Manifest       *DatasetManifest `json:"manifest,omitempty"`
}

type CreateDatasetRequest struct {
	Slug        string `json:"slug" binding:"required,min=3,max=100"`
	Name        string `json:"name" binding:"required,min=3,max=200"`
	Description string `json:"description"`
	Version     string `json:"version" binding:"required"`
	Source      string `json:"source" binding:"required,oneof=manifest custom imported"`
	IsDefault   bool   `json:"is_default"`
}

type ImportDatasetRequest struct {
	DatasetType  string `json:"dataset_type" binding:"required,oneof=questions geography"`
	ManifestURL  string `json:"manifest_url" binding:"required,url"`
	SetAsDefault bool   `json:"set_as_default"`
	Force        bool   `json:"force"`
}

type UpdateDatasetRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

type DeleteDatasetRequest struct {
	Force bool `json:"force"`
}

type DatasetInfo struct {
	ID        uuid.UUID
	Type      string
	Slug      string
	Name      string
	IsActive  bool
	IsDefault bool
}

type UnifiedDataset struct {
	ID                     uuid.UUID              `json:"id"`
	Type                   string                 `json:"type"`
	Slug                   string                 `json:"slug"`
	Name                   string                 `json:"name"`
	Description            string                 `json:"description"`
	Version                string                 `json:"version"`
	Source                 string                 `json:"source"`
	ManifestURL            string                 `json:"manifest_url"`
	IsActive               bool                   `json:"is_active"`
	IsDefault              bool                   `json:"is_default"`
	ImportedAt             interface{}            `json:"imported_at"`
	QuestionCount          int                    `json:"question_count,omitempty"`
	ThemeCount             int                    `json:"theme_count,omitempty"`
	CountryCount           int                    `json:"country_count,omitempty"`
	ContinentCount         int                    `json:"continent_count,omitempty"`
	RegionCount            int                    `json:"region_count,omitempty"`
	Stats                  map[string]interface{} `json:"stats,omitempty"`
	UpdateAvailable        bool                   `json:"update_available"`
	LatestAvailableVersion string                 `json:"latest_available_version,omitempty"`
}

type DatasetUpdateResult struct {
	DatasetID   uuid.UUID
	DatasetType string
	DatasetName string
	OldVersion  string
	NewVersion  string
}
