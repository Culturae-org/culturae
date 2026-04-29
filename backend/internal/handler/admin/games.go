// backend/internal/handler/admin/games.go

package admin

import (
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/service"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminGamesHandler struct {
	Usecase        *adminUsecase.AdminGameUsecase
	LoggingService service.LoggingServiceInterface
	logger         *zap.Logger
}

func NewAdminGamesHandler(
	usecase *adminUsecase.AdminGameUsecase,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AdminGamesHandler {
	return &AdminGamesHandler{
		Usecase:        usecase,
		LoggingService: loggingService,
		logger:         logger,
	}
}

// -----------------------------------------------------
// Admin Games Handlers
//
// - ListGames
// - GetGameStats
// - GetGameByID
// - GetGamePlayers
// - GetGameQuestions
// - GetGameAnswers
// - AdminCancelGame
// - DeleteGameByID
// - ArchiveGame
// - UnarchiveGame
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
// - GetUserGameHistory
// - GetGameEventLogs
// -----------------------------------------------------

func (gc *AdminGamesHandler) ListGames(c *gin.Context) {
	pagination := pagination.Parse(c, pagination.AdminConfig())
	status := c.Query("status")
	mode := c.Query("mode")
	search := c.Query("search")
	archived := c.Query("archived")

	games, total, err := gc.Usecase.ListGames(status, mode, search, archived, pagination.Limit, pagination.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, games, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (gc *AdminGamesHandler) GetGameStats(c *gin.Context) {
	stats, err := gc.Usecase.GetGameStats()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (gc *AdminGamesHandler) GetGameByID(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	game, err := gc.Usecase.GetGameByID(gameID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, game)
}

func (gc *AdminGamesHandler) GetGamePlayers(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	players, err := gc.Usecase.GetGamePlayers(gameID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, players)
}

func (gc *AdminGamesHandler) GetGameQuestions(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	questions, err := gc.Usecase.GetGameQuestions(gameID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, questions)
}

func (gc *AdminGamesHandler) GetGameAnswers(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	answers, err := gc.Usecase.GetGameAnswers(gameID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, answers)
}

func (gc *AdminGamesHandler) AdminCancelGame(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeInvalidToken, "Invalid admin")
		return
	}

	if err := gc.Usecase.AdminCancelGame(c, gameID, adminID); err != nil {
		errMsg := err.Error()
		_ = gc.LoggingService.LogAdminAction(adminID, c.GetString("username"), "cancel_game", "game", &gameID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"game_id": gameID}, false, &errMsg)
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	_ = gc.LoggingService.LogAdminAction(adminID, c.GetString("username"), "cancel_game", "game", &gameID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"game_id": gameID}, true, nil)
	httputil.SuccessWithMessage(c, http.StatusOK, "Game cancelled by admin", nil)
}

func (gc *AdminGamesHandler) DeleteGameByID(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	gameDetails, err := gc.Usecase.GetGameByID(gameID)
	if err != nil {
		gameDetails = nil
	}

	if err := gc.Usecase.DeleteGame(gameID); err != nil {
		adminID := httputil.GetUserIDFromContext(c)
		adminName := c.GetString("username")
		errorMessage := err.Error()
		details := map[string]interface{}{
			"game_id": gameID,
		}
		if gameDetails != nil {
			details["game_mode"] = gameDetails.Mode
			details["game_status"] = gameDetails.Status
			details["creator_id"] = gameDetails.CreatorID
			details["player_count"] = len(gameDetails.Players)
			details["question_count"] = gameDetails.QuestionCount
		}
		_ = gc.LoggingService.LogAdminAction(adminID, adminName, "delete_game", "game", &gameID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, false, &errorMessage)
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	details := map[string]interface{}{
		"game_id": gameID,
	}
	if gameDetails != nil {
		details["game_mode"] = gameDetails.Mode
		details["game_status"] = gameDetails.Status
		details["creator_id"] = gameDetails.CreatorID
		details["player_count"] = len(gameDetails.Players)
		details["question_count"] = gameDetails.QuestionCount
	}
	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "delete_game", "game", &gameID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)

	httputil.SuccessWithMessage(c, http.StatusOK, "Game deleted successfully", nil)
}

func (gc *AdminGamesHandler) ArchiveGame(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	if err := gc.Usecase.ArchiveGame(gameID); err != nil {
		gc.logger.Error("Failed to archive game", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to archive game")
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	details := map[string]interface{}{"game_id": gameID}
	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "archive_game", "game", &gameID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)

	httputil.SuccessWithMessage(c, http.StatusOK, "Game archived successfully", nil)
}

func (gc *AdminGamesHandler) UnarchiveGame(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	if err := gc.Usecase.UnarchiveGame(gameID); err != nil {
		gc.logger.Error("Failed to unarchive game", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to unarchive game")
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	details := map[string]interface{}{"game_id": gameID}
	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "unarchive_game", "game", &gameID, httputil.GetRealIP(c), httputil.GetUserAgent(c), details, true, nil)

	httputil.SuccessWithMessage(c, http.StatusOK, "Game restored successfully", nil)
}

func (gc *AdminGamesHandler) ListGameInvites(c *gin.Context) {
	pagination := pagination.Parse(c, pagination.AdminConfig())
	status := c.Query("status")

	invites, total, err := gc.Usecase.ListGameInvites(status, pagination.Limit, pagination.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, invites, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (gc *AdminGamesHandler) ListPendingInvites(c *gin.Context) {
	pagination := pagination.Parse(c, pagination.AdminConfig())

	invites, total, err := gc.Usecase.ListPendingInvites(pagination.Limit, pagination.Offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, invites, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (gc *AdminGamesHandler) DeleteGameInvite(c *gin.Context) {
	inviteIDStr := c.Param("inviteID")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid invite ID")
		return
	}

	if err := gc.Usecase.DeleteGameInvite(inviteID); err != nil {
		adminID := httputil.GetUserIDFromContext(c)
		adminName := c.GetString("username")
		errorMessage := err.Error()
		_ = gc.LoggingService.LogAdminAction(adminID, adminName, "delete_game_invite", "game_invite", &inviteID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"invite_id": inviteID,
		}, false, &errorMessage)
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "delete_game_invite", "game_invite", &inviteID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
		"invite_id": inviteID,
	}, true, nil)

	httputil.SuccessWithMessage(c, http.StatusOK, "Game invite deleted successfully", nil)
}

func (gc *AdminGamesHandler) CancelGameInvite(c *gin.Context) {
	inviteIDStr := c.Param("inviteID")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid invite ID")
		return
	}

	if err := gc.Usecase.CancelGameInvite(inviteID); err != nil {
		adminID := httputil.GetUserIDFromContext(c)
		adminName := c.GetString("username")
		errorMessage := err.Error()
		_ = gc.LoggingService.LogAdminAction(adminID, adminName, "cancel_game_invite", "game_invite", &inviteID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"invite_id": inviteID,
		}, false, &errorMessage)
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "cancel_game_invite", "game_invite", &inviteID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
		"invite_id": inviteID,
	}, true, nil)

	httputil.SuccessWithMessage(c, http.StatusOK, "Game invite cancelled successfully", nil)
}

func (gc *AdminGamesHandler) GetGameModeStats(c *gin.Context) {
	stats, err := gc.Usecase.GetGameModeStats()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (gc *AdminGamesHandler) GetDailyGameStats(c *gin.Context) {
	var startDate, endDate *time.Time
	var mode *string

	if start := c.Query("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = &t
		}
	}
	if m := c.Query("mode"); m != "" {
		mode = &m
	}

	stats, err := gc.Usecase.GetDailyGameStats(startDate, endDate, mode)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (gc *AdminGamesHandler) GetUserGameStats(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user ID")
		return
	}

	stats, err := gc.Usecase.GetUserGameStats(userID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (gc *AdminGamesHandler) GetGamePerformanceStats(c *gin.Context) {
	stats, err := gc.Usecase.GetGamePerformanceStats()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (gc *AdminGamesHandler) CleanupAbandonedGames(c *gin.Context) {
	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")

	result, err := gc.Usecase.CleanupAbandonedGames(adminID)
	if err != nil {
		errorMessage := err.Error()
		_ = gc.LoggingService.LogAdminAction(adminID, adminName, "cleanup_abandoned_games", "game", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"action": "cleanup_abandoned_games",
		}, false, &errorMessage)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "cleanup_abandoned_games", "game", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), result, true, nil)

	httputil.Success(c, http.StatusOK, result)
}

func (gc *AdminGamesHandler) RunGameMaintenance(c *gin.Context) {
	adminID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")

	result, err := gc.Usecase.RunGameMaintenance(adminID)
	if err != nil {
		errorMessage := err.Error()
		_ = gc.LoggingService.LogAdminAction(adminID, adminName, "run_game_maintenance", "game", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
			"action": "run_game_maintenance",
		}, false, &errorMessage)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	_ = gc.LoggingService.LogAdminAction(adminID, adminName, "run_game_maintenance", "game", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), result, true, nil)

	httputil.Success(c, http.StatusOK, result)
}

func (gc *AdminGamesHandler) GetUserGameHistory(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user ID")
		return
	}

	pagination := pagination.Parse(c, pagination.AdminConfig())
	status := c.Query("status")
	mode := c.Query("mode")

	history, err := gc.Usecase.GetGameHistory(userID, pagination.Limit, pagination.Offset, status, mode)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, history)
}

func (h *AdminGamesHandler) GetGameEventLogs(c *gin.Context) {
	idStr := c.Param("id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID")
		return
	}

	events, err := h.Usecase.GetGameEventLogs(gameID)
	if err != nil {
		h.logger.Error("Failed to get game event logs", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to get game event logs")
		return
	}

	httputil.Success(c, http.StatusOK, events)
}
