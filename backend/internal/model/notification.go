// backend/internal/model/notification.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"-"`
	Type      string         `gorm:"type:varchar(100);not null" json:"type"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Body      string         `gorm:"type:text" json:"body"`
	Data      datatypes.JSON `gorm:"type:jsonb" json:"data,omitempty"`
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time      `json:"created_at"`
}
