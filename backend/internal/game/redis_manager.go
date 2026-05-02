// backend/internal/game/redis_manager.go

package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/bsm/redislock"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RedisGameManager struct {
	redisService cache.RedisClientInterface
	locker       *redislock.Client
	userRepo     repository.UserRepositoryInterface
	gameRepo     repository.GameRepositoryInterface
	logsRepo     adminRepo.AdminLogsRepositoryInterface
	xpCalculator *XPCalculator
	eloCalc      *ELOCalculator
	userNotifier UserNotifier
	logger       *zap.Logger
	eventChan    chan GameEvent
	ctx          context.Context

	archiveQueue chan GameEngine
	stopWorkers  chan struct{}
	stopOnce     sync.Once
}

func NewRedisGameManager(
	appCtx context.Context,
	redisService cache.RedisClientInterface,
	userRepo repository.UserRepositoryInterface,
	gameRepo repository.GameRepositoryInterface,
	logsRepo adminRepo.AdminLogsRepositoryInterface,
	logger *zap.Logger,
) *RedisGameManager {
	return &RedisGameManager{
		redisService: redisService,
		locker:       redislock.New(redisService.NativeClient()),
		userRepo:     userRepo,
		gameRepo:     gameRepo,
		logsRepo:     logsRepo,
		xpCalculator: NewXPCalculator(redisService),
		eloCalc:      NewELOCalculator(redisService),
		logger:       logger,
		eventChan:    make(chan GameEvent, 100),
		ctx:          appCtx,
		archiveQueue: make(chan GameEngine, 500),
		stopWorkers:  make(chan struct{}),
	}
}

func (rgm *RedisGameManager) opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(rgm.ctx, 5*time.Second)
}

func (rgm *RedisGameManager) scanCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(rgm.ctx, 15*time.Second)
}

func (rgm *RedisGameManager) SetUserNotifier(n UserNotifier) {
	rgm.userNotifier = n
}

func (rgm *RedisGameManager) gameKey(gameID uuid.UUID) string {
	return fmt.Sprintf("game:active:%s", gameID.String())
}

func determineMultiplayerWinner(players []Player) *uuid.UUID {
	bestScore := -1
	var winnerID *uuid.UUID
	for _, p := range players {
		if p.Score > bestScore {
			bestScore = p.Score
			id := p.UserID
			winnerID = &id
		} else if p.Score == bestScore {
			winnerID = nil
		}
	}
	return winnerID
}

func buildPlayersFinalPayload(players []Player) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(players))
	for _, p := range players {
		result = append(result, map[string]interface{}{
			"user_public_id": p.PublicID,
			"username":       p.Username,
			"score":          p.Score,
		})
	}
	return result
}

const gameConfigKey = "system:game:config"

const gameLockTTL = 10 * time.Second

func (rgm *RedisGameManager) withGameLock(gameID uuid.UUID, fn func() error) error {
	lockKey := fmt.Sprintf("game:lock:%s", gameID.String())

	ctx, cancel := context.WithTimeout(rgm.ctx, gameLockTTL)
	defer cancel()

	lock, err := rgm.locker.Obtain(ctx, lockKey, gameLockTTL, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 3),
	})
	if errors.Is(err, redislock.ErrNotObtained) {
		return errors.New("game is busy, please retry")
	}
	if err != nil {
		return fmt.Errorf("lock acquisition error: %w", err)
	}
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(rgm.ctx, 2*time.Second)
		defer releaseCancel()
		_ = lock.Release(releaseCtx)
	}()

	return fn()
}

func (rgm *RedisGameManager) gameStateTTL(status model.GameStatus) time.Duration {
	cfg := rgm.loadGameConfig()
	switch status {
	case model.GameStatusCompleted, model.GameStatusCancelled, model.GameStatusAbandoned:
		return cfg.FinishedTTL()
	default:
		return cfg.ActiveTTL()
	}
}

func (rgm *RedisGameManager) loadGameConfig() model.GameConfig {
	ctx, cancel := rgm.opCtx()
	defer cancel()
	var cfg model.GameConfig
	if err := rgm.redisService.GetJSON(ctx, gameConfigKey, &cfg); err != nil ||
		cfg.ActiveTTLMinutes <= 0 || cfg.FinishedTTLMinutes <= 0 {
		cfg = model.DefaultGameConfig()
	}
	return cfg
}

func (rgm *RedisGameManager) loadCountdownConfig() model.CountdownConfig {
	ctx, cancel := rgm.opCtx()
	defer cancel()
	var cfg model.CountdownConfig
	if err := rgm.redisService.GetJSON(ctx, "system:game:countdown", &cfg); err != nil {
		return model.DefaultCountdownConfig()
	}
	defaults := model.DefaultCountdownConfig()
	if cfg.PreGameCountdownSeconds <= 0 {
		cfg.PreGameCountdownSeconds = defaults.PreGameCountdownSeconds
	}
	if cfg.ReconnectGracePeriodSeconds <= 0 {
		cfg.ReconnectGracePeriodSeconds = defaults.ReconnectGracePeriodSeconds
	}
	return cfg
}

type GameStateData struct {
	ID          uuid.UUID         `json:"id"`
	PublicID    string            `json:"public_id"`
	Mode        model.GameMode    `json:"mode"`
	Status      model.GameStatus  `json:"status"`
	Players     []Player          `json:"players"`
	Settings    GameSettings      `json:"settings"`
	Questions   []*model.Question `json:"questions"`
	CurrentQ    int               `json:"current_q"`
	WinnerID    *uuid.UUID        `json:"winner_id,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Category    string            `json:"category,omitempty"`
	FlagVariant string            `json:"flag_variant,omitempty"`
	FoundISOs   []string          `json:"found_isos,omitempty"`
	Paused            bool       `json:"paused,omitempty"`
	PausedAt          *time.Time `json:"paused_at,omitempty"`
	ReconnectDeadline *time.Time `json:"reconnect_deadline,omitempty"`
}

func (rgm *RedisGameManager) SaveGame(game GameEngine) error {
	pausedAt := game.GetPausedAt()
	var pausedAtPtr *time.Time
	if !pausedAt.IsZero() {
		pausedAtPtr = &pausedAt
	}
	state := GameStateData{
		ID:          game.GetID(),
		PublicID:    game.GetPublicID(),
		Mode:        game.GetMode(),
		Status:      game.GetStatus(),
		Players:     game.GetPlayers(),
		Settings:    game.GetSettings(),
		CurrentQ:    game.GetQuestionNumber(),
		WinnerID:    game.GetWinnerID(),
		StartedAt:   game.GetStartedAt(),
		CompletedAt: game.GetCompletedAt(),
		Paused:            game.GetPaused(),
		PausedAt:          pausedAtPtr,
		ReconnectDeadline: game.GetReconnectDeadline(),
	}

	switch g := game.(type) {
	case *VersusGame:
		state.Questions = g.GetQuestions()
	}

	key := rgm.gameKey(game.GetID())
	ctx, cancel := rgm.opCtx()
	defer cancel()
	if err := rgm.redisService.SetJSON(ctx, key, state, rgm.gameStateTTL(game.GetStatus())); err != nil {
		rgm.logger.Error("Failed to save game to Redis",
			zap.String("game_id", game.GetID().String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

func (rgm *RedisGameManager) LoadGame(gameID uuid.UUID) (GameEngine, error) {
	key := rgm.gameKey(gameID)

	ctx, cancel := rgm.opCtx()
	defer cancel()
	var state GameStateData
	if err := rgm.redisService.GetJSON(ctx, key, &state); err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	var game GameEngine
	switch state.Mode {
	case model.GameModeSolo, model.GameMode1v1, model.GameModeMulti:
		game = rgm.reconstructVersusGame(state)
	default:
		return nil, errors.New("invalid game mode")
	}

	return game, nil
}

func (rgm *RedisGameManager) reconstructVersusGame(state GameStateData) *VersusGame {
	settings := state.Settings
	if settings.MaxPlayers == 0 || settings.MinPlayers == 0 {
		switch state.Mode {
		case model.GameModeSolo:
			settings.MaxPlayers = 1
			settings.MinPlayers = 1
			if settings.ScoreMode == "" {
				settings.ScoreMode = ScoreModeTimeBonus
			}
		case model.GameMode1v1:
			settings.MaxPlayers = 2
			settings.MinPlayers = 2
			if settings.ScoreMode == "" {
				settings.ScoreMode = ScoreModeFastestWins
			}
		}
	}

	countdownCfg := rgm.loadCountdownConfig()
	game := NewVersusGame(
		state.ID,
		state.PublicID,
		state.Mode,
		settings,
		state.Questions,
		rgm.logger,
		rgm.logsRepo,
		&countdownCfg,
	)

	game.status = state.Status
	game.currentQ = state.CurrentQ - 1
	if game.currentQ < 0 {
		game.currentQ = 0
	}
	game.winnerID = state.WinnerID
	game.startedAt = state.StartedAt
	game.completedAt = state.CompletedAt
	if state.Paused {
		pausedAt := time.Time{}
		if state.PausedAt != nil {
			pausedAt = *state.PausedAt
		}
		game.SetPausedState(true, pausedAt)
	}
	game.SetReconnectDeadline(state.ReconnectDeadline)

	for _, player := range state.Players {
		playerCopy := player
		game.players[player.UserID] = &playerCopy
	}

	return game
}

func (rgm *RedisGameManager) CreateGame(
	gameID uuid.UUID,
	publicID string,
	mode model.GameMode,
	settings GameSettings,
	questions []*model.Question,
) (GameEngine, error) {
	existsCtx, existsCancel := rgm.opCtx()
	defer existsCancel()
	exists, err := rgm.redisService.Exists(existsCtx, rgm.gameKey(gameID))
	if err != nil {
		return nil, fmt.Errorf("failed to check game existence: %w", err)
	}
	if exists {
		return nil, errors.New("game already exists")
	}

	var gameEngine GameEngine

	countdownCfg := rgm.loadCountdownConfig()

	switch mode {
	case model.GameModeSolo, model.GameMode1v1, model.GameModeMulti:
		gameEngine = NewVersusGame(
			gameID,
			publicID,
			mode,
			settings,
			questions,
			rgm.logger,
			rgm.logsRepo,
			&countdownCfg,
		)
	default:
		return nil, errors.New("invalid game mode")
	}

	if err := rgm.SaveGame(gameEngine); err != nil {
		return nil, err
	}

	rgm.logger.Info("Game created in Redis",
		zap.String("game_id", gameID.String()),
		zap.String("public_id", publicID),
		zap.String("mode", string(mode)),
	)

	rgm.EmitEvent(GameEvent{
		Type:     EventGameCreated,
		GameID:   gameID,
		PublicID: publicID,
		Data: map[string]interface{}{
			"mode": mode,
		},
	})

	return gameEngine, nil
}

func (rgm *RedisGameManager) CreateGameWithPlayers(
	gameID uuid.UUID,
	publicID string,
	mode model.GameMode,
	settings GameSettings,
	questions []*model.Question,
	playerIDs []uuid.UUID,
) (GameEngine, error) {
	engine, err := rgm.CreateGame(gameID, publicID, mode, settings, questions)
	if err != nil {
		return nil, err
	}

	for _, pid := range playerIDs {
		if err := rgm.AddPlayerToGame(gameID, pid); err != nil {
			rgm.logger.Warn("Failed to auto-add player to match", zap.String("game_id", gameID.String()), zap.String("user_id", pid.String()), zap.Error(err))
		}
	}

	return engine, nil
}

func (rgm *RedisGameManager) GetGame(gameID uuid.UUID) (GameEngine, error) {
	return rgm.LoadGame(gameID)
}

func (rgm *RedisGameManager) RemoveGame(gameID uuid.UUID) error {
	game, err := rgm.GetGame(gameID)
	if err == nil {
		game.StopGoroutine()
	}

	key := rgm.gameKey(gameID)
	delCtx, delCancel := rgm.opCtx()
	defer delCancel()
	if err := rgm.redisService.Delete(delCtx, key); err != nil {
		return fmt.Errorf("failed to remove game: %w", err)
	}

	_ = rgm.RemoveQuestionTimeout(gameID)

	rgm.logger.Info("Game removed from Redis",
		zap.String("game_id", gameID.String()),
	)

	return nil
}

func (rgm *RedisGameManager) AddPlayerToGame(gameID, userID uuid.UUID) error {
	return rgm.withGameLock(gameID, func() error {
		return rgm.doAddPlayerToGame(gameID, userID)
	})
}

func (rgm *RedisGameManager) doAddPlayerToGame(gameID, userID uuid.UUID) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		return err
	}

	if err := game.AddPlayer(userID); err != nil {
		return err
	}

	var userPublicID, username string
	if user, err := rgm.userRepo.GetByID(userID.String()); err == nil {
		_ = game.SetPlayerPublicID(userID, user.PublicID)
		_ = game.SetPlayerUsername(userID, user.Username)
		userPublicID = user.PublicID
		username = user.Username
	}

	if err := rgm.SaveGame(game); err != nil {
		return err
	}

	rgm.logger.Info("Player added to game",
		zap.String("game_id", gameID.String()),
		zap.String("user_id", userID.String()),
	)

	rgm.EmitEvent(GameEvent{
		Type:     EventPlayerJoined,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data: map[string]interface{}{
			"user_public_id": userPublicID,
			"username":       username,
		},
	})

	if game.GetStatus() == model.GameStatusReady {
		rgm.EmitEvent(GameEvent{
			Type:     EventGameReady,
			GameID:   gameID,
			PublicID: game.GetPublicID(),
			Data:     map[string]interface{}{},
		})
	}

	return nil
}

func (rgm *RedisGameManager) RemovePlayerFromGame(gameID, userID uuid.UUID) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		return err
	}

	var userPublicID string
	if player, err := game.GetPlayer(userID); err == nil {
		userPublicID = player.PublicID
	}

	if err := game.RemovePlayer(userID); err != nil {
		return err
	}

	if err := rgm.SaveGame(game); err != nil {
		return err
	}

	if game.GetStatus() == model.GameStatusCompleted || game.GetStatus() == model.GameStatusCancelled {
		_ = rgm.RemoveQuestionTimeout(gameID)
	}

	rgm.logger.Info("Player removed from game",
		zap.String("game_id", gameID.String()),
		zap.String("user_id", userID.String()),
	)

	rgm.EmitEvent(GameEvent{
		Type:     EventPlayerLeft,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data: map[string]interface{}{
			"user_public_id": userPublicID,
		},
	})

	return nil
}

func (rgm *RedisGameManager) MarkPlayerDisconnected(gameID, userID uuid.UUID) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		rgm.logger.Warn("MarkPlayerDisconnected: game not found",
			zap.String("game_id", gameID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil
	}
	if err := game.DisconnectPlayer(userID); err != nil {
		rgm.logger.Warn("MarkPlayerDisconnected: DisconnectPlayer failed",
			zap.String("game_id", gameID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil
	}

	shouldPause := game.GetStatus() == model.GameStatusInProgress && game.GetMode() != model.GameModeSolo
	var graceSecs int
	if shouldPause {
		countdownCfg := rgm.loadCountdownConfig()
		graceSecs = countdownCfg.ReconnectGracePeriodSeconds
		now := time.Now()
		deadline := now.Add(time.Duration(graceSecs) * time.Second)
		game.SetPausedState(true, now)
		game.SetReconnectDeadline(&deadline)
		_ = rgm.RemoveQuestionTimeout(gameID)
	}

	if err := rgm.SaveGame(game); err != nil {
		return err
	}

	userPublicID := ""
	if user, err := rgm.userRepo.GetByID(userID.String()); err == nil {
		userPublicID = user.PublicID
	}
	publicID := game.GetPublicID()

	rgm.EmitEvent(GameEvent{
		Type:     EventPlayerDisconnected,
		GameID:   gameID,
		PublicID: publicID,
		Data:     map[string]interface{}{"user_public_id": userPublicID},
	})

	if shouldPause {
		rgm.EmitEvent(GameEvent{
			Type:     EventGamePaused,
			GameID:   gameID,
			PublicID: publicID,
			Data:     map[string]interface{}{keyCountdownSecs: graceSecs},
		})
		go func() {
			time.Sleep(time.Duration(graceSecs) * time.Second)
			rgm.doHandleReconnectTimeout(gameID, userID)
		}()
	}

	return nil
}

func (rgm *RedisGameManager) MarkPlayerReconnected(gameID, userID uuid.UUID) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		rgm.logger.Warn("MarkPlayerReconnected: game not found",
			zap.String("game_id", gameID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil
	}
	if err := game.ReconnectPlayer(userID); err != nil {
		rgm.logger.Warn("MarkPlayerReconnected: ReconnectPlayer failed",
			zap.String("game_id", gameID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil
	}

	wasPaused := game.GetPaused()
	pausedAt := game.GetPausedAt()

	if wasPaused {
		pauseDuration := time.Since(pausedAt)
		game.AdjustQuestionTimeForPause(pauseDuration)
		game.SetPausedState(false, time.Time{})
		game.SetReconnectDeadline(nil)
	}

	if err := rgm.SaveGame(game); err != nil {
		return err
	}

	var questionRemainingMs int64
	if wasPaused {
		if currentQ, qErr := game.GetCurrentQuestion(); qErr == nil && currentQ != nil {
			var latestSentAt time.Time
			for _, p := range game.GetPlayers() {
				if p.CurrentQuestionSentAt.After(latestSentAt) {
					latestSentAt = p.CurrentQuestionSentAt
				}
			}
			if !latestSentAt.IsZero() {
				questionDuration := time.Duration(currentQ.EstimatedSeconds) * time.Second
				remaining := questionDuration - time.Since(latestSentAt)
				if remaining <= 0 {
					remaining = time.Second
				}
				questionRemainingMs = remaining.Milliseconds()
				if err := rgm.RegisterQuestionTimeout(gameID, time.Now().Add(remaining)); err != nil {
					rgm.logger.Warn("Failed to re-register question timeout after reconnect",
						zap.String("game_id", gameID.String()),
						zap.Error(err),
					)
				}
			}
		}
	}

	userPublicID := ""
	if user, err := rgm.userRepo.GetByID(userID.String()); err == nil {
		userPublicID = user.PublicID
	}
	publicID := game.GetPublicID()

	rgm.EmitEvent(GameEvent{
		Type:     EventPlayerReconnected,
		GameID:   gameID,
		PublicID: publicID,
		Data:     map[string]interface{}{"user_public_id": userPublicID},
	})

	if wasPaused {
		rgm.EmitEvent(GameEvent{
			Type:     EventGameResumed,
			GameID:   gameID,
			PublicID: publicID,
			Data: map[string]interface{}{
				"question_remaining_ms": questionRemainingMs,
			},
		})
	}

	return nil
}

func (rgm *RedisGameManager) doHandleReconnectTimeout(gameID, disconnectedUserID uuid.UUID) {
	_ = rgm.withGameLock(gameID, func() error {
		game, err := rgm.GetGame(gameID)
		if err != nil {
			return nil
		}
		if game.GetStatus() != model.GameStatusInProgress {
			return nil
		}

		player, pErr := game.GetPlayer(disconnectedUserID)
		if pErr != nil || player.Status != model.PlayerStatusDisconnected {
			return nil
		}

		game.SetPausedState(false, time.Time{})
		game.SetReconnectDeadline(nil)

		publicID := game.GetPublicID()
		players := game.GetPlayers()

		var activePlayers []Player
		for _, p := range players {
			if p.UserID != disconnectedUserID && p.Status == model.PlayerStatusActive {
				activePlayers = append(activePlayers, p)
			}
		}

		if len(activePlayers) == 0 {
			_ = rgm.RemoveQuestionTimeout(gameID)
			_ = game.Cancel()
			_ = rgm.SaveGame(game)
			rgm.EmitEvent(GameEvent{
				Type:     EventGameCancelled,
				GameID:   gameID,
				PublicID: publicID,
				Data:     map[string]interface{}{"reason": "reconnect_timeout"},
			})
		} else if len(activePlayers) == 1 || game.GetMode() == model.GameMode1v1 {
			_ = rgm.RemoveQuestionTimeout(gameID)
			winnerID := activePlayers[0].UserID
			winnerPublicID := activePlayers[0].PublicID
			_ = game.End(&winnerID)
			_ = rgm.SaveGame(game)
			completedData := map[string]interface{}{
				"game_id":          publicID,
				"winner_public_id": winnerPublicID,
				"players_final":    buildPlayersFinalPayload(players),
				"reason":           "reconnect_timeout",
			}
			rgm.EmitEvent(GameEvent{
				Type:     EventGameCompleted,
				GameID:   gameID,
				PublicID: publicID,
				Data:     completedData,
			})
			rgm.applyPostGameRewards(game, players, &winnerID, game.GetStartedAt())
		} else {
			discoPublicID := ""
			for _, p := range players {
				if p.UserID == disconnectedUserID {
					discoPublicID = p.PublicID
					break
				}
			}
			_ = game.RemovePlayer(disconnectedUserID)
			_ = rgm.SaveGame(game)

			var questionRemainingMs int64
			if currentQ, qErr := game.GetCurrentQuestion(); qErr == nil && currentQ != nil {
				var latestSentAt time.Time
				for _, p := range game.GetPlayers() {
					if p.CurrentQuestionSentAt.After(latestSentAt) {
						latestSentAt = p.CurrentQuestionSentAt
					}
				}
				if !latestSentAt.IsZero() {
					questionDuration := time.Duration(currentQ.EstimatedSeconds) * time.Second
					remaining := questionDuration - time.Since(latestSentAt)
					if remaining <= 0 {
						remaining = time.Second
					}
					questionRemainingMs = remaining.Milliseconds()
					if err := rgm.RegisterQuestionTimeout(gameID, time.Now().Add(remaining)); err != nil {
						rgm.logger.Error("Failed to re-register question timeout after reconnect timeout",
							zap.String("game_id", gameID.String()),
							zap.Error(err),
						)
					}
				}
			}

			rgm.EmitEvent(GameEvent{
				Type:     EventPlayerLeft,
				GameID:   gameID,
				PublicID: publicID,
				Data:     map[string]interface{}{"user_public_id": discoPublicID, "reason": "reconnect_timeout"},
			})
			rgm.EmitEvent(GameEvent{
				Type:     EventGameResumed,
				GameID:   gameID,
				PublicID: publicID,
				Data: map[string]interface{}{
					"question_remaining_ms": questionRemainingMs,
				},
			})
		}

		return nil
	})
}

func (rgm *RedisGameManager) SetPlayerReady(gameID, userID uuid.UUID, ready bool) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		return err
	}

	if err := game.SetPlayerReady(userID, ready); err != nil {
		return err
	}

	if err := rgm.SaveGame(game); err != nil {
		return err
	}

	rgm.logger.Info("Player ready status updated",
		zap.String("game_id", gameID.String()),
		zap.String("user_id", userID.String()),
		zap.Bool("ready", ready),
	)

	var userPublicID string
	if player, err := game.GetPlayer(userID); err == nil {
		userPublicID = player.PublicID
	}

	rgm.EmitEvent(GameEvent{
		Type:     EventPlayerReady,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data: map[string]interface{}{
			"user_public_id": userPublicID,
			"ready":          ready,
		},
	})

	return nil
}

func (rgm *RedisGameManager) StartGame(gameID uuid.UUID) error {
	return rgm.withGameLock(gameID, func() error {
		game, err := rgm.GetGame(gameID)
		if err != nil {
			return err
		}

		if !game.CanStart() {
			return errors.New("game cannot start")
		}

		isMultiplayer := game.GetMode() != model.GameModeSolo

		if isMultiplayer {
			countdownCfg := rgm.loadCountdownConfig()
			countdownSecs := countdownCfg.PreGameCountdownSeconds

			var playerPayloads []map[string]interface{}
			for _, p := range game.GetPlayers() {
				playerPayloads = append(playerPayloads, map[string]interface{}{
					"user_public_id": p.PublicID,
					"username":       p.Username,
				})
			}
			rgm.EmitEvent(GameEvent{
				Type:     EventGameStarting,
				GameID:   gameID,
				PublicID: game.GetPublicID(),
				Data: map[string]interface{}{
					keyCountdownSecs: countdownSecs,
					"players":        playerPayloads,
				},
			})

			go func() {
				time.Sleep(time.Duration(countdownSecs) * time.Second)
				if err := rgm.withGameLock(gameID, func() error {
					freshGame, err := rgm.GetGame(gameID)
					if err != nil {
						return nil
					}
					if freshGame.GetStatus() != model.GameStatusReady {
						return nil
					}
					return rgm.doStartGame(gameID, freshGame)
				}); err != nil {
					rgm.logger.Error("Failed to start game after countdown",
						zap.String("game_id", gameID.String()),
						zap.Error(err),
					)
				}
			}()
			return nil
		}

		return rgm.doStartGame(gameID, game)
	})
}

func (rgm *RedisGameManager) doStartGame(gameID uuid.UUID, game GameEngine) error {
	if err := game.Start(); err != nil {
		rgm.logger.Warn("doStartGame: game.Start failed", zap.String("game_id", gameID.String()), zap.Error(err))
		return err
	}

	type questionTimeSetter interface {
		SetAllPlayersQuestionStartTime(t time.Time)
	}
	if qts, ok := game.(questionTimeSetter); ok {
		qts.SetAllPlayersQuestionStartTime(time.Now())
	}

	if err := rgm.SaveGame(game); err != nil {
		rgm.logger.Warn("doStartGame: SaveGame failed", zap.String("game_id", gameID.String()), zap.Error(err))
		return err
	}

	rgm.logger.Info("Game started",
		zap.String("game_id", gameID.String()),
	)

	rgm.EmitEvent(GameEvent{
		Type:     EventGameStarted,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data: map[string]interface{}{
			"players": buildPlayersFinalPayload(game.GetPlayers()),
		},
	})

	question, err := game.GetCurrentQuestion()
	if err == nil && question != nil {
		estimatedSecs := question.EstimatedSeconds
		if estimatedSecs <= 0 {
			estimatedSecs = 30
		}
		expiresAt := time.Now().Add(time.Duration(estimatedSecs) * time.Second)
		if err := rgm.RegisterQuestionTimeout(gameID, expiresAt); err != nil {
			rgm.logger.Warn("Failed to register question timeout",
				zap.String("game_id", gameID.String()),
				zap.Error(err),
			)
		}

		rgm.EmitEvent(GameEvent{
			Type:     EventQuestionSent,
			GameID:   gameID,
			PublicID: game.GetPublicID(),
			Data: map[string]interface{}{
				"question":        QuestionToPayload(question),
				"question_number": game.GetQuestionNumber(),
				"total_questions": game.GetTotalQuestions(),
				"time_limit":      estimatedSecs,
			},
		})
	}

	return nil
}

func (rgm *RedisGameManager) SubmitAnswer(gameID, userID uuid.UUID, answer Answer) error {
	return rgm.withGameLock(gameID, func() error {
		return rgm.doSubmitAnswer(gameID, userID, answer)
	})
}

func (rgm *RedisGameManager) doSubmitAnswer(gameID, userID uuid.UUID, answer Answer) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		return err
	}

	prevQNumber := game.GetQuestionNumber()
	prevQIndex := prevQNumber - 1

	prevQuestion, _ := game.GetCurrentQuestion()

	if err := game.SubmitAnswer(userID, answer); err != nil {
		return err
	}

	if err := rgm.SaveGame(game); err != nil {
		rgm.logger.Warn("Failed to save game state after answer", zap.Error(err))
	}

	player, playerErr := game.GetPlayer(userID)
	var validatedAnswer Answer
	if playerErr == nil && len(player.Answers) > prevQIndex {
		validatedAnswer = player.Answers[prevQIndex]
	} else {
		validatedAnswer = answer
	}

	resultPayload := BuildAnswerResultPayload(validatedAnswer, prevQuestion)

	settings := game.GetSettings()
	interRoundDelay := time.Duration(settings.InterRoundDelayMs) * time.Millisecond
	newQNumber := game.GetQuestionNumber()
	willAdvance := newQNumber != prevQNumber

	userPublicID := ""
	if playerErr == nil {
		userPublicID = player.PublicID
	}

	answerReceivedData := map[string]interface{}{
		"user_public_id":  userPublicID,
		"question_number": prevQNumber,
		"submitted_slug":  validatedAnswer.GetAnswerSlug(),
		"is_correct":      resultPayload.IsCorrect,
		"points":          resultPayload.Points,
		"time_spent_ms":   resultPayload.TimeSpentMs,
		"correct_answer":  resultPayload.CorrectAnswer,
	}
	if willAdvance && interRoundDelay > 0 {
		answerReceivedData["next_question_at"] = time.Now().Add(interRoundDelay).UnixMilli()
	}
	rgm.EmitEvent(GameEvent{
		Type:     EventAnswerReceived,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data:     answerReceivedData,
	})

	if playerErr == nil {
		rgm.EmitEvent(GameEvent{
			Type:     EventScoreUpdated,
			GameID:   gameID,
			PublicID: game.GetPublicID(),
			Data: map[string]interface{}{
				"user_public_id": userPublicID,
				"score":          player.Score,
			},
		})
	}

	if willAdvance {
		_ = rgm.RemoveQuestionTimeout(gameID)

		for _, p := range game.GetPlayers() {
			rgm.EmitEvent(GameEvent{
				Type:     EventScoreUpdated,
				GameID:   gameID,
				PublicID: game.GetPublicID(),
				Data: map[string]interface{}{
					"user_public_id": p.PublicID,
					"score":          p.Score,
				},
			})
		}

		doAdvance := func(g GameEngine) {
			nextQuestion, qErr := g.GetCurrentQuestion()
			if qErr == nil {
				nextEstimatedSecs := nextQuestion.EstimatedSeconds
				if nextEstimatedSecs <= 0 {
					nextEstimatedSecs = 30
				}
				expiresAt := time.Now().Add(time.Duration(nextEstimatedSecs) * time.Second)
				if err := rgm.RegisterQuestionTimeout(gameID, expiresAt); err != nil {
					rgm.logger.Warn("Failed to register question timeout",
						zap.String("game_id", gameID.String()),
						zap.Error(err),
					)
				}

				rgm.EmitEvent(GameEvent{
					Type:     EventQuestionSent,
					GameID:   gameID,
					PublicID: g.GetPublicID(),
					Data: map[string]interface{}{
						"question":        QuestionToPayload(nextQuestion),
						"question_number": g.GetQuestionNumber(),
						"total_questions": g.GetTotalQuestions(),
						"time_limit":      nextEstimatedSecs,
					},
				})
			} else {
				players := g.GetPlayers()

				var winnerID *uuid.UUID
				if g.GetMode() != model.GameModeSolo {
					winnerID = determineMultiplayerWinner(players)
				}

				_ = g.End(winnerID)
				if saveErr := rgm.SaveGame(g); saveErr != nil {
					rgm.logger.Warn("Failed to save completed game state", zap.Error(saveErr))
				}
				completedData := map[string]interface{}{
					"game_id":       gameID,
					"players_final": buildPlayersFinalPayload(players),
				}
				for _, p := range players {
					if winnerID != nil && p.UserID == *winnerID {
						completedData["winner_public_id"] = p.PublicID
						break
					}
				}
				rgm.EmitEvent(GameEvent{
					Type:     EventGameCompleted,
					GameID:   gameID,
					PublicID: g.GetPublicID(),
					Data:     completedData,
				})
				rgm.applyPostGameRewards(g, players, winnerID, g.GetStartedAt())
			}
		}

		capturedNewQNumber := newQNumber

		if interRoundDelay > 0 {
			go func() {
				time.Sleep(interRoundDelay)
				_ = rgm.withGameLock(gameID, func() error {
					freshGame, err := rgm.GetGame(gameID)
					if err != nil {
						return nil
					}
					if freshGame.GetQuestionNumber() != capturedNewQNumber {
						return nil
					}
					doAdvance(freshGame)
					return nil
				})
			}()
		} else {
			doAdvance(game)
		}
	}

	rgm.logger.Info("Answer submitted",
		zap.String("game_id", gameID.String()),
		zap.String("user_id", userID.String()),
		zap.Bool("is_correct", validatedAnswer.IsCorrect),
		zap.Int("points", validatedAnswer.Points),
	)

	return nil
}

func (rgm *RedisGameManager) CancelGame(gameID uuid.UUID) error {
	game, err := rgm.GetGame(gameID)
	if err != nil {
		return err
	}

	if err := game.Cancel(); err != nil {
		return err
	}

	if err := rgm.SaveGame(game); err != nil {
		return err
	}

	rgm.logger.Info("Game cancelled",
		zap.String("game_id", gameID.String()),
	)

	rgm.EmitEvent(GameEvent{
		Type:     EventGameCancelled,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data:     map[string]interface{}{"reason": "cancelled"},
	})

	return nil
}

func (rgm *RedisGameManager) GetActiveGamesCount() int {
	pattern := "game:active:*"

	keys, err := rgm.scanKeys(pattern)
	if err != nil {
		rgm.logger.Error("Failed to scan active game keys", zap.Error(err))
		return 0
	}

	return len(keys)
}

func (rgm *RedisGameManager) GetActiveGameIDs() []uuid.UUID {
	pattern := "game:active:*"
	keys, err := rgm.scanKeys(pattern)
	if err != nil {
		rgm.logger.Error("Failed to scan active game keys", zap.Error(err))
		return []uuid.UUID{}
	}

	gameIDs := make([]uuid.UUID, 0, len(keys))
	for _, key := range keys {
		parts := key[len("game:active:"):]
		if gameID, err := uuid.Parse(parts); err == nil {
			gameIDs = append(gameIDs, gameID)
		}
	}

	return gameIDs
}

func (rgm *RedisGameManager) GetActiveGamesByMode() map[string]int {
	keys, err := rgm.scanKeys("game:active:*")
	if err != nil {
		rgm.logger.Error("Failed to scan active game keys for mode stats", zap.Error(err))
		return map[string]int{}
	}

	if len(keys) == 0 {
		return map[string]int{}
	}

	ctx, cancel := rgm.scanCtx()
	defer cancel()
	values, err := rgm.redisService.MGet(ctx, keys...)
	if err != nil {
		rgm.logger.Error("Failed to mget active game data for mode stats", zap.Error(err))
		return map[string]int{}
	}

	counts := make(map[string]int)
	for _, raw := range values {
		if raw == "" {
			continue
		}
		var partial struct {
			Mode string `json:"mode"`
		}
		if err := json.Unmarshal([]byte(raw), &partial); err != nil || partial.Mode == "" {
			continue
		}
		counts[partial.Mode]++
	}
	return counts
}

func (rgm *RedisGameManager) scanKeys(pattern string) ([]string, error) {
	ctx, cancel := rgm.scanCtx()
	defer cancel()
	return rgm.redisService.Scan(ctx, pattern)
}

func (rgm *RedisGameManager) EmitEvent(event GameEvent) {
	event.Timestamp = time.Now()
	select {
	case rgm.eventChan <- event:
	default:
		rgm.logger.Warn("Event channel full, dropping event",
			zap.String("event_type", event.Type),
			zap.String("game_id", event.GameID.String()),
		)
	}
}

func (rgm *RedisGameManager) GetEventChannel() <-chan GameEvent {
	return rgm.eventChan
}

func (rgm *RedisGameManager) CleanupFinishedGames() (int, error) {
	gameIDs := rgm.GetActiveGameIDs()

	cleaned := 0
	for _, gameID := range gameIDs {
		game, err := rgm.GetGame(gameID)
		if err != nil {
			continue
		}

		if game.IsFinished() {
			if err := rgm.RemoveGame(gameID); err != nil {
				rgm.logger.Error("Failed to remove finished game",
					zap.String("game_id", gameID.String()),
					zap.Error(err),
				)
				continue
			}
			cleaned++
		}
	}

	if cleaned > 0 {
		rgm.logger.Info("Cleaned up finished games from Redis",
			zap.Int("count", cleaned),
		)
	}

	return cleaned, nil
}

const questionTimeoutZSetKey = "game:question_timeouts"

func (rgm *RedisGameManager) RegisterQuestionTimeout(gameID uuid.UUID, expiresAt time.Time) error {
	ctx, cancel := rgm.opCtx()
	defer cancel()
	score := float64(expiresAt.Unix())
	member := gameID.String()
	err := rgm.redisService.ZAdd(ctx, questionTimeoutZSetKey, cache.ZSetMember{
		Score:  score,
		Member: member,
	})
	if err != nil {
		rgm.logger.Error("Failed to register question timeout",
			zap.String("game_id", gameID.String()),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (rgm *RedisGameManager) RemoveQuestionTimeout(gameID uuid.UUID) error {
	ctx, cancel := rgm.opCtx()
	defer cancel()
	err := rgm.redisService.ZRem(ctx, questionTimeoutZSetKey, gameID.String())
	if err != nil {
		rgm.logger.Warn("Failed to remove question timeout",
			zap.String("game_id", gameID.String()),
			zap.Error(err),
		)
	}
	return nil
}

func (rgm *RedisGameManager) GetExpiredGameIDs() []uuid.UUID {
	ctx, cancel := rgm.opCtx()
	defer cancel()
	now := time.Now()
	nowUnix := now.Unix()

	results, err := rgm.redisService.ZRangeByScore(ctx, questionTimeoutZSetKey, cache.ZRangeByScoreOptions{
		Min:   "0",
		Max:   fmt.Sprintf("%d", nowUnix),
		Count: 100,
	})
	if err != nil {
		rgm.logger.Error("Failed to get expired games from ZSET", zap.Error(err))
		return nil
	}

	gameIDs := make([]uuid.UUID, 0, len(results))
	for _, s := range results {
		if gameID, err := uuid.Parse(s); err == nil {
			gameIDs = append(gameIDs, gameID)
		}
	}
	return gameIDs
}

func (rgm *RedisGameManager) StartQuestionTimeoutChecker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				rgm.checkQuestionTimeouts(now)
			}
		}
	}()
}

func (rgm *RedisGameManager) checkQuestionTimeouts(now time.Time) {
	expiredGameIDs := rgm.GetExpiredGameIDs()
	if len(expiredGameIDs) == 0 {
		return
	}

	for _, gameID := range expiredGameIDs {
		rgm.handleSingleGameTimeout(gameID, now)
	}
}

func (rgm *RedisGameManager) handleSingleGameTimeout(gameID uuid.UUID, now time.Time) {
	if err := rgm.withGameLock(gameID, func() error {
		rgm.doHandleSingleGameTimeout(gameID, now)
		return nil
	}); err != nil {
		rgm.logger.Warn("Failed to acquire lock for question timeout",
			zap.String("game_id", gameID.String()),
			zap.Error(err),
		)
	}
}

func (rgm *RedisGameManager) doHandleSingleGameTimeout(gameID uuid.UUID, now time.Time) {
	game, err := rgm.LoadGame(gameID)
	if err != nil {
		_ = rgm.RemoveQuestionTimeout(gameID)
		return
	}
	if game.GetStatus() != model.GameStatusInProgress {
		_ = rgm.RemoveQuestionTimeout(gameID)
		return
	}

	prevQNumber := game.GetQuestionNumber()
	if !game.CheckAndAdvanceTimeout(now) {
		return
	}

	_ = rgm.RemoveQuestionTimeout(gameID)

	nextQuestion, qErr := game.GetCurrentQuestion()
	gameOver := qErr != nil || nextQuestion == nil

	if gameOver {
		players := game.GetPlayers()
		var winnerID *uuid.UUID
		if game.GetMode() != model.GameModeSolo {
			winnerID = determineMultiplayerWinner(players)
		}
		_ = game.End(winnerID)
	}

	if saveErr := rgm.SaveGame(game); saveErr != nil {
		rgm.logger.Warn("Failed to save game after question timeout",
			zap.String("game_id", gameID.String()),
			zap.Error(saveErr),
		)
	}

	rgm.EmitEvent(GameEvent{
		Type:     EventQuestionTimeout,
		GameID:   gameID,
		PublicID: game.GetPublicID(),
		Data:     map[string]interface{}{"question_number": prevQNumber},
	})

	if !gameOver {
		expiresAt := time.Now().Add(time.Duration(nextQuestion.EstimatedSeconds) * time.Second)
		if err := rgm.RegisterQuestionTimeout(gameID, expiresAt); err != nil {
			rgm.logger.Error("Failed to register question timeout on advance",
				zap.String("game_id", gameID.String()),
				zap.Error(err),
			)
		}

		rgm.EmitEvent(GameEvent{
			Type:     EventQuestionSent,
			GameID:   gameID,
			PublicID: game.GetPublicID(),
			Data: map[string]interface{}{
				"question":        QuestionToPayload(nextQuestion),
				"question_number": game.GetQuestionNumber(),
				"total_questions": game.GetTotalQuestions(),
				"time_limit":      nextQuestion.EstimatedSeconds,
			},
		})
	} else {
		players := game.GetPlayers()
		timeoutCompletedData := map[string]interface{}{
			"game_id":       game.GetPublicID(),
			"players_final": buildPlayersFinalPayload(players),
		}
		if winnerID := game.GetWinnerID(); winnerID != nil {
			for _, p := range players {
				if p.UserID == *winnerID {
					timeoutCompletedData["winner_public_id"] = p.PublicID
					break
				}
			}
		}
		rgm.EmitEvent(GameEvent{
			Type:     EventGameCompleted,
			GameID:   gameID,
			PublicID: game.GetPublicID(),
			Data:     timeoutCompletedData,
		})
		winnerID := game.GetWinnerID()
		rgm.applyPostGameRewards(game, players, winnerID, game.GetStartedAt())
	}
}

func (rgm *RedisGameManager) ArchiveGame(game GameEngine) error {
	winnerID := game.GetWinnerID()
	settings := game.GetSettings()
	players := game.GetPlayers()
	startedAt := game.GetStartedAt()

	creatorID := uuid.UUID{}
	if len(players) > 0 {
		creatorID = players[0].UserID
	}

	createdAt := time.Now()
	if startedAt != nil {
		createdAt = *startedAt
	}

	pgGame := &model.Game{
		ID:               game.GetID(),
		PublicID:         "g-" + game.GetID().String()[:8],
		Mode:             game.GetMode(),
		Status:           model.GameStatusCompleted,
		CreatorID:        creatorID,
		QuestionCount:    settings.QuestionCount,
		PointsPerCorrect: settings.PointsPerCorrect,
		TimeBonus:        settings.TimeBonus,
		DatasetID:        settings.DatasetID,
		WinnerID:         winnerID,
		StartedAt:        startedAt,
		CompletedAt:      game.GetCompletedAt(),
		CreatedAt:        createdAt,
		UpdatedAt:        time.Now(),
	}

	if g, ok := game.(*VersusGame); ok {
		pgGame.PublicID = g.publicID
	}

	isNewGame := false
	existingGame, err := rgm.gameRepo.GetGameByID(pgGame.ID)
	if err != nil {
		if err := rgm.gameRepo.CreateGame(pgGame); err != nil {
			return fmt.Errorf("failed to create game record: %w", err)
		}
		isNewGame = true
	} else {
		existingGame.Status = pgGame.Status
		existingGame.WinnerID = pgGame.WinnerID
		existingGame.CompletedAt = pgGame.CompletedAt
		existingGame.UpdatedAt = pgGame.UpdatedAt
		if err := rgm.gameRepo.UpdateGame(existingGame); err != nil {
			return fmt.Errorf("failed to update game record: %w", err)
		}
	}

	existingPlayers, err := rgm.gameRepo.GetGamePlayers(pgGame.ID)
	if err != nil {
		rgm.logger.Warn("Failed to get existing game players", zap.Error(err))
		existingPlayers = []model.GamePlayer{}
	}

	existingPlayerMap := make(map[uuid.UUID]*model.GamePlayer)
	for i := range existingPlayers {
		existingPlayerMap[existingPlayers[i].UserID] = &existingPlayers[i]
	}

	joinedAt := time.Now()
	if startedAt != nil {
		joinedAt = *startedAt
	}

	for _, p := range players {
		if existingPlayer, exists := existingPlayerMap[p.UserID]; exists {
			if err := rgm.gameRepo.UpdatePlayerScore(existingPlayer.ID, p.Score); err != nil {
				rgm.logger.Warn("Failed to update player score", zap.Error(err))
			}
		} else {
			pgPlayer := &model.GamePlayer{
				GameID:   pgGame.ID,
				UserID:   p.UserID,
				Score:    p.Score,
				IsReady:  true,
				Status:   model.PlayerStatusActive,
				JoinedAt: joinedAt,
			}
			if err := rgm.gameRepo.AddPlayerToGame(pgPlayer); err != nil {
				rgm.logger.Warn("Failed to archive game player", zap.Error(err))
			}
		}
	}

	if isNewGame {
		rgm.applyPostGameRewards(game, players, winnerID, startedAt)
	}

	return nil
}

func (rgm *RedisGameManager) applyPostGameRewards(
	game GameEngine,
	players []Player,
	winnerID *uuid.UUID,
	startedAt *time.Time,
) {
	if rgm.userRepo == nil {
		return
	}

	mode := game.GetMode()

	var durationSecs int64
	if startedAt != nil && game.GetCompletedAt() != nil {
		durationSecs = int64(game.GetCompletedAt().Sub(*startedAt).Seconds())
	}

	xpCtx, xpCancel := rgm.opCtx()
	cfg := rgm.xpCalculator.LoadConfig(xpCtx)
	xpCancel()

	isDrawn := winnerID == nil && mode != model.GameModeSolo

	for _, p := range players {
		isWinner := winnerID != nil && p.UserID == *winnerID

		xp := rgm.xpCalculator.CalculateXPWithConfig(mode, p.Score, isWinner, cfg)
		if xp > 0 {
			if err := rgm.userRepo.AddExperience(p.UserID, xp, cfg); err != nil {
				rgm.logger.Warn("Failed to add experience after game",
					zap.String("user_id", p.UserID.String()),
					zap.Int64("xp", xp),
					zap.Error(err),
				)
			} else if rgm.userNotifier != nil {
				newUser, userErr := rgm.userRepo.GetByID(p.UserID.String())
				notifData := map[string]interface{}{
					"xp_gained": xp,
					"is_winner": isWinner,
					"game_mode": string(mode),
					"public_id": game.GetPublicID(),
				}
				if userErr == nil {
					notifData["new_level"] = newUser.Level
					notifData["new_rank"] = newUser.Rank
					notifData["total_xp"] = newUser.Experience
				}
				if err := rgm.userNotifier.SendToUser(p.UserID, map[string]interface{}{
					"type": "xp_updated",
					keyData: notifData,
				}); err != nil {
					rgm.logger.Warn("Failed to send XP notification",
						zap.String("user_id", p.UserID.String()),
						zap.Error(err),
					)
				}
			}
		}

		if err := rgm.userRepo.UpdateGameStats(p.UserID, isWinner, isDrawn, p.Score, durationSecs, string(mode)); err != nil {
			rgm.logger.Warn("Failed to update game stats after game",
				zap.String("user_id", p.UserID.String()),
				zap.Error(err),
			)
		}

		if rgm.userNotifier != nil {
			_ = rgm.userNotifier.SendToUser(p.UserID, map[string]interface{}{
				"type":      "game_results",
				"public_id": game.GetPublicID(),
				keyData: map[string]interface{}{
					"score":     p.Score,
					"xp_gained": xp,
					"is_winner": isWinner,
				},
			})
		}
	}

	if mode == model.GameMode1v1 {
		if winnerID != nil && len(players) >= 2 {
			rgm.applyEloUpdate(players, *winnerID)
		} else if isDrawn && len(players) == 2 {
			rgm.applyEloUpdateDraw(players)
		}
	}
}

func (rgm *RedisGameManager) applyEloUpdate(players []Player, winnerID uuid.UUID) {
	var winnerPlayer Player
	var loserPlayers []Player
	for _, p := range players {
		if p.UserID == winnerID {
			winnerPlayer = p
		} else {
			loserPlayers = append(loserPlayers, p)
		}
	}
	if len(loserPlayers) == 0 {
		return
	}

	winnerUser, err := rgm.userRepo.GetByID(winnerPlayer.UserID.String())
	if err != nil {
		rgm.logger.Warn("Failed to load winner for ELO update",
			zap.String("user_id", winnerPlayer.UserID.String()),
			zap.Error(err),
		)
		return
	}

	eloCtx, eloCancel := rgm.opCtx()
	cfg := rgm.eloCalc.LoadConfig(eloCtx)
	eloCancel()
	runningWinnerElo := winnerUser.EloRating

	for _, loserPlayer := range loserPlayers {
		loserUser, err := rgm.userRepo.GetByID(loserPlayer.UserID.String())
		if err != nil {
			rgm.logger.Warn("Failed to load loser for ELO update",
				zap.String("user_id", loserPlayer.UserID.String()),
				zap.Error(err),
			)
			continue
		}

		newWinnerElo, newLoserElo := rgm.eloCalc.CalculateElo(
			runningWinnerElo, loserUser.EloRating,
			winnerUser.EloGamesPlayed, loserUser.EloGamesPlayed,
			cfg,
		)
		runningWinnerElo = newWinnerElo

		if err := rgm.userRepo.UpdateEloRating(loserPlayer.UserID, newLoserElo); err != nil {
			rgm.logger.Warn("Failed to update loser ELO",
				zap.String("user_id", loserPlayer.UserID.String()),
				zap.Int("new_elo", newLoserElo),
				zap.Error(err),
			)
		} else {
			rgm.logger.Info("Loser ELO updated",
				zap.String("user_id", loserPlayer.UserID.String()),
				zap.Int("old_elo", loserUser.EloRating),
				zap.Int("new_elo", newLoserElo),
			)
			if rgm.userNotifier != nil {
				if err := rgm.userNotifier.SendToUser(loserPlayer.UserID, map[string]interface{}{
					"type": keyEloUpdated,
					keyData: map[string]interface{}{
						keyEloDelta: newLoserElo - loserUser.EloRating,
						"new_elo":   newLoserElo,
						"result":    "loss",
					},
				}); err != nil {
					rgm.logger.Warn("Failed to send ELO notification to loser",
						zap.String("user_id", loserPlayer.UserID.String()),
						zap.Error(err),
					)
				}
			}
		}
	}

	if err := rgm.userRepo.UpdateEloRating(winnerPlayer.UserID, runningWinnerElo); err != nil {
		rgm.logger.Warn("Failed to update winner ELO",
			zap.String("user_id", winnerPlayer.UserID.String()),
			zap.Int("new_elo", runningWinnerElo),
			zap.Error(err),
		)
	} else {
		rgm.logger.Info("Winner ELO updated",
			zap.String("user_id", winnerPlayer.UserID.String()),
			zap.Int("old_elo", winnerUser.EloRating),
			zap.Int("new_elo", runningWinnerElo),
		)
		if rgm.userNotifier != nil {
			if err := rgm.userNotifier.SendToUser(winnerPlayer.UserID, map[string]interface{}{
				"type": keyEloUpdated,
				keyData: map[string]interface{}{
					keyEloDelta: runningWinnerElo - winnerUser.EloRating,
					"new_elo":   runningWinnerElo,
					"result":    "win",
				},
			}); err != nil {
				rgm.logger.Warn("Failed to send ELO notification to winner",
					zap.String("user_id", winnerPlayer.UserID.String()),
					zap.Error(err),
				)
			}
		}
	}
}

func (rgm *RedisGameManager) applyEloUpdateDraw(players []Player) {
	if len(players) != 2 {
		return
	}

	userA, err := rgm.userRepo.GetByID(players[0].UserID.String())
	if err != nil {
		rgm.logger.Warn("Failed to load player A for draw ELO", zap.String("user_id", players[0].UserID.String()), zap.Error(err))
		return
	}
	userB, err := rgm.userRepo.GetByID(players[1].UserID.String())
	if err != nil {
		rgm.logger.Warn("Failed to load player B for draw ELO", zap.String("user_id", players[1].UserID.String()), zap.Error(err))
		return
	}

	eloCtx, eloCancel := rgm.opCtx()
	cfg := rgm.eloCalc.LoadConfig(eloCtx)
	eloCancel()

	newA, newB := rgm.eloCalc.CalculateEloDraw(
		userA.EloRating, userB.EloRating,
		userA.EloGamesPlayed, userB.EloGamesPlayed,
		cfg,
	)

	for _, pair := range []struct {
		player  Player
		oldElo  int
		newElo  int
	}{
		{players[0], userA.EloRating, newA},
		{players[1], userB.EloRating, newB},
	} {
		if err := rgm.userRepo.UpdateEloRating(pair.player.UserID, pair.newElo); err != nil {
			rgm.logger.Warn("Failed to update ELO on draw",
				zap.String("user_id", pair.player.UserID.String()),
				zap.Error(err),
			)
			continue
		}
		rgm.logger.Info("Draw ELO updated",
			zap.String("user_id", pair.player.UserID.String()),
			zap.Int("old_elo", pair.oldElo),
			zap.Int("new_elo", pair.newElo),
		)
		if rgm.userNotifier != nil {
			_ = rgm.userNotifier.SendToUser(pair.player.UserID, map[string]interface{}{
				"type": keyEloUpdated,
				keyData: map[string]interface{}{
					keyEloDelta: pair.newElo - pair.oldElo,
					"new_elo":   pair.newElo,
					"result":    "draw",
				},
			})
		}
	}
}

func (rgm *RedisGameManager) QueueForArchiving(game GameEngine) {
	select {
	case rgm.archiveQueue <- game:
		rgm.logger.Info("Game queued for async archiving",
			zap.String("game_id", game.GetID().String()),
		)
	default:
		rgm.logger.Warn("Archive queue full, archiving synchronously",
			zap.String("game_id", game.GetID().String()),
		)
		if err := rgm.ArchiveGame(game); err != nil {
			rgm.logger.Error("Synchronous archive failed",
				zap.String("game_id", game.GetID().String()),
				zap.Error(err),
			)
		}
	}
}

func (rgm *RedisGameManager) archiveWithRetry(game GameEngine) {
	const maxAttempts = 3
	var lastErr error
	for attempt := range maxAttempts {
		lastErr = rgm.ArchiveGame(game)
		if lastErr == nil {
			rgm.logger.Info("Game archived successfully",
				zap.String("game_id", game.GetID().String()),
			)
			return
		}
		if attempt < maxAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
		}
	}
	rgm.logger.Error("Failed to archive game after retries",
		zap.String("game_id", game.GetID().String()),
		zap.Int("attempts", maxAttempts),
		zap.Error(lastErr),
	)
}

func (rgm *RedisGameManager) StartArchiveWorker(ctx context.Context) {
	const numWorkers = 2
	worker := func() {
		for {
			select {
			case <-ctx.Done():
				rgm.drainArchiveQueue()
				return
			case <-rgm.stopWorkers:
				rgm.drainArchiveQueue()
				return
			case game := <-rgm.archiveQueue:
				rgm.archiveWithRetry(game)
			}
		}
	}
	for range numWorkers {
		go worker()
	}
}

func (rgm *RedisGameManager) drainArchiveQueue() {
	for {
		select {
		case game := <-rgm.archiveQueue:
			rgm.archiveWithRetry(game)
		default:
			return
		}
	}
}

func (rgm *RedisGameManager) StopWorkers() {
	rgm.stopOnce.Do(func() { close(rgm.stopWorkers) })
}
