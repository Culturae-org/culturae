// backend/internal/service/ws_pubsub.go

package service

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const wsPodRelayChannel = "ws:pod:relay"

type PubSubMessageType string

const (
	PubSubMsgTypeUser      PubSubMessageType = "user"
	PubSubMsgTypeGame      PubSubMessageType = "game"
	PubSubMsgTypeAdmin     PubSubMessageType = "admin"
	PubSubMsgTypeBroadcast PubSubMessageType = "broadcast"
)

type PubSubRelayMessage struct {
	SenderPodID   string            `json:"pod"`
	MessageType   PubSubMessageType `json:"mt"`
	TargetID      string            `json:"tid"`
	ExcludeUserID *string           `json:"eid,omitempty"`
	Payload       json.RawMessage   `json:"p"`
}

type PubSubRelay struct {
	podID        string
	redisService cache.RedisClientInterface
	logger       *zap.Logger
	mu           sync.Mutex
	subCancel    context.CancelFunc
	started      bool
}

func NewPubSubRelay(podID string, redisService cache.RedisClientInterface, logger *zap.Logger) *PubSubRelay {
	return &PubSubRelay{
		podID:        podID,
		redisService: redisService,
		logger:       logger,
	}
}

func (r *PubSubRelay) PodID() string {
	return r.podID
}

func (r *PubSubRelay) Start(parentCtx context.Context, handler func(msg PubSubRelayMessage)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return
	}
	r.started = true

	ctx, cancel := context.WithCancel(parentCtx)
	r.subCancel = cancel

	pubsub := r.redisService.Subscribe(ctx, wsPodRelayChannel)
	ch := pubsub.Channel()

	go func() {
		defer func() {
			_ = pubsub.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				r.handleMessage(msg.Payload, handler)
			}
		}
	}()

	r.logger.Info("Pub/Sub relay started", zap.String("channel", wsPodRelayChannel), zap.String("pod_id", r.podID))
}

func (r *PubSubRelay) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return
	}
	r.started = false

	if r.subCancel != nil {
		r.subCancel()
	}

	r.logger.Info("Pub/Sub relay stopped")
}

func (r *PubSubRelay) publish(ctx context.Context, msg PubSubRelayMessage) error {
	msg.SenderPodID = r.podID
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.redisService.PublishRaw(ctx, wsPodRelayChannel, string(data))
}

func (r *PubSubRelay) PublishUserMessage(ctx context.Context, userID uuid.UUID, payload []byte) error {
	return r.publish(ctx, PubSubRelayMessage{
		MessageType: PubSubMsgTypeUser,
		TargetID:    userID.String(),
		Payload:     payload,
	})
}

func (r *PubSubRelay) PublishGameMessage(ctx context.Context, gamePublicID string, payload []byte, excludeUserID *uuid.UUID) error {
	msg := PubSubRelayMessage{
		MessageType: PubSubMsgTypeGame,
		TargetID:    gamePublicID,
		Payload:     payload,
	}
	if excludeUserID != nil {
		eid := excludeUserID.String()
		msg.ExcludeUserID = &eid
	}
	return r.publish(ctx, msg)
}

func (r *PubSubRelay) PublishAdminMessage(ctx context.Context, payload []byte) error {
	return r.publish(ctx, PubSubRelayMessage{
		MessageType: PubSubMsgTypeAdmin,
		Payload:     payload,
	})
}

func (r *PubSubRelay) PublishBroadcastMessage(ctx context.Context, payload []byte) error {
	return r.publish(ctx, PubSubRelayMessage{
		MessageType: PubSubMsgTypeBroadcast,
		Payload:     payload,
	})
}

func (r *PubSubRelay) handleMessage(payload string, handler func(msg PubSubRelayMessage)) {
	var msg PubSubRelayMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		r.logger.Warn("Failed to unmarshal Pub/Sub relay message", zap.Error(err))
		return
	}

	if msg.SenderPodID == r.podID {
		return
	}

	handler(msg)
}
