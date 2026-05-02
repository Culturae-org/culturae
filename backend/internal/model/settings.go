// backend/internal/model/settings.go

package model

import (
	"math"
	"time"
)

const defaultHourlyCron = "0 * * * *"

type WebSocketConfig struct {
	WriteWaitSeconds int   `json:"write_wait_seconds"`
	PongWaitSeconds  int   `json:"pong_wait_seconds"`
	MaxMessageSizeKB int64 `json:"max_message_size_kb"`

	AllowedOrigins []string `json:"allowed_origins"`

	ReconnectGracePeriodSeconds int `json:"reconnect_grace_period_seconds"`

	MessageRateLimit         int `json:"message_rate_limit"`
	MessageRateWindowSeconds int `json:"message_rate_window_seconds"`
}

func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		WriteWaitSeconds:            10,
		PongWaitSeconds:             60,
		MaxMessageSizeKB:            512,
		AllowedOrigins:              []string{},
		ReconnectGracePeriodSeconds: 180,
		MessageRateLimit:            0,
		MessageRateWindowSeconds:    60,
	}
}

type AvatarConfig struct {
	MaxFileSizeMB     int      `json:"max_file_size_mb"`
	AllowedMimeTypes  []string `json:"allowed_mime_types"`
	AllowedExtensions []string `json:"allowed_extensions"`
}

func DefaultAvatarConfig() AvatarConfig {
	return AvatarConfig{
		MaxFileSizeMB:     5,
		AllowedMimeTypes:  []string{"image/jpeg", "image/png"},
		AllowedExtensions: []string{".png", ".jpeg", ".jpg"},
	}
}

func (cfg AvatarConfig) MaxFileSizeBytes() int64 {
	return int64(cfg.MaxFileSizeMB) * 1024 * 1024
}

type RankDefinition struct {
	Name     string `json:"name"`
	MinLevel int    `json:"min_level"`
}

type XPConfig struct {
	BaseXP     float64 `json:"base_xp"`
	GrowthRate float64 `json:"growth_rate"`

	SoloMultiplier     float64 `json:"solo_multiplier"`
	OneVsOneMultiplier float64 `json:"onevone_multiplier"`
	MultiMultiplier    float64 `json:"multi_multiplier"`

	WinnerBonus int64 `json:"winner_bonus"`

	Ranks []RankDefinition `json:"ranks"`
}

func DefaultXPConfig() XPConfig {
	return XPConfig{
		BaseXP:                   2000,
		GrowthRate:               1.5,
		SoloMultiplier:       0.5,
		OneVsOneMultiplier:   1.0,
		MultiMultiplier:      1.0,
		WinnerBonus:          100,
		Ranks: []RankDefinition{
			{Name: "Beginner", MinLevel: 0},
			{Name: "Intermediate", MinLevel: 5},
			{Name: "Pro", MinLevel: 10},
			{Name: "Expert", MinLevel: 15},
			{Name: "Legend", MinLevel: 20},
		},
	}
}

func (cfg XPConfig) CalculateLevel(totalXP int64) int {
	if totalXP <= 0 {
		return 0
	}
	return int(math.Log(float64(totalXP)/cfg.BaseXP+1.0) / math.Log(cfg.GrowthRate))
}

func (cfg XPConfig) RankFromLevel(level int) string {
	best := "Beginner"
	bestMin := -1
	for _, r := range cfg.Ranks {
		if level >= r.MinLevel && r.MinLevel > bestMin {
			best = r.Name
			bestMin = r.MinLevel
		}
	}
	return best
}

func (cfg XPConfig) MultiplierForMode(mode GameMode) float64 {
	switch mode {
	case GameModeSolo:
		return cfg.SoloMultiplier
	case GameMode1v1:
		return cfg.OneVsOneMultiplier
	case GameModeMulti:
		return cfg.MultiMultiplier
	default:
		return 1.0
	}
}

type GameConfig struct {
	ActiveTTLMinutes   int `json:"active_ttl_minutes"`
	FinishedTTLMinutes int `json:"finished_ttl_minutes"`
}

func DefaultGameConfig() GameConfig {
	return GameConfig{
		ActiveTTLMinutes:   1440,
		FinishedTTLMinutes: 120,
	}
}

func (cfg GameConfig) ActiveTTL() time.Duration {
	return time.Duration(cfg.ActiveTTLMinutes) * time.Minute
}

func (cfg GameConfig) FinishedTTL() time.Duration {
	return time.Duration(cfg.FinishedTTLMinutes) * time.Minute
}

type SystemConfig struct {
	UserCacheTTLMinutes int `json:"user_cache_ttl_minutes"`

	CleanupIntervalMinutes int `json:"cleanup_interval_minutes"`

	OfflineDelaySeconds   int `json:"offline_delay_seconds"`
	GameLeaveDelaySeconds int `json:"game_leave_delay_seconds"`

	AnalyticsActiveDays  int `json:"analytics_active_days"`
	AnalyticsArchiveDays int `json:"analytics_archive_days"`

	DatasetCheckEnabled bool   `json:"dataset_check_enabled"`
	DatasetCheckCron    string `json:"dataset_check_cron"`
	VersionCheckEnabled bool   `json:"version_check_enabled"`
	VersionCheckCron    string `json:"version_check_cron"`

	SessionCleanupEnabled bool   `json:"session_cleanup_enabled"`
	SessionCleanupCron    string `json:"session_cleanup_cron"`
	GameCleanupEnabled    bool   `json:"game_cleanup_enabled"`
	GameCleanupCron       string `json:"game_cleanup_cron"`
}

func DefaultSystemConfig() SystemConfig {
	return SystemConfig{
		UserCacheTTLMinutes:         1440,
		CleanupIntervalMinutes:      5,
		OfflineDelaySeconds:         2,
		GameLeaveDelaySeconds:       30,
		AnalyticsActiveDays:         1,
		AnalyticsArchiveDays:        30,
		DatasetCheckEnabled: false,
		DatasetCheckCron:    defaultHourlyCron,
		VersionCheckEnabled: false,
		VersionCheckCron:    defaultHourlyCron,

		SessionCleanupEnabled: true,
		SessionCleanupCron:    defaultHourlyCron,
		GameCleanupEnabled:    true,
		GameCleanupCron:       "*/5 * * * *",
	}
}

func (cfg SystemConfig) UserCacheTTL() time.Duration {
	return time.Duration(cfg.UserCacheTTLMinutes) * time.Minute
}

func (cfg SystemConfig) CleanupInterval() time.Duration {
	return time.Duration(cfg.CleanupIntervalMinutes) * time.Minute
}

func (cfg SystemConfig) OfflineDelay() time.Duration {
	return time.Duration(cfg.OfflineDelaySeconds) * time.Second
}

func (cfg SystemConfig) GameLeaveDelay() time.Duration {
	return time.Duration(cfg.GameLeaveDelaySeconds) * time.Second
}

type ELOConfig struct {
	KFactorLowGames  int `json:"k_factor_low_games"`
	KFactorHighGames int `json:"k_factor_high_games"`
	KFactorThreshold int `json:"k_factor_threshold"`
	MinRating        int `json:"min_rating"`
	MaxRating        int `json:"max_rating"`
}

func DefaultELOConfig() ELOConfig {
	return ELOConfig{
		KFactorLowGames:  32,
		KFactorHighGames: 16,
		KFactorThreshold: 30,
		MinRating:        0,
		MaxRating:        9999,
	}
}

type CountdownConfig struct {
	PreGameCountdownSeconds     int `json:"pre_game_countdown_seconds"`
	ReconnectGracePeriodSeconds int `json:"reconnect_grace_period_seconds"`
}

func DefaultCountdownConfig() CountdownConfig {
	return CountdownConfig{
		PreGameCountdownSeconds:     3,
		ReconnectGracePeriodSeconds: 30,
	}
}

type AuthConfig struct {
	AccessTokenTTLMinutes  int `json:"access_token_ttl_minutes"`
	RefreshTokenTTLDays    int `json:"refresh_token_ttl_days"`
	SessionTTLDays         int `json:"session_ttl_days"`
	MaxConcurrentSessions  int `json:"max_concurrent_sessions"`
	FailedLoginAttempts    int `json:"failed_login_attempts"`
	LoginLockoutMinutes    int `json:"login_lockout_minutes"`
}

func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLDays:   7,
		SessionTTLDays:        30,
		MaxConcurrentSessions: 5,
		FailedLoginAttempts:   5,
		LoginLockoutMinutes:   15,
	}
}
