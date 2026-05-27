// backend/internal/service/ws_pubsub.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	wsPodBroadcastChannel = "ws:pod:broadcast"
	userPodMappingKey     = "ws:user:pods"
)

func podRelayChannel(podID string) string {
	return fmt.Sprintf("ws:pod:%s:relay", podID)
}

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

	myChannel := podRelayChannel(r.podID)
	pubsub := r.redisService.Subscribe(ctx, myChannel, wsPodBroadcastChannel)
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

	r.logger.Info("Pub/Sub relay started",
		zap.String("pod_channel", myChannel),
		zap.String("broadcast_channel", wsPodBroadcastChannel),
		zap.String("pod_id", r.podID),
	)
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

func (r *PubSubRelay) publishTo(ctx context.Context, channel string, msg PubSubRelayMessage) error {
	msg.SenderPodID = r.podID
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.redisService.PublishRaw(ctx, channel, string(data))
}

func (r *PubSubRelay) PublishUserMessage(ctx context.Context, userID uuid.UUID, payload []byte) error {
	msg := PubSubRelayMessage{
		MessageType: PubSubMsgTypeUser,
		TargetID:    userID.String(),
		Payload:     payload,
	}

	if r.redisService != nil {
		targetPod, err := r.redisService.HGet(ctx, userPodMappingKey, userID.String())
		if err == nil && targetPod != "" && targetPod != r.podID {
			return r.publishTo(ctx, podRelayChannel(targetPod), msg)
		}
	}

	return r.publishTo(ctx, wsPodBroadcastChannel, msg)
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
	return r.publishTo(ctx, wsPodBroadcastChannel, msg)
}

func (r *PubSubRelay) PublishAdminMessage(ctx context.Context, payload []byte) error {
	return r.publishTo(ctx, wsPodBroadcastChannel, PubSubRelayMessage{
		MessageType: PubSubMsgTypeAdmin,
		Payload:     payload,
	})
}

func (r *PubSubRelay) PublishBroadcastMessage(ctx context.Context, payload []byte) error {
	return r.publishTo(ctx, wsPodBroadcastChannel, PubSubRelayMessage{
		MessageType: PubSubMsgTypeBroadcast,
		Payload:     payload,
	})
}

func (r *PubSubRelay) RegisterUserPod(ctx context.Context, userID uuid.UUID) {
	if r.redisService == nil {
		return
	}
	if err := r.redisService.HSet(ctx, userPodMappingKey, userID.String(), r.podID); err != nil {
		r.logger.Warn("Failed to register user pod mapping",
			zap.String("user_id", userID.String()), zap.Error(err))
	}
}

func (r *PubSubRelay) UnregisterUserPod(ctx context.Context, userID uuid.UUID) {
	if r.redisService == nil {
		return
	}
	if err := r.redisService.HDel(ctx, userPodMappingKey, userID.String()); err != nil {
		r.logger.Warn("Failed to unregister user pod mapping",
			zap.String("user_id", userID.String()), zap.Error(err))
	}
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
