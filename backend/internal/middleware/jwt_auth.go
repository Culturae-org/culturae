// backend/internal/middleware/jwt_auth.go

package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/token"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService     *token.JWTService
	sessionService service.SessionServiceInterface
	jwtSecret      string
}

func NewAuthMiddleware(
	jwtService *token.JWTService,
	sessionService service.SessionServiceInterface,
	jwtSecret string,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:     jwtService,
		sessionService: sessionService,
		jwtSecret:      jwtSecret,
	}
}

func (am *AuthMiddleware) JWTAuthWithSessions() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			var err error
			tokenStr, err = c.Cookie("culturae_token")
			if err != nil {
				proto := c.GetHeader("Sec-WebSocket-Protocol")
				if proto != "" {
					for _, p := range strings.Split(proto, ",") {
						p = strings.TrimSpace(p)
						if strings.HasPrefix(p, "Bearer ") {
							tokenStr = strings.TrimPrefix(p, "Bearer ")
							break
						}
						if strings.HasPrefix(p, "access_token:") {
							tokenStr = strings.TrimPrefix(p, "access_token:")
							break
						}

						if p != "" && p != "Bearer" {
							tokenStr = p
							break
						}
					}
				}

				if tokenStr == "" {
					httputil.AbortWithError(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "missing token")
					return
				}
			}
		}

		if tokenStr == "" {
			httputil.AbortWithError(c, http.StatusUnauthorized, httputil.ErrCodeEmptyToken, "empty token")
			return
		}

		if am.jwtService == nil || am.sessionService == nil {
			httputil.AbortWithError(c, http.StatusInternalServerError, httputil.ErrCodeServiceUnavailable, "authentication services unavailable")
			return
		}

		claims, err := am.jwtService.ValidateToken(tokenStr)
		if err != nil {
			httputil.AbortWithError(c, http.StatusUnauthorized, httputil.ErrCodeInvalidToken, "invalid token")
			return
		}

		session, err := am.sessionService.GetSession(claims.ID)
		if err != nil {
			httputil.AbortWithError(c, http.StatusUnauthorized, httputil.ErrCodeInvalidToken, "invalid session")
			return
		}

		if session.User.AccountStatus == model.AccountStatusBanned && session.User.BannedUntil != nil && time.Now().After(*session.User.BannedUntil) {
			session.User.AccountStatus = model.AccountStatusActive
			session.User.BannedUntil = nil
			session.User.BanReason = ""
		}

		if session.User.AccountStatus == model.AccountStatusDeleted {
			_ = am.sessionService.RevokeAllUserSessions(session.UserID)
			httputil.AbortWithError(c, http.StatusForbidden, httputil.ErrCodeAccountDeleted, "This account has been deleted")
			return
		}

		if session.User.AccountStatus != model.AccountStatusActive {
			if err := am.sessionService.RevokeAllUserSessions(session.UserID); err != nil {
				log.Printf("Failed to revoke sessions for user %s: %v", session.UserID.String(), err)
			}
			if session.User.AccountStatus == model.AccountStatusBanned && session.User.BannedUntil != nil {
				httputil.AbortWithErrorDetails(c, http.StatusForbidden, httputil.ErrCodeAccountBanned, "Your account has been banned", map[string]interface{}{
					"banned_until": session.User.BannedUntil,
					"ban_reason":   session.User.BanReason,
				})
			} else {
				httputil.AbortWithError(c, http.StatusForbidden, httputil.ErrCodeForbidden, "account status prevents access")
			}
			return
		}

		c.Set("user_id", session.UserID)
		c.Set("user", session.User)
		c.Set("session", session)
		c.Set("token_id", claims.ID)
		c.Set("username", claims.USN)
		c.Set("public_id", session.User.PublicID)
		c.Next()
	}
}

func (am *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			httputil.AbortWithError(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "user not found in context")
			return
		}

		userModel, ok := user.(model.User)
		if !ok {
			httputil.AbortWithError(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "invalid user type")
			return
		}

		if userModel.Role != "administrator" {
			httputil.AbortWithError(c, http.StatusForbidden, httputil.ErrCodeForbidden, "admin access required")
			return
		}

		c.Next()
	}
}
