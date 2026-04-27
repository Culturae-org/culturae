// backend/internal/repository/game.go

package repository

import (
	"errors"
	"time"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func populateGameCreatorPublicID(db *gorm.DB, game *model.Game) error {
	if game.CreatorID == uuid.Nil {
		return nil
	}
	var creator model.User
	if err := db.Select("public_id").Where("id = ?", game.CreatorID).First(&creator).Error; err != nil {
		return err
	}
	game.CreatorPublicID = creator.PublicID
	return nil
}

func populateGamesCreatorPublicID(db *gorm.DB, games []model.Game) error {
	for i := range games {
		if games[i].CreatorID != uuid.Nil {
			var creator model.User
			if err := db.Select("public_id").Where("id = ?", games[i].CreatorID).First(&creator).Error; err != nil {
				return err
			}
			games[i].CreatorPublicID = creator.PublicID
		}
	}
	return nil
}

type GameRepositoryInterface interface {
	CreateGame(game *model.Game) error
	GetGameByID(gameID uuid.UUID) (*model.Game, error)
	GetGameByPublicID(publicID string) (*model.Game, error)
	PublicIDExists(publicID string) (bool, error)
	UpdateGame(game *model.Game) error
	GetUserActiveGames(userID uuid.UUID) ([]model.Game, error)
	GetUserGameHistory(userID uuid.UUID, limit, offset int, status, mode string) ([]model.Game, error)
	CountUserGameHistory(userID uuid.UUID, status, mode string) (int64, error)
	GetGamesByStatus(status model.GameStatus, limit, offset int) ([]model.Game, error)
	GetGamesByStatusWithCount(status model.GameStatus, limit, offset int) ([]model.Game, int64, error)
	GetGamesByStatusesWithCount(statuses []model.GameStatus, limit, offset int) ([]model.Game, int64, error)
	AddPlayerToGame(gamePlayer *model.GamePlayer) error
	RemovePlayerFromGame(gameID, userID uuid.UUID) error
	GetGamePlayers(gameID uuid.UUID) ([]model.GamePlayer, error)
	UpdatePlayerScore(gamePlayerID uuid.UUID, score int) error
	UpdatePlayerReady(gameID, userID uuid.UUID, ready bool) error
	GetPlayerReadyStatus(gameID, userID uuid.UUID) (bool, error)
	CreateGameInvite(invite *model.GameInvite) error
	GetGameInviteByID(inviteID uuid.UUID) (*model.GameInvite, error)
	UpdateGameInviteStatus(inviteID uuid.UUID, status model.GameInviteStatus) error
	GetUserGameInvites(userID uuid.UUID, status model.GameInviteStatus) ([]model.GameInvite, error)
	GetPendingInvitesByGameID(gameID uuid.UUID) ([]model.GameInvite, error)
	AddQuestionToGame(gameQuestion *model.GameQuestion) error
	SaveGameAnswer(answer *model.GameAnswer) error
	GetPlayerAnswers(gameID, playerID uuid.UUID) ([]model.GameAnswer, error)
	GetGameAnswers(gameID uuid.UUID) ([]model.GameAnswer, error)
	GetUserGamesByMode(userID uuid.UUID) ([]model.GameModeStats, error)
	GetUserRecentGames(userID uuid.UUID, limit int) ([]model.RecentGameInfo, error)
	GetUserStatsByPeriod(userID uuid.UUID, since time.Time, limit int) (*model.UserStatsByPeriod, error)
	GetLeaderboardByTimeRange(since time.Time, mode string, limit, offset int) ([]model.LeaderboardEntry, error)

	WithTransaction(fn func(txRepo GameRepositoryInterface) error) error
	DeletePlayerFromGame(gameID, userID uuid.UUID) error
	CountPlayersInGame(gameID uuid.UUID) (int64, error)
	UpdatePlayerStatus(gameID, userID uuid.UUID, status string) error
	FindActivePlayersInGame(gameID uuid.UUID, excludeStatus string) ([]model.GamePlayer, error)
	GetGameByIDSimple(gameID uuid.UUID) (*model.Game, error)
	SaveGame(game *model.Game) error
	DeleteGameRecord(game *model.Game) error
	FindPlayersByGame(gameID uuid.UUID) ([]model.GamePlayer, error)
	UpdateUserGamesWon(userID uuid.UUID, mode string) error
	UpdateUserGamesLost(userID uuid.UUID, mode string) error
	SetPlayerReady(gameID, userID uuid.UUID) error
	CancelUserWaitingGames(userID uuid.UUID) ([]uuid.UUID, error)
	GetGameQuestionByOrder(gameID uuid.UUID, orderNumber int) (*model.GameQuestion, error)
}

type GameRepository struct {
	DB     *gorm.DB
	logger *zap.Logger
}

func NewGameRepository(
	db *gorm.DB,
	logger *zap.Logger,
) *GameRepository {
	return &GameRepository{
		DB:     db,
		logger: logger,
	}
}

func (r *GameRepository) CreateGame(game *model.Game) error {
	return r.DB.Create(game).Error
}

func (r *GameRepository) GetGameByID(gameID uuid.UUID) (*model.Game, error) {
	var game model.Game
	err := r.DB.
		Preload("Players").
		Preload("Players.User").
		Preload("Invites").
		Preload("Questions").
		Preload("Answers").
		Where("id = ?", gameID).
		First(&game).Error

	if err != nil {
		return nil, err
	}

	if err := populateGameCreatorPublicID(r.DB, &game); err != nil {
		return nil, err
	}

	return &game, nil
}

func (r *GameRepository) GetGameByPublicID(publicID string) (*model.Game, error) {
	var game model.Game
	err := r.DB.
		Preload("Players").
		Preload("Players.User").
		Preload("Invites").
		Preload("Questions").
		Preload("Questions.Question").
		Preload("Answers").
		Where("public_id = ?", publicID).
		First(&game).Error

	if err != nil {
		return nil, err
	}

	if err := populateGameCreatorPublicID(r.DB, &game); err != nil {
		return nil, err
	}

	return &game, nil
}

func (r *GameRepository) PublicIDExists(publicID string) (bool, error) {
	var count int64
	err := r.DB.Model(&model.Game{}).Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *GameRepository) UpdateGame(game *model.Game) error {
	return r.DB.Save(game).Error
}

func (r *GameRepository) DeleteGame(gameID uuid.UUID) error {
	return r.DB.Delete(&model.Game{}, "id = ?", gameID).Error
}

func (r *GameRepository) ArchiveGame(gameID uuid.UUID) error {
	now := time.Now()
	return r.DB.Model(&model.Game{}).Where("id = ?", gameID).Update("deleted_at", now).Error
}

func (r *GameRepository) UnarchiveGame(gameID uuid.UUID) error {
	return r.DB.Model(&model.Game{}).Where("id = ?", gameID).Update("deleted_at", nil).Error
}

func (r *GameRepository) GetUserActiveGames(userID uuid.UUID) ([]model.Game, error) {
	var games []model.Game

	staleCutoff := time.Now().Add(-30 * time.Minute)

	err := r.DB.
		Joins("JOIN game_players ON games.id = game_players.game_id").
		Where("game_players.user_id = ?", userID).
		Where(
			r.DB.Where("games.status IN ?", []model.GameStatus{
				model.GameStatusReady,
				model.GameStatusInProgress,
			}).Or(
				"games.status = ? AND games.created_at >= ?",
				model.GameStatusWaiting, staleCutoff,
			),
		).
		Preload("Players").
		Preload("Players.User").
		Order("CASE games.status WHEN 'in_progress' THEN 0 WHEN 'ready' THEN 1 ELSE 2 END ASC, games.created_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, err
	}

	if err := populateGamesCreatorPublicID(r.DB, games); err != nil {
		return nil, err
	}

	return games, nil
}

func (r *GameRepository) CancelUserWaitingGames(userID uuid.UUID) ([]uuid.UUID, error) {
	var games []model.Game
	err := r.DB.
		Joins("JOIN game_players ON games.id = game_players.game_id").
		Where("game_players.user_id = ? AND games.status = ?", userID, model.GameStatusWaiting).
		Select("games.id").
		Find(&games).Error
	if err != nil || len(games) == 0 {
		return nil, err
	}

	ids := make([]uuid.UUID, len(games))
	for i, g := range games {
		ids[i] = g.ID
	}

	now := time.Now()
	if err := r.DB.Model(&model.Game{}).
		Where("id IN ? AND status = ?", ids, model.GameStatusWaiting).
		Updates(map[string]interface{}{"status": model.GameStatusCancelled, "updated_at": now}).Error; err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *GameRepository) GetUserGameHistory(userID uuid.UUID, limit, offset int, status, mode string) ([]model.Game, error) {
	var games []model.Game

	query := r.DB.
		Distinct("games.*").
		Joins("JOIN game_players ON games.id = game_players.game_id").
		Where("game_players.user_id = ?", userID).
		Order("games.created_at DESC").
		Preload("Players").
		Preload("Players.User")

	if status != "" {
		query = query.Where("games.status = ?", status)
	}

	if mode != "" {
		query = query.Where("games.mode = ?", mode)
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&games).Error
	if err != nil {
		return nil, err
	}

	if err := populateGamesCreatorPublicID(r.DB, games); err != nil {
		return nil, err
	}

	return games, nil
}

func (r *GameRepository) CountUserGameHistory(userID uuid.UUID, status, mode string) (int64, error) {
	var count int64
	query := r.DB.Model(&model.Game{}).
		Distinct("games.id").
		Joins("JOIN game_players ON games.id = game_players.game_id").
		Where("game_players.user_id = ?", userID)

	if status != "" {
		query = query.Where("games.status = ?", status)
	}

	if mode != "" {
		query = query.Where("games.mode = ?", mode)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *GameRepository) GetGamesByStatus(status model.GameStatus, limit, offset int) ([]model.Game, error) {
	var games []model.Game

	query := r.DB.
		Where("status = ?", status).
		Order("created_at DESC").
		Preload("Players").
		Preload("Players.User")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&games).Error
	if err != nil {
		return nil, err
	}

	if err := populateGamesCreatorPublicID(r.DB, games); err != nil {
		return nil, err
	}

	return games, nil
}

func (r *GameRepository) GetGamesByStatusWithCount(status model.GameStatus, limit, offset int) ([]model.Game, int64, error) {
	var games []model.Game
	var total int64

	r.DB.Model(&model.Game{}).Where("status = ?", status).Count(&total)

	query := r.DB.
		Where("status = ?", status).
		Order("created_at DESC").
		Preload("Players").
		Preload("Players.User")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&games).Error
	if err != nil {
		return nil, 0, err
	}

	if err := populateGamesCreatorPublicID(r.DB, games); err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *GameRepository) GetGamesByStatusesWithCount(statuses []model.GameStatus, limit, offset int) ([]model.Game, int64, error) {
	var games []model.Game
	var total int64

	r.DB.Model(&model.Game{}).Where("status IN ?", statuses).Count(&total)

	query := r.DB.
		Where("status IN ?", statuses).
		Order("created_at DESC").
		Preload("Players").
		Preload("Players.User")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&games).Error
	if err != nil {
		return nil, 0, err
	}

	if err := populateGamesCreatorPublicID(r.DB, games); err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *GameRepository) AddPlayerToGame(gamePlayer *model.GamePlayer) error {
	var count int64
	r.DB.Model(&model.GamePlayer{}).
		Where("game_id = ? AND user_id = ?", gamePlayer.GameID, gamePlayer.UserID).
		Count(&count)

	if count > 0 {
		return errors.New("player already in game")
	}

	return r.DB.Create(gamePlayer).Error
}

func (r *GameRepository) RemovePlayerFromGame(gameID, userID uuid.UUID) error {
	return r.DB.
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Delete(&model.GamePlayer{}).Error
}

func (r *GameRepository) GetGamePlayers(gameID uuid.UUID) ([]model.GamePlayer, error) {
	var players []model.GamePlayer
	err := r.DB.
		Where("game_id = ?", gameID).
		Preload("User").
		Find(&players).Error
	if err != nil {
		return nil, err
	}

	for i := range players {
		if players[i].User != nil {
			players[i].UserPublicID = players[i].User.PublicID
		}
	}

	return players, nil
}

func (r *GameRepository) UpdatePlayerScore(gamePlayerID uuid.UUID, score int) error {
	return r.DB.Model(&model.GamePlayer{}).
		Where("id = ?", gamePlayerID).
		Update("score", score).Error
}

func (r *GameRepository) UpdatePlayerReady(gameID, userID uuid.UUID, ready bool) error {
	return r.DB.Model(&model.GamePlayer{}).
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Update("is_ready", ready).Error
}

func (r *GameRepository) GetPlayerReadyStatus(gameID, userID uuid.UUID) (bool, error) {
	var gamePlayer model.GamePlayer
	err := r.DB.
		Select("is_ready").
		Where("game_id = ? AND user_id = ?", gameID, userID).
		First(&gamePlayer).Error

	if err != nil {
		return false, err
	}

	return gamePlayer.IsReady, nil
}

func (r *GameRepository) CreateGameInvite(invite *model.GameInvite) error {
	var count int64
	r.DB.Model(&model.GameInvite{}).
		Where("game_id = ? AND to_user_id = ? AND status = ?",
			invite.GameID, invite.ToUserID, model.GameInviteStatusPending).
		Count(&count)

	if count > 0 {
		return errors.New("invite already pending for this user")
	}

	return r.DB.Create(invite).Error
}

func (r *GameRepository) GetGameInviteByID(inviteID uuid.UUID) (*model.GameInvite, error) {
	var invite model.GameInvite
	err := r.DB.
		Preload("Game").
		Preload("FromUser").
		Preload("ToUser").
		Where("id = ?", inviteID).
		First(&invite).Error

	if err != nil {
		return nil, err
	}

	if invite.FromUser != nil {
		invite.FromUserPublicID = invite.FromUser.PublicID
	}
	if invite.ToUser != nil {
		invite.ToUserPublicID = invite.ToUser.PublicID
	}

	return &invite, nil
}

func (r *GameRepository) UpdateGameInviteStatus(inviteID uuid.UUID, status model.GameInviteStatus) error {
	return r.DB.Model(&model.GameInvite{}).
		Where("id = ?", inviteID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (r *GameRepository) GetUserGameInvites(userID uuid.UUID, status model.GameInviteStatus) ([]model.GameInvite, error) {
	var invites []model.GameInvite

	query := r.DB.
		Where("to_user_id = ?", userID).
		Preload("Game").
		Preload("FromUser").
		Order("created_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&invites).Error
	if err != nil {
		return nil, err
	}

	for i := range invites {
		if invites[i].FromUser != nil {
			invites[i].FromUserPublicID = invites[i].FromUser.PublicID
		}
	}

	return invites, nil
}

func (r *GameRepository) GetPendingInvitesByGameID(gameID uuid.UUID) ([]model.GameInvite, error) {
	var invites []model.GameInvite
	err := r.DB.
		Where("game_id = ? AND status = ?", gameID, model.GameInviteStatusPending).
		Find(&invites).Error
	return invites, err
}

func (r *GameRepository) DeleteGameInvite(inviteID uuid.UUID) error {
	return r.DB.Delete(&model.GameInvite{}, "id = ?", inviteID).Error
}

func (r *GameRepository) AddQuestionToGame(gameQuestion *model.GameQuestion) error {
	return r.DB.Create(gameQuestion).Error
}

func (r *GameRepository) GetGameQuestionByOrder(gameID uuid.UUID, orderNumber int) (*model.GameQuestion, error) {
	var gq model.GameQuestion
	err := r.DB.
		Where("game_id = ? AND order_number = ?", gameID, orderNumber).
		Preload("Question").
		First(&gq).Error
	if err != nil {
		return nil, err
	}
	return &gq, nil
}

func (r *GameRepository) GetGameQuestions(gameID uuid.UUID) ([]model.GameQuestion, error) {
	var questions []model.GameQuestion
	err := r.DB.
		Where("game_id = ?", gameID).
		Order("order_number ASC").
		Preload("Question").
		Find(&questions).Error
	return questions, err
}

func (r *GameRepository) SaveGameAnswer(answer *model.GameAnswer) error {
	return r.DB.Create(answer).Error
}

func (r *GameRepository) GetPlayerAnswers(gameID, playerID uuid.UUID) ([]model.GameAnswer, error) {
	var answers []model.GameAnswer
	err := r.DB.
		Where("game_id = ? AND player_id = ?", gameID, playerID).
		Order("answered_at ASC").
		Find(&answers).Error
	return answers, err
}

func (r *GameRepository) GetGameAnswers(gameID uuid.UUID) ([]model.GameAnswer, error) {
	var answers []model.GameAnswer
	err := r.DB.
		Where("game_id = ?", gameID).
		Order("answered_at ASC").
		Find(&answers).Error
	return answers, err
}

func (r *GameRepository) GetUserGamesByMode(userID uuid.UUID) ([]model.GameModeStats, error) {
	var stats []model.GameModeStats
	err := r.DB.Raw(`
		SELECT g.mode, COUNT(*) as total_games
		FROM games g
		JOIN game_players gp ON gp.game_id = g.id
		WHERE gp.user_id = ? AND g.status = 'completed'
		GROUP BY g.mode
	`, userID).Scan(&stats).Error
	return stats, err
}

func (r *GameRepository) GetUserRecentGames(userID uuid.UUID, limit int) ([]model.RecentGameInfo, error) {
	var games []model.RecentGameInfo
	err := r.DB.Raw(`
		SELECT g.public_id, g.mode, g.status, gp.score,
			   (g.winner_id = ?) as is_winner, g.completed_at
		FROM games g
		JOIN game_players gp ON gp.game_id = g.id
		WHERE gp.user_id = ? AND g.status = 'completed'
		ORDER BY g.completed_at DESC
		LIMIT ?
	`, userID, userID, limit).Scan(&games).Error
	return games, err
}

func (r *GameRepository) GetUserStatsByPeriod(userID uuid.UUID, since time.Time, limit int) (*model.UserStatsByPeriod, error) {
	type periodRow struct {
		TotalGames   int     `db:"total_games"`
		GamesWon     int     `db:"games_won"`
		GamesLost    int     `db:"games_lost"`
		GamesDrawn   int     `db:"games_drawn"`
		TotalScore   int64   `db:"total_score"`
		AverageScore float64 `db:"average_score"`
		PlayTime     int64   `db:"play_time"`
	}
	var row periodRow
	err := r.DB.Raw(`
		SELECT
			COUNT(*) AS total_games,
			SUM(CASE WHEN g.winner_id = ? THEN 1 ELSE 0 END) AS games_won,
			SUM(CASE WHEN g.winner_id IS NOT NULL AND g.winner_id != ? THEN 1 ELSE 0 END) AS games_lost,
			SUM(CASE WHEN g.winner_id IS NULL AND g.status = 'completed' THEN 1 ELSE 0 END) AS games_drawn,
			COALESCE(SUM(gp.score), 0) AS total_score,
			COALESCE(AVG(gp.score), 0) AS average_score,
			COALESCE(SUM(EXTRACT(EPOCH FROM (g.completed_at - g.started_at))::int), 0) AS play_time
		FROM game_players gp
		JOIN games g ON g.id = gp.game_id
		WHERE gp.user_id = ? AND g.status = 'completed' AND g.completed_at >= ?
	`, userID, userID, userID, since).Scan(&row).Error
	if err != nil {
		return nil, err
	}

	var recentGames []model.RecentGameInfo
	r.DB.Raw(`
		SELECT g.public_id, g.mode, g.status, gp.score,
			   (g.winner_id = ?) as is_winner, g.completed_at
		FROM games g
		JOIN game_players gp ON gp.game_id = g.id
		WHERE gp.user_id = ? AND g.status = 'completed' AND g.completed_at >= ?
		ORDER BY g.completed_at DESC
		LIMIT ?
	`, userID, userID, since, limit).Scan(&recentGames)

	if recentGames == nil {
		recentGames = []model.RecentGameInfo{}
	}

	return &model.UserStatsByPeriod{
		TotalGames:   row.TotalGames,
		GamesWon:     row.GamesWon,
		GamesLost:    row.GamesLost,
		GamesDrawn:   row.GamesDrawn,
		TotalScore:   row.TotalScore,
		AverageScore: row.AverageScore,
		PlayTime:     row.PlayTime,
		RecentGames:  recentGames,
	}, nil
}

func (r *GameRepository) GetLeaderboardByTimeRange(since time.Time, mode string, limit, offset int) ([]model.LeaderboardEntry, error) {
	var entries []model.LeaderboardEntry
	query := `
		SELECT u.username, u.public_id, u.has_avatar,
			   SUM(gp.score) as score, u.elo_rating
		FROM game_players gp
		JOIN games g ON g.id = gp.game_id
		JOIN users u ON u.id = gp.user_id
		WHERE g.status = 'completed' AND g.completed_at >= ? AND u.account_status = 'active' AND u.is_profile_public = true
	`
	args := []interface{}{since}

	if mode != "" && mode != "all" {
		query += " AND g.mode = ?"
		args = append(args, mode)
	}

	query += " GROUP BY u.username, u.public_id, u.has_avatar, u.elo_rating ORDER BY score DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err := r.DB.Raw(query, args...).Scan(&entries).Error
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Rank = offset + i + 1
	}
	return entries, nil
}

func (r *GameRepository) ListGamesWithFilters(status, mode, search, archived string, limit, offset int) ([]model.Game, int64, error) {
	var games []model.Game
	var total int64

	query := r.DB.Model(&model.Game{})

	switch archived {
	case "true":
		query = query.Where("deleted_at IS NOT NULL")
	case "false", "":
		query = query.Where("deleted_at IS NULL")
	}

	if status == "active" {
		query = query.Where("status IN ?", []model.GameStatus{
			model.GameStatusWaiting, model.GameStatusReady, model.GameStatusInProgress,
		})
	} else if status != "" {
		query = query.Where("status = ?", status)
	}
	if mode != "" {
		query = query.Where("mode = ?", mode)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("id::text LIKE ? OR public_id::text LIKE ?", searchPattern, searchPattern)
	}
	query.Count(&total)

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Players").
		Preload("Players.User").
		Find(&games).Error

	if err != nil {
		return nil, 0, err
	}

	if err := populateGamesCreatorPublicID(r.DB, games); err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *GameRepository) GetGameStats() (*model.GameStatsResult, error) {
	var stats model.GameStatsResult

	r.DB.Model(&model.Game{}).Where("status IN ?",
		[]model.GameStatus{model.GameStatusWaiting, model.GameStatusReady, model.GameStatusInProgress}).
		Count(&stats.ActiveGames)
	r.DB.Model(&model.Game{}).Where("status = ?", model.GameStatusCompleted).Count(&stats.CompletedGames)
	r.DB.Model(&model.Game{}).Where("status = ?", model.GameStatusCancelled).Count(&stats.CancelledGames)
	r.DB.Model(&model.Game{}).Where("status = ?", model.GameStatusAbandoned).Count(&stats.AbandonedGames)
	r.DB.Model(&model.GamePlayer{}).Count(&stats.TotalPlayers)
	r.DB.Model(&model.GameInvite{}).Count(&stats.TotalInvites)
	r.DB.Model(&model.GameInvite{}).Where("status = ?", model.GameInviteStatusPending).Count(&stats.PendingInvites)
	r.DB.Model(&model.GameInvite{}).Where("status = ?", model.GameInviteStatusAccepted).Count(&stats.AcceptedInvites)
	r.DB.Model(&model.GameInvite{}).Where("status = ?", model.GameInviteStatusRejected).Count(&stats.RejectedInvites)

	return &stats, nil
}

func (r *GameRepository) ListGameInvitesWithFilters(status string, limit, offset int) ([]model.GameInvite, int64, error) {
	var invites []model.GameInvite
	var total int64

	query := r.DB.Model(&model.GameInvite{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Game").
		Preload("FromUser").
		Preload("ToUser").
		Find(&invites).Error

	if err != nil {
		return nil, 0, err
	}

	for i := range invites {
		if invites[i].FromUser != nil {
			invites[i].FromUserPublicID = invites[i].FromUser.PublicID
		}
		if invites[i].ToUser != nil {
			invites[i].ToUserPublicID = invites[i].ToUser.PublicID
		}
	}

	return invites, total, nil
}

func (r *GameRepository) GetGameModeStats() ([]model.GameModeStatResult, error) {
	var stats []model.GameModeStatResult
	err := r.DB.Model(&model.Game{}).
		Select("mode, COUNT(*) as count").
		Group("mode").
		Find(&stats).Error
	return stats, err
}

func (r *GameRepository) GetDailyGameStats(startDate, endDate *time.Time, mode *string) ([]model.DailyGameStatResult, error) {
	var results []model.DailyGameStatResult

	now := time.Now()
	defaultDays := 30

	if startDate == nil {
		d := now.AddDate(0, 0, -defaultDays)
		startDate = &d
	}
	if endDate == nil {
		endDate = &now
	}

	startOfStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	for d := startOfStart; !d.After(endOfEnd); d = d.AddDate(0, 0, 1) {
		startOfDay := d
		endOfDay := d.Add(24 * time.Hour)

		var stat model.DailyGameStatResult
		stat.Date = d.Format("2006-01-02")

		query := r.DB.Model(&model.Game{}).Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)
		if mode != nil && *mode != "" {
			query = query.Where("mode = ?", *mode)
		}
		query.Count(&stat.TotalGames)

		query = r.DB.Model(&model.Game{}).Where("created_at >= ? AND created_at < ? AND status = ?", startOfDay, endOfDay, model.GameStatusCompleted)
		if mode != nil && *mode != "" {
			query = query.Where("mode = ?", *mode)
		}
		query.Count(&stat.CompletedGames)

		query = r.DB.Model(&model.Game{}).Where("created_at >= ? AND created_at < ? AND status = ?", startOfDay, endOfDay, model.GameStatusCancelled)
		if mode != nil && *mode != "" {
			query = query.Where("mode = ?", *mode)
		}
		query.Count(&stat.CancelledGames)

		playersQuery := r.DB.Model(&model.GamePlayer{}).Joins("JOIN games ON game_players.game_id = games.id").Where("games.created_at >= ? AND games.created_at < ?", startOfDay, endOfDay)
		if mode != nil && *mode != "" {
			playersQuery = playersQuery.Where("games.mode = ?", *mode)
		}
		playersQuery.Count(&stat.TotalPlayers)

		results = append(results, stat)
	}

	return results, nil
}

func (r *GameRepository) GetUserGamePlayers(userID uuid.UUID) ([]model.GamePlayer, error) {
	var gamePlayers []model.GamePlayer
	err := r.DB.
		Where("user_id = ?", userID).
		Joins("JOIN games ON game_players.game_id = games.id").
		Where("games.status = ?", model.GameStatusCompleted).
		Find(&gamePlayers).Error
	return gamePlayers, err
}

func (r *GameRepository) GetGamePerformanceStats() (*model.GamePerformanceResult, error) {
	var stats model.GamePerformanceResult

	var completedGames []model.Game
	r.DB.
		Where("status = ? AND started_at IS NOT NULL AND completed_at IS NOT NULL", model.GameStatusCompleted).
		Find(&completedGames)

	if len(completedGames) > 0 {
		totalDuration := int64(0)
		for _, game := range completedGames {
			duration := game.CompletedAt.Sub(*game.StartedAt).Seconds()
			totalDuration += int64(duration)
		}
		stats.AvgGameDuration = float64(totalDuration) / float64(len(completedGames))
	}

	r.DB.Model(&model.GameQuestion{}).Count(&stats.TotalQuestionsUsed)
	totalGames := int64(0)
	r.DB.Model(&model.Game{}).Count(&totalGames)
	if totalGames > 0 {
		stats.AvgQuestionsPerGame = float64(stats.TotalQuestionsUsed) / float64(totalGames)
	}

	totalPlayers := int64(0)
	r.DB.Model(&model.GamePlayer{}).Count(&totalPlayers)
	if totalGames > 0 {
		stats.AvgPlayersPerGame = float64(totalPlayers) / float64(totalGames)
	}

	var modeStats struct {
		Mode  string
		Count int64
	}
	r.DB.Model(&model.Game{}).
		Select("mode, COUNT(*) as count").
		Group("mode").
		Order("count DESC").
		Limit(1).
		Scan(&modeStats)
	stats.MostPopularMode = modeStats.Mode

	return &stats, nil
}

func (r *GameRepository) FindAbandonedGames(cutoffTime time.Time) ([]model.Game, error) {
	staleInProgressCutoff := time.Now().Add(-15 * time.Minute)
	var games []model.Game
	err := r.DB.
		Where(
			"(status IN ? AND updated_at < ?) OR (status = ? AND updated_at < ?)",
			[]model.GameStatus{model.GameStatusWaiting, model.GameStatusReady}, cutoffTime,
			model.GameStatusInProgress, staleInProgressCutoff,
		).
		Find(&games).Error
	return games, err
}

func (r *GameRepository) FindOldCompletedGames(cutoffTime time.Time) ([]model.Game, error) {
	var games []model.Game
	err := r.DB.
		Where("status = ? AND completed_at < ?", model.GameStatusCompleted, cutoffTime).
		Find(&games).Error
	return games, err
}

func (r *GameRepository) FindExpiredInvites(cutoffTime time.Time) ([]model.GameInvite, error) {
	var invites []model.GameInvite
	err := r.DB.
		Where("status = ? AND created_at < ?", model.GameInviteStatusPending, cutoffTime).
		Find(&invites).Error
	return invites, err
}

func (r *GameRepository) WithTransaction(fn func(txRepo GameRepositoryInterface) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &GameRepository{DB: tx, logger: r.logger}
		return fn(txRepo)
	})
}

func (r *GameRepository) DeletePlayerFromGame(gameID, userID uuid.UUID) error {
	return r.DB.Where("game_id = ? AND user_id = ?", gameID, userID).Delete(&model.GamePlayer{}).Error
}

func (r *GameRepository) CountPlayersInGame(gameID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.GamePlayer{}).Where("game_id = ?", gameID).Count(&count).Error
	return count, err
}

func (r *GameRepository) UpdatePlayerStatus(gameID, userID uuid.UUID, status string) error {
	return r.DB.Model(&model.GamePlayer{}).Where("game_id = ? AND user_id = ?", gameID, userID).Update("status", status).Error
}

func (r *GameRepository) FindActivePlayersInGame(gameID uuid.UUID, excludeStatus string) ([]model.GamePlayer, error) {
	var players []model.GamePlayer
	err := r.DB.Where("game_id = ? AND status != ?", gameID, excludeStatus).Find(&players).Error
	return players, err
}

func (r *GameRepository) GetGameByIDSimple(gameID uuid.UUID) (*model.Game, error) {
	var game model.Game
	err := r.DB.Where("id = ?", gameID).First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *GameRepository) SaveGame(game *model.Game) error {
	return r.DB.Save(game).Error
}

func (r *GameRepository) DeleteGameRecord(game *model.Game) error {
	return r.DB.Delete(game).Error
}

func (r *GameRepository) FindPlayersByGame(gameID uuid.UUID) ([]model.GamePlayer, error) {
	var players []model.GamePlayer
	err := r.DB.Where("game_id = ?", gameID).Find(&players).Error
	return players, err
}

func (r *GameRepository) UpdateUserGamesWon(userID uuid.UUID, mode string) error {
	if mode == string(model.GameModeSolo) {
		return nil
	}
	now := time.Now()
	return r.DB.Exec(`
		INSERT INTO user_game_stats (id, user_id, games_won, total_games, updated_at)
		VALUES (uuid_generate_v4(), ?, 1, 1, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			games_won = user_game_stats.games_won + 1,
			total_games = user_game_stats.total_games + 1,
			updated_at = EXCLUDED.updated_at
	`, userID, now).Error
}

func (r *GameRepository) UpdateUserGamesLost(userID uuid.UUID, mode string) error {
	if mode == string(model.GameModeSolo) {
		return nil
	}
	now := time.Now()
	return r.DB.Exec(`
		INSERT INTO user_game_stats (id, user_id, games_lost, total_games, updated_at)
		VALUES (uuid_generate_v4(), ?, 1, 1, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			games_lost = user_game_stats.games_lost + 1,
			total_games = user_game_stats.total_games + 1,
			updated_at = EXCLUDED.updated_at
	`, userID, now).Error
}

func (r *GameRepository) SetPlayerReady(gameID, userID uuid.UUID) error {
	return r.DB.Model(&model.GamePlayer{}).Where("game_id = ? AND user_id = ?", gameID, userID).Update("is_ready", true).Error
}
