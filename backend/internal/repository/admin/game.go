// backend/internal/repository/admin/game.go

package admin

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
)

// AdminGameRepositoryInterface contains the methods needed by admin game operations.
// The concrete *repository.GameRepository satisfies both this interface and
// repository.GameRepositoryInterface.
type AdminGameRepositoryInterface interface {
	// Shared with GameRepositoryInterface
	GetGameByID(gameID uuid.UUID) (*model.Game, error)
	UpdateGame(game *model.Game) error
	UpdateGameInviteStatus(inviteID uuid.UUID, status model.GameInviteStatus) error
	GetUserGameHistory(userID uuid.UUID, limit, offset int, status, mode string) ([]model.Game, error)
	GetGamePlayers(gameID uuid.UUID) ([]model.GamePlayer, error)

	// Admin-only methods
	ListGamesWithFilters(status, mode, search, archived string, limit, offset int) ([]model.Game, int64, error)
	GetGameStats() (*model.GameStatsResult, error)
	GetGameQuestions(gameID uuid.UUID) ([]model.GameQuestion, error)
	GetGameAnswers(gameID uuid.UUID) ([]model.GameAnswer, error)
	ListGameInvitesWithFilters(status string, limit, offset int) ([]model.GameInvite, int64, error)
	GetGameModeStats() ([]model.GameModeStatResult, error)
	GetDailyGameStats(startDate, endDate *time.Time, mode *string) ([]model.DailyGameStatResult, error)
	GetUserGamePlayers(userID uuid.UUID) ([]model.GamePlayer, error)
	GetGamePerformanceStats() (*model.GamePerformanceResult, error)
	FindAbandonedGames(cutoffTime time.Time) ([]model.Game, error)
	FindOldCompletedGames(cutoffTime time.Time) ([]model.Game, error)
	FindExpiredInvites(cutoffTime time.Time) ([]model.GameInvite, error)
	DeleteGame(gameID uuid.UUID) error
	DeleteGameInvite(inviteID uuid.UUID) error
	ArchiveGame(gameID uuid.UUID) error
	UnarchiveGame(gameID uuid.UUID) error
}
