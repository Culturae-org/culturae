// backend/internal/model/connections.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type UserActionStats struct {
	TotalActions  int64            `json:"total_actions"`
	SuccessRate   float64          `json:"success_rate"`
	ActionsByType map[string]int64 `json:"actions_by_type"`
	ActionsByUser map[string]int64 `json:"actions_by_user"`
	DailyAverage  float64          `json:"daily_average"`
}

type UserConnectionLog struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID        *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	SessionID     *uuid.UUID     `gorm:"type:uuid;index" json:"session_id"`
	IPAddress     string         `gorm:"type:varchar(45);not null" json:"ip_address"`
	UserAgent     string         `gorm:"type:text;not null" json:"user_agent"`
	DeviceInfo    datatypes.JSON `gorm:"type:jsonb" json:"device_info"`
	Location      string         `gorm:"type:varchar(100)" json:"location"`
	IsSuccess     bool           `gorm:"default:false" json:"is_success"`
	FailureReason *string        `gorm:"type:varchar(100)" json:"failure_reason"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

type DeviceInfo struct {
	DeviceType string `json:"device_type"`
	OS         string `json:"os"`
	Browser    string `json:"browser"`
}

type AdminActionLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AdminID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	AdminName  string         `gorm:"type:varchar(100);not null"`
	Action     string         `gorm:"type:varchar(100);not null"`
	Resource   string         `gorm:"type:varchar(100);not null"`
	ResourceID *uuid.UUID     `gorm:"type:uuid;index"`
	IPAddress  string         `gorm:"type:varchar(45)"`
	UserAgent  string         `gorm:"type:text"`
	Details    datatypes.JSON `gorm:"type:jsonb"`
	IsSuccess  bool           `gorm:""`
	ErrorMsg   *string        `gorm:"type:text"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
}

type UserActionLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index"`
	Username   string         `gorm:"type:varchar(100);not null"`
	Action     string         `gorm:"type:varchar(100);not null"`
	Resource   string         `gorm:"type:varchar(100);not null"`
	ResourceID *uuid.UUID     `gorm:"type:uuid;index"`
	IPAddress  string         `gorm:"type:varchar(45)"`
	UserAgent  string         `gorm:"type:text"`
	Details    datatypes.JSON `gorm:"type:jsonb"`
	IsSuccess  bool           `gorm:""`
	ErrorMsg   *string        `gorm:"type:text"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
}

type SecurityEventLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index"`
	EventType string         `gorm:"type:varchar(100);not null"`
	IPAddress string         `gorm:"type:varchar(45)"`
	UserAgent string         `gorm:"type:text"`
	Details   datatypes.JSON `gorm:"type:jsonb"`
	IsSuccess bool           `gorm:""`
	ErrorMsg  *string        `gorm:"type:text"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
}

type APIRequestLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Method       string     `gorm:"type:varchar(10);not null"`
	Path         string     `gorm:"type:varchar(500);not null"`
	StatusCode   int        `gorm:"not null"`
	UserID       *uuid.UUID `gorm:"type:uuid;index"`
	IPAddress    string     `gorm:"type:varchar(45)"`
	UserAgent    string     `gorm:"type:text"`
	RequestSize  int64      `gorm:"default:0"`
	ResponseSize int64      `gorm:"default:0"`
	Duration     int64      `gorm:"default:0"`
	IsError      bool       `gorm:"default:false"`
	ErrorMsg     *string    `gorm:"type:text"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}

type AdminActionStats struct {
	TotalActions      int64            `json:"total_actions"`
	SuccessRate       float64          `json:"success_rate"`
	ActionsByType     map[string]int64 `json:"actions_by_type"`
	ActionsByResource map[string]int64 `json:"actions_by_resource"`
	ActionsByAdmin    map[string]int64 `json:"actions_by_admin"`
	DailyAverage      float64          `json:"daily_average"`
}

type APIRequestStats struct {
	TotalRequests    int64            `json:"total_requests"`
	ErrorRate        float64          `json:"error_rate"`
	AvgResponseTime  float64          `json:"avg_response_time"`
	RequestsByStatus map[string]int64 `json:"requests_by_status"`
	RequestsByMethod map[string]int64 `json:"requests_by_method"`
	RequestsByPath   map[string]int64 `json:"requests_by_path"`
	DailyAverage     float64          `json:"daily_average"`
}

