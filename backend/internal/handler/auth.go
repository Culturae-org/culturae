// backend/internal/handler/auth.go

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/config"
	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/crypto"
	"github.com/Culturae-org/culturae/internal/pkg/env"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/validation"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/token"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AuthHandler struct {
	Usecase        *usecase.UserUsecase
	JWTSecret      string
	RedisService   cache.RedisClientInterface
	SessionService service.SessionServiceInterface
	JWTService     *token.JWTService
	SessionConfig  *model.SessionConfig
	Config         *config.Config
	LoggingService service.LoggingServiceInterface
	wsService      service.WebSocketServiceInterface
	logger         *zap.Logger
}

func NewAuthHandler(
	cfg *config.Config,
	uc *usecase.UserUsecase,
	secret string,
	redisService cache.RedisClientInterface,
	sessionService service.SessionServiceInterface,
	jwtService *token.JWTService,
	sessionConfig *model.SessionConfig,
	loggingService service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		Usecase:        uc,
		JWTSecret:      secret,
		RedisService:   redisService,
		SessionService: sessionService,
		JWTService:     jwtService,
		SessionConfig:  sessionConfig,
		Config:         cfg,
		LoggingService: loggingService,
		wsService:      wsService,
		logger:         logger,
	}
}

// -----------------------------------------------------
// Authentication Handlers
//
// - Register
// - Login
// - RefreshToken
// - Logout
// -----------------------------------------------------

func (ac *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if len(req.Username) < 3 {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "username_too_short")
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Username must be at least 3 characters long.")
		return
	}

	if len(req.Username) > 20 {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "username_too_long")
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Username must not exceed 20 characters.")
		return
	}

	if !validation.IsValidUsername(req.Username) {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "invalid_username_format")
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Username can only contain letters, numbers, and hyphens.")
		return
	}

	if !validation.IsValidEmail(req.Email) {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "invalid_email_format")
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid email format.")
		return
	}

	if !crypto.IsValidPassword(req.Password) {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "password_not_strong_enough")
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Password must be at least 8 characters long and contain at least one uppercase letter, one digit, and one special character.")
		return
	}

	ctx := c.Request.Context()
	clientIP := httputil.GetRealIP(c)

	if ac.Config.Env != env.Development && ac.RedisService != nil {
		allowed, _, resetAt, err := ac.RedisService.CheckRateLimit(ctx, fmt.Sprintf("register:%s", clientIP), 3, time.Hour)
		if err != nil {
			ac.logger.Warn("Redis rate limit error", zap.Error(err))
		} else if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 0 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			httputil.ErrorWithDetails(c, http.StatusTooManyRequests, httputil.ErrCodeRateLimited, "Too many registration attempts. Please try again later.", map[string]interface{}{
				"retry_after": retryAfter,
			})
			return
		}
	}

	if ac.Usecase.CheckUserExists(req.Email, req.Username) {
		existingUser, err := ac.Usecase.GetByIdentifier(req.Email)
		userID := (*uuid.UUID)(nil)
		if err == nil {
			userID = &existingUser.ID
		}
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, userID, "user_already_exists")
		httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, "User already exists")
		return
	}

	user := model.User{
		Email:         req.Email,
		Username:      req.Username,
		Password:      req.Password,
		Role:          "user",
		AccountStatus: model.AccountStatusActive,
	}

	if err := ac.Usecase.CreateUser(&user); err != nil {
		errorMsg := err.Error()
		_ = ac.LoggingService.LogUserAction(
			uuid.Nil, "register",
			httputil.GetRealIP(c),
			httputil.GetUserAgent(c),
			map[string]string{"email": req.Email,
				"username": req.Username},
			false,
			&errorMsg,
		)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to create user")
		return
	}

	createdUser, err := ac.Usecase.GetByIdentifier(user.Email)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch created user")
		return
	}

	if ac.SessionService == nil || ac.JWTService == nil {
		httputil.Error(c, http.StatusServiceUnavailable, httputil.ErrCodeServiceUnavailable, "Session service not available")
		return
	}

	sessionVariables := datatypes.JSON(`{
		"first_login": true
	}`)

	session, err := ac.SessionService.CreateSession(
		createdUser,
		clientIP,
		httputil.GetUserAgent(c),
		sessionVariables,
	)
	if err != nil {
		ac.logger.Error("Failed to create session", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to create session")
		return
	}

	accessToken, err := ac.JWTService.GenerateAccessToken(session, ac.SessionConfig.AccessTokenDuration)
	if err != nil {
		ac.logger.Error("Failed to generate access token", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to generate access token")
		return
	}

	httputil.SetAuthCookie(c, accessToken, ac.SessionConfig.AccessTokenDuration, ac.Config.CookieSecure)
	httputil.SetRefreshCookie(c, session.RefreshToken, ac.SessionConfig.RefreshTokenDuration, ac.Config.CookieSecure)

	response := model.SessionResponse{
		Created:        session.Created,
		TokenType:      "Bearer",
		Token:          accessToken,
		RefreshToken:   session.RefreshToken,
		ExpiresAt:      session.ExpiresAt.Unix(),
		TokenExpiresAt: time.Now().Add(ac.SessionConfig.AccessTokenDuration).Unix(),
		User: model.UserAuth{
			PublicID: createdUser.PublicID,
			Username: createdUser.Username,
			Email:    createdUser.Email,
			Language: createdUser.Language,
			Role:     createdUser.Role,
		},
	}

	httputil.LogSuccessfulAuthAttempt(ac.LoggingService, c, createdUser.ID, &session.ID)

	ac.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "user_registered",
		Data: map[string]interface{}{
			"username": createdUser.Username,
		},
		EntityType: "user",
		EntityID:   createdUser.PublicID,
		ActionURL:  "/users/" + createdUser.PublicID,
	})

	httputil.SuccessWithMessage(c, http.StatusCreated, "User created successfully", response)
}

func (ac *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	ctx := c.Request.Context()
	clientIP := httputil.GetRealIP(c)

	if ac.Config.Env != env.Development && ac.RedisService != nil {
		allowed, _, resetAt, err := ac.RedisService.CheckRateLimit(ctx, fmt.Sprintf("login:%s", clientIP), 10, time.Minute*15)
		if err != nil {
			ac.logger.Warn("Redis rate limit error", zap.Error(err))
		} else if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 0 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			httputil.ErrorWithDetails(c, http.StatusTooManyRequests, httputil.ErrCodeRateLimited, "Too many login attempts. Please try again later.", map[string]interface{}{
				"retry_after": retryAfter,
			})
			return
		}
	}
	user, err := ac.Usecase.Authenticate(req.Identifier, req.Password)
	if err != nil {
		{
			var e *model.ErrAccountStatus
			switch {
			case errors.As(err, &e):
				if httputil.HandleAccountStatus(c, e.Status, e.UserID, func(reason string) {
					httputil.LogFailedAuthAttempt(ac.LoggingService, c, &e.UserID, reason)
				}) {
					return
				}
				httputil.Error(c, http.StatusForbidden, httputil.ErrCodeForbidden, "Account status prevents login")
				return
			default:
				if errors.Is(err, model.ErrUserNotFound) {
					httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "user_not_found")
					httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeInvalidCredentials, "Invalid credentials")
					return
				} else if errors.Is(err, model.ErrInvalidCredentials) {
					var uid *uuid.UUID
					if user != nil {
						u := user.ID
						uid = &u
					}
					httputil.LogFailedAuthAttempt(ac.LoggingService, c, uid, "wrong_password")
					httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeInvalidCredentials, "Invalid credentials")
					return
				}
				ac.logger.Error("Authentication error", zap.Error(err))
				httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Internal server error")
				return
			}
		}
	}

	if ac.SessionService == nil || ac.JWTService == nil {
		httputil.Error(c, http.StatusServiceUnavailable, httputil.ErrCodeServiceUnavailable, "Session service not available")
		return
	}

	sessionVariables := datatypes.JSON(`{
		"login_method": "password"
	}`)

	session, err := ac.SessionService.CreateSession(
		user,
		clientIP,
		httputil.GetUserAgent(c),
		sessionVariables,
	)
	if err != nil {
		ac.logger.Error("Failed to create session", zap.Error(err))
		errorMsg := err.Error()
		_ = ac.LoggingService.LogUserAction(user.ID, "login", httputil.GetRealIP(c), httputil.GetUserAgent(c), nil, false, &errorMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to create session")
		return
	}

	accessToken, err := ac.JWTService.GenerateAccessToken(session, ac.SessionConfig.AccessTokenDuration)
	if err != nil {
		ac.logger.Error("Failed to generate access token", zap.Error(err))
		errorMsg := err.Error()
		_ = ac.LoggingService.LogUserAction(user.ID, "login", httputil.GetRealIP(c), httputil.GetUserAgent(c), nil, false, &errorMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to generate access token")
		return
	}

	httputil.SetAuthCookie(c, accessToken, ac.SessionConfig.AccessTokenDuration, ac.Config.CookieSecure)
	httputil.SetRefreshCookie(c, session.RefreshToken, ac.SessionConfig.RefreshTokenDuration, ac.Config.CookieSecure)

	response := model.SessionResponse{
		Created:        session.Created,
		TokenType:      "Bearer",
		Token:          accessToken,
		RefreshToken:   session.RefreshToken,
		ExpiresAt:      session.ExpiresAt.Unix(),
		TokenExpiresAt: time.Now().Add(ac.SessionConfig.AccessTokenDuration).Unix(),
		User: model.UserAuth{
			PublicID: user.PublicID,
			Username: user.Username,
			Email:    user.Email,
			Language: user.Language,
			Role:     user.Role,
		},
	}

	httputil.LogSuccessfulAuthAttempt(ac.LoggingService, c, user.ID, &session.ID)

	httputil.SuccessWithMessage(c, http.StatusOK, "Login successful", response)
}

func (ac *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		refreshCookie, err := c.Cookie("culturae_refresh")
		if err != nil {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "refresh_token is required in body or cookie")
			return
		}
		req.RefreshToken = refreshCookie
	}

	if ac.SessionService == nil || ac.JWTService == nil {
		httputil.Error(c, http.StatusServiceUnavailable, httputil.ErrCodeServiceUnavailable, "Session service not available")
		return
	}

	session, err := ac.SessionService.RefreshSession(
		req.RefreshToken,
		httputil.GetRealIP(c),
		httputil.GetUserAgent(c),
	)
	if err != nil {
		if errors.Is(err, service.ErrRefreshTokenReused) {
			errorMsg := err.Error()
			_ = ac.LoggingService.LogSecurityEvent(nil, "refresh_token_reuse_detected", map[string]string{keyError: errorMsg}, httputil.GetRealIP(c), httputil.GetUserAgent(c), false, &errorMsg)
			httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeInvalidToken, "Invalid refresh token")
			return
		} else if errors.Is(err, service.ErrInvalidRefreshToken) {
			httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "invalid_refresh_token")
			errorMsg := err.Error()
			_ = ac.LoggingService.LogSecurityEvent(nil, "refresh_token_failed", map[string]string{keyError: errorMsg}, httputil.GetRealIP(c), httputil.GetUserAgent(c), false, &errorMsg)
			httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeInvalidToken, "Invalid refresh token")
			return
		} else if errors.Is(err, service.ErrRefreshTokenExpired) {
			httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "refresh_token_expired")
			errorMsg := err.Error()
			_ = ac.LoggingService.LogSecurityEvent(nil, "refresh_token_failed", map[string]string{keyError: errorMsg}, httputil.GetRealIP(c), httputil.GetUserAgent(c), false, &errorMsg)
			httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeRefreshExpired, "Refresh token expired")
			return
		} else if errors.Is(err, service.ErrSessionRevoked) {
			httputil.LogFailedAuthAttempt(ac.LoggingService, c, nil, "refresh_token_revoked")
			errorMsg := err.Error()
			_ = ac.LoggingService.LogSecurityEvent(nil, "refresh_token_failed", map[string]string{keyError: errorMsg}, httputil.GetRealIP(c), httputil.GetUserAgent(c), false, &errorMsg)
			httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeSessionRevoked, "Refresh token revoked")
			return
		}

		ac.logger.Error("Failed to refresh session", zap.Error(err))
		errorMsg := err.Error()
		_ = ac.LoggingService.LogSecurityEvent(nil, "refresh_token_failed", map[string]string{keyError: errorMsg}, httputil.GetRealIP(c), httputil.GetUserAgent(c), false, &errorMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to refresh token")
		return
	}

	if httputil.HandleAccountStatus(c, session.User.AccountStatus, session.User.ID, func(reason string) {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, &session.User.ID, reason)
	}) {
		return
	}

	accessToken, err := ac.JWTService.GenerateAccessToken(session, ac.SessionConfig.AccessTokenDuration)
	if err != nil {
		errorMsg := err.Error()
		_ = ac.LoggingService.LogSecurityEvent(&session.User.ID, "token_refresh_failed", map[string]string{keyError: errorMsg}, httputil.GetRealIP(c), httputil.GetUserAgent(c), false, &errorMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to generate access token")
		return
	}

	httputil.SetAuthCookie(c, accessToken, ac.SessionConfig.AccessTokenDuration, ac.Config.CookieSecure)
	httputil.SetRefreshCookie(c, session.RefreshToken, ac.SessionConfig.RefreshTokenDuration, ac.Config.CookieSecure)

	response := model.SessionResponse{
		Created:        session.Created,
		TokenType:      "Bearer",
		Token:          accessToken,
		RefreshToken:   session.RefreshToken,
		ExpiresAt:      session.ExpiresAt.Unix(),
		TokenExpiresAt: time.Now().Add(ac.SessionConfig.AccessTokenDuration).Unix(),
		User: model.UserAuth{
			PublicID: session.User.PublicID,
			Username: session.User.Username,
			Email:    session.User.Email,
			Language: session.User.Language,
			Role:     session.User.Role,
		},
	}

	_ = ac.LoggingService.LogSecurityEvent(&session.User.ID, "token_refreshed", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), true, nil)

	httputil.SuccessWithMessage(c, http.StatusOK, "Token refreshed successfully", response)
}

func (ac *AuthHandler) Logout(c *gin.Context) {
	tokenID := c.GetString("token_id")

	if tokenID != "" && ac.SessionService != nil {
		if err := ac.SessionService.RevokeSession(tokenID); err != nil {
			ac.logger.Warn("Error revoking session", zap.Error(err))
		}
		if userID := httputil.GetUserIDFromContext(c); userID != uuid.Nil {
			_ = ac.LoggingService.LogSecurityEvent(&userID, "session_revoked", map[string]string{"token_id": tokenID}, httputil.GetRealIP(c), httputil.GetUserAgent(c), true, nil)
		}
	} else {
		if userID := httputil.GetUserIDFromContext(c); userID != uuid.Nil && ac.RedisService != nil {
			ctx := c.Request.Context()
			if err := ac.RedisService.Delete(ctx, fmt.Sprintf("user:%s", userID.String())); err != nil {
				ac.logger.Warn("Error deleting user from cache", zap.Error(err))
			}
		}
	}

		httputil.ClearAuthCookie(c, ac.Config.CookieSecure)
	httputil.ClearRefreshCookie(c, ac.Config.CookieSecure)

	httputil.SuccessWithMessage(c, http.StatusOK, "Logged out successfully", nil)
}
