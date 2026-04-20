// backend/internal/model/session.go

package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SessionTokenClaims struct {
	USN   string         `json:"usn"`
	SID   string         `json:"sid"`
	Roles []string       `json:"roles"`
	VRS   datatypes.JSON `json:"vrs"`
	jwt.RegisteredClaims
}

type Session struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID            uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	TokenID           string         `json:"token_id" gorm:"uniqueIndex;not null"`
	RefreshToken      string         `json:"refresh_token" gorm:"uniqueIndex;not null"`
	PreviousRefresh   string         `json:"previous_refresh" gorm:"index"`
	Created           time.Time      `json:"created_at" gorm:"not null"`
	ExpiresAt         time.Time      `json:"expires_at" gorm:"not null"`
	LastUsed          time.Time      `json:"last_used" gorm:"not null"`
	IPAddress         string         `json:"ip_address"`
	UserAgent         string         `json:"user_agent"`
	DeviceFingerprint string         `json:"device_fingerprint"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	IsRevoked         bool           `json:"is_revoked" gorm:"default:false"`
	RevokedAt         *time.Time     `json:"revoked_at"`
	RevokedReason     string         `json:"revoked_reason"`
	Variables         datatypes.JSON `json:"variables" gorm:"type:jsonb"`
	CreatedAt         time.Time      `json:"created_at_db"`
	UpdatedAt         time.Time      `json:"updated_at"`
	User              User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type SessionResponse struct {
	Created        time.Time `json:"created"`
	TokenType      string    `json:"token_type"`
	Token          string    `json:"token"`
	RefreshToken   string    `json:"refresh_token"`
	ExpiresAt      int64     `json:"expires_at"`
	TokenExpiresAt int64     `json:"token_expires_at"`
	User           UserAuth  `json:"user"`
}

type SessionInfo struct {
	ID                uuid.UUID      `json:"id"`
	Created           time.Time      `json:"created"`
	ExpiresAt         time.Time      `json:"expires_at"`
	LastUsed          time.Time      `json:"last_used"`
	IPAddress         string         `json:"ip_address"`
	UserAgent         string         `json:"user_agent"`
	DeviceFingerprint string         `json:"device_fingerprint"`
	IsActive          bool           `json:"is_active"`
	IsRevoked         bool           `json:"is_revoked"`
	RevokedAt         *time.Time     `json:"revoked_at"`
	RevokedReason     string         `json:"revoked_reason"`
	Variables         datatypes.JSON `json:"variables"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type SessionHookType string

const (
	SessionHookAfterLogin           SessionHookType = "after_login"
	SessionHookAfterRefresh         SessionHookType = "after_refresh"
	SessionHookAfterLogout          SessionHookType = "after_logout"
	SessionHookBeforeExpire         SessionHookType = "before_expire"
	SessionHookAfterRevoke          SessionHookType = "after_revoke"
	SessionHookRefreshReuseDetected SessionHookType = "refresh_reuse_detected"
)

type SessionHookContext struct {
	Type      SessionHookType `json:"type"`
	Session   *Session        `json:"session"`
	User      *User           `json:"user"`
	IPAddress string          `json:"ip_address"`
	UserAgent string          `json:"user_agent"`
	Variables datatypes.JSON  `json:"variables"`
	Error     error           `json:"error,omitempty"`
}

type SessionHookFunc func(ctx *SessionHookContext) error

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type SessionConfig struct {
	AccessTokenDuration    time.Duration `json:"access_token_duration"`
	RefreshTokenDuration   time.Duration `json:"refresh_token_duration"`
	MaxActiveSessions      int           `json:"max_active_sessions"`
	CleanupInterval        time.Duration `json:"cleanup_interval"`
	RevokeOnPasswordChange bool          `json:"revoke_on_password_change"`
	EnableRotation         bool          `json:"enable_rotation"`
	EnableReuseDetection   bool          `json:"enable_reuse_detection"`
	RateLimitRequests      int           `json:"rate_limit_requests"`
	RateLimitWindow        time.Duration `json:"rate_limit_window"`
}

func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		AccessTokenDuration:    time.Minute * 15,
		RefreshTokenDuration:   time.Hour * 24 * 30,
		MaxActiveSessions:      5,
		CleanupInterval:        time.Hour,
		RevokeOnPasswordChange: true,
		EnableRotation:         true,
		EnableReuseDetection:   true,
		RateLimitRequests:      10,
		RateLimitWindow:        time.Minute,
	}
}

func DevSessionConfig() *SessionConfig {
	return &SessionConfig{
		AccessTokenDuration:    time.Minute * 5,
		RefreshTokenDuration:   time.Hour * 24 * 7,
		MaxActiveSessions:      3,
		CleanupInterval:        time.Minute * 30,
		RevokeOnPasswordChange: true,
		EnableRotation:         true,
		EnableReuseDetection:   true,
		RateLimitRequests:      20,
		RateLimitWindow:        time.Minute,
	}
}

func ProductionSessionConfig() *SessionConfig {
	return &SessionConfig{
		AccessTokenDuration:    time.Minute * 10,
		RefreshTokenDuration:   time.Hour * 24 * 60,
		MaxActiveSessions:      3,
		CleanupInterval:        time.Hour * 6,
		RevokeOnPasswordChange: true,
		EnableRotation:         true,
		EnableReuseDetection:   true,
		RateLimitRequests:      5,
		RateLimitWindow:        time.Minute,
	}
}
