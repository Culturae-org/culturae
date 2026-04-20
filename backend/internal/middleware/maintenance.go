// backend/internal/middleware/maintenance.go

package middleware

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"

	"github.com/gin-gonic/gin"
)

const maintenanceRedisKey = "system:maintenance:enabled"

type MaintenanceMiddleware struct {
	redisClient cache.RedisClientInterface
}

func NewMaintenanceMiddleware(
	redisClient cache.RedisClientInterface,
) *MaintenanceMiddleware {
	return &MaintenanceMiddleware{
		redisClient: redisClient,
	}
}

func (m *MaintenanceMiddleware) CheckMaintenance() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		enabled, err := m.redisClient.Exists(ctx, maintenanceRedisKey)
		if err != nil || !enabled {
			c.Next()
			return
		}

		if user, exists := c.Get("user"); exists {
			if userModel, ok := user.(model.User); ok && userModel.Role == "administrator" {
				c.Next()
				return
			}
		}

		httputil.AbortWithError(c, http.StatusServiceUnavailable, httputil.ErrCodeMaintenance, "The service is temporarily unavailable for maintenance. Please try again later.")
	}
}
