// backend/internal/game/manager_interface.go

package game

import (
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
)

type GameManagerInterface interface {
	CreateGame(gameID uuid.UUID, publicID string, mode model.GameMode, settings GameSettings, questions []*model.Question) (GameEngine, error)
	CreateGameWithPlayers(gameID uuid.UUID, publicID string, mode model.GameMode, settings GameSettings, questions []*model.Question, playerIDs []uuid.UUID) (GameEngine, error)
	GetGame(gameID uuid.UUID) (GameEngine, error)
	RemoveGame(gameID uuid.UUID) error

	AddPlayerToGame(gameID, userID uuid.UUID) error
	RemovePlayerFromGame(gameID, userID uuid.UUID) error
	SetPlayerReady(gameID, userID uuid.UUID, ready bool) error
	MarkPlayerDisconnected(gameID, userID uuid.UUID) error
	MarkPlayerReconnected(gameID, userID uuid.UUID) error

	StartGame(gameID uuid.UUID) error
	SubmitAnswer(gameID, userID uuid.UUID, answer Answer) error
	CancelGame(gameID uuid.UUID) error

	GetActiveGamesCount() int
	GetActiveGameIDs() []uuid.UUID
	GetActiveGamesByMode() map[string]int
	GetEventChannel() <-chan GameEvent
	CleanupFinishedGames() (int, error)
}
