// backend/internal/game/xp_calculator.go

package game

import (
	"context"
	"math"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
)

const xpConfigKey = "system:xp:config"

type XPCalculator struct {
	redisService cache.RedisClientInterface
}

func NewXPCalculator(redisService cache.RedisClientInterface) *XPCalculator {
	return &XPCalculator{redisService: redisService}
}

func (c *XPCalculator) LoadConfig(ctx context.Context) model.XPConfig {
	if c.redisService == nil {
		return model.DefaultXPConfig()
	}
	var cfg model.XPConfig
	if err := c.redisService.GetJSON(ctx, xpConfigKey, &cfg); err != nil {
		return model.DefaultXPConfig()
	}
	if cfg.BaseXP <= 0 || cfg.GrowthRate <= 0 {
		return model.DefaultXPConfig()
	}
	return cfg
}

func (c *XPCalculator) CalculateXPWithConfig(mode model.GameMode, score int, isWinner bool, cfg model.XPConfig) int64 {
	xp := float64(score) * cfg.MultiplierForMode(mode)
	if isWinner && (mode == model.GameMode1v1 || mode == model.GameModeMulti) {
		xp += float64(cfg.WinnerBonus)
	}
	return int64(math.Round(xp))
}

func (c *XPCalculator) CalculateXPWithTemplateMultiplier(mode model.GameMode, score int, isWinner bool, cfg model.XPConfig, templateMultiplier *float64) int64 {
	xp := c.CalculateXPWithConfig(mode, score, isWinner, cfg)
	if templateMultiplier != nil && *templateMultiplier > 0 {
		return int64(math.Round(float64(xp) * *templateMultiplier))
	}
	return xp
}

func (c *XPCalculator) CalculateXP(mode model.GameMode, score int, isWinner bool) int64 {
	return c.CalculateXPWithConfig(mode, score, isWinner, model.DefaultXPConfig())
}

func (c *XPCalculator) CalculateLevel(totalXP int64) int {
	return model.DefaultXPConfig().CalculateLevel(totalXP)
}

func (c *XPCalculator) CalculateRank(totalXP int64) string {
	return RankFromLevel(c.CalculateLevel(totalXP))
}

func RankFromLevel(level int) string {
	cfg := model.DefaultXPConfig()
	return cfg.RankFromLevel(level)
}
