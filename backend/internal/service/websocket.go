// backend/internal/service/websocket.go

package service

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WebSocketServiceInterface interface {
	UpgradeConnection(c *gin.Context, userID uuid.UUID) error
	UpgradeAdminConnection(c *gin.Context, userID uuid.UUID) error
	GetConnectedClients() int
	GetOnlineUsers() int
	SendToUser(userID uuid.UUID, message interface{}) error
	SendToGame(gamePublicID string, message interface{}) error
	BroadcastToGame(gamePublicID string, message interface{}, excludeUserID *uuid.UUID) error
	BroadcastToAdmins(message interface{}) error
	IsUserOnline(userID uuid.UUID) bool
	SetGameActionHandler(handler GameActionHandler)
	BroadcastAdminNotification(notif AdminNotification)
	BroadcastToAll(message interface{}) error
	StartRelay(ctx context.Context)
	StopRelay()
	IsMultiPod() bool
	PodID() (string, error)
}

type UserStatusUpdater interface {
	UpdateUserOnlineStatus(userID uuid.UUID, isOnline bool) error
}

type GameActionHandler interface {
	HandleSubmitAnswer(userID uuid.UUID, gamePublicID string, answer map[string]interface{}) error
	HandlePlayerReady(userID uuid.UUID, gamePublicID string, ready bool) error
	HandleStartGame(userID uuid.UUID, gamePublicID string) error
	HandleLeaveGame(userID uuid.UUID, gamePublicID string) error
	GetCurrentQuestionPayload(gamePublicID string) (map[string]interface{}, error)
	GetGameStateForReconnect(gamePublicID string) (map[string]interface{}, error)
	MarkPlayerDisconnected(userID uuid.UUID, gamePublicID string) error
	MarkPlayerReconnected(userID uuid.UUID, gamePublicID string) error
	OnUserConnected(userID uuid.UUID)
}

type WSClient struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	PublicID string
	Conn     *websocket.Conn
	IsClosed atomic.Bool
	SendChan chan []byte
	cancel   context.CancelFunc
	GameID   *string
	IsAdmin  bool

	reconnectGrace time.Duration

	msgRateLimit  int
	msgRateWindow time.Duration
	rateMu        sync.Mutex
	msgCount      int
	msgWindowEnd  time.Time
}

const wsConfigKey = "system:websocket:config"

type WebSocketService struct {
	clients              map[uuid.UUID]*WSClient
	userClients          map[uuid.UUID][]*WSClient
	gameClients          map[string][]*WSClient
	mutex                sync.RWMutex
	userStatusUpdater    UserStatusUpdater
	gameActionHandler    GameActionHandler
	redisService         cache.RedisClientInterface
	logger               *zap.Logger
	offlineTimers        map[uuid.UUID]*time.Timer
	gameDisconnectTimers map[uuid.UUID]*time.Timer
	relay                *PubSubRelay
	multiPod             bool
}

func NewWebSocketService(
	userStatusUpdater UserStatusUpdater,
	redisService cache.RedisClientInterface,
	logger *zap.Logger,
) *WebSocketService {
	return NewWebSocketServiceWithMode(userStatusUpdater, redisService, logger, redisService != nil)
}

func NewWebSocketServiceWithMode(
	userStatusUpdater UserStatusUpdater,
	redisService cache.RedisClientInterface,
	logger *zap.Logger,
	multiPod bool,
) *WebSocketService {
	var relay *PubSubRelay
	if multiPod && redisService != nil {
		relay = NewPubSubRelay(uuid.New().String(), redisService, logger)
	}
	return &WebSocketService{
		clients:              make(map[uuid.UUID]*WSClient),
		userClients:          make(map[uuid.UUID][]*WSClient),
		gameClients:          make(map[string][]*WSClient),
		offlineTimers:        make(map[uuid.UUID]*time.Timer),
		gameDisconnectTimers: make(map[uuid.UUID]*time.Timer),
		userStatusUpdater:    userStatusUpdater,
		redisService:         redisService,
		logger:               logger,
		relay:                relay,
		multiPod:             multiPod,
	}
}

func (ws *WebSocketService) PodID() (string, error) {
	if ws.relay != nil {
		return ws.relay.PodID(), nil
	}
	return "", nil
}

func (ws *WebSocketService) StartRelay(ctx context.Context) {
	if ws.relay == nil {
		return
	}
	ws.relay.Start(ctx, ws.handleRelayMessage)
}

func (ws *WebSocketService) StopRelay() {
	if ws.relay == nil {
		return
	}
	ws.relay.Stop()
}

func (ws *WebSocketService) IsMultiPod() bool {
	return ws.multiPod
}

func (ws *WebSocketService) loadWSConfig(ctx context.Context) model.WebSocketConfig {
	var cfg model.WebSocketConfig
	if ws.redisService == nil {
		return model.DefaultWebSocketConfig()
	}
	if err := ws.redisService.GetJSON(ctx, wsConfigKey, &cfg); err != nil {
		return model.DefaultWebSocketConfig()
	}
	if cfg.WriteWaitSeconds <= 0 || cfg.PongWaitSeconds <= 0 || cfg.MaxMessageSizeKB <= 0 {
		return model.DefaultWebSocketConfig()
	}
	return cfg
}

func (ws *WebSocketService) SetGameActionHandler(handler GameActionHandler) {
	ws.gameActionHandler = handler
}

func (ws *WebSocketService) IsUserOnline(userID uuid.UUID) bool {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	clients, exists := ws.userClients[userID]
	return exists && len(clients) > 0
}

func (ws *WebSocketService) UpgradeConnection(c *gin.Context, userID uuid.UUID) error {
	return ws.upgradeConnectionWithRole(c, userID, false)
}

func (ws *WebSocketService) UpgradeAdminConnection(c *gin.Context, userID uuid.UUID) error {
	return ws.upgradeConnectionWithRole(c, userID, true)
}

func (ws *WebSocketService) upgradeConnectionWithRole(c *gin.Context, userID uuid.UUID, isAdmin bool) error {
	wsCfg := ws.loadWSConfig(c.Request.Context())

	opts := &websocket.AcceptOptions{
		Subprotocols: []string{"Bearer"},
	}
	if len(wsCfg.AllowedOrigins) == 0 {
		opts.InsecureSkipVerify = true
	} else {
		opts.OriginPatterns = wsCfg.AllowedOrigins
	}

	conn, err := websocket.Accept(c.Writer, c.Request, opts)
	if err != nil {
		return err
	}

	reconnectGrace := time.Duration(wsCfg.ReconnectGracePeriodSeconds) * time.Second
	if reconnectGrace <= 0 {
		reconnectGrace = 180 * time.Second
	}
	msgRateWindow := time.Duration(wsCfg.MessageRateWindowSeconds) * time.Second
	if msgRateWindow <= 0 {
		msgRateWindow = 60 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec // cancel is stored in WSClient and called from disconnectClient

	clientID := uuid.New()
	client := &WSClient{
		ID:             clientID,
		UserID:         userID,
		PublicID:       ginContextString(c, "public_id"),
		Conn:           conn,
		SendChan:       make(chan []byte, 256),
		cancel:         cancel,
		IsAdmin:        isAdmin,
		reconnectGrace: reconnectGrace,
		msgRateLimit:   wsCfg.MessageRateLimit,
		msgRateWindow:  msgRateWindow,
	}

	ws.mutex.Lock()
	isFirstConnection := len(ws.userClients[userID]) == 0
	ws.clients[clientID] = client
	ws.userClients[userID] = append(ws.userClients[userID], client)
	if t, ok := ws.offlineTimers[userID]; ok {
		if t.Stop() {
			ws.logger.Info("Cancelled pending offline timer", zap.String("user_id", userID.String()))
		}
		delete(ws.offlineTimers, userID)
	}
	count := len(ws.userClients[userID])
	ws.mutex.Unlock()

	if isFirstConnection {
		ws.incrOnlineUser(userID)
	}

	ws.logger.Info("WebSocket new client added", zap.String("user_id", userID.String()), zap.Int("user_connections", count))

	hello := HelloMessage{
		Type:              "hello",
		ProtocolVersion:   1,
		ServerTime:        time.Now().Format(time.RFC3339),
		HeartbeatInterval: 30000,
	}
	if helloData, err := json.Marshal(hello); err == nil {
		ws.trySend(client, helloData)
	}

	if err := ws.updateUserStatus(userID, true); err != nil {
		ws.logger.Error("Failed to update user status", zap.Error(err))
	}

	if !isAdmin && ws.gameActionHandler != nil {
		go ws.gameActionHandler.OnUserConnected(userID)
	}

	go ws.handleConnection(client, ctx, wsCfg)
	go ws.writePump(client, ctx, wsCfg)

	ws.logger.Info("WebSocket client connected", zap.String("clientID", clientID.String()), zap.String("userID", userID.String()))
	return nil
}

func (ws *WebSocketService) handleConnection(client *WSClient, ctx context.Context, cfg model.WebSocketConfig) {
	defer func() {
		ws.disconnectClient(client)
	}()

	client.Conn.SetReadLimit(cfg.MaxMessageSizeKB * 1024)

	for {
		_, message, err := client.Conn.Read(ctx)
		if err != nil {
			closeStatus := websocket.CloseStatus(err)
			if closeStatus != websocket.StatusNormalClosure &&
				closeStatus != websocket.StatusGoingAway &&
				ctx.Err() == nil {
				ws.logger.Warn("WebSocket unexpected close", zap.String("client_id", client.ID.String()), zap.Error(err))
			}
			break
		}
		ws.handleMessage(client, message)
	}
}

func (ws *WebSocketService) writePump(client *WSClient, ctx context.Context, cfg model.WebSocketConfig) {
	writeWait := time.Duration(cfg.WriteWaitSeconds) * time.Second
	pongWait := time.Duration(cfg.PongWaitSeconds) * time.Second
	pingPeriod := (pongWait * 9) / 10
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		ws.disconnectClient(client)
	}()

	for {
		select {
		case message := <-client.SendChan:
			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := client.Conn.Write(writeCtx, websocket.MessageText, message)
			cancel()
			if err != nil {
				ws.logger.Error("Error writing message", zap.Error(err))
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := client.Conn.Ping(pingCtx)
			cancel()
			if err != nil {
				ws.logger.Debug("Ping failed, closing connection", zap.String("client_id", client.ID.String()), zap.Error(err))
				return
			}

		case <-ctx.Done():
			_ = client.Conn.Close(websocket.StatusNormalClosure, "")
			return
		}
	}
}

func (ws *WebSocketService) handleMessage(client *WSClient, message []byte) {
	if client.msgRateLimit > 0 {
		client.rateMu.Lock()
		now := time.Now()
		if now.After(client.msgWindowEnd) {
			client.msgCount = 0
			client.msgWindowEnd = now.Add(client.msgRateWindow)
		}
		client.msgCount++
		limited := client.msgCount > client.msgRateLimit
		client.rateMu.Unlock()

		if limited {
			ws.logger.Warn("Client rate limited", zap.String("client_id", client.ID.String()))
			errMsg := "rate limit exceeded"
			data, _ := json.Marshal(AckMessage{
				Type:    "ack",
				Action:  "message",
				Success: false,
				Error:   &errMsg,
			})
			ws.trySend(client, data)
			return
		}
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		ws.logger.Warn("Invalid message format", zap.Error(err))
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		ws.logger.Warn("Message missing type field")
		return
	}

	correlationID := ws.extractCorrelationID(msg)

	switch msgType {
	case "join_game":
		var joinMsg JoinGameMessage
		if err := ws.unmarshalMessage(message, &joinMsg); err != nil {
			ws.sendAck(client, "join_game", false, "invalid message format", correlationID)
			return
		}
		ws.addClientToGame(client, joinMsg.GamePublicID)
		ws.sendAck(client, "join_game", true, "", correlationID)

		if ws.gameActionHandler != nil {
			if payload, err := ws.gameActionHandler.GetCurrentQuestionPayload(joinMsg.GamePublicID); err == nil && payload != nil {
				qNum := 0
				switch v := payload["question_number"].(type) {
				case float64:
					qNum = int(v)
				case int:
					qNum = v
				}
				ws.logger.Debug("Sending current question to joining player",
					zap.String("game_public_id", joinMsg.GamePublicID),
					zap.String("user_id", client.UserID.String()),
					zap.Int("question_number", qNum),
				)
				data, _ := json.Marshal(map[string]interface{}{
					"type":      "question_sent",
					"public_id": joinMsg.GamePublicID,
					"data":      payload,
					"timestamp": time.Now(),
				})
				ws.trySend(client, data)
			} else if err != nil {
				ws.logger.Warn("GetCurrentQuestionPayload failed for joining player",
					zap.String("game_public_id", joinMsg.GamePublicID),
					zap.String("user_id", client.UserID.String()),
					zap.Error(err),
				)
			}

			if statePayload, err := ws.gameActionHandler.GetGameStateForReconnect(joinMsg.GamePublicID); err == nil && statePayload != nil {
				ws.logger.Debug("Sending game state to joining player",
					zap.String("game_public_id", joinMsg.GamePublicID),
					zap.String("user_id", client.UserID.String()),
				)
				data, _ := json.Marshal(map[string]interface{}{
					"type":      "game_state",
					"public_id": joinMsg.GamePublicID,
					"data":      statePayload,
					"timestamp": time.Now(),
				})
				ws.trySend(client, data)
			}
		}

	case "leave_game":
		var leaveMsg LeaveGameMessage
		if err := ws.unmarshalMessage(message, &leaveMsg); err != nil {
			ws.sendAck(client, "leave_game", false, "invalid message format", correlationID)
			return
		}

		if leaveMsg.GamePublicID != nil {
			ws.removeClientFromGame(client, *leaveMsg.GamePublicID)
			if ws.gameActionHandler != nil {
				if err := ws.gameActionHandler.HandleLeaveGame(client.UserID, *leaveMsg.GamePublicID); err != nil {
					ws.sendAck(client, "leave_game", false, err.Error(), correlationID)
				} else {
					ws.sendAck(client, "leave_game", true, "", correlationID)
				}
			} else {
				ws.sendAck(client, "leave_game", true, "", correlationID)
			}
		} else {
			ws.mutex.RLock()
			gameIDPtr := client.GameID
			var gameIDCopy string
			if gameIDPtr != nil {
				gameIDCopy = *gameIDPtr
			}
			ws.mutex.RUnlock()

			if gameIDPtr != nil {
				ws.removeClientFromGame(client, gameIDCopy)
				if ws.gameActionHandler != nil {
					if err := ws.gameActionHandler.HandleLeaveGame(client.UserID, gameIDCopy); err != nil {
						ws.sendAck(client, "leave_game", false, err.Error(), correlationID)
					} else {
						ws.sendAck(client, "leave_game", true, "", correlationID)
					}
				} else {
					ws.sendAck(client, "leave_game", true, "", correlationID)
				}
			} else {
				ws.sendAck(client, "leave_game", false, "not in a game", correlationID)
			}
		}

	case "submit_answer":
		ws.handleSubmitAnswer(client, message, correlationID)

	case "player_ready":
		var readyMsg PlayerReadyMessage
		if err := ws.unmarshalMessage(message, &readyMsg); err != nil {
			ws.sendAck(client, "player_ready", false, "invalid message format", correlationID)
			return
		}
		ws.handleGameAction(client, msg, "player_ready", correlationID, func(gamePublicID string) error {
			return ws.gameActionHandler.HandlePlayerReady(client.UserID, gamePublicID, readyMsg.Ready)
		})

	case "start_game":
		ws.handleGameAction(client, msg, "start_game", correlationID, func(gamePublicID string) error {
			return ws.gameActionHandler.HandleStartGame(client.UserID, gamePublicID)
		})

	case "ping":
		ws.sendPong(client)

	default:
		ws.logger.Debug("Unknown message type", zap.String("type", msgType))
		ws.sendAck(client, msgType, false, "unknown message type", correlationID)
	}
}

func (ws *WebSocketService) extractCorrelationID(msg map[string]interface{}) *string {
	if correlationID, ok := msg["correlation_id"].(string); ok && correlationID != "" {
		return &correlationID
	}
	generatedID := uuid.New().String()
	return &generatedID
}

func (ws *WebSocketService) unmarshalMessage(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (ws *WebSocketService) handleSubmitAnswer(client *WSClient, message []byte, correlationID *string) {
	if ws.gameActionHandler == nil {
		ws.sendAck(client, "submit_answer", false, "game actions not available", correlationID)
		return
	}

	var submitMsg SubmitAnswerMessage
	if err := ws.unmarshalMessage(message, &submitMsg); err != nil {
		ws.sendAck(client, "submit_answer", false, "invalid message format", correlationID)
		return
	}

	if submitMsg.GamePublicID == "" {
		ws.sendAck(client, "submit_answer", false, "game_public_id required", correlationID)
		return
	}

	if err := ws.gameActionHandler.HandleSubmitAnswer(client.UserID, submitMsg.GamePublicID, map[string]interface{}{
		"type":        submitMsg.Answer.Type,
		"answer":      submitMsg.Answer.Answer,
		"question_id": submitMsg.Answer.QuestionID,
		"time_spent":  submitMsg.Answer.TimeSpent,
	}); err != nil {
		ws.logger.Warn("Submit answer failed",
			zap.String("action", "submit_answer"),
			zap.String("game_public_id", submitMsg.GamePublicID),
			zap.Error(err),
		)
		ws.sendAck(client, "submit_answer", false, err.Error(), correlationID)
	} else {
		ws.sendAck(client, "submit_answer", true, "", correlationID)
	}
}

func (ws *WebSocketService) handleGameAction(client *WSClient, msg map[string]interface{}, actionType string, correlationID *string, action func(gamePublicID string) error) {
	if ws.gameActionHandler == nil {
		ws.sendAck(client, actionType, false, "game actions not available", correlationID)
		return
	}

	gamePublicID, ok := msg["game_public_id"].(string)
	if !ok {
		ws.mutex.RLock()
		gameIDPtr := client.GameID
		if gameIDPtr != nil {
			gamePublicID = *gameIDPtr
		}
		ws.mutex.RUnlock()
		if gameIDPtr == nil {
			ws.sendAck(client, actionType, false, "game_public_id required", correlationID)
			return
		}
	}

	if err := action(gamePublicID); err != nil {
		ws.logger.Warn("Game action failed",
			zap.String("action", actionType),
			zap.String("game_public_id", gamePublicID),
			zap.Error(err),
		)
		ws.sendAck(client, actionType, false, err.Error(), correlationID)
	} else {
		ws.sendAck(client, actionType, true, "", correlationID)
	}
}

func (ws *WebSocketService) trySend(client *WSClient, data []byte) {
	if client.IsClosed.Load() {
		return
	}
	select {
	case client.SendChan <- data:
	default:
		ws.logger.Warn("Client send channel full, dropping message", zap.String("client_id", client.ID.String()))
	}
}

func (ws *WebSocketService) sendAck(client *WSClient, action string, success bool, errMsg string, correlationID *string) {
	ack := AckMessage{
		Type:          "ack",
		Action:        action,
		Success:       success,
		CorrelationID: correlationID,
	}
	if errMsg != "" {
		ack.Error = &errMsg
	}
	data, _ := json.Marshal(ack)
	ws.trySend(client, data)
}

func (ws *WebSocketService) sendPong(client *WSClient) {
	data, _ := json.Marshal(PongMessage{Type: "pong"})
	ws.trySend(client, data)
}

func (ws *WebSocketService) disconnectClient(client *WSClient) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !client.IsClosed.CompareAndSwap(false, true) {
		return
	}

	client.cancel()

	delete(ws.clients, client.ID)

	if clients, exists := ws.userClients[client.UserID]; exists {
		for i, c := range clients {
			if c.ID == client.ID {
				ws.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		if len(ws.userClients[client.UserID]) == 0 {
			delete(ws.userClients, client.UserID)

			ws.decrOnlineUser(client.UserID)

			const offlineDelay = 2 * time.Second
			userID := client.UserID
			timer := time.AfterFunc(offlineDelay, func() {
				ws.mutex.Lock()
				clients, exists := ws.userClients[userID]
				if exists && len(clients) > 0 {
					delete(ws.offlineTimers, userID)
					ws.mutex.Unlock()
					ws.logger.Info("Offline timer aborted, user reconnected", zap.String("user_id", userID.String()))
					return
				}
				delete(ws.offlineTimers, userID)
				ws.mutex.Unlock()

				if err := ws.updateUserStatus(userID, false); err != nil {
					ws.logger.Error("Failed to update user status (delayed)", zap.Error(err))
				} else {
					ws.logger.Info("Offline timer fired, user set offline", zap.String("user_id", userID.String()))
				}
			})
			ws.offlineTimers[userID] = timer
			ws.logger.Info("Scheduled delayed offline update", zap.String("user_id", userID.String()))
		}
	}

	if client.GameID != nil {
		gamePublicID := *client.GameID

		if clients, exists := ws.gameClients[gamePublicID]; exists {
			for i, c := range clients {
				if c.ID == client.ID {
					ws.gameClients[gamePublicID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}

			if len(ws.gameClients[gamePublicID]) == 0 {
				delete(ws.gameClients, gamePublicID)
			}
		}

		gameLeaveDelay := client.reconnectGrace
		if gameLeaveDelay <= 0 {
			gameLeaveDelay = 180 * time.Second
		}
		userID := client.UserID
		if ws.gameActionHandler != nil {
			go func() {
				if err := ws.gameActionHandler.MarkPlayerDisconnected(userID, gamePublicID); err != nil {
					ws.logger.Warn("Failed to mark player disconnected",
						zap.String("user_id", userID.String()),
						zap.String("game_public_id", gamePublicID),
						zap.Error(err),
					)
				}
			}()
		}
		if existing, ok := ws.gameDisconnectTimers[userID]; ok {
			existing.Stop()
		}
		ws.gameDisconnectTimers[userID] = time.AfterFunc(gameLeaveDelay, func() {
			ws.mutex.Lock()
			_, stillConnected := ws.userClients[userID]
			delete(ws.gameDisconnectTimers, userID)
			ws.mutex.Unlock()

			if stillConnected {
				return
			}
			if ws.gameActionHandler == nil {
				return
			}
			if err := ws.gameActionHandler.HandleLeaveGame(userID, gamePublicID); err != nil {
				ws.logger.Warn("Failed to auto-leave game after disconnect",
					zap.String("user_id", userID.String()),
					zap.String("game_public_id", gamePublicID),
					zap.Error(err),
				)
			} else {
				ws.logger.Info("Auto-left game after disconnect grace period",
					zap.String("user_id", userID.String()),
					zap.String("game_public_id", gamePublicID),
				)
			}
		})
	}

	_ = client.Conn.CloseNow()
	ws.logger.Info("WebSocket client disconnected", zap.String("clientID", client.ID.String()))
}

func (ws *WebSocketService) updateUserStatus(userID uuid.UUID, isOnline bool) error {
	if ws.userStatusUpdater != nil {
		if err := ws.userStatusUpdater.UpdateUserOnlineStatus(userID, isOnline); err != nil {
			return err
		}

		publicID := ""
		ws.mutex.RLock()
		if clients, ok := ws.userClients[userID]; ok && len(clients) > 0 {
			publicID = clients[0].PublicID
		}
		ws.mutex.RUnlock()

		msg := map[string]interface{}{
			"type":           "presence",
			"user_public_id": publicID,
			"is_online":      isOnline,
		}
		ws.mutex.RLock()
		connectedClients := len(ws.clients)
		ws.mutex.RUnlock()
		ws.logger.Info("Broadcasting presence update", zap.String("user_id", userID.String()), zap.Bool("is_online", isOnline), zap.Int("connected_clients", connectedClients))
		_ = ws.BroadcastToAll(msg)
		return nil
	}
	ws.logger.Warn("Warning: No UserStatusUpdater provided, status update skipped")
	return nil
}

func (ws *WebSocketService) BroadcastToAdmins(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ws.mutex.RLock()
	clients := make([]*WSClient, 0)
	for _, c := range ws.clients {
		if c.IsAdmin {
			clients = append(clients, c)
		}
	}
	ws.mutex.RUnlock()

	for _, client := range clients {
		if !client.IsClosed.Load() {
			ws.trySend(client, data)
		}
	}

	if ws.relay != nil {
		pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer pubCancel()
		if err := ws.relay.PublishAdminMessage(pubCtx, data); err != nil {
			ws.logger.Warn("Failed to publish admin message to relay", zap.Error(err))
		}
	}

	return nil
}

func (ws *WebSocketService) BroadcastToAll(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ws.mutex.RLock()
	clients := make([]*WSClient, 0, len(ws.clients))
	for _, c := range ws.clients {
		clients = append(clients, c)
	}
	ws.mutex.RUnlock()

	for _, client := range clients {
		if !client.IsClosed.Load() {
			ws.trySend(client, data)
		}
	}

	if ws.relay != nil {
		pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer pubCancel()
		if err := ws.relay.PublishBroadcastMessage(pubCtx, data); err != nil {
			ws.logger.Warn("Failed to publish broadcast message to relay", zap.Error(err))
		}
	}

	return nil
}

func (ws *WebSocketService) GetConnectedClients() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
}

func (ws *WebSocketService) GetOnlineUsers() int {
	if ws.multiPod && ws.redisService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		count, err := ws.redisService.HLen(ctx, onlineUsersRedisKey)
		if err == nil {
			return int(count)
		}
		ws.logger.Warn("Failed to get online users from Redis, falling back to local", zap.Error(err))
	}
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.userClients)
}

const onlineUsersRedisKey = "system:online:users"

func (ws *WebSocketService) incrOnlineUser(userID uuid.UUID) {
	if !ws.multiPod || ws.redisService == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	newCount, err := ws.redisService.HIncrBy(ctx, onlineUsersRedisKey, userID.String(), 1)
	if err != nil {
		ws.logger.Warn("Failed to increment online user counter", zap.String("user_id", userID.String()), zap.Error(err))
		return
	}
	if newCount == 1 {
		ws.logger.Debug("User came online (Redis)", zap.String("user_id", userID.String()))
	}
}

func (ws *WebSocketService) decrOnlineUser(userID uuid.UUID) {
	if !ws.multiPod || ws.redisService == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	newCount, err := ws.redisService.HIncrBy(ctx, onlineUsersRedisKey, userID.String(), -1)
	if err != nil {
		ws.logger.Warn("Failed to decrement online user counter", zap.String("user_id", userID.String()), zap.Error(err))
		return
	}
	if newCount <= 0 {
		_ = ws.redisService.HDel(ctx, onlineUsersRedisKey, userID.String())
		ws.logger.Debug("User went offline (Redis)", zap.String("user_id", userID.String()))
	}
}

func (ws *WebSocketService) SendToUser(userID uuid.UUID, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ws.mutex.RLock()
	src, exists := ws.userClients[userID]
	if exists && len(src) > 0 {
		clients := make([]*WSClient, len(src))
		copy(clients, src)
		ws.mutex.RUnlock()
		for _, client := range clients {
			if !client.IsClosed.Load() {
				ws.trySend(client, data)
			}
		}
	} else {
		ws.mutex.RUnlock()
	}

	if ws.relay != nil {
		pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer pubCancel()
		if err := ws.relay.PublishUserMessage(pubCtx, userID, data); err != nil {
			ws.logger.Warn("Failed to publish user message to relay",
				zap.String("user_id", userID.String()), zap.Error(err))
		}
	}

	return nil
}

func (ws *WebSocketService) SendToGame(gamePublicID string, message interface{}) error {
	return ws.BroadcastToGame(gamePublicID, message, nil)
}

func (ws *WebSocketService) BroadcastToGame(gamePublicID string, message interface{}, excludeUserID *uuid.UUID) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ws.mutex.RLock()
	src, exists := ws.gameClients[gamePublicID]
	if exists && len(src) > 0 {
		clients := make([]*WSClient, len(src))
		copy(clients, src)
		ws.mutex.RUnlock()
		for _, client := range clients {
			if excludeUserID != nil && client.UserID == *excludeUserID {
				continue
			}
			if !client.IsClosed.Load() {
				ws.trySend(client, data)
			}
		}
	} else {
		ws.mutex.RUnlock()
	}

	if ws.relay != nil {
		pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer pubCancel()
		if err := ws.relay.PublishGameMessage(pubCtx, gamePublicID, data, excludeUserID); err != nil {
			ws.logger.Warn("Failed to publish game message to relay",
				zap.String("game_public_id", gamePublicID), zap.Error(err))
		}
	}

	return nil
}

func (ws *WebSocketService) addClientToGame(client *WSClient, gamePublicID string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	for _, c := range ws.gameClients[gamePublicID] {
		if c.ID == client.ID {
			return
		}
	}

	client.GameID = &gamePublicID
	ws.gameClients[gamePublicID] = append(ws.gameClients[gamePublicID], client)

	if timer, ok := ws.gameDisconnectTimers[client.UserID]; ok {
		timer.Stop()
		delete(ws.gameDisconnectTimers, client.UserID)
		ws.logger.Info("Game disconnect timer cancelled, player rejoined",
			zap.String("user_id", client.UserID.String()),
			zap.String("game_public_id", gamePublicID),
		)
		gamePublicIDCopy := gamePublicID
		userIDCopy := client.UserID
		if ws.gameActionHandler != nil {
			go func() {
				_ = ws.gameActionHandler.MarkPlayerReconnected(userIDCopy, gamePublicIDCopy)
			}()
		}
	}

	ws.logger.Info("Client joined game room",
		zap.String("client_id", client.ID.String()),
		zap.String("game_public_id", gamePublicID),
	)
}

func (ws *WebSocketService) removeClientFromGame(client *WSClient, gamePublicID string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if clients, exists := ws.gameClients[gamePublicID]; exists {
		for i, c := range clients {
			if c.ID == client.ID {
				ws.gameClients[gamePublicID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		if len(ws.gameClients[gamePublicID]) == 0 {
			delete(ws.gameClients, gamePublicID)
		}
	}

	client.GameID = nil

	ws.logger.Info("Client left game room",
		zap.String("client_id", client.ID.String()),
		zap.String("game_public_id", gamePublicID),
	)
}

func (ws *WebSocketService) BroadcastAdminNotification(notif AdminNotification) {
	msg := map[string]interface{}{
		"type":      "admin_notification",
		"event":     notif.Event,
		"data":      notif.Data,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	if notif.EntityType != "" {
		msg["entity_type"] = notif.EntityType
	}
	if notif.EntityID != "" {
		msg["entity_id"] = notif.EntityID
	}
	if notif.ActionURL != "" {
		msg["action_url"] = notif.ActionURL
	}
	if err := ws.BroadcastToAdmins(msg); err != nil {
		ws.logger.Warn("Failed to broadcast admin notification", zap.String("event", notif.Event), zap.Error(err))
	}
}

func (ws *WebSocketService) handleRelayMessage(msg PubSubRelayMessage) {
	switch msg.MessageType {
	case PubSubMsgTypeUser:
		userID, err := uuid.Parse(msg.TargetID)
		if err != nil {
			return
		}
		ws.deliverToUserLocally(userID, msg.Payload)

	case PubSubMsgTypeGame:
		var excludeUserID *uuid.UUID
		if msg.ExcludeUserID != nil {
			if uid, err := uuid.Parse(*msg.ExcludeUserID); err == nil {
				excludeUserID = &uid
			}
		}
		ws.deliverToGameLocally(msg.TargetID, msg.Payload, excludeUserID)

	case PubSubMsgTypeAdmin:
		ws.deliverToAdminsLocally(msg.Payload)

	case PubSubMsgTypeBroadcast:
		ws.deliverToAllLocally(msg.Payload)
	}
}

func (ws *WebSocketService) deliverToUserLocally(userID uuid.UUID, data json.RawMessage) {
	ws.mutex.RLock()
	src, exists := ws.userClients[userID]
	if !exists || len(src) == 0 {
		ws.mutex.RUnlock()
		return
	}
	clients := make([]*WSClient, len(src))
	copy(clients, src)
	ws.mutex.RUnlock()

	for _, client := range clients {
		if !client.IsClosed.Load() {
			ws.trySend(client, data)
		}
	}
}

func (ws *WebSocketService) deliverToGameLocally(gamePublicID string, data json.RawMessage, excludeUserID *uuid.UUID) {
	ws.mutex.RLock()
	src, exists := ws.gameClients[gamePublicID]
	if !exists || len(src) == 0 {
		ws.mutex.RUnlock()
		return
	}
	clients := make([]*WSClient, len(src))
	copy(clients, src)
	ws.mutex.RUnlock()

	for _, client := range clients {
		if excludeUserID != nil && client.UserID == *excludeUserID {
			continue
		}
		if !client.IsClosed.Load() {
			ws.trySend(client, data)
		}
	}
}

func (ws *WebSocketService) deliverToAdminsLocally(data json.RawMessage) {
	ws.mutex.RLock()
	clients := make([]*WSClient, 0)
	for _, c := range ws.clients {
		if c.IsAdmin {
			clients = append(clients, c)
		}
	}
	ws.mutex.RUnlock()

	for _, client := range clients {
		if !client.IsClosed.Load() {
			ws.trySend(client, data)
		}
	}
}

func (ws *WebSocketService) deliverToAllLocally(data json.RawMessage) {
	ws.mutex.RLock()
	clients := make([]*WSClient, 0, len(ws.clients))
	for _, c := range ws.clients {
		clients = append(clients, c)
	}
	ws.mutex.RUnlock()

	for _, client := range clients {
		if !client.IsClosed.Load() {
			ws.trySend(client, data)
		}
	}
}

func ginContextString(c *gin.Context, key string) string {
	val, exists := c.Get(key)
	if !exists {
		return ""
	}
	s, _ := val.(string)
	return s
}
