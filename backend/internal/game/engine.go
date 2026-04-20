// backend/internal/game/engine.go

package game

import (
	"time"

	"github.com/Culturae-org/culturae/internal/game/types"
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
)

type GameCommand struct {
	Type    string
	Payload interface{}
}

type GameEvent struct {
	Type      string
	GameID    uuid.UUID
	PublicID  string
	Data      map[string]interface{}
	Timestamp time.Time
}

type SubmitAnswerPayload struct {
	UserID uuid.UUID
	Answer Answer
}

type AddPlayerPayload struct {
	UserID uuid.UUID
}

type GameEngine interface {
	GetID() uuid.UUID
	GetPublicID() string
	GetMode() model.GameMode
	GetStatus() model.GameStatus
	GetPlayers() []Player
	GetSettings() GameSettings

	AddPlayer(userID uuid.UUID) error
	RemovePlayer(userID uuid.UUID) error
	SetPlayerReady(userID uuid.UUID, ready bool) error
	SetPlayerPublicID(userID uuid.UUID, publicID string) error
	SetPlayerUsername(userID uuid.UUID, username string) error
	GetPlayer(userID uuid.UUID) (*Player, error)

	CanStart() bool
	Start() error
	End(winnerID *uuid.UUID) error
	Cancel() error

	GetCurrentQuestion() (*model.Question, error)
	GetQuestionNumber() int
	GetTotalQuestions() int

	SubmitAnswer(userID uuid.UUID, answer Answer) error

	IsFinished() bool
	AllPlayersFinished() bool
	GetWinnerID() *uuid.UUID
	GetStartedAt() *time.Time
	GetCompletedAt() *time.Time

	CheckAndAdvanceTimeout(now time.Time) bool

	DisconnectPlayer(userID uuid.UUID) error
	ReconnectPlayer(userID uuid.UUID) error

	GetPaused() bool
	GetPausedAt() time.Time
	SetPausedState(paused bool, pausedAt time.Time)
	AdjustQuestionTimeForPause(pauseDuration time.Duration)

	GetReconnectDeadline() *time.Time
	SetReconnectDeadline(deadline *time.Time)

	StartGoroutine()
	StopGoroutine()
	SendCommand(cmd GameCommand)
	Events() <-chan GameEvent
	SetSaveCallback(callback func() error)
}

type Player struct {
	ID                    uuid.UUID          `json:"id"`
	UserID                uuid.UUID          `json:"user_id"`
	PublicID              string             `json:"public_id"`
	Username              string             `json:"username"`
	Score                 int                `json:"score"`
	IsReady               bool               `json:"is_ready"`
	Status                model.PlayerStatus `json:"status"`
	Answers               []Answer           `json:"answers"`
	JoinedAt              time.Time          `json:"joined_at"`
	CurrentQuestionSentAt time.Time          `json:"current_question_sent_at"`
}

type Answer = types.Answer

var (
	NewMCQAnswer      = types.NewMCQAnswer
	NewTextAnswer     = types.NewTextAnswer
	NewBooleanAnswer  = types.NewBooleanAnswer
	NewMultipleAnswer = types.NewMultipleAnswer
)

type GameSettings struct {
	QuestionCount      int
	PointsPerCorrect   int
	TimeBonus          bool
	DatasetID          *uuid.UUID
	Category           string
	FlagVariant        string
	Continent          string
	IncludeTerritories bool
	Language           string
	InterRoundDelayMs  int

	MaxPlayers int
	MinPlayers int
	ScoreMode  string
	TemplateID *uuid.UUID
}

func DefaultGameSettings() GameSettings {
	return GameSettings{
		QuestionCount:     10,
		PointsPerCorrect:  100,
		TimeBonus:         true,
		DatasetID:         nil,
		InterRoundDelayMs: 1500,
		MaxPlayers:        1,
		MinPlayers:        1,
		ScoreMode:         ScoreModeTimeBonus,
	}
}

const (
	ScoreModeTimeBonus   = "time_bonus"
	ScoreModeFastestWins = "fastest_wins"
)

const (
	EventCountryFound        = "country_found"
	EventCountryNotFound     = "country_not_found"
	EventCountryAlreadyFound = "country_already_found"
	EventGameCreated        = "game_created"
	EventGameReady          = "game_ready"
	EventGameStarting       = "game_starting"
	EventGameStarted        = "game_started"
	EventGameCompleted      = "game_completed"
	EventGameCancelled      = "game_cancelled"
	EventGamePaused         = "game_paused"
	EventGameResumed        = "game_resumed"
	EventPlayerJoined        = "player_joined"
	EventPlayerLeft          = "player_left"
	EventPlayerReady         = "player_ready"
	EventPlayerDisconnected  = "player_disconnected"
	EventPlayerReconnected   = "player_reconnected"
	EventQuestionSent       = "question_sent"
	EventAnswerReceived  = "answer_received"
	EventScoreUpdated    = "score_updated"
	EventQuestionTimeout = "question_timeout"
)

const (
	CmdStartGame      = "start_game"
	CmdEndGame        = "end_game"
	CmdSubmitAnswer   = "submit_answer"
	CmdAddPlayer      = "add_player"
	CmdRemovePlayer   = "remove_player"
	CmdSetPlayerReady = "set_player_ready"
	CmdCancelGame          = "cancel_game"
	CmdNextQuestion        = "next_question"
	CmdPlayerDisconnected  = "player_disconnected"
	CmdPlayerReconnected   = "player_reconnected"
	CmdReconnectTimeout    = "reconnect_timeout"
)

const MinAnswerTimeMs = 50

type RemovePlayerPayload struct {
	UserID uuid.UUID
}

type ReconnectPlayerPayload struct {
	UserID uuid.UUID
}

type SetPlayerReadyPayload struct {
	UserID uuid.UUID
	Ready  bool
}

type NextQuestionPayload struct {
	QuestionIndex int
}

type UserNotifier interface {
	SendToUser(userID uuid.UUID, message interface{}) error
}
