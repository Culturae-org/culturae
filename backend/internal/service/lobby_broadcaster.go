// backend/internal/service/lobby_broadcaster.go

package service

import (
	"time"

	"go.uber.org/zap"
)

type GameStatsProvider interface {
	GetActiveGamesByMode() map[string]int
}

type QueueStatsProvider interface {
	GetQueueStats() (map[string]int64, error)
}

type LobbyBroadcaster struct {
	gameStats   GameStatsProvider
	queueStats  QueueStatsProvider
	wsService   WebSocketServiceInterface
	logger      *zap.Logger
	stop        chan struct{}
}

func NewLobbyBroadcaster(
	gameStats GameStatsProvider,
	queueStats QueueStatsProvider,
	wsService WebSocketServiceInterface,
	logger *zap.Logger,
) *LobbyBroadcaster {
	return &LobbyBroadcaster{
		gameStats:  gameStats,
		queueStats: queueStats,
		wsService:  wsService,
		logger:     logger,
		stop:       make(chan struct{}),
	}
}

func (lb *LobbyBroadcaster) Start() {
	go lb.run()
}

func (lb *LobbyBroadcaster) Stop() {
	close(lb.stop)
}

func (lb *LobbyBroadcaster) Broadcast() {
	stats := lb.buildStats()
	msg := map[string]interface{}{
		keyType: "lobby_stats",
		keyData: stats,
	}
	if err := lb.wsService.BroadcastToAll(msg); err != nil {
		lb.logger.Warn("Failed to broadcast lobby_stats", zap.Error(err))
	}
}

func (lb *LobbyBroadcaster) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-lb.stop:
			return
		case <-ticker.C:
			lb.Broadcast()
		}
	}
}

type lobbyStatsPayload struct {
	OnlineUsers int            `json:"online_users"`
	ActiveGames map[string]int `json:"active_games"`
	InQueue     map[string]int `json:"in_queue"`
}

func (lb *LobbyBroadcaster) buildStats() lobbyStatsPayload {
	activeGames := lb.gameStats.GetActiveGamesByMode()

	queueStats, _ := lb.queueStats.GetQueueStats()
	inQueue := make(map[string]int, len(queueStats))
	for mode, count := range queueStats {
		inQueue[mode] = int(count)
	}

	return lobbyStatsPayload{
		OnlineUsers: lb.wsService.GetOnlineUsers(),
		ActiveGames: activeGames,
		InQueue:     inQueue,
	}
}
