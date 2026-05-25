// backend/internal/handler/leaderboard.go

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
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
	lbType := c.DefaultQuery("type", "global")
	mode := c.DefaultQuery("mode", "all")
	pag := pagination.Parse(c)

	type leaderboardCache struct {
		Entries []model.LeaderboardEntry `json:"entries"`
	}

	if lbType != "global" && lbType != "daily" && lbType != "weekly" && lbType != "monthly" && lbType != "elo" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid leaderboard type. Use: global, daily, weekly, monthly, elo")
		return
	}

	cacheKey := fmt.Sprintf("leaderboard:%s:%s:%d:%d", lbType, mode, pag.Limit, pag.Offset)
	if lc.redisService != nil {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()
		cached, err := lc.redisService.Get(cacheCtx, cacheKey)
		if err == nil && cached != "" {
			var cachedData leaderboardCache
			if json.Unmarshal([]byte(cached), &cachedData) == nil {
				entries := cachedData.Entries
				if len(entries) == pag.Limit {
					pag.WithTotal(int64(pag.Offset + len(entries) + 1))
				} else {
					pag.WithTotal(int64(pag.Offset + len(entries)))
				}
				httputil.SuccessList(c, entries, &pag)
				return
			}
		}
	}

	entries, err := lc.usecase.GetEntries(lbType, mode, pag.Limit, pag.Offset)
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

	if len(entries) == pag.Limit {
		pag.WithTotal(int64(pag.Offset + len(entries) + 1))
	} else {
		pag.WithTotal(int64(pag.Offset + len(entries)))
	}
	httputil.SuccessList(c, entries, &pag)
}
