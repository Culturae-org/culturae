// backend/internal/service/game_events.go

package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GameEventPersister interface {
	SaveEvent(gameID uuid.UUID, eventType string, data map[string]interface{}, occurredAt time.Time) error
}

type RawGameEvent struct {
	Type      string
	GameID    uuid.UUID
	PublicID  string
	Data      map[string]interface{}
	Timestamp time.Time
}

type GameEventSource interface {
	GetRawEventChannel() <-chan RawGameEvent
}

type GameEventBroadcaster struct {
	eventChan    <-chan RawGameEvent
	wsService    WebSocketServiceInterface
	eventLogRepo GameEventPersister
	logger       *zap.Logger
	stopChan     chan struct{}
}

func NewGameEventBroadcaster(eventChan <-chan RawGameEvent, wsService WebSocketServiceInterface, logger *zap.Logger) *GameEventBroadcaster {
	return &GameEventBroadcaster{
		eventChan: eventChan,
		wsService: wsService,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}
}

func (geb *GameEventBroadcaster) SetEventLogRepo(repo GameEventPersister) {
	geb.eventLogRepo = repo
}

func (geb *GameEventBroadcaster) Start() {
	go geb.run()
}

func (geb *GameEventBroadcaster) Stop() {
	close(geb.stopChan)
}

func (geb *GameEventBroadcaster) run() {
	for {
		select {
		case event, ok := <-geb.eventChan:
			if !ok {
				geb.logger.Info("Game event channel closed")
				return
			}
			geb.handleEvent(event)
		case <-geb.stopChan:
			geb.logger.Info("Game event broadcaster stopped")
			return
		}
	}
}

func (geb *GameEventBroadcaster) handleEvent(event RawGameEvent) {
	geb.logger.Debug("Broadcasting game event",
		zap.String("type", event.Type),
		zap.String(keyGamePublicID, event.PublicID),
	)

	if err := geb.wsService.SendToGame(event.PublicID, map[string]interface{}{
		keyType:      event.Type,
		keyPublicID: event.PublicID,
		keyData:      event.Data,
		keyTimestamp: event.Timestamp,
	}); err != nil {
		geb.logger.Error("Failed to broadcast game event",
			zap.String("type", event.Type),
			zap.String(keyGamePublicID, event.PublicID),
			zap.Error(err),
		)
	}

	if geb.eventLogRepo != nil && event.GameID != uuid.Nil {
		if err := geb.eventLogRepo.SaveEvent(event.GameID, event.Type, event.Data, event.Timestamp); err != nil {
			geb.logger.Error("Failed to persist game event",
				zap.String("type", event.Type),
				zap.String("game_id", event.GameID.String()),
				zap.Error(err),
			)
		}
	}
}
