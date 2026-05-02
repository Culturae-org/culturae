// backend/internal/game/matchmaking.go

package game

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MatchFoundCallback func(user1, user2 uuid.UUID, mode model.GameMode, gameParams map[string]interface{}) error

type MatchmakingService struct {
	redisService       cache.RedisClientInterface
	userRepo           repository.UserRepositoryInterface
	logger             *zap.Logger
	appCtx             context.Context
	matchFoundCallback MatchFoundCallback
	userNotifier       UserNotifier
	podDiscovery       service.PodDiscoveryServiceInterface
	podID              string
}

func NewMatchmakingService(
	appCtx context.Context,
	redisService cache.RedisClientInterface,
	userRepo repository.UserRepositoryInterface,
	logger *zap.Logger,
) *MatchmakingService {
	return &MatchmakingService{
		redisService: redisService,
		userRepo:     userRepo,
		logger:       logger,
		appCtx:       appCtx,
	}
}

func (s *MatchmakingService) opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(s.appCtx, 5*time.Second)
}

func (s *MatchmakingService) scanCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(s.appCtx, 15*time.Second)
}

func (s *MatchmakingService) SetMatchFoundCallback(cb MatchFoundCallback) {
	s.matchFoundCallback = cb
}

func (s *MatchmakingService) SetUserNotifier(n UserNotifier) {
	s.userNotifier = n
}

func (s *MatchmakingService) SetPodDiscovery(podDiscovery service.PodDiscoveryServiceInterface, podID string) {
	s.podDiscovery = podDiscovery
	s.podID = podID
}

func (s *MatchmakingService) queueKeyForParams(mode model.GameMode, params map[string]interface{}) string {
	category := ""
	flagVariant := ""
	language := ""
	questionType := ""

	if v, ok := params["category"].(string); ok {
		category = v
	}
	if v, ok := params["flag_variant"].(string); ok {
		flagVariant = v
	}
	if v, ok := params["language"].(string); ok {
		language = v
	}
	if v, ok := params["question_type"].(string); ok {
		questionType = v
	}

	key := fmt.Sprintf("matchmaking:queue:%s", mode)
	if category != "" {
		key += fmt.Sprintf(":%s", category)
	}
	if flagVariant != "" {
		key += fmt.Sprintf(":%s", flagVariant)
	}
	if language != "" {
		key += fmt.Sprintf(":%s", language)
	}
	if questionType != "" {
		key += fmt.Sprintf(":%s", questionType)
	}
	return key
}

func (s *MatchmakingService) tsKey(mode model.GameMode, userID uuid.UUID) string {
	return fmt.Sprintf("matchmaking:ts:%s:%s", mode, userID.String())
}

const atomicPopTwoScript = `
local v1 = redis.call('LPOP', KEYS[1])
if v1 == false then return {} end
local v2 = redis.call('LPOP', KEYS[1])
if v2 == false then
    redis.call('RPUSH', KEYS[1], v1)
    return {}
end
if v1 == v2 then
    redis.call('RPUSH', KEYS[1], v1)
    return {}
end
return {v1, v2}
`

func (s *MatchmakingService) JoinQueue(userID uuid.UUID, mode model.GameMode, gameParams map[string]interface{}) error {
	paramsKey := fmt.Sprintf("matchmaking:params:%s:%s", mode, userID.String())
	paramsJSON, _ := json.Marshal(gameParams)

	setnxCtx, setnxCancel := s.opCtx()
	inserted, err := s.redisService.SetNX(setnxCtx, paramsKey, string(paramsJSON), 5*time.Minute)
	setnxCancel()
	if err != nil {
		return fmt.Errorf("failed to join queue: %w", err)
	}
	if !inserted {
		return fmt.Errorf("user already in queue for mode %s", mode)
	}

	queueKey := s.queueKeyForParams(mode, gameParams)
	userStr := userID.String()

	ctx, cancel := s.opCtx()
	pushErr := s.redisService.RPush(ctx, queueKey, userStr)
	cancel()
	if pushErr != nil {
		delCtx, delCancel := s.opCtx()
		_ = s.redisService.Delete(delCtx, paramsKey)
		delCancel()
		return fmt.Errorf("failed to join queue: %w", pushErr)
	}

	expCtx, expCancel := s.opCtx()
	_ = s.redisService.Expire(expCtx, queueKey, 10*time.Minute)
	expCancel()

	tsKey := s.tsKey(mode, userID)
	tsCtx, tsCancel := s.opCtx()
	_ = s.redisService.Set(tsCtx, tsKey, fmt.Sprintf("%d", time.Now().Unix()), 5*time.Minute)
	tsCancel()

	s.logger.Info("User joined matchmaking queue",
		zap.String("user_id", userStr),
		zap.String("mode", string(mode)),
		zap.Any("params", gameParams),
	)

	if s.userNotifier != nil {
		_ = s.userNotifier.SendToUser(userID, map[string]interface{}{
			"type": "queue_joined",
			keyData: map[string]interface{}{
				"mode":   string(mode),
				"params": gameParams,
			},
		})
	}
	go s.ProcessQueueForParams(mode, gameParams)

	return nil
}

func (s *MatchmakingService) LeaveQueue(userID uuid.UUID, mode model.GameMode) error {
	paramsKey := fmt.Sprintf("matchmaking:params:%s:%s", mode, userID.String())

	var resolvedKey string
	getCtx, getCancel := s.opCtx()
	if data, err := s.redisService.Get(getCtx, paramsKey); err == nil && data != "" {
		var gameParams map[string]interface{}
		if err := json.Unmarshal([]byte(data), &gameParams); err == nil {
			resolvedKey = s.queueKeyForParams(mode, gameParams)
		}
	}
	getCancel()

	if resolvedKey != "" {
		lremCtx, lremCancel := s.opCtx()
		_ = s.redisService.LRem(lremCtx, resolvedKey, 0, userID.String())
		lremCancel()
	} else {
		pattern := fmt.Sprintf("matchmaking:queue:%s*", mode)
		scanCtx, scanCancel := s.scanCtx()
		if keys, err := s.redisService.Scan(scanCtx, pattern); err == nil {
			for _, key := range keys {
				lremCtx, lremCancel := s.opCtx()
				_ = s.redisService.LRem(lremCtx, key, 0, userID.String())
				lremCancel()
			}
		}
		scanCancel()
	}

	delCtx, delCancel := s.opCtx()
	_ = s.redisService.Delete(delCtx, paramsKey)
	_ = s.redisService.Delete(delCtx, s.tsKey(mode, userID))
	delCancel()

	s.logger.Info("User left matchmaking queue",
		zap.String("user_id", userID.String()),
		zap.String("mode", string(mode)),
	)

	if s.userNotifier != nil {
		_ = s.userNotifier.SendToUser(userID, map[string]interface{}{
			"type": "queue_left",
			keyData: map[string]interface{}{
				"mode": string(mode),
			},
		})
	}

	return nil
}

func (s *MatchmakingService) ProcessQueueForParams(mode model.GameMode, gameParams map[string]interface{}) {
	key := s.queueKeyForParams(mode, gameParams)

	for {
		popCtx, popCancel := s.opCtx()
		result, err := s.redisService.Eval(popCtx, atomicPopTwoScript, []string{key})
		popCancel()
		if err != nil {
			s.logger.Error("Failed to atomically pop from matchmaking queue", zap.Error(err))
			return
		}

		items, ok := result.([]interface{})
		if !ok || len(items) < 2 {
			break
		}

		user1IDStr, ok1 := items[0].(string)
		user2IDStr, ok2 := items[1].(string)
		if !ok1 || !ok2 || user1IDStr == "" || user2IDStr == "" {
			break
		}

		user1, err1 := uuid.Parse(user1IDStr)
		user2, err2 := uuid.Parse(user2IDStr)
		if err1 != nil || err2 != nil {
			s.logger.Warn("Invalid UUID in matchmaking queue",
				zap.String("u1", user1IDStr),
				zap.String("u2", user2IDStr),
			)
			continue
		}

		s.createMatchWithParams(user1, user2, mode, gameParams)
	}
}

func (s *MatchmakingService) createMatchWithParams(user1, user2 uuid.UUID, mode model.GameMode, gameParams map[string]interface{}) {
	if s.matchFoundCallback == nil {
		s.logger.Error("matchFoundCallback not set — re-queuing players",
			zap.String("u1", user1.String()),
			zap.String("u2", user2.String()),
		)
		key := s.queueKeyForParams(mode, gameParams)
		pushCtx, pushCancel := s.opCtx()
		_ = s.redisService.RPush(pushCtx, key, user1.String())
		_ = s.redisService.RPush(pushCtx, key, user2.String())
		pushCancel()
		return
	}

	bestPodID := ""
	if s.podDiscovery != nil {
		bestPod, err := s.podDiscovery.GetBestPodForGame(mode, "")
		if err != nil {
			s.logger.Warn("Failed to get best pod for game, falling back to local",
				zap.Error(err),
			)
		} else {
			bestPodID = bestPod.PodID
			s.logger.Info("Selected best pod for matchmaking game",
				zap.String("best_pod", bestPodID),
				zap.String("current_pod", s.podID),
				zap.Int64("clients", bestPod.ConnectedClients),
			)
		}
	}

	if bestPodID != "" && bestPodID != s.podID {
		s.logger.Info("Delegating game creation to better pod",
			zap.String("target_pod", bestPodID),
			zap.String("current_pod", s.podID),
			zap.String("u1", user1.String()),
			zap.String("u2", user2.String()),
		)
		delegateReq := DelegateGameRequest{
			User1:      user1,
			User2:      user2,
			Mode:       mode,
			GameParams: gameParams,
		}
		delegateJSON, _ := json.Marshal(delegateReq)
		delegateKey := fmt.Sprintf("game:delegate:%s", bestPodID)
		ctx, cancel := context.WithTimeout(s.appCtx, 2*time.Second)
		_ = s.redisService.RPush(ctx, delegateKey, string(delegateJSON))
		_ = s.redisService.Expire(ctx, delegateKey, 30*time.Second)
		cancel()
		return
	}

	if err := s.matchFoundCallback(user1, user2, mode, gameParams); err != nil {
		s.logger.Error("Failed to create matchmaked game — re-queuing players",
			zap.String("u1", user1.String()),
			zap.String("u2", user2.String()),
			zap.Error(err),
		)
		key := s.queueKeyForParams(mode, gameParams)
		pushCtx, pushCancel := s.opCtx()
		_ = s.redisService.RPush(pushCtx, key, user1.String())
		_ = s.redisService.RPush(pushCtx, key, user2.String())
		pushCancel()
		return
	}

	if s.userNotifier != nil {
		matchEvent := map[string]interface{}{
			"type": "match_found",
			keyData: map[string]interface{}{
				"mode": string(mode),
			},
		}
		_ = s.userNotifier.SendToUser(user1, matchEvent)
		_ = s.userNotifier.SendToUser(user2, matchEvent)
	}

	s.logger.Info("Match created and delegated",
		zap.String("u1", user1.String()),
		zap.String("u2", user2.String()),
		zap.String("mode", string(mode)),
		zap.Any("params", gameParams),
	)
}

func (s *MatchmakingService) GetQueueStats() (map[string]int64, error) {
	scanCtx, scanCancel := s.scanCtx()
	defer scanCancel()
	keys, err := s.redisService.Scan(scanCtx, "matchmaking:queue:*")
	if err != nil {
		return nil, fmt.Errorf("failed to scan matchmaking queue keys: %w", err)
	}

	stats := make(map[string]int64)
	const prefix = "matchmaking:queue:"
	for _, key := range keys {
		suffix := strings.TrimPrefix(key, prefix)
		mode := strings.SplitN(suffix, ":", 2)[0]
		if mode == "" {
			continue
		}
		lenCtx, lenCancel := s.opCtx()
		count, err := s.redisService.LLen(lenCtx, key)
		lenCancel()
		if err != nil {
			continue
		}
		stats[mode] += count
	}

	return stats, nil
}

func (s *MatchmakingService) ClearQueue(mode model.GameMode) error {
	paramsPattern := fmt.Sprintf("matchmaking:params:%s:*", mode)
	pScanCtx, pScanCancel := s.scanCtx()
	paramsKeys, _ := s.redisService.Scan(pScanCtx, paramsPattern)
	pScanCancel()
	s.logger.Info("ClearQueue: found params keys", zap.Int("count", len(paramsKeys)), zap.Strings("keys", paramsKeys), zap.Bool("has_notifier", s.userNotifier != nil))
	prefix := fmt.Sprintf("matchmaking:params:%s:", mode)
	for _, key := range paramsKeys {
		userID, err := uuid.Parse(strings.TrimPrefix(key, prefix))
		if err != nil {
			continue
		}
		if s.userNotifier != nil {
			_ = s.userNotifier.SendToUser(userID, map[string]interface{}{
				"type": "queue_left",
				keyData: map[string]interface{}{
					"mode":   string(mode),
					"reason": "queue_cleared",
				},
			})
		}
		delCtx, delCancel := s.opCtx()
		_ = s.redisService.Delete(delCtx, key)
		_ = s.redisService.Delete(delCtx, s.tsKey(mode, userID))
		delCancel()
	}

	pattern := fmt.Sprintf("matchmaking:queue:%s*", mode)
	qScanCtx, qScanCancel := s.scanCtx()
	keys, err := s.redisService.Scan(qScanCtx, pattern)
	qScanCancel()
	if err != nil {
		return fmt.Errorf("failed to scan queue keys for %s: %w", mode, err)
	}
	for _, key := range keys {
		delCtx, delCancel := s.opCtx()
		_ = s.redisService.Delete(delCtx, key)
		delCancel()
	}
	s.logger.Info("Matchmaking queue cleared", zap.String("mode", string(mode)), zap.Int("keys_deleted", len(keys)))
	return nil
}

type DelegateGameRequest struct {
	User1       uuid.UUID                `json:"user1"`
	User2       uuid.UUID                `json:"user2"`
	Mode        model.GameMode           `json:"mode"`
	GameParams  map[string]interface{} `json:"game_params"`
}

func (s *MatchmakingService) Start(ctx context.Context) {
	if s.podID == "" {
		return
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.drainDelegateQueue(ctx)
		}
	}
}

func (s *MatchmakingService) drainDelegateQueue(ctx context.Context) {
	delegateKey := fmt.Sprintf("game:delegate:%s", s.podID)
	for {
		popCtx, popCancel := context.WithTimeout(ctx, 2*time.Second)
		delegateJSON, err := s.redisService.LPop(popCtx, delegateKey)
		popCancel()
		if err != nil || delegateJSON == "" {
			return
		}

		var req DelegateGameRequest
		if err := json.Unmarshal([]byte(delegateJSON), &req); err != nil {
			s.logger.Warn("Failed to unmarshal delegate request", zap.Error(err))
			continue
		}

		s.logger.Info("Processing delegated game creation",
			zap.String("user1", req.User1.String()),
			zap.String("user2", req.User2.String()),
			zap.String("mode", string(req.Mode)),
		)

		if s.matchFoundCallback == nil {
			s.logger.Error("matchFoundCallback not set for delegated game — dropping")
			continue
		}

		if err := s.matchFoundCallback(req.User1, req.User2, req.Mode, req.GameParams); err != nil {
			s.logger.Error("Failed to create delegated game", zap.Error(err))
			continue
		}

		if s.userNotifier != nil {
			matchEvent := map[string]interface{}{
				"type": "match_found",
				keyData: map[string]interface{}{
					"mode": string(req.Mode),
				},
			}
			_ = s.userNotifier.SendToUser(req.User1, matchEvent)
			_ = s.userNotifier.SendToUser(req.User2, matchEvent)
		}
	}
}
