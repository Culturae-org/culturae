// backend/internal/usecase/leaderboard.go

package usecase

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"

	"github.com/google/uuid"
)

type LeaderboardUsecaseInterface interface {
	GetEntries(lbType, mode string, limit, offset int) ([]model.LeaderboardEntry, error)
	GetUserRank(userID uuid.UUID, lbType string) (*model.LeaderboardEntry, error)
}

type LeaderboardUsecase struct {
	userRepo repository.UserRepositoryInterface
	gameRepo repository.GameRepositoryInterface
}

func NewLeaderboardUsecase(
	userRepo repository.UserRepositoryInterface,
	gameRepo repository.GameRepositoryInterface,
) *LeaderboardUsecase {
	return &LeaderboardUsecase{
		userRepo: userRepo,
		gameRepo: gameRepo,
	}
}

func (u *LeaderboardUsecase) GetEntries(lbType, mode string, limit, offset int) ([]model.LeaderboardEntry, error) {
	switch lbType {
	case "global":
		return u.userRepo.GetLeaderboardGlobal(limit, offset)
	case "daily":
		since := time.Now().AddDate(0, 0, -1)
		return u.gameRepo.GetLeaderboardByTimeRange(since, mode, limit, offset)
	case "weekly":
		since := time.Now().AddDate(0, 0, -7)
		return u.gameRepo.GetLeaderboardByTimeRange(since, mode, limit, offset)
	case "monthly":
		since := time.Now().AddDate(0, -1, 0)
		return u.gameRepo.GetLeaderboardByTimeRange(since, mode, limit, offset)
	case "elo":
		return u.userRepo.GetLeaderboardByElo(limit, offset)
	}
	return nil, nil
}

func (u *LeaderboardUsecase) GetUserRank(userID uuid.UUID, lbType string) (*model.LeaderboardEntry, error) {
	user, err := u.userRepo.GetByID(userID.String())
	if err != nil {
		return nil, err
	}

	var rank int
	switch lbType {
	case "elo":
		rank, _ = u.userRepo.GetUserRankByElo(userID)
	default:
		rank, _ = u.userRepo.GetUserRankByScore(userID)
	}

	if rank <= 0 {
		return nil, nil
	}

	var totalScore int64
	if gameStats, err := u.userRepo.GetUserGameStats(userID); err == nil && gameStats != nil {
		totalScore = gameStats.TotalScore
	}

	return &model.LeaderboardEntry{
		Rank:      rank,
		PublicID:  user.PublicID,
		Username:  user.Username,
		HasAvatar: user.HasAvatar,
		Score:     totalScore,
		EloRating: user.EloRating,
	}, nil
}

var _ LeaderboardUsecaseInterface = (*LeaderboardUsecase)(nil)
