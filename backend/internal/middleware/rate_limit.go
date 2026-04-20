// backend/internal/middleware/rate_limit.go

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/config"
	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	redisClient cache.RedisClientInterface
	config      *config.Config
}

func NewRateLimitMiddleware(
	redisClient cache.RedisClientInterface,
	cfg *config.Config,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
		config:      cfg,
	}
}

func (rl *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if path == "/health" {
			c.Next()
			return
		}

		if strings.HasSuffix(path, "/realtime") {
			c.Next()
			return
		}

		isAdmin := strings.Contains(path, "/admin/")

		enabled, applyToAdmin, limit, window := rl.getRuntimeConfig(c.Request.Context())
		if !enabled {
			c.Next()
			return
		}

		if isAdmin && !applyToAdmin {
			c.Next()
			return
		}

		keyPrefix := "ratelimit"
		if isAdmin {
			keyPrefix = "ratelimit:admin"
		}

		ip := rl.getClientIP(c)
		key := fmt.Sprintf("%s:%s", keyPrefix, ip)

		allowed, remaining, resetAt := rl.checkRateLimitWithParams(c.Request.Context(), key, limit, window)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt))

		if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 0 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			httputil.AbortWithErrorDetails(c, http.StatusTooManyRequests, httputil.ErrCodeRateLimited, "Too many requests", map[string]interface{}{
				"retry_after": retryAfter,
			})
			return
		}

		c.Next()
	}
}

func (rl *RateLimitMiddleware) AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.RateLimitEnabled {
			c.Next()
			return
		}

		ip := rl.getClientIP(c)
		key := fmt.Sprintf("ratelimit:auth:%s", ip)

		limit := 50
		window := time.Minute

		allowed, remaining, resetAt := rl.checkRateLimitWithParams(c.Request.Context(), key, limit, window)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt))

		if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 0 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			httputil.AbortWithErrorDetails(c, http.StatusTooManyRequests, httputil.ErrCodeRateLimited, "Too many requests. Please try again later.", map[string]interface{}{
				"retry_after": retryAfter,
			})
			return
		}

		c.Next()
	}
}

func (rl *RateLimitMiddleware) getClientIP(c *gin.Context) string {
	if cfip := c.GetHeader("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	return c.ClientIP()
}

func (rl *RateLimitMiddleware) checkRateLimitWithParams(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt int64) {
	if rl.redisClient == nil {
		resetAt = time.Now().Add(window).Unix()
		rem := limit - 1
		if rem < 0 {
			rem = 0
		}
		return true, rem, resetAt
	}

	allowed64, remaining64, resetAt64, err := rl.redisClient.CheckRateLimit(ctx, key, int64(limit), window)
	if err != nil {
		resetAt = time.Now().Add(window).Unix()
		rem := limit - 1
		if rem < 0 {
			rem = 0
		}
		return true, rem, resetAt
	}

	resetAt = resetAt64
	remaining = int(remaining64)
	return allowed64, remaining, resetAt
}

func (rl *RateLimitMiddleware) getRuntimeConfig(ctx context.Context) (enabled bool, applyToAdmin bool, limit int, window time.Duration) {
	enabled = rl.config.RateLimitEnabled
	limit = rl.config.RateLimitRequests
	window = rl.config.RateLimitWindow

	if rl.redisClient == nil {
		return
	}

	data, err := rl.redisClient.Get(ctx, "system:ratelimit:config")
	if err != nil || data == "" {
		return
	}

	var cfg struct {
		Enabled       bool `json:"enabled"`
		ApplyToAdmin  bool `json:"apply_to_admin"`
		MaxRequests   int  `json:"max_requests"`
		WindowSeconds int  `json:"window_seconds"`
	}
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return
	}

	enabled = cfg.Enabled
	applyToAdmin = cfg.ApplyToAdmin
	if cfg.MaxRequests > 0 {
		limit = cfg.MaxRequests
	}
	if cfg.WindowSeconds > 0 {
		window = time.Duration(cfg.WindowSeconds) * time.Second
	}
	return
}
