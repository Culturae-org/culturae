// backend/internal/game/elo.go

package game

import (
	"context"
	"math"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
)

const eloConfigKey = "system:elo:config"

type ELOCalculator struct {
	redisService cache.RedisClientInterface
}

func NewELOCalculator(redisService cache.RedisClientInterface) *ELOCalculator {
	return &ELOCalculator{redisService: redisService}
}

func (e *ELOCalculator) LoadConfig(ctx context.Context) model.ELOConfig {
	if e.redisService == nil {
		return model.DefaultELOConfig()
	}
	var cfg model.ELOConfig
	if err := e.redisService.GetJSON(ctx, eloConfigKey, &cfg); err != nil {
		return model.DefaultELOConfig()
	}
	if cfg.KFactorLowGames <= 0 || cfg.KFactorHighGames <= 0 || cfg.KFactorThreshold <= 0 {
		return model.DefaultELOConfig()
	}
	return cfg
}

func (e *ELOCalculator) expectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}

func (e *ELOCalculator) kFactor(gamesPlayed int, cfg model.ELOConfig) float64 {
	if gamesPlayed < cfg.KFactorThreshold {
		return float64(cfg.KFactorLowGames)
	}
	return float64(cfg.KFactorHighGames)
}

func (e *ELOCalculator) CalculateEloDraw(ratingA, ratingB, gamesA, gamesB int, cfg model.ELOConfig) (int, int) {
	expectedA := e.expectedScore(ratingA, ratingB)
	expectedB := e.expectedScore(ratingB, ratingA)

	kA := e.kFactor(gamesA, cfg)
	kB := e.kFactor(gamesB, cfg)

	newA := float64(ratingA) + kA*(0.5-expectedA)
	newB := float64(ratingB) + kB*(0.5-expectedB)

	na := int(math.Round(newA))
	nb := int(math.Round(newB))

	if cfg.MinRating > 0 && na < cfg.MinRating {
		na = cfg.MinRating
	}
	if cfg.MinRating > 0 && nb < cfg.MinRating {
		nb = cfg.MinRating
	}
	if cfg.MaxRating > 0 && na > cfg.MaxRating {
		na = cfg.MaxRating
	}
	if cfg.MaxRating > 0 && nb > cfg.MaxRating {
		nb = cfg.MaxRating
	}

	return na, nb
}

func (e *ELOCalculator) CalculateElo(winnerRating, loserRating, winnerGames, loserGames int, cfg model.ELOConfig) (int, int) {
	expectedWinner := e.expectedScore(winnerRating, loserRating)
	expectedLoser := e.expectedScore(loserRating, winnerRating)

	kWinner := e.kFactor(winnerGames, cfg)
	kLoser := e.kFactor(loserGames, cfg)

	newWinner := float64(winnerRating) + kWinner*(1.0-expectedWinner)
	newLoser := float64(loserRating) + kLoser*(0.0-expectedLoser)

	nw := int(math.Round(newWinner))
	nl := int(math.Round(newLoser))

	if cfg.MinRating > 0 && nw < cfg.MinRating {
		nw = cfg.MinRating
	}
	if cfg.MinRating > 0 && nl < cfg.MinRating {
		nl = cfg.MinRating
	}
	if cfg.MaxRating > 0 && nw > cfg.MaxRating {
		nw = cfg.MaxRating
	}
	if cfg.MaxRating > 0 && nl > cfg.MaxRating {
		nl = cfg.MaxRating
	}

	return nw, nl
}
