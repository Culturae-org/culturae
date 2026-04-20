// backend/internal/model/geography.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Country struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	DatasetID uuid.UUID `gorm:"type:uuid;not null;index" json:"dataset_id"`

	Slug       string `gorm:"type:varchar(10);not null;index" json:"slug"`
	ISOAlpha2  string `gorm:"type:varchar(2);index" json:"iso_alpha2"`
	ISOAlpha3  string `gorm:"type:varchar(3);index" json:"iso_alpha3"`
	ISONumeric string `gorm:"type:varchar(3)" json:"iso_numeric"`

	Name         datatypes.JSON `gorm:"type:jsonb" json:"name"`
	OfficialName datatypes.JSON `gorm:"type:jsonb" json:"official_name"`
	Capital      datatypes.JSON `gorm:"type:jsonb" json:"capital"`

	Continent string  `gorm:"type:varchar(50);index" json:"continent"`
	Region    string  `gorm:"type:varchar(100);index" json:"region"`
	Latitude  float64 `gorm:"type:decimal(10,6)" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(10,6)" json:"longitude"`

	Flag       string `gorm:"type:varchar(10)" json:"flag"`
	Population int64  `gorm:"default:0" json:"population"`
	AreaKm2    int64  `gorm:"default:0" json:"area_km2"`

	Currency  datatypes.JSON `gorm:"type:jsonb" json:"currency"`
	Languages datatypes.JSON `gorm:"type:jsonb" json:"languages"`
	Neighbors datatypes.JSON `gorm:"type:jsonb" json:"neighbors"`

	TLD         string `gorm:"type:varchar(10)" json:"tld"`
	PhoneCode   string `gorm:"type:varchar(20)" json:"phone_code"`
	DrivingSide string `gorm:"type:varchar(10)" json:"driving_side"`
	Independent bool   `gorm:"default:false" json:"independent"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Country) TableName() string {
	return "countries"
}

type Continent struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	DatasetID uuid.UUID `gorm:"type:uuid;not null;index" json:"dataset_id"`

	Slug string `gorm:"type:varchar(50);not null;index" json:"slug"`

	Name datatypes.JSON `gorm:"type:jsonb" json:"name"`

	Countries  datatypes.JSON `gorm:"type:jsonb" json:"countries"`
	AreaKm2    int64          `gorm:"default:0" json:"area_km2"`
	Population int64          `gorm:"default:0" json:"population"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Continent) TableName() string {
	return "continents"
}

type Region struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	DatasetID uuid.UUID `gorm:"type:uuid;not null;index" json:"dataset_id"`

	Slug string `gorm:"type:varchar(100);not null;index" json:"slug"`

	Name datatypes.JSON `gorm:"type:jsonb" json:"name"`

	Continent string         `gorm:"type:varchar(50);index" json:"continent"`
	Countries datatypes.JSON `gorm:"type:jsonb" json:"countries"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Region) TableName() string {
	return "regions"
}

type CountryRaw struct {
	Slug       string            `json:"slug"`
	ISOAlpha2  string            `json:"iso_alpha2"`
	ISOAlpha3  string            `json:"iso_alpha3"`
	ISONumeric string            `json:"iso_numeric"`
	Name       map[string]string `json:"name"`
	Official   map[string]string `json:"official_name"`
	Capital    map[string]string `json:"capital"`
	Continent  string            `json:"continent"`
	Region     string            `json:"region"`
	Coords     struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinates"`
	Flag        string   `json:"flag"`
	Population  int64    `json:"population"`
	AreaKm2     float64  `json:"area_km2"`
	Currency    Currency `json:"currency"`
	Languages   []string `json:"languages"`
	Neighbors   []string `json:"neighbors"`
	TLD         string   `json:"tld"`
	PhoneCode   string   `json:"phone_code"`
	DrivingSide string   `json:"driving_side"`
	Independent bool     `json:"independent"`
}

type Currency struct {
	Code   string            `json:"code"`
	Name   map[string]string `json:"name"`
	Symbol string            `json:"symbol"`
}

type ContinentRaw struct {
	Slug       string            `json:"slug"`
	Name       map[string]string `json:"name"`
	Countries  []string          `json:"countries"`
	AreaKm2    int64             `json:"area_km2"`
	Population int64             `json:"population"`
}

type RegionRaw struct {
	Slug      string            `json:"slug"`
	Name      map[string]string `json:"name"`
	Continent string            `json:"continent"`
	Countries []string          `json:"countries"`
}

type GeographyImportResult struct {
	Success           bool     `json:"success"`
	Message           string   `json:"message"`
	CountriesAdded    int      `json:"countries_added"`
	CountriesUpdated  int      `json:"countries_updated"`
	CountriesSkipped  int      `json:"countries_skipped"`
	ContinentsAdded   int      `json:"continents_added"`
	ContinentsUpdated int      `json:"continents_updated"`
	RegionsAdded      int      `json:"regions_added"`
	RegionsUpdated    int      `json:"regions_updated"`
	FlagsPNG512Added  int      `json:"flags_png512_added"`
	FlagsPNG1024Added int      `json:"flags_png1024_added"`
	Errors            []string `json:"errors,omitempty"`
}

type CountryFilters struct {
	Continent     string `form:"continent"`
	Region        string `form:"region"`
	PopulationMin *int64 `form:"population_min"`
	PopulationMax *int64 `form:"population_max"`
	AreaMin       *int64 `form:"area_min"`
	AreaMax       *int64 `form:"area_max"`
	Independent   *bool  `form:"independent"`
	DrivingSide   string `form:"driving_side"`
}
