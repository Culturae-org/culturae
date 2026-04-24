// backend/internal/pkg/httputil/cookie.go

package httputil

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetAuthCookie(c *gin.Context, token string, maxAge time.Duration, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "culturae_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
		Secure:   secure,
	})
}

func SetRefreshCookie(c *gin.Context, token string, maxAge time.Duration, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "culturae_refresh",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
		Secure:   secure,
	})
}

func ClearAuthCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "culturae_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Secure:   secure,
	})
}

func ClearRefreshCookie(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "culturae_refresh",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Secure:   secure,
	})
}
