// backend/internal/database/migrations.go

package database

import (
	"github.com/Culturae-org/culturae/internal/model"

	"gorm.io/gorm"
)

func createExtensions(db *gorm.DB) error {
	extensions := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`, // UUID generation functions
		`CREATE EXTENSION IF NOT EXISTS "citext"`,    // Case-insensitive text type
	}

	for _, ext := range extensions {
		if err := db.Exec(ext).Error; err != nil {
			return err
		}
	}

	return nil
}

func RunMigrations(db *gorm.DB) error {
	if err := createExtensions(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.UserGameStats{},
		&model.UserGameStatsByMode{},
		&model.Session{},
		&model.UserConnectionLog{},
		&model.AdminActionLog{},
		&model.UserActionLog{},
		&model.APIRequestLog{},
		&model.SecurityEventLog{},
		&model.Theme{},
		&model.Friend{},
		&model.FriendRequest{},
		&model.QuestionDataset{},
		&model.Question{},
		&model.ImportJob{},
		&model.ImportQuestionLog{},
		&model.Game{},
		&model.GamePlayer{},
		&model.GameInvite{},
		&model.GameQuestion{},
		&model.GameAnswer{},
		&model.GeographyDataset{},
		&model.Country{},
		&model.Continent{},
		&model.Region{},
		&model.QuestionReport{},
		&model.GameTemplate{},
		&model.GameEventLog{},
	); err != nil {
		return err
	}

	return nil
}
