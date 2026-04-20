// backend/internal/handler/admin/settings.go

package admin

import (
	"fmt"
	"net/http"

	"github.com/Culturae-org/culturae/internal/config"
	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	maintenanceKey  = "system:maintenance:enabled"
	rateLimitKey    = "system:ratelimit:config"
	wsConfigKey     = "system:websocket:config"
	avatarConfigKey = "system:avatar:config"
	xpConfigKey     = "system:xp:config"
	eloConfigKey    = "system:elo:config"
	gameConfigKey   = "system:game:config"
	systemConfigKey = "system:config"
)

type AdminSettingsHandler struct {
	RedisService      cache.RedisClientInterface
	Config            *config.Config
	LoggingService    service.LoggingServiceInterface
	Logger            *zap.Logger
	RestartSchedulers func()
}

func NewAdminSettingsHandler(
	redisService cache.RedisClientInterface,
	cfg *config.Config,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
	restartSchedulers func(),
) *AdminSettingsHandler {
	return &AdminSettingsHandler{
		RedisService:      redisService,
		Config:            cfg,
		LoggingService:    loggingService,
		Logger:            logger,
		RestartSchedulers: restartSchedulers,
	}
}

// -----------------------------------------------------
// Admin Settings Handlers
//
// - GetMaintenanceStatus
// - SetMaintenanceMode
// - GetRateLimitConfig
// - UpdateRateLimitConfig
// - GetWebSocketConfig
// - UpdateWebSocketConfig
// - GetAvatarConfig
// - UpdateAvatarConfig
// - GetXPConfig
// - UpdateXPConfig
// - GetELOConfig
// - UpdateELOConfig
// - GetGamesConfig
// - UpdateGamesConfig
// - ClearCache
// - GetGameCountdownConfig
// - UpdateGameCountdownConfig
// - GetSystemConfig
// - UpdateSystemConfig
// -----------------------------------------------------

type rateLimitConfig struct {
	Enabled        bool `json:"enabled"`
	ApplyToAdmin   bool `json:"apply_to_admin"`
	MaxRequests    int  `json:"max_requests"`
	WindowSeconds  int  `json:"window_seconds"`
}

func (sc *AdminSettingsHandler) GetMaintenanceStatus(c *gin.Context) {
	ctx := c.Request.Context()

	enabled, err := sc.RedisService.Exists(ctx, maintenanceKey)
	if err != nil {
		sc.Logger.Error("Failed to check maintenance status", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to check maintenance status")
		return
	}

	httputil.Success(c, http.StatusOK, enabled)
}

func (sc *AdminSettingsHandler) SetMaintenanceMode(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	ctx := c.Request.Context()

	if req.Enabled {
		if err := sc.RedisService.Set(ctx, maintenanceKey, "1", 0); err != nil {
			sc.Logger.Error("Failed to enable maintenance mode", zap.Error(err))
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to enable maintenance mode")
			return
		}
	} else {
		if err := sc.RedisService.Delete(ctx, maintenanceKey); err != nil {
			sc.Logger.Error("Failed to disable maintenance mode", zap.Error(err))
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to disable maintenance mode")
			return
		}
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "set_maintenance", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"enabled": req.Enabled}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req.Enabled)
}

func (sc *AdminSettingsHandler) GetRateLimitConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg rateLimitConfig
	if err := sc.RedisService.GetJSON(ctx, rateLimitKey, &cfg); err != nil {
		cfg = rateLimitConfig{
			Enabled:       sc.Config.RateLimitEnabled,
			MaxRequests:   sc.Config.RateLimitRequests,
			WindowSeconds: int(sc.Config.RateLimitWindow.Seconds()),
		}
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateRateLimitConfig(c *gin.Context) {
	var req rateLimitConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.Enabled && (req.MaxRequests <= 0 || req.WindowSeconds <= 0) {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "max_requests and window_seconds must be positive when rate limiting is enabled")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, rateLimitKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update rate limit config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update rate limit config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_rate_limit", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"enabled":        req.Enabled,
			"max_requests":   req.MaxRequests,
			"window_seconds": req.WindowSeconds,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetWebSocketConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.WebSocketConfig
	if err := sc.RedisService.GetJSON(ctx, wsConfigKey, &cfg); err != nil ||
		cfg.WriteWaitSeconds <= 0 || cfg.PongWaitSeconds <= 0 || cfg.MaxMessageSizeKB <= 0 {
		cfg = model.DefaultWebSocketConfig()
	} else {
		if cfg.ReconnectGracePeriodSeconds <= 0 {
			cfg.ReconnectGracePeriodSeconds = 180
		}
		if cfg.MessageRateWindowSeconds <= 0 {
			cfg.MessageRateWindowSeconds = 60
		}
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateWebSocketConfig(c *gin.Context) {
	var req model.WebSocketConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.WriteWaitSeconds <= 0 || req.PongWaitSeconds <= 0 || req.MaxMessageSizeKB <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "write_wait_seconds, pong_wait_seconds and max_message_size_kb must be positive")
		return
	}
	if req.ReconnectGracePeriodSeconds <= 0 {
		req.ReconnectGracePeriodSeconds = 180
	}
	if req.MessageRateWindowSeconds <= 0 {
		req.MessageRateWindowSeconds = 60
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, wsConfigKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update websocket config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update websocket config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_websocket_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"write_wait_seconds":              req.WriteWaitSeconds,
			"pong_wait_seconds":               req.PongWaitSeconds,
			"max_message_size_kb":             req.MaxMessageSizeKB,
			"allowed_origins":                 req.AllowedOrigins,
			"reconnect_grace_period_seconds":  req.ReconnectGracePeriodSeconds,
			"message_rate_limit":              req.MessageRateLimit,
			"message_rate_window_seconds":     req.MessageRateWindowSeconds,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetAvatarConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.AvatarConfig
	if err := sc.RedisService.GetJSON(ctx, avatarConfigKey, &cfg); err != nil ||
		cfg.MaxFileSizeMB <= 0 || len(cfg.AllowedMimeTypes) == 0 || len(cfg.AllowedExtensions) == 0 {
		cfg = model.DefaultAvatarConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateAvatarConfig(c *gin.Context) {
	var req model.AvatarConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.MaxFileSizeMB <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "max_file_size_mb must be positive")
		return
	}
	if len(req.AllowedMimeTypes) == 0 || len(req.AllowedExtensions) == 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "allowed_mime_types and allowed_extensions must have at least one entry")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, avatarConfigKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update avatar config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update avatar config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_avatar_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"max_file_size_mb":   req.MaxFileSizeMB,
			"allowed_mime_types": req.AllowedMimeTypes,
			"allowed_extensions": req.AllowedExtensions,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetXPConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.XPConfig
	if err := sc.RedisService.GetJSON(ctx, xpConfigKey, &cfg); err != nil ||
		cfg.BaseXP <= 0 || cfg.GrowthRate <= 0 {
		cfg = model.DefaultXPConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateXPConfig(c *gin.Context) {
	var req model.XPConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.BaseXP <= 0 || req.GrowthRate <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "base_xp and growth_rate must be positive")
		return
	}
	if len(req.Ranks) == 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "at least one rank definition is required")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, xpConfigKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update XP config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update XP config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_xp_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"base_xp":     req.BaseXP,
			"growth_rate": req.GrowthRate,
			"ranks":       len(req.Ranks),
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetELOConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.ELOConfig
	if err := sc.RedisService.GetJSON(ctx, eloConfigKey, &cfg); err != nil ||
		cfg.KFactorLowGames <= 0 || cfg.KFactorHighGames <= 0 || cfg.KFactorThreshold <= 0 {
		cfg = model.DefaultELOConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateELOConfig(c *gin.Context) {
	var req model.ELOConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.KFactorLowGames <= 0 || req.KFactorHighGames <= 0 || req.KFactorThreshold <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "k_factor_low_games, k_factor_high_games and k_factor_threshold must be positive")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, eloConfigKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update ELO config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update ELO config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_elo_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"k_factor_low_games":  req.KFactorLowGames,
			"k_factor_high_games": req.KFactorHighGames,
			"k_factor_threshold":  req.KFactorThreshold,
			"min_rating":          req.MinRating,
			"max_rating":          req.MaxRating,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetGamesConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.GameConfig
	if err := sc.RedisService.GetJSON(ctx, gameConfigKey, &cfg); err != nil ||
		cfg.ActiveTTLMinutes <= 0 || cfg.FinishedTTLMinutes <= 0 {
		cfg = model.DefaultGameConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateGamesConfig(c *gin.Context) {
	var req model.GameConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.ActiveTTLMinutes <= 0 || req.FinishedTTLMinutes <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "active_ttl_minutes and finished_ttl_minutes must be positive")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, gameConfigKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update game config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update game config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_game_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"active_ttl_minutes":   req.ActiveTTLMinutes,
			"finished_ttl_minutes": req.FinishedTTLMinutes,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) ClearCache(c *gin.Context) {
	ctx := c.Request.Context()

	preservePrefixes := []string{"system:", "session:", "rate_limit:"}

	keys, err := sc.RedisService.Scan(ctx, "*")
	if err != nil {
		sc.Logger.Error("Failed to scan Redis keys", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to clear cache")
		return
	}

	deleted := 0
	for _, key := range keys {
		preserve := false
		for _, prefix := range preservePrefixes {
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				preserve = true
				break
			}
		}
		if !preserve {
			if err := sc.RedisService.Delete(ctx, key); err != nil {
				sc.Logger.Warn("Failed to delete key", zap.String("key", key), zap.Error(err))
				continue
			}
			deleted++
		}
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "clear_cache", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"deleted_keys": deleted}, true, nil)
	}()

	httputil.SuccessWithMessage(c, http.StatusOK, "Cache cleared successfully", gin.H{"deleted_keys": deleted})
}

func (sc *AdminSettingsHandler) GetGameCountdownConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.CountdownConfig
	if err := sc.RedisService.GetJSON(ctx, "system:game:countdown", &cfg); err != nil ||
		cfg.PreGameCountdownSeconds <= 0 {
		cfg = model.DefaultCountdownConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateGameCountdownConfig(c *gin.Context) {
	var req model.CountdownConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.PreGameCountdownSeconds <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "pre_game_countdown_seconds must be positive")
		return
	}
	if req.ReconnectGracePeriodSeconds <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "reconnect_grace_period_seconds must be positive")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, "system:game:countdown", req, 0); err != nil {
		sc.Logger.Error("Failed to update game countdown config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update game countdown config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_game_countdown", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"pre_game_countdown_seconds":      req.PreGameCountdownSeconds,
			"reconnect_grace_period_seconds":  req.ReconnectGracePeriodSeconds,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetSystemConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.SystemConfig
	if err := sc.RedisService.GetJSON(ctx, systemConfigKey, &cfg); err != nil ||
		cfg.UserCacheTTLMinutes <= 0 || cfg.CleanupIntervalMinutes <= 0 {
		cfg = model.DefaultSystemConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateSystemConfig(c *gin.Context) {
	var req model.SystemConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.UserCacheTTLMinutes <= 0 || req.CleanupIntervalMinutes <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "user_cache_ttl_minutes and cleanup_interval_minutes must be positive")
		return
	}
	cronDefaults := map[*string]string{
		&req.DatasetCheckCron:  "0 * * * *",
		&req.SessionCleanupCron: "0 * * * *",
		&req.GameCleanupCron:   "*/5 * * * *",
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	for field, def := range cronDefaults {
		if *field == "" {
			*field = def
		} else if _, err := parser.Parse(*field); err != nil {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation,
				fmt.Sprintf("invalid cron expression %q: %v", *field, err))
			return
		}
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, systemConfigKey, req, 0); err != nil {
		sc.Logger.Error("Failed to update system config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update system config")
		return
	}

	if sc.RestartSchedulers != nil {
		go sc.RestartSchedulers()
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_system_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"user_cache_ttl_minutes":              req.UserCacheTTLMinutes,
			"cleanup_interval_minutes":            req.CleanupIntervalMinutes,
			"offline_delay_seconds":               req.OfflineDelaySeconds,
			"game_leave_delay_seconds":            req.GameLeaveDelaySeconds,
			"analytics_active_days":               req.AnalyticsActiveDays,
			"analytics_archive_days":              req.AnalyticsArchiveDays,
			"dataset_check_enabled":  req.DatasetCheckEnabled,
			"dataset_check_cron":     req.DatasetCheckCron,
			"session_cleanup_enabled": req.SessionCleanupEnabled,
			"session_cleanup_cron":   req.SessionCleanupCron,
			"game_cleanup_enabled":   req.GameCleanupEnabled,
			"game_cleanup_cron":      req.GameCleanupCron,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetAuthConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var cfg model.AuthConfig
	if err := sc.RedisService.GetJSON(ctx, "system:auth", &cfg); err != nil ||
		cfg.AccessTokenTTLMinutes <= 0 {
		cfg = model.DefaultAuthConfig()
	}

	httputil.Success(c, http.StatusOK, cfg)
}

func (sc *AdminSettingsHandler) UpdateAuthConfig(c *gin.Context) {
	var req model.AuthConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid request body")
		return
	}

	if req.AccessTokenTTLMinutes <= 0 || req.RefreshTokenTTLDays <= 0 || req.SessionTTLDays <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "All TTL values must be positive")
		return
	}
	if req.MaxConcurrentSessions <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "max_concurrent_sessions must be positive")
		return
	}
	if req.FailedLoginAttempts <= 0 || req.LoginLockoutMinutes <= 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "failed_login_attempts and login_lockout_minutes must be positive")
		return
	}

	ctx := c.Request.Context()

	if err := sc.RedisService.SetJSON(ctx, "system:auth", req, 0); err != nil {
		sc.Logger.Error("Failed to update auth config", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to update auth config")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	go func() {
		httputil.LogAdminAction(sc.LoggingService, adminUUID, adminName, "update_auth_config", "settings", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"access_token_ttl_minutes":  req.AccessTokenTTLMinutes,
			"refresh_token_ttl_days":    req.RefreshTokenTTLDays,
			"session_ttl_days":          req.SessionTTLDays,
			"max_concurrent_sessions":   req.MaxConcurrentSessions,
			"failed_login_attempts":     req.FailedLoginAttempts,
			"login_lockout_minutes":    req.LoginLockoutMinutes,
		}, true, nil)
	}()

	httputil.Success(c, http.StatusOK, req)
}

func (sc *AdminSettingsHandler) GetVersionStatus(c *gin.Context) {
	ctx := c.Request.Context()

	var status struct {
		CurrentVersion string `json:"current_version"`
		LatestVersion  string `json:"latest_version"`
		IsUpToDate     bool   `json:"is_up_to_date"`
		CheckedAt      string `json:"checked_at"`
	}

	if err := sc.RedisService.GetJSON(ctx, "system:version:status", &status); err != nil {
		httputil.Success(c, http.StatusOK, gin.H{"is_up_to_date": true, "checked_at": nil})
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{
		"current_version": status.CurrentVersion,
		"latest_version":  status.LatestVersion,
		"is_up_to_date":   status.IsUpToDate,
		"checked_at":      status.CheckedAt,
	})
}
