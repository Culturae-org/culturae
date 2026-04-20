// backend/internal/handler/websocket.go

package handler

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	wsService      service.WebSocketServiceInterface
	logger         *zap.Logger
	LoggingService service.LoggingServiceInterface
}

func NewWebSocketHandler(
	wsService service.WebSocketServiceInterface,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		wsService:      wsService,
		logger:         logger,
		LoggingService: loggingService,
	}
}

// -----------------------------------------------------
// WebSocket Handlers
//
// - HandleWebSocket
// -----------------------------------------------------

func (wsc *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User Not Authenticated")
		return
	}

	err := wsc.wsService.UpgradeConnection(c, userID)
	if err != nil {
		wsc.logger.Error("Error upgrading WebSocket", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to establish WebSocket connection")
		return
	}
}

func (wsc *WebSocketHandler) HandleAdminWebSocket(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "User Not Authenticated")
		return
	}

	err := wsc.wsService.UpgradeAdminConnection(c, userID)
	if err != nil {
		wsc.logger.Error("Error upgrading admin WebSocket", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to establish WebSocket connection")
		return
	}
}
