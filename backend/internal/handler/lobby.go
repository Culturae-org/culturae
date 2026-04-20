// backend/internal/handler/lobby.go

package handler

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
)

type LobbyHandler struct {
	gameStats  service.GameStatsProvider
	queueStats service.QueueStatsProvider
	wsService  service.WebSocketServiceInterface
}

func NewLobbyHandler(
	gameStats service.GameStatsProvider,
	queueStats service.QueueStatsProvider,
	wsService service.WebSocketServiceInterface,
) *LobbyHandler {
	return &LobbyHandler{
		gameStats:  gameStats,
		queueStats: queueStats,
		wsService:  wsService,
	}
}

// -----------------------------------------------------
// Lobby Handlers
//
// - GetLobbyStats
// -----------------------------------------------------

func (h *LobbyHandler) GetLobbyStats(c *gin.Context) {
	activeGames := h.gameStats.GetActiveGamesByMode()

	queueStats, _ := h.queueStats.GetQueueStats()
	inQueue := make(map[string]int, len(queueStats))
	for mode, count := range queueStats {
		inQueue[mode] = int(count)
	}

	httputil.Success(c, http.StatusOK, map[string]interface{}{
		"online_users": h.wsService.GetOnlineUsers(),
		"active_games": activeGames,
		"in_queue":     inQueue,
	})
}
