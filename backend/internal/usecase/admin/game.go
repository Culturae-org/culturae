// backend/internal/usecase/admin/game.go

package admin

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/Culturae-org/culturae/internal/game"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/repository"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminGameUsecase struct {
	gameRepo       adminRepo.AdminGameRepositoryInterface
	userRepo       repository.UserRepositoryInterface
	questionRepo   repository.QuestionRepositoryInterface
	gameManager    game.GameManagerInterface
	loggingService service.LoggingServiceInterface
	eventLogRepo   adminRepo.GameEventLogRepositoryInterface
	logger         *zap.Logger
}

func NewAdminGameUsecase(
	gameRepo adminRepo.AdminGameRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	questionRepo repository.QuestionRepositoryInterface,
	gameManager game.GameManagerInterface,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AdminGameUsecase {
	return &AdminGameUsecase{
		gameRepo:       gameRepo,
		userRepo:       userRepo,
		questionRepo:   questionRepo,
		gameManager:    gameManager,
		loggingService: loggingService,
		logger:         logger,
	}
}

func (u *AdminGameUsecase) SetEventLogRepo(repo adminRepo.GameEventLogRepositoryInterface) {
	u.eventLogRepo = repo
}

// -----------------------------------------------------
// Admin Game Usecase
//
// - ListGames
// - GetGameStats
// - GetGameByID
// - GetGamePlayers
// - GetGameQuestions
// - GetGameAnswers
// - AdminCancelGame
// - DeleteGame
// - ListGameInvites
// - ListPendingInvites
// - DeleteGameInvite
// - CancelGameInvite
// - GetGameModeStats
// - GetDailyGameStats
// - GetUserGameStats
// - GetGamePerformanceStats
// - CleanupAbandonedGames
// - RunGameMaintenance
// - GetGameHistory
// -----------------------------------------------------

func (u *AdminGameUsecase) ListGames(status, mode, search, archived string, limit, offset int) ([]model.Game, int64, error) {
	return u.gameRepo.ListGamesWithFilters(status, mode, search, archived, limit, offset)
}

func (u *AdminGameUsecase) GetGameStats() (map[string]interface{}, error) {
	stats, err := u.gameRepo.GetGameStats()
	if err != nil {
		return nil, err
	}

	totalGames := stats.ActiveGames + stats.CompletedGames + stats.CancelledGames + stats.AbandonedGames

	return map[string]interface{}{
		keyTotalGames:      totalGames,
		"active_games":     stats.ActiveGames,
		"completed_games":  stats.CompletedGames,
		"cancelled_games":  stats.CancelledGames,
		"abandoned_games":  stats.AbandonedGames,
		"total_players":    stats.TotalPlayers,
		"total_invites":    stats.TotalInvites,
		"pending_invites":  stats.PendingInvites,
		"accepted_invites": stats.AcceptedInvites,
		"rejected_invites": stats.RejectedInvites,
	}, nil
}

func (u *AdminGameUsecase) GetGameByID(gameID uuid.UUID) (*model.Game, error) {
	return u.gameRepo.GetGameByID(gameID)
}

func (u *AdminGameUsecase) GetGamePlayers(gameID uuid.UUID) ([]model.GamePlayer, error) {
	return u.gameRepo.GetGamePlayers(gameID)
}

func (u *AdminGameUsecase) GetGameQuestions(gameID uuid.UUID) ([]model.GameQuestion, error) {
	return u.gameRepo.GetGameQuestions(gameID)
}

func (u *AdminGameUsecase) GetGameAnswers(gameID uuid.UUID) ([]model.GameAnswerDetail, error) {
	rawAnswers, err := u.gameRepo.GetGameAnswers(gameID)
	if err != nil {
		return nil, err
	}

	questionCache := make(map[uuid.UUID]*model.Question)
	for _, a := range rawAnswers {
		if a.QuestionID != nil {
			questionCache[*a.QuestionID] = nil
		}
	}
	for qID := range questionCache {
		q, qErr := u.questionRepo.GetByID(qID)
		if qErr == nil {
			questionCache[qID] = q
		}
	}

	gqSlice, _ := u.gameRepo.GetGameQuestions(gameID)
	gqByOrder := make(map[int]*model.GameQuestion, len(gqSlice))
	for i := range gqSlice {
		gqByOrder[gqSlice[i].OrderNumber] = &gqSlice[i]
	}

	type ia struct {
		id uuid.UUID
		at time.Time
	}
	perPlayer := make(map[uuid.UUID][]ia)
	for _, a := range rawAnswers {
		perPlayer[a.PlayerID] = append(perPlayer[a.PlayerID], ia{a.ID, a.AnsweredAt})
	}
	answerOrder := make(map[uuid.UUID]int, len(rawAnswers))
	for _, list := range perPlayer {
		sort.Slice(list, func(i, j int) bool { return list[i].at.Before(list[j].at) })
		for idx, item := range list {
			answerOrder[item.id] = idx + 1
		}
	}

	details := make([]model.GameAnswerDetail, 0, len(rawAnswers))
	for _, a := range rawAnswers {
		detail := model.GameAnswerDetail{
			ID:         a.ID,
			GameID:     a.GameID,
			PlayerID:   a.PlayerID,
			QuestionID: a.QuestionID,
			AnswerSlug: a.AnswerSlug,
			Data:       a.Data,
			IsCorrect:  a.IsCorrect,
			TimeSpent:  a.TimeSpent,
			Points:     a.Points,
			AnsweredAt: a.AnsweredAt,
		}

		if a.QuestionID != nil {
			if q, ok := questionCache[*a.QuestionID]; ok && q != nil {
				detail.QuestionSlug = q.Slug
				detail.QuestionType = q.QType

				var qI18n map[string]model.QuestionI18n
				if json.Unmarshal(q.I18n, &qI18n) == nil {
					if en, ok := qI18n["en"]; ok {
						if en.Stem != "" {
							detail.QuestionTitle = en.Stem
						} else {
							detail.QuestionTitle = en.Title
						}
					}
				}

				var answers []model.Answer
				if json.Unmarshal(q.Answers, &answers) == nil {
					for _, ans := range answers {
						if ans.IsCorrect {
							detail.CorrectAnswerSlug = ans.Slug
							if i18n, ok := ans.I18n["en"]; ok {
								detail.CorrectAnswerLabel = i18n.Label
							}
						}
						if ans.Slug == a.AnswerSlug {
							if i18n, ok := ans.I18n["en"]; ok {
								detail.AnswerLabel = i18n.Label
							}
						}
					}
				}
			}
		}

		if detail.QuestionType == "" || detail.CorrectAnswerSlug == "" || detail.AnswerLabel == "" {
			var ansData map[string]interface{}
			if json.Unmarshal(a.Data, &ansData) == nil {
				if detail.QuestionType == "" {
					if qt, ok := ansData["question_type"].(string); ok {
						detail.QuestionType = qt
					}
				}
				if detail.CorrectAnswerSlug == "" {
					if cas, ok := ansData["correct_answer_slug"].(string); ok {
						detail.CorrectAnswerSlug = cas
					}
				}
				if detail.AnswerLabel == "" && detail.QuestionType == model.QTypeTextInput {
					if ua, ok := ansData["user_answer"].(string); ok && ua != "" {
						detail.AnswerLabel = ua
					}
				}
			}
		}

		if detail.CorrectAnswerLabel == "" || detail.AnswerLabel == "" || detail.QuestionTitle == "" {
			if order, ok := answerOrder[a.ID]; ok {
				if gq, ok := gqByOrder[order]; ok {
					if detail.QuestionType == "" {
						detail.QuestionType = gq.Type
					}
					var gqData map[string]interface{}
					if json.Unmarshal(gq.Data, &gqData) == nil {
						if ca, ok := gqData["correct_answer"].(map[string]interface{}); ok {
							if detail.CorrectAnswerSlug == "" {
								if slug, ok := ca["slug"].(string); ok {
									detail.CorrectAnswerSlug = slug
								}
							}
							if names, ok := ca["name"].(map[string]interface{}); ok {
								if en, ok := names["en"].(string); ok {
									if detail.CorrectAnswerLabel == "" {
										detail.CorrectAnswerLabel = en
									}
									if detail.QuestionTitle == "" {
										detail.QuestionTitle = en
									}
								}
							}
						}
						if opts, ok := gqData["options"].([]interface{}); ok {
							for _, opt := range opts {
								optMap, ok := opt.(map[string]interface{})
								if !ok {
									continue
								}
								slug, _ := optMap["slug"].(string)
								names, hasNames := optMap["name"].(map[string]interface{})
								if !hasNames {
									continue
								}
								en, _ := names["en"].(string)
								if slug == a.AnswerSlug && detail.AnswerLabel == "" {
									detail.AnswerLabel = en
								}
								if slug == detail.CorrectAnswerSlug && detail.CorrectAnswerLabel == "" {
									detail.CorrectAnswerLabel = en
								}
							}
						}
						if detail.AnswerLabel == "" && gq.Type == model.QTypeTextInput {
							var ansData map[string]interface{}
							if json.Unmarshal(a.Data, &ansData) == nil {
								if ua, ok := ansData["user_answer"].(string); ok && ua != "" {
									detail.AnswerLabel = ua
								}
							}
						}
					}
				}
			}
		}

		details = append(details, detail)
	}

	return details, nil
}

func (u *AdminGameUsecase) AdminCancelGame(c *gin.Context, gameID, adminID uuid.UUID) error {
	gameModel, err := u.gameRepo.GetGameByID(gameID)
	if err != nil {
		return err
	}

	if gameModel.Status == model.GameStatusCompleted || gameModel.Status == model.GameStatusCancelled {
		return errors.New("game is already finished")
	}

	if err := u.gameManager.CancelGame(gameID); err != nil {
		return err
	}

	now := time.Now()
	gameModel.Status = model.GameStatusCancelled
	gameModel.UpdatedAt = now

	if err := u.gameRepo.UpdateGame(gameModel); err != nil {
		return err
	}

	admin, err := u.userRepo.GetByID(adminID.String())
	adminName := "admin"
	if err == nil && admin != nil {
		adminName = admin.Username
	}

	_ = u.loggingService.LogAdminAction(adminID, adminName, "cancel_game", "game", &gameID, httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyGameID:      gameID,
		"game_status":  gameModel.Status,
		"creator_id":   gameModel.CreatorID,
		"player_count": len(gameModel.Players),
		"reason":       "admin cancellation",
	}, true, nil)

	return nil
}

func (u *AdminGameUsecase) DeleteGame(gameID uuid.UUID) error {
	u.logger.Info("Admin deleting game",
		zap.String(keyGameID, gameID.String()),
		zap.String("action", "delete_game"),
	)

	if err := u.gameManager.RemoveGame(gameID); err != nil {
		return err
	}

	return u.gameRepo.DeleteGame(gameID)
}

func (u *AdminGameUsecase) ArchiveGame(gameID uuid.UUID) error {
	u.logger.Info("Admin archiving game",
		zap.String(keyGameID, gameID.String()),
		zap.String("action", "archive_game"),
	)

	return u.gameRepo.ArchiveGame(gameID)
}

func (u *AdminGameUsecase) UnarchiveGame(gameID uuid.UUID) error {
	u.logger.Info("Admin unarchiving game",
		zap.String(keyGameID, gameID.String()),
		zap.String("action", "unarchive_game"),
	)

	return u.gameRepo.UnarchiveGame(gameID)
}

func (u *AdminGameUsecase) ListGameInvites(status string, limit, offset int) ([]model.GameInvite, int64, error) {
	return u.gameRepo.ListGameInvitesWithFilters(status, limit, offset)
}

func (u *AdminGameUsecase) ListPendingInvites(limit, offset int) ([]model.GameInvite, int64, error) {
	return u.ListGameInvites(string(model.GameInviteStatusPending), limit, offset)
}

func (u *AdminGameUsecase) DeleteGameInvite(inviteID uuid.UUID) error {
	return u.gameRepo.DeleteGameInvite(inviteID)
}

func (u *AdminGameUsecase) CancelGameInvite(inviteID uuid.UUID) error {
	return u.gameRepo.UpdateGameInviteStatus(inviteID, model.GameInviteStatusCancelled)
}

func (u *AdminGameUsecase) GetGameModeStats() (map[string]interface{}, error) {
	stats, err := u.gameRepo.GetGameModeStats()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, stat := range stats {
		result[stat.Mode] = stat.Count
	}

	return result, nil
}

func (u *AdminGameUsecase) GetDailyGameStats(startDate, endDate *time.Time, mode *string) ([]map[string]interface{}, error) {
	dailyStats, err := u.gameRepo.GetDailyGameStats(startDate, endDate, mode)
	if err != nil {
		return nil, err
	}

	var stats []map[string]interface{}
	for _, ds := range dailyStats {
		stats = append(stats, map[string]interface{}{
			"date":            ds.Date,
			keyTotalGames:     ds.TotalGames,
			"completed_games": ds.CompletedGames,
			"cancelled_games": ds.CancelledGames,
			"total_players":   ds.TotalPlayers,
		})
	}

	return stats, nil
}

func (u *AdminGameUsecase) GetUserGameStats(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := u.userRepo.GetByID(userID.String())
	if err != nil {
		return nil, err
	}

	gameStats, err := u.userRepo.GetUserGameStats(userID)
	if err != nil || gameStats == nil {
		gameStats = &model.UserGameStats{}
	}

	var bestScore int
	gamePlayers, _ := u.gameRepo.GetUserGamePlayers(userID)
	for _, gp := range gamePlayers {
		if gp.Score > bestScore {
			bestScore = gp.Score
		}
	}

	winRate := 0.0
	if gameStats.TotalGames > 0 {
		winRate = float64(gameStats.GamesWon) / float64(gameStats.TotalGames) * 100
	}

	return map[string]interface{}{
		"user_id":       userID,
		"username":      user.Username,
		keyTotalGames:   gameStats.TotalGames,
		"wins":          gameStats.GamesWon,
		"losses":        gameStats.GamesLost,
		"win_rate":      winRate,
		"total_score":   gameStats.TotalScore,
		"average_score": gameStats.AverageScore,
		"best_score":    bestScore,
	}, nil
}

func (u *AdminGameUsecase) GetGamePerformanceStats() (map[string]interface{}, error) {
	stats, err := u.gameRepo.GetGamePerformanceStats()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"avg_game_duration_seconds": stats.AvgGameDuration,
		"avg_questions_per_game":    stats.AvgQuestionsPerGame,
		"total_questions_used":      stats.TotalQuestionsUsed,
		"avg_players_per_game":      stats.AvgPlayersPerGame,
		"most_popular_mode":         stats.MostPopularMode,
	}, nil
}

func (u *AdminGameUsecase) CleanupAbandonedGames(adminID uuid.UUID) (map[string]interface{}, error) {
	cutoffTime := time.Now().Add(-24 * time.Hour)

	abandonedGames, err := u.gameRepo.FindAbandonedGames(cutoffTime)
	if err != nil {
		return nil, err
	}

	cancelledCount := 0
	for _, g := range abandonedGames {
		g.Status = model.GameStatusAbandoned
		g.UpdatedAt = time.Now()
		if err := u.gameRepo.UpdateGame(&g); err != nil {
			continue
		}
		if err := u.gameManager.CancelGame(g.ID); err != nil {
			continue
		}
		cancelledCount++
	}

	return map[string]interface{}{
		"abandoned_games_cancelled": cancelledCount,
		keyMessage:                   "Cleanup completed successfully",
	}, nil
}

func (u *AdminGameUsecase) RunGameMaintenance(adminID uuid.UUID) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	cutoffTime := time.Now().Add(-30 * 24 * time.Hour)

	oldGames, err := u.gameRepo.FindOldCompletedGames(cutoffTime)
	if err != nil {
		return nil, err
	}

	deletedCount := 0
	for _, g := range oldGames {
		if err := u.gameRepo.DeleteGame(g.ID); err != nil {
			continue
		}
		deletedCount++
	}

	inviteCutoffTime := time.Now().Add(-7 * 24 * time.Hour)

	expiredInvites, err := u.gameRepo.FindExpiredInvites(inviteCutoffTime)
	if err != nil {
		return nil, err
	}

	expiredCount := 0
	for _, invite := range expiredInvites {
		if err := u.gameRepo.UpdateGameInviteStatus(invite.ID, model.GameInviteStatusCancelled); err != nil {
			continue
		}
		expiredCount++
	}

	result["old_games_deleted"] = deletedCount
	result["expired_invites_cancelled"] = expiredCount
	result[keyMessage] = "Maintenance completed successfully"

	return result, nil
}

func (u *AdminGameUsecase) GetGameHistory(userID uuid.UUID, limit, offset int, status, mode string) ([]model.GameHistoryResponse, error) {
	games, err := u.gameRepo.GetUserGameHistory(userID, limit, offset, status, mode)
	if err != nil {
		return nil, err
	}

	history := make([]model.GameHistoryResponse, 0, len(games))
	for _, g := range games {
		var userScore, opponentScore int
		isWinner := false

		for _, p := range g.Players {
			if p.UserID == userID {
				userScore = p.Score
				if g.WinnerID != nil && *g.WinnerID == userID {
					isWinner = true
				}
			} else {
				opponentScore = p.Score
			}
		}

		history = append(history, model.GameHistoryResponse{
			Game:          &g,
			Players:       g.Players,
			UserScore:     userScore,
			OpponentScore: opponentScore,
			IsWinner:      isWinner,
		})
	}

	return history, nil
}

func (u *AdminGameUsecase) GetGameEventLogs(gameID uuid.UUID) ([]model.GameEventLog, error) {
	if u.eventLogRepo == nil {
		return []model.GameEventLog{}, nil
	}
	return u.eventLogRepo.GetEventsByGameID(gameID)
}
