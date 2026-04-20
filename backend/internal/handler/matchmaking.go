// backend/internal/handler/matchmaking.go

package handler

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/game"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MatchmakingHandler struct {
	Service *game.MatchmakingService
	logger  *zap.Logger
}

func NewMatchmakingHandler(
	service *game.MatchmakingService,
	logger *zap.Logger,
) *MatchmakingHandler {
	return &MatchmakingHandler{
		Service: service,
		logger:  logger,
	}
}

// -----------------------------------------------------
// Matchmaking Handlers
//
// - JoinQueue
// - LeaveQueue
// -----------------------------------------------------

func (mc *MatchmakingHandler) JoinQueue(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	var req struct {
		Mode          model.GameMode `json:"mode" binding:"required"`
		Category      string         `json:"category"`
		FlagVariant   string         `json:"flag_variant"`
		Language      string         `json:"language"`
		QuestionCount int            `json:"question_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if req.Mode != model.GameMode1v1 {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Only 1v1 mode is supported for matchmaking currently")
		return
	}

	gameParams := map[string]interface{}{
		"category":       req.Category,
		"flag_variant":   req.FlagVariant,
		"language":       req.Language,
		"question_count": req.QuestionCount,
	}

	if err := mc.Service.JoinQueue(userID, req.Mode, gameParams); err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Joined matchmaking queue", nil)
}

func (mc *MatchmakingHandler) LeaveQueue(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)
	if userID == uuid.Nil {
		httputil.Error(c, http.StatusUnauthorized, httputil.ErrCodeMissingToken, "Unauthorized")
		return
	}

	var req struct {
		Mode model.GameMode `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	if err := mc.Service.LeaveQueue(userID, req.Mode); err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Left matchmaking queue", nil)
}
