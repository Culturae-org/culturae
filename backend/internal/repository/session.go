// backend/internal/repository/session.go

package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SessionRepositoryInterface interface {
	CreateSession(session *model.Session) error
	UpdateSession(session *model.Session) error
	GetSessionByID(sessionID uuid.UUID) (*model.Session, error)
	GetSessionByTokenID(tokenID string) (*model.Session, error)
	GetSessionByRefreshToken(refreshToken string) (*model.Session, error)
	GetSessionByPreviousRefreshToken(previousRefresh string) (*model.Session, error)
	GetActiveSessionsByUserID(userID uuid.UUID) ([]model.Session, error)
	UpdateSessionLastUsed(tokenID string) error
	RevokeSession(tokenID string) error
	RevokeSessionByRefreshToken(refreshToken string) error
	RevokeAllUserSessions(userID uuid.UUID) error
	RevokeOldestUserSessions(userID uuid.UUID, keepCount int) error
	CleanupExpiredSessions() error
	GetSessionStats(userID uuid.UUID) (map[string]interface{}, error)
	ValidateSession(tokenID string) (*model.Session, error)
	RefreshSessionAtomic(oldTokenID string, newSession *model.Session) error
}

type SessionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSessionRepository(
	db *gorm.DB,
	logger *zap.Logger,
) *SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

func (sr *SessionRepository) CreateSession(session *model.Session) error {
	return sr.db.Create(session).Error
}

func (sr *SessionRepository) UpdateSession(session *model.Session) error {
	return sr.db.Save(session).Error
}

func (sr *SessionRepository) GetSessionByID(sessionID uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := sr.db.Where("id = ?", sessionID).
		Preload("User").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (sr *SessionRepository) GetSessionByTokenID(tokenID string) (*model.Session, error) {
	var session model.Session
	err := sr.db.Where("token_id = ? AND is_active = ?", tokenID, true).
		Preload("User").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (sr *SessionRepository) GetSessionByRefreshToken(refreshToken string) (*model.Session, error) {
	var session model.Session
	err := sr.db.Where("refresh_token = ? AND is_active = ?", refreshToken, true).
		Preload("User").
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (sr *SessionRepository) GetSessionByPreviousRefreshToken(previousRefresh string) (*model.Session, error) {
	var session model.Session
	err := sr.db.Where("previous_refresh = ? AND is_active = ?", previousRefresh, true).
		Preload("User").
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (sr *SessionRepository) GetActiveSessionsByUserID(userID uuid.UUID) ([]model.Session, error) {
	var sessions []model.Session
	err := sr.db.Where("user_id = ? AND is_active = ? AND expires_at > ?",
		userID, true, time.Now()).
		Order("last_used DESC").
		Find(&sessions).Error
	return sessions, err
}

func (sr *SessionRepository) UpdateSessionLastUsed(tokenID string) error {
	return sr.db.Model(&model.Session{}).
		Where("token_id = ?", tokenID).
		Update("last_used", time.Now()).Error
}

func (sr *SessionRepository) RevokeSession(tokenID string) error {
	return sr.db.Model(&model.Session{}).
		Where("token_id = ?", tokenID).
		Update("is_active", false).Error
}

func (sr *SessionRepository) RevokeSessionByRefreshToken(refreshToken string) error {
	return sr.db.Model(&model.Session{}).
		Where("refresh_token = ?", refreshToken).
		Update("is_active", false).Error
}

func (sr *SessionRepository) RevokeAllUserSessions(userID uuid.UUID) error {
	return sr.db.Model(&model.Session{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

func (sr *SessionRepository) RevokeOldestUserSessions(userID uuid.UUID, keepCount int) error {
	var sessions []model.Session

	err := sr.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used DESC").
		Find(&sessions).Error
	if err != nil {
		return err
	}

	if len(sessions) > keepCount {
		sessionsToRevoke := sessions[keepCount:]
		tokenIDs := make([]string, len(sessionsToRevoke))
		for i, session := range sessionsToRevoke {
			tokenIDs[i] = session.TokenID
		}

		return sr.db.Model(&model.Session{}).
			Where("token_id IN ?", tokenIDs).
			Update("is_active", false).Error
	}

	return nil
}

func (sr *SessionRepository) CleanupExpiredSessions() error {

	err := sr.db.Model(&model.Session{}).
		Where("expires_at < ? AND is_active = ?", time.Now(), true).
		Update("is_active", false).Error
	if err != nil {
		return fmt.Errorf("failed to mark expired sessions as inactive: %w", err)
	}

	err = sr.db.Where("expires_at < ? AND is_active = ?",
		time.Now().Add(-7*24*time.Hour), false).
		Delete(&model.Session{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete old sessions: %w", err)
	}

	return nil
}

func (sr *SessionRepository) GetSessionStats(userID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var activeCount int64
	err := sr.db.Model(&model.Session{}).
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Count(&activeCount).Error
	if err != nil {
		return nil, err
	}
	stats["active_sessions"] = activeCount

	var totalCount int64
	err = sr.db.Model(&model.Session{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	stats["total_sessions"] = totalCount

	var lastSession model.Session
	err = sr.db.Where("user_id = ?", userID).
		Order("last_used DESC").
		First(&lastSession).Error
	if err == nil {
		stats["last_activity"] = lastSession.LastUsed
	}

	return stats, nil
}

func (sr *SessionRepository) ValidateSession(tokenID string) (*model.Session, error) {
	session, err := sr.GetSessionByTokenID(tokenID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		if err := sr.RevokeSession(tokenID); err != nil {
			sr.logger.Error("Failed to revoke expired session", zap.Error(err))
		}
		return nil, fmt.Errorf("session expired")
	}

	if err := sr.UpdateSessionLastUsed(tokenID); err != nil {
		sr.logger.Error("Failed to update session last used", zap.Error(err))
	}

	return session, nil
}

func (sr *SessionRepository) RefreshSessionAtomic(oldTokenID string, newSession *model.Session) error {
	return sr.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Session{}).
			Where("token_id = ?", oldTokenID).
			Updates(map[string]interface{}{
				"is_active":      false,
				"is_revoked":     true,
				"revoked_at":     time.Now(),
				"revoked_reason": "refreshed",
			}).Error; err != nil {
			return fmt.Errorf("failed to revoke old session: %w", err)
		}

		if err := tx.Create(newSession).Error; err != nil {
			return fmt.Errorf("failed to create new session: %w", err)
		}

		return nil
	})
}
