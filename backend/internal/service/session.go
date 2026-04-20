// backend/internal/service/session.go

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"

	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
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

type SessionServiceInterface interface {
	AddHook(hookType model.SessionHookType, hook model.SessionHookFunc)
	CreateSession(user *model.User, ipAddress string, userAgent string, variables datatypes.JSON) (*model.Session, error)
	GetSession(tokenID string) (*model.Session, error)
	GetSessionByID(sessionID uuid.UUID) (*model.Session, error)
	RefreshSession(refreshToken string, ipAddress string, userAgent string) (*model.Session, error)
	RevokeSession(tokenID string) error
	RevokeAllUserSessions(userID uuid.UUID) error
	CleanupExpiredSessions() error
	GetUserSessions(userID uuid.UUID) ([]model.Session, error)
	GetSessionStats(userID uuid.UUID) (map[string]interface{}, error)
}

type SessionService struct {
	sessionRepo SessionRepositoryInterface
	redisClient *cache.RedisClient
	config      *model.SessionConfig
	hooks       map[model.SessionHookType][]model.SessionHookFunc
	logger      *zap.Logger
}

func NewSessionService(
	sessionRepo SessionRepositoryInterface,
	redisClient *cache.RedisClient,
	config *model.SessionConfig,
	logger *zap.Logger,
) *SessionService {
	if config == nil {
		config = model.DefaultSessionConfig()
	}

	return &SessionService{
		sessionRepo: sessionRepo,
		redisClient: redisClient,
		config:      config,
		hooks:       make(map[model.SessionHookType][]model.SessionHookFunc),
		logger:      logger,
	}
}

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrRefreshTokenReused  = errors.New("refresh token reuse detected")
)

func (ss *SessionService) redisCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (ss *SessionService) AddHook(hookType model.SessionHookType, hook model.SessionHookFunc) {
	ss.hooks[hookType] = append(ss.hooks[hookType], hook)
}

func (ss *SessionService) executeHooks(ctx *model.SessionHookContext) {
	if hooks, exists := ss.hooks[ctx.Type]; exists {
		for _, hook := range hooks {
			if err := hook(ctx); err != nil {
				ss.logger.Error("Hook error", zap.String("type", string(ctx.Type)), zap.Error(err))
			}
		}
	}
}

func (ss *SessionService) generateTokenID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		ss.logger.Error("Failed to generate random bytes for token ID", zap.Error(err))
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func (ss *SessionService) generateRefreshToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		ss.logger.Error("Failed to generate random bytes for refresh token", zap.Error(err))
		return fmt.Sprintf("%d%s", time.Now().UnixNano(), uuid.New().String())
	}
	return hex.EncodeToString(bytes)
}

func (ss *SessionService) CreateSession(
	user *model.User,
	ipAddress string,
	userAgent string,
	variables datatypes.JSON,
) (*model.Session, error) {
	tokenID := ss.generateTokenID()
	refreshToken := ss.generateRefreshToken()

	deviceFingerprint := fmt.Sprintf("%x",
		[]byte(userAgent + ipAddress)[:min(32, len(userAgent+ipAddress))])

	session := &model.Session{
		ID:                uuid.New(),
		UserID:            user.ID,
		TokenID:           tokenID,
		RefreshToken:      refreshToken,
		PreviousRefresh:   "",
		Created:           time.Now(),
		ExpiresAt:         time.Now().Add(ss.config.RefreshTokenDuration),
		LastUsed:          time.Now(),
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		DeviceFingerprint: deviceFingerprint,
		IsActive:          true,
		IsRevoked:         false,
		RevokedAt:         nil,
		RevokedReason:     "",
		Variables:         variables,
		User:              *user,
	}

	if ss.config.MaxActiveSessions > 0 {
		activeSessions, err := ss.sessionRepo.GetActiveSessionsByUserID(user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check active sessions: %w", err)
		}

		if len(activeSessions) >= ss.config.MaxActiveSessions {
			keepCount := ss.config.MaxActiveSessions - 1
			err = ss.sessionRepo.RevokeOldestUserSessions(user.ID, keepCount)
			if err != nil {
				ss.logger.Error("Failed to revoke oldest sessions", zap.Error(err))
			}
		}
	}

	if err := ss.sessionRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if ss.redisClient != nil {
		ctx, cancel := ss.redisCtx()
		cacheKey := fmt.Sprintf("session:%s", tokenID)
		if err := ss.redisClient.SetJSON(ctx, cacheKey, session, ss.config.AccessTokenDuration); err != nil {
			ss.logger.Error("Failed to cache session", zap.Error(err))
		}
		cancel()
	}

	return session, nil
}

func (ss *SessionService) GetSession(tokenID string) (*model.Session, error) {
	if ss.redisClient != nil {
		cacheKey := fmt.Sprintf("session:%s", tokenID)
		var cachedSession model.Session
		getCtx, getCancel := ss.redisCtx()
		err := ss.redisClient.GetJSON(getCtx, cacheKey, &cachedSession)
		getCancel()
		if err == nil {
			return &cachedSession, nil
		}
	}

	session, err := ss.sessionRepo.ValidateSession(tokenID)
	if err != nil {
		return nil, err
	}

	if ss.redisClient != nil {
		cacheKey := fmt.Sprintf("session:%s", tokenID)
		setCtx, setCancel := ss.redisCtx()
		if err := ss.redisClient.SetJSON(setCtx, cacheKey, session, ss.config.AccessTokenDuration); err != nil {
			ss.logger.Error("Failed to cache session", zap.Error(err))
		}
		setCancel()
	}

	return session, nil
}

func (ss *SessionService) RefreshSession(
	refreshToken string,
	ipAddress string,
	userAgent string,
) (*model.Session, error) {
	session, err := ss.sessionRepo.GetSessionByRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, model.ErrSessionNotFound) {
			if ss.config.EnableReuseDetection {
				if prevSession, prevErr := ss.sessionRepo.GetSessionByPreviousRefreshToken(refreshToken); prevErr == nil {
					ss.logger.Warn("Refresh token reuse detected, revoking compromised session",
						zap.String("user_id", prevSession.UserID.String()),
						zap.String("session_id", prevSession.ID.String()),
						zap.String("ip", prevSession.IPAddress),
					)
					_ = ss.sessionRepo.RevokeSession(prevSession.TokenID)
					if ss.redisClient != nil {
						delCtx, delCancel := ss.redisCtx()
						_ = ss.redisClient.Delete(delCtx, fmt.Sprintf("session:%s", prevSession.TokenID))
						delCancel()
					}
					ss.executeHooks(&model.SessionHookContext{
						Type:      model.SessionHookRefreshReuseDetected,
						Session:   prevSession,
						User:      &prevSession.User,
						IPAddress: prevSession.IPAddress,
						UserAgent: prevSession.UserAgent,
					})
					return nil, ErrRefreshTokenReused
				}
			}
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("failed to fetch session by refresh token: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		if err := ss.sessionRepo.RevokeSessionByRefreshToken(refreshToken); err != nil {
			ss.logger.Error("Failed to revoke expired session", zap.Error(err))
		}
		return nil, ErrRefreshTokenExpired
	}

	if session.IsRevoked {
		return nil, ErrSessionRevoked
	}

	newTokenID := ss.generateTokenID()
	newRefreshToken := ss.generateRefreshToken()

	oldTokenID := session.TokenID

	deviceFingerprint := fmt.Sprintf("%x",
		[]byte(userAgent + ipAddress)[:min(32, len(userAgent+ipAddress))])

	newSession := &model.Session{
		ID:                uuid.New(),
		UserID:            session.UserID,
		TokenID:           newTokenID,
		RefreshToken:      newRefreshToken,
		PreviousRefresh:   session.RefreshToken,
		Created:           time.Now(),
		ExpiresAt:         time.Now().Add(ss.config.RefreshTokenDuration),
		LastUsed:          time.Now(),
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		DeviceFingerprint: deviceFingerprint,
		IsActive:          true,
		Variables:         session.Variables,
		User:              session.User,
	}

	if err := ss.sessionRepo.RefreshSessionAtomic(oldTokenID, newSession); err != nil {
		return nil, fmt.Errorf("failed to refresh session atomically: %w", err)
	}

	if ss.redisClient != nil {
		setCtx, setCancel := ss.redisCtx()
		cacheKey := fmt.Sprintf("session:%s", newTokenID)
		if err := ss.redisClient.SetJSON(setCtx, cacheKey, newSession, ss.config.AccessTokenDuration); err != nil {
			ss.logger.Error("Failed to cache refreshed session", zap.Error(err))
		}
		setCancel()

		delCtx, delCancel := ss.redisCtx()
		oldCacheKey := fmt.Sprintf("session:%s", oldTokenID)
		if err := ss.redisClient.Delete(delCtx, oldCacheKey); err != nil {
			ss.logger.Error("Failed to delete old session cache", zap.Error(err))
		}
		delCancel()
	}

	ss.executeHooks(&model.SessionHookContext{
		Type:      model.SessionHookAfterRefresh,
		Session:   newSession,
		User:      &newSession.User,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Variables: newSession.Variables,
	})

	return newSession, nil
}

func (ss *SessionService) RevokeSession(tokenID string) error {
	session, _ := ss.sessionRepo.GetSessionByTokenID(tokenID)

	if err := ss.sessionRepo.RevokeSession(tokenID); err != nil {
		return err
	}

	if ss.redisClient != nil {
		delCtx, delCancel := ss.redisCtx()
		cacheKey := fmt.Sprintf("session:%s", tokenID)
		if err := ss.redisClient.Delete(delCtx, cacheKey); err != nil {
			ss.logger.Error("Failed to delete session from cache", zap.Error(err))
		}
		delCancel()
	}

	if session != nil {
		ss.executeHooks(&model.SessionHookContext{
			Type:      model.SessionHookAfterRevoke,
			Session:   session,
			User:      &session.User,
			IPAddress: session.IPAddress,
			UserAgent: session.UserAgent,
			Variables: session.Variables,
		})
	}

	return nil
}

func (ss *SessionService) RevokeAllUserSessions(userID uuid.UUID) error {
	sessions, _ := ss.sessionRepo.GetActiveSessionsByUserID(userID)

	if err := ss.sessionRepo.RevokeAllUserSessions(userID); err != nil {
		return err
	}

	for _, session := range sessions {
		if ss.redisClient != nil {
			delCtx, delCancel := ss.redisCtx()
			cacheKey := fmt.Sprintf("session:%s", session.TokenID)
			if err := ss.redisClient.Delete(delCtx, cacheKey); err != nil {
				ss.logger.Error("Failed to delete session from cache", zap.Error(err))
			}
			delCancel()
		}

		ss.executeHooks(&model.SessionHookContext{
			Type:      model.SessionHookAfterRevoke,
			Session:   &session,
			User:      &session.User,
			IPAddress: session.IPAddress,
			UserAgent: session.UserAgent,
			Variables: session.Variables,
		})
	}

	return nil
}

func (ss *SessionService) CleanupExpiredSessions() error {
	return ss.sessionRepo.CleanupExpiredSessions()
}

func (ss *SessionService) GetUserSessions(userID uuid.UUID) ([]model.Session, error) {
	return ss.sessionRepo.GetActiveSessionsByUserID(userID)
}

func (ss *SessionService) GetSessionByID(sessionID uuid.UUID) (*model.Session, error) {
	return ss.sessionRepo.GetSessionByID(sessionID)
}

func (ss *SessionService) GetSessionStats(userID uuid.UUID) (map[string]interface{}, error) {
	return ss.sessionRepo.GetSessionStats(userID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
