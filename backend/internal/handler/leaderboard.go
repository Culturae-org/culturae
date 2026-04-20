// backend/internal/handler/leaderboard.go

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeaderboardHandler struct {
	usecase      usecase.LeaderboardUsecaseInterface
	redisService cache.RedisClientInterface
}

func NewLeaderboardHandler(
	uc usecase.LeaderboardUsecaseInterface,
	redisService cache.RedisClientInterface,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		usecase:      uc,
		redisService: redisService,
	}
}

// -----------------------------------------------------
// Leaderboard Handlers
//
// - GetLeaderboard
// -----------------------------------------------------

func (lc *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	userID := httputil.GetUserIDFromContext(c)

	lbType := c.DefaultQuery("type", "global")
	mode := c.DefaultQuery("mode", "all")
	limit, offset := httputil.ParsePagination(c, 20, 100)

	type leaderboardCache struct {
		Entries []model.LeaderboardEntry `json:"entries"`
	}

	if lbType != "global" && lbType != "daily" && lbType != "weekly" && lbType != "monthly" && lbType != "elo" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid leaderboard type. Use: global, daily, weekly, monthly, elo")
		return
	}

	cacheKey := "leaderboard:" + lbType + ":" + mode + ":" + strconv.Itoa(limit) + ":" + strconv.Itoa(offset)
	if lc.redisService != nil {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()
		cached, err := lc.redisService.Get(cacheCtx, cacheKey)
		if err == nil && cached != "" {
			var cachedData leaderboardCache
			if json.Unmarshal([]byte(cached), &cachedData) == nil {
				var userRank *model.LeaderboardEntry
				if userID != uuid.Nil {
					userRank, _ = lc.usecase.GetUserRank(userID, lbType)
				}
				httputil.SuccessRaw(c, http.StatusOK, gin.H{
					"data":       cachedData.Entries,
					"pagination": httputil.Pagination{Limit: limit, Offset: offset, HasMore: len(cachedData.Entries) == limit},
					"type":       lbType,
					"mode":       mode,
					"user_rank":  userRank,
				})
				return
			}
		}
	}

	entries, err := lc.usecase.GetEntries(lbType, mode, limit, offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch leaderboard")
		return
	}

	if entries == nil {
		entries = []model.LeaderboardEntry{}
	}

	if lc.redisService != nil {
		if data, err := json.Marshal(leaderboardCache{Entries: entries}); err == nil {
			setCtx, setCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer setCancel()
			_ = lc.redisService.Set(setCtx, cacheKey, string(data), 5*time.Minute)
		}
	}

	var userRank *model.LeaderboardEntry
	if userID != uuid.Nil {
		userRank, _ = lc.usecase.GetUserRank(userID, lbType)
	}

	httputil.SuccessRaw(c, http.StatusOK, gin.H{
		"data":       entries,
		"pagination": httputil.Pagination{Limit: limit, Offset: offset, HasMore: len(entries) == limit},
		"type":       lbType,
		"mode":       mode,
		"user_rank":  userRank,
	})
}
