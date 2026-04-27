// backend/internal/handler/games.go

package handler

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/validation"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GamesHandler struct {
	Usecase *usecase.GameUsecase
	logger  *zap.Logger
}

func NewGamesHandler(
	usecase *usecase.GameUsecase,
	logger *zap.Logger,
) *GamesHandler {
	return &GamesHandler{
		Usecase: usecase,
		logger:  logger,
	}
}

// -----------------------------------------------------
// Games Handlers
//
// - CreateGame
// - InviteToGame
// - AcceptGameInvite
// - RejectGameInvite
// - LeaveGame
// - StartGame
// - SubmitAnswer
// - GetGameStatus
// - GetActiveGames
// - GetGameHistory
// - GetUserGameInvites
// - CancelGame
// - GetGameTemplates
// - GetGameResults
// - CancelUserGameInvite
// -----------------------------------------------------

func (gc *GamesHandler) CreateGame(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	var req model.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if req.QuestionCount != nil && (*req.QuestionCount < 1 || *req.QuestionCount > 50) {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "QuestionCount must be between 1 and 50")
		return
	}

	game, err := gc.Usecase.CreateGame(c, userID, req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusCreated, "Game created successfully", game)
}

func (gc *GamesHandler) InviteToGame(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	gamePublicID := c.Param("gameID")
	if len(gamePublicID) == 0 || len(gamePublicID) > 20 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID format")
		return
	}
	game, err := gc.Usecase.GetGameByPublicID(gamePublicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, "Game not found")
		return
	}

	toUserPublicID := c.Param("toUserPublicID")
	if toUserPublicID == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeMissingField, "User public ID is required")
		return
	}
	sanitizedToUserID, err := validation.SanitizeSlug(toUserPublicID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid user public ID format")
		return
	}

	invite, err := gc.Usecase.InviteToGame(c, game.ID, userID, sanitizedToUserID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusCreated, "Invitation sent successfully", invite)
}

func (gc *GamesHandler) AcceptGameInvite(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	inviteIDStr := c.Param("inviteID")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid invite ID")
		return
	}

	invite, err := gc.Usecase.AcceptGameInvite(c, inviteID, userID)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	gamePublicID := ""
	if invite != nil && invite.Game != nil {
		gamePublicID = invite.Game.PublicID
	}
	httputil.SuccessWithMessage(c, http.StatusOK, "Invitation accepted", map[string]string{"game_public_id": gamePublicID})
}

func (gc *GamesHandler) RejectGameInvite(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	inviteIDStr := c.Param("inviteID")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid invite ID")
		return
	}

	if err := gc.Usecase.RejectGameInvite(c, inviteID, userID); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Invitation rejected", nil)
}

func (gc *GamesHandler) JoinGameByCode(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	var req model.JoinGameByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "code is required")
		return
	}

	if err := gc.Usecase.JoinGameByCode(c, req.Code, userID); err != nil {
		switch err.Error() {
		case "game not found":
			httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, err.Error())
		case "game is full":
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeGameFull, err.Error())
		case "you are already in this game":
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, err.Error())
		default:
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		}
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "joined game"})
}

func (gc *GamesHandler) LeaveGame(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	gamePublicID := c.Param("gameID")
	if len(gamePublicID) == 0 || len(gamePublicID) > 20 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID format")
		return
	}
	game, err := gc.Usecase.GetGameByPublicID(gamePublicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, "Game not found")
		return
	}

	if err := gc.Usecase.LeaveGame(c, game.ID, userID); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Left game successfully", nil)
}

func (gc *GamesHandler) StartGame(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	gamePublicID := c.Param("gameID")
	if len(gamePublicID) > 15 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID format")
		return
	}
	game, err := gc.Usecase.GetGameByPublicID(gamePublicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, "Game not found")
		return
	}

	if err := gc.Usecase.StartGame(c, game.ID, userID); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Ready to start", nil)
}

func (gc *GamesHandler) SubmitAnswer(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	gamePublicID := c.Param("gameID")
	if len(gamePublicID) == 0 || len(gamePublicID) > 20 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID format")
		return
	}
	game, err := gc.Usecase.GetGameByPublicID(gamePublicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, "Game not found")
		return
	}

	var req model.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if req.TimeSpent < 0 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "TimeSpent cannot be negative")
		return
	}

	sanitizedAnswerSlug, err := validation.SanitizeSlug(req.AnswerSlug)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid answer slug format")
		return
	}
	req.AnswerSlug = sanitizedAnswerSlug

	if err := gc.Usecase.SubmitAnswer(c, game.ID, userID, req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Answer submitted", nil)
}

func (gc *GamesHandler) GetGameStatus(c *gin.Context) {
	gamePublicID := c.Param("gameID")
	if len(gamePublicID) == 0 || len(gamePublicID) > 20 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID format")
		return
	}
	game, err := gc.Usecase.GetGameByPublicID(gamePublicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, "Game not found")
		return
	}

	status, err := gc.Usecase.GetGameStatus(game.ID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, status)
}

func (gc *GamesHandler) GetActiveGames(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	games, err := gc.Usecase.GetActiveGames(userID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, games)
}

func (gc *GamesHandler) GetGameHistory(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	limit, offset := httputil.ParsePagination(c, 20, 100)

	total, err := gc.Usecase.CountGameHistory(userID, "", "")
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	history, err := gc.Usecase.GetGameHistory(userID, limit, offset, "", "")
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.SuccessList(c, history, httputil.ParamsToPagination(total, limit, offset))
}

func (gc *GamesHandler) GetUserGameInvites(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	statusStr := c.DefaultQuery("status", "pending")
	var status model.GameInviteStatus
	switch statusStr {
	case "pending":
		status = model.GameInviteStatusPending
	case "accepted":
		status = model.GameInviteStatusAccepted
	case "rejected":
		status = model.GameInviteStatusRejected
	case "cancelled":
		status = model.GameInviteStatusCancelled
	case "all":
		status = ""
	default:
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid status. Use: pending, accepted, rejected, cancelled, all")
		return
	}

	invites, err := gc.Usecase.GetUserGameInvites(userID, status)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch invites")
		return
	}

	httputil.Success(c, http.StatusOK, invites)
}

func (gc *GamesHandler) CancelGame(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	gamePublicID := c.Param("gameID")
	if len(gamePublicID) > 15 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid game ID format")
		return
	}
	game, err := gc.Usecase.GetGameByPublicID(gamePublicID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeGameNotFound, "Game not found")
		return
	}

	if err := gc.Usecase.CancelGame(c, game.ID, userID); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Game cancelled", nil)
}

func (gc *GamesHandler) GetGameTemplates(c *gin.Context) {
	templates, err := gc.Usecase.ListActiveTemplates()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch game templates")
		return
	}

	httputil.Success(c, http.StatusOK, templates)
}

func (gc *GamesHandler) GetGameResults(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	gameID := c.Param("gameID")
	results, err := gc.Usecase.GetGameResults(gameID, userID)
	if err != nil {
		switch err.Error() {
		case "game not found":
			httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, err.Error())
		case "game is not completed":
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		case "forbidden":
			httputil.Error(c, http.StatusForbidden, httputil.ErrCodeForbidden, "You are not a player in this game")
		default:
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch results")
		}
		return
	}

	httputil.Success(c, http.StatusOK, results)
}

func (gc *GamesHandler) CancelUserGameInvite(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	inviteID, err := uuid.Parse(c.Param("inviteID"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid invite ID")
		return
	}

	if err := gc.Usecase.CancelUserGameInvite(c, inviteID, userID); err != nil {
		switch err.Error() {
		case "invite not found":
			httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, err.Error())
		case "forbidden":
			httputil.Error(c, http.StatusForbidden, httputil.ErrCodeForbidden, "You cannot cancel this invite")
		case "invite is not pending":
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		default:
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to cancel invite")
		}
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "invite cancelled"})
}
