// backend/internal/service/pod_discovery.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	podKeyPrefix      = "pod:"
	podKeySuffix     = ":heartbeat"
	heartbeatTTL     = 15 * time.Second
	heartbeatInterval = 5 * time.Second
)

type PodDiscoveryServiceInterface interface {
	Start(ctx context.Context)
	Stop()
	PodID() string
	PodType() string
	PublishHeartbeat() error
	GetAllPods(ctx context.Context) ([]model.PodInfo, error)
	GetPodInfo(ctx context.Context, podID string) (*model.PodInfo, error)
	GetBestPodForGame(mode model.GameMode, preferPodID string) (*model.PodInfo, error)
	GetPodsByLoad() ([]model.PodInfo, error)
}

type GameCountProvider interface {
	GetActiveGamesCount() int
}

type PodDiscoveryService struct {
	podID        string
	podType      string
	redisService cache.RedisClientInterface
	logger       *zap.Logger
	wsService    WebSocketServiceInterface
	gameManager  GameCountProvider
	stopChan     chan struct{}
	startedAt    time.Time
}

func NewPodDiscoveryService(
	podType string,
	redisService cache.RedisClientInterface,
	logger *zap.Logger,
	wsService WebSocketServiceInterface,
	gameManager GameCountProvider,
) *PodDiscoveryService {
	return &PodDiscoveryService{
		podID:        uuid.New().String(),
		podType:      podType,
		redisService: redisService,
		logger:       logger,
		wsService:    wsService,
		gameManager:  gameManager,
		stopChan:     make(chan struct{}),
		startedAt:    time.Now(),
	}
}

func (p *PodDiscoveryService) PodID() string {
	return p.podID
}

func (p *PodDiscoveryService) PodType() string {
	return p.podType
}

func (p *PodDiscoveryService) Start(ctx context.Context) {
	p.logger.Info("Starting pod discovery service",
		zap.String("pod_id", p.podID),
		zap.String("pod_type", p.podType),
	)

	if err := p.PublishHeartbeat(); err != nil {
		p.logger.Warn("Failed to publish initial heartbeat", zap.Error(err))
	}

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.Stop()
			return
		case <-p.stopChan:
			p.removePod()
			return
		case <-ticker.C:
			if err := p.PublishHeartbeat(); err != nil {
				p.logger.Warn("Failed to publish heartbeat", zap.Error(err))
			}
		}
	}
}

func (p *PodDiscoveryService) Stop() {
	select {
	case <-p.stopChan:
		return
	default:
		close(p.stopChan)
	}
	p.removePod()
	p.logger.Info("Pod discovery service stopped", zap.String("pod_id", p.podID))
}

func (p *PodDiscoveryService) removePod() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := podKeyPrefix + p.podID + podKeySuffix
	if err := p.redisService.Delete(ctx, key); err != nil {
		p.logger.Warn("Failed to remove pod from Redis", zap.Error(err))
	}
}

func (p *PodDiscoveryService) PublishHeartbeat() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connectedClients := 0
	onlineUsers := 0
	activeGames := 0

	if p.wsService != nil {
		connectedClients = p.wsService.GetConnectedClients()
		onlineUsers = p.wsService.GetLocalOnlineUsers()
	}

	if p.gameManager != nil {
		activeGames = p.gameManager.GetActiveGamesCount()
	}

	podInfo := model.PodInfo{
		PodID:            p.podID,
		PodType:          model.PodType(p.podType),
		Status:           model.PodStatusHealthy,
		IsCurrent:        false,
		ConnectedClients: int64(connectedClients),
		OnlineUsers:      int64(onlineUsers),
		ActiveGames:      int64(activeGames),
		LastHeartbeat:    time.Now(),
		StartedAt:        p.startedAt,
	}

	data, err := json.Marshal(podInfo)
	if err != nil {
		return err
	}

	key := podKeyPrefix + p.podID + podKeySuffix
	return p.redisService.Set(ctx, key, string(data), heartbeatTTL)
}

func (p *PodDiscoveryService) GetAllPods(ctx context.Context) ([]model.PodInfo, error) {
	keys, err := p.redisService.Scan(ctx, podKeyPrefix+"*"+podKeySuffix)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []model.PodInfo{}, nil
	}

	values, err := p.redisService.MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}

	var pods []model.PodInfo
	for i, value := range values {
		if value == "" {
			continue
		}
		var pod model.PodInfo
		if err := json.Unmarshal([]byte(value), &pod); err != nil {
			p.logger.Warn("Failed to unmarshal pod info",
				zap.String("key", keys[i]),
				zap.Error(err),
			)
			continue
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (p *PodDiscoveryService) GetPodInfo(ctx context.Context, podID string) (*model.PodInfo, error) {
	key := podKeyPrefix + podID + podKeySuffix
	value, err := p.redisService.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var pod model.PodInfo
	if err := json.Unmarshal([]byte(value), &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

func (p *PodDiscoveryService) GetBestPodForGame(mode model.GameMode, preferPodID string) (*model.PodInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := p.GetAllPods(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods available")
	}

	var availablePods []model.PodInfo
	for _, pod := range pods {
		if pod.Status != model.PodStatusHealthy {
			continue
		}
		if pod.PodType != model.PodTypeHeadless && pod.PodType != model.PodTypeMain {
			continue
		}
		if preferPodID != "" && pod.PodID == preferPodID {
			p.logger.Debug("Preferring requested pod for game",
				zap.String("prefer_pod", preferPodID),
				zap.String("pod_id", pod.PodID),
				zap.Int64("clients", pod.ConnectedClients),
			)
			return &pod, nil
		}
		availablePods = append(availablePods, pod)
	}

	if len(availablePods) == 0 {
		return nil, fmt.Errorf("no healthy pods available")
	}

	podLoad := func(p model.PodInfo) int64 {
		return p.ConnectedClients + p.ActiveGames*2
	}
	sort.Slice(availablePods, func(i, j int) bool {
		return podLoad(availablePods[i]) < podLoad(availablePods[j])
	})

	bestPod := availablePods[0]
	p.logger.Debug("Selected best pod for game",
		zap.String("mode", string(mode)),
		zap.String("pod_id", bestPod.PodID),
		zap.Int64("clients", bestPod.ConnectedClients),
		zap.Int64("games", bestPod.ActiveGames),
		zap.Int64("load_score", podLoad(bestPod)),
	)

	return &bestPod, nil
}

func (p *PodDiscoveryService) GetPodsByLoad() ([]model.PodInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := p.GetAllPods(ctx)
	if err != nil {
		return nil, err
	}

	var availablePods []model.PodInfo
	for _, pod := range pods {
		if pod.Status == model.PodStatusHealthy &&
			(pod.PodType == model.PodTypeHeadless || pod.PodType == model.PodTypeMain) {
			availablePods = append(availablePods, pod)
		}
	}

	sort.Slice(availablePods, func(i, j int) bool {
		loadI := availablePods[i].ConnectedClients + availablePods[i].ActiveGames*2
		loadJ := availablePods[j].ConnectedClients + availablePods[j].ActiveGames*2
		return loadI < loadJ
	})

	return availablePods, nil
}