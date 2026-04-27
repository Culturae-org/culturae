// backend/internal/model/game.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type GameMode string

const (
	GameModeSolo  GameMode = "solo"
	GameMode1v1   GameMode = "1v1"
	GameModeMulti GameMode = "multi"
)

type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusReady      GameStatus = "ready"
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusCompleted  GameStatus = "completed"
	GameStatusCancelled  GameStatus = "cancelled"
	GameStatusAbandoned  GameStatus = "abandoned"
)

type GameInviteStatus string

const (
	GameInviteStatusPending   GameInviteStatus = "pending"
	GameInviteStatusAccepted  GameInviteStatus = "accepted"
	GameInviteStatusRejected  GameInviteStatus = "rejected"
	GameInviteStatusCancelled GameInviteStatus = "cancelled"
)

type PlayerStatus string

const (
	PlayerStatusActive       PlayerStatus = "active"
	PlayerStatusLeft         PlayerStatus = "left"
	PlayerStatusDisconnected PlayerStatus = "disconnected"
)

type Game struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PublicID  string     `gorm:"unique;not null;index" json:"public_id"`
	Mode      GameMode   `gorm:"type:citext;not null" json:"mode"`
	Status    GameStatus `gorm:"type:citext;not null;default:'waiting'" json:"status"`
	CreatorID uuid.UUID  `gorm:"type:uuid;not null;index" json:"-"`

	CreatorPublicID string `json:"creator_public_id"`

	TemplateID       *uuid.UUID `gorm:"type:uuid;index" json:"template_id,omitempty"`
	TemplateSnapshot string     `gorm:"type:jsonb" json:"template_snapshot,omitempty"`
	MaxPlayers       int        `gorm:"not null;default:1" json:"max_players"`
	MinPlayers       int        `gorm:"not null;default:1" json:"min_players"`
	QuestionCount    int        `gorm:"not null;default:10" json:"question_count"`
	PointsPerCorrect int        `gorm:"not null;default:100" json:"points_per_correct"`
	TimeBonus        bool       `gorm:"default:true" json:"time_bonus"`

	DatasetID    *uuid.UUID `gorm:"type:uuid;index" json:"dataset_id,omitempty"`
	Category     string     `gorm:"type:varchar(50);default:'general'" json:"category,omitempty"`
	FlagVariant  string     `gorm:"type:varchar(50)" json:"flag_variant,omitempty"`
	QuestionType string     `gorm:"type:varchar(50)" json:"question_type,omitempty"`
	Language     string     `gorm:"type:varchar(10);default:'en'" json:"language,omitempty"`
	Continent    string     `gorm:"type:varchar(100);default:''" json:"continent,omitempty"`

	WinnerID *uuid.UUID `gorm:"type:uuid;index" json:"winner_id,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"-"`

	Players   []GamePlayer   `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"players,omitempty"`
	Invites   []GameInvite   `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"invites,omitempty"`
	Questions []GameQuestion `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
	Answers   []GameAnswer   `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

type GamePlayer struct {
	ID           uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	GameID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"game_id"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"-"`
	UserPublicID string       `json:"user_public_id"`
	Score        int          `gorm:"default:0" json:"score"`
	IsReady      bool         `gorm:"default:false" json:"is_ready"`
	Status       PlayerStatus `gorm:"type:citext;default:'active'" json:"status"`
	JoinedAt     time.Time    `json:"joined_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type GameInvite struct {
	ID         uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	GameID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"game_id"`
	FromUserID uuid.UUID        `gorm:"type:uuid;not null;index" json:"-"`
	ToUserID   uuid.UUID        `gorm:"type:uuid;not null;index" json:"-"`
	Status     GameInviteStatus `gorm:"type:citext;not null;default:'pending'" json:"status"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`

	Game             *Game  `gorm:"foreignKey:GameID" json:"game,omitempty"`
	FromUser         *User  `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser           *User  `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
	FromUserPublicID string `json:"from_user_public_id"`
	ToUserPublicID   string `json:"to_user_public_id"`
}

type GameQuestion struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	GameID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"game_id"`
	QuestionID  *uuid.UUID `gorm:"type:uuid;index" json:"question_id,omitempty"`
	OrderNumber int        `gorm:"not null" json:"order_number"`
	EntityKey   string     `gorm:"type:varchar(200);index" json:"-"`

	Type     string         `gorm:"type:varchar(50)" json:"type,omitempty"`
	Data     datatypes.JSON `gorm:"type:jsonb" json:"data,omitempty"`
	Question *Question      `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

type GameAnswer struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	GameID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"game_id"`
	PlayerID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"player_id"`
	QuestionID     *uuid.UUID     `gorm:"type:uuid;index" json:"question_id,omitempty"`
	GameQuestionID *uuid.UUID     `gorm:"type:uuid;index" json:"-"`
	AnswerSlug     string         `gorm:"not null" json:"answer_slug"`
	Data           datatypes.JSON `gorm:"type:jsonb" json:"data,omitempty"`
	IsCorrect      bool           `gorm:"not null" json:"is_correct"`
	TimeSpent      int            `gorm:"not null" json:"time_spent"`
	Points         int            `gorm:"not null" json:"points"`
	AnsweredAt     time.Time      `json:"answered_at"`
}

type CreateGameRequest struct {
	Mode               GameMode   `json:"mode" binding:"required,oneof=solo 1v1 multi"`
	Category           string     `json:"category,omitempty"`
	FlagVariant        string     `json:"flag_variant,omitempty"`
	QuestionType       string     `json:"question_type,omitempty"`
	Continent          string     `json:"continent,omitempty"`
	IncludeTerritories bool       `json:"include_territories,omitempty"`
	QuestionCount      *int       `json:"question_count,omitempty"`
	PointsPerCorrect   *int       `json:"points_per_correct,omitempty"`
	TimeBonus          *bool      `json:"time_bonus,omitempty"`
	DatasetID          *uuid.UUID `json:"dataset_id,omitempty"`
	Language           string     `json:"language,omitempty"`
	TemplateID         *uuid.UUID `json:"template_id,omitempty"`
	TemplateSnapshot   string     `json:"-"`
}

type JoinGameByCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

type SubmitAnswerRequest struct {
	QuestionID uuid.UUID      `json:"question_id"`
	AnswerSlug string         `json:"answer_slug"`
	Data       datatypes.JSON `json:"data"`
	TimeSpent  int            `json:"time_spent" binding:"required,min=0"`
}

type GameStatusResponse struct {
	Game            *Game        `json:"game"`
	Players         []GamePlayer `json:"players"`
	CurrentQuestion *Question    `json:"current_question,omitempty"`
	QuestionNumber  int          `json:"question_number"`
	TotalQuestions  int          `json:"total_questions"`
	TimeRemaining   int          `json:"time_remaining,omitempty"`
}

type GameHistoryResponse struct {
	Game          *Game        `json:"game"`
	Players       []GamePlayer `json:"players"`
	UserScore     int          `json:"user_score"`
	OpponentScore int          `json:"opponent_score,omitempty"`
	IsWinner      bool         `json:"is_winner"`
}

type GameAnswerDetail struct {
	ID         uuid.UUID  `json:"id"`
	GameID     uuid.UUID  `json:"game_id"`
	PlayerID   uuid.UUID  `json:"player_id"`
	QuestionID *uuid.UUID `json:"question_id,omitempty"`

	QuestionSlug  string `json:"question_slug,omitempty"`
	QuestionTitle string `json:"question_title,omitempty"`
	QuestionType  string `json:"question_type,omitempty"`

	AnswerSlug string `json:"answer_slug"`
	AnswerLabel string `json:"answer_label,omitempty"`

	CorrectAnswerSlug  string `json:"correct_answer_slug,omitempty"`
	CorrectAnswerLabel string `json:"correct_answer_label,omitempty"`

	Data       datatypes.JSON `json:"data,omitempty"`
	IsCorrect  bool           `json:"is_correct"`
	TimeSpent  int            `json:"time_spent"`
	Points     int            `json:"points"`
	AnsweredAt time.Time      `json:"answered_at"`
}

type GameStatsResult struct {
	ActiveGames     int64
	CompletedGames  int64
	CancelledGames  int64
	AbandonedGames  int64
	TotalPlayers    int64
	TotalInvites    int64
	PendingInvites  int64
	AcceptedInvites int64
	RejectedInvites int64
}

type GameModeStatResult struct {
	Mode  string
	Count int64
}

type DailyGameStatResult struct {
	Date           string
	TotalGames     int64
	CompletedGames int64
	CancelledGames int64
	TotalPlayers   int64
}

type GamePerformanceResult struct {
	AvgGameDuration     float64
	AvgQuestionsPerGame float64
	TotalQuestionsUsed  int64
	AvgPlayersPerGame   float64
	MostPopularMode     string
}

const (
	CategoryGeneral   = "general"
	CategoryFlags     = "flags"
	CategoryGeography = "geography"

	FlagVariantMix            = "mix"
	FlagVariantFlagToName4    = "flag_to_name_4"
	FlagVariantFlagToName2    = "flag_to_name_2"
	FlagVariantNameToFlag4    = "name_to_flag_4"
	FlagVariantNameToFlag2    = "name_to_flag_2"
	FlagVariantFlagToText     = "flag_to_text"
	FlagVariantFlagToCapital  = "flag_to_capital"
	FlagVariantCapitalToFlag4 = "capital_to_flag_4"
	FlagVariantCapitalToFlag2 = "capital_to_flag_2"
	FlagVariantCapitalToName4 = "capital_to_name_4"
	FlagVariantCapitalToName2 = "capital_to_name_2"
)

var ValidFlagVariants = map[string]bool{
	FlagVariantMix:            true,
	FlagVariantFlagToName4:    true,
	FlagVariantFlagToName2:    true,
	FlagVariantNameToFlag4:    true,
	FlagVariantNameToFlag2:    true,
	FlagVariantFlagToText:     true,
	FlagVariantFlagToCapital:  true,
	FlagVariantCapitalToFlag4: true,
	FlagVariantCapitalToFlag2: true,
	FlagVariantCapitalToName4: true,
	FlagVariantCapitalToName2: true,
}

type GameResultsResponse struct {
	Game      GameSummary           `json:"game"`
	Players   []PlayerResultSummary `json:"players"`
	Questions []QuestionResult      `json:"questions"`
}

type GameSummary struct {
	PublicID       string     `json:"public_id"`
	Mode           string     `json:"mode"`
	Status         string     `json:"status"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	WinnerPublicID *string    `json:"winner_public_id,omitempty"`
}

type PlayerResultSummary struct {
	PublicID string `json:"public_id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	IsWinner bool   `json:"is_winner"`
}

type QuestionResult struct {
	OrderNumber   int                  `json:"order_number"`
	QuestionTitle string               `json:"question_title"`
	QuestionType  string               `json:"question_type"`
	CorrectAnswer string               `json:"correct_answer"`
	Answers       []PlayerAnswerResult `json:"answers"`
}

type PlayerAnswerResult struct {
	PlayerPublicID string `json:"player_public_id"`
	Answer         string `json:"answer"`
	IsCorrect      bool   `json:"is_correct"`
	Points         int    `json:"points"`
	TimeSpentMs    int    `json:"time_spent_ms"`
}
