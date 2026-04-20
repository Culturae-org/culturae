// backend/internal/handler/admin/auth.go

package admin

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/config"
	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/env"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/token"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AdminAuthHandler struct {
	Usecase        *usecase.UserUsecase
	JWTSecret      string
	RedisService   cache.RedisClientInterface
	SessionService service.SessionServiceInterface
	JWTService     *token.JWTService
	SessionConfig  *model.SessionConfig
	Config         *config.Config
	LoggingService service.LoggingServiceInterface
}

func NewAdminAuthHandler(
	cfg *config.Config,
	uc *usecase.UserUsecase,
	secret string,
	redisService cache.RedisClientInterface,
	sessionService service.SessionServiceInterface,
	jwtService *token.JWTService,
	sessionConfig *model.SessionConfig,
	loggingService service.LoggingServiceInterface,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		Usecase:        uc,
		JWTSecret:      secret,
		RedisService:   redisService,
		SessionService: sessionService,
		JWTService:     jwtService,
		SessionConfig:  sessionConfig,
		Config:         cfg,
		LoggingService: loggingService,
	}
}

// -----------------------------------------------------
// Admin Auth Handlers
//
// - LoginAdmin
// -----------------------------------------------------

func (ac *AdminAuthHandler) LoginAdmin(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	ctx := c.Request.Context()
	clientIP := httputil.GetRealIP(c)

	if ac.Config.Env != env.Development && ac.RedisService != nil {
		allowed, remaining, resetAt, err := ac.RedisService.CheckRateLimit(ctx, fmt.Sprintf("login:%s", clientIP), 10, time.Minute*15)
		if err != nil {
			errorMsg := fmt.Sprintf("Redis rate limit error: %v", err)
			_ = ac.LoggingService.LogAPIRequest("POST", c.Request.URL.Path, http.StatusInternalServerError, nil, clientIP, httputil.GetUserAgent(c), 0, 0, 0, true, &errorMsg)
		} else if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 0 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			_ = remaining
			httputil.Error(c, http.StatusTooManyRequests, httputil.ErrCodeRateLimited, fmt.Sprintf("Too many login attempts. Please try again later. Retry after: %d", retryAfter))
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
				_ = ac.LoggingService.LogAPIRequest("POST", c.Request.URL.Path, http.StatusInternalServerError, nil, clientIP, httputil.GetUserAgent(c), 0, 0, 0, true, nil)
				httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Internal server error")
				return
			}
		}
	}

	if user.Role != model.RoleAdministrator {
		httputil.LogFailedAuthAttempt(ac.LoggingService, c, &user.ID, "access_denied_not_admin")
		httputil.Error(c, http.StatusForbidden, httputil.ErrCodeForbidden, "Access denied: administrator only")
		return
	}

	if ac.SessionService == nil || ac.JWTService == nil {
		httputil.Error(c, http.StatusServiceUnavailable, httputil.ErrCodeInternal, "Session service not available")
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
		errorMsg := fmt.Sprintf("Failed to create session: %v", err)
		_ = ac.LoggingService.LogAPIRequest("POST", c.Request.URL.Path, http.StatusInternalServerError, &user.ID, clientIP, httputil.GetUserAgent(c), 0, 0, 0, true, &errorMsg)
		_ = ac.LoggingService.LogUserAction(user.ID, "admin_login", clientIP, httputil.GetUserAgent(c), nil, false, &errorMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to create session")
		return
	}

	accessToken, err := ac.JWTService.GenerateAccessToken(session, ac.SessionConfig.AccessTokenDuration)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to generate access token: %v", err)
		_ = ac.LoggingService.LogAPIRequest("POST", c.Request.URL.Path, http.StatusInternalServerError, &user.ID, clientIP, httputil.GetUserAgent(c), 0, 0, 0, true, &errorMsg)
		_ = ac.LoggingService.LogUserAction(user.ID, "admin_login", clientIP, httputil.GetUserAgent(c), nil, false, &errorMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to generate access token")
		return
	}

	httputil.SetAuthCookie(c, accessToken, ac.SessionConfig.AccessTokenDuration, ac.Config.CookieSecure)
	httputil.SetRefreshCookie(c, session.RefreshToken, ac.SessionConfig.RefreshTokenDuration)

	response := model.SessionResponse{
		Created:        session.Created,
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
