// backend/internal/handler/admin/matchmaking.go

package admin

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/game"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"

	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminMatchmakingHandler struct {
	Service        *game.MatchmakingService
	LoggingService service.LoggingServiceInterface
	logger         *zap.Logger
}

func NewAdminMatchmakingHandler(
	service *game.MatchmakingService,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AdminMatchmakingHandler {
	return &AdminMatchmakingHandler{
		Service:        service,
		LoggingService: loggingService,
		logger:         logger,
	}
}

// -----------------------------------------------------
// Admin Matchmaking Handlers
//
// - GetQueueStats
// - ClearQueue
// -----------------------------------------------------

func (mc *AdminMatchmakingHandler) GetQueueStats(c *gin.Context) {
	stats, err := mc.Service.GetQueueStats()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (mc *AdminMatchmakingHandler) ClearQueue(c *gin.Context) {
	modeStr := c.Param("mode")
	mode := model.GameMode(modeStr)

	if mode != model.GameMode1v1 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid or unsupported game mode")
		return
	}

	if err := mc.Service.ClearQueue(mode); err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		httputil.LogAdminAction(
			mc.LoggingService,
			adminID,
			c.GetString("username"),
			"clear_queue",
			"matchmaking",
			nil,
			httputil.GetRealIP(c),
			httputil.GetUserAgent(c),
			map[string]string{"mode": string(mode)},
			true,
			nil,
		)
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Queue cleared successfully", mode)
}
