// backend/internal/model/user.go

package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	AccountStatusActive    = "active"
	AccountStatusSuspended = "suspended"
	AccountStatusBanned    = "banned"
	AccountStatusInactive  = "inactive"
	AccountStatusDeleted   = "deleted"
)

const (
	RoleUser          = "user"
	RoleAdministrator = "administrator"
	RoleModerator     = "moderator"
)

type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"-"`
	Email         string     `gorm:"uniqueIndex" json:"email"`
	Password      string     `json:"-"`
	Username      string     `gorm:"index" json:"username"`
	PublicID      string     `gorm:"unique;not null" json:"public_id"`
	Role          string     `json:"role"`
	AccountStatus string     `json:"account_status" gorm:"default:'active'"`
	BannedUntil   *time.Time `json:"banned_until,omitempty"`
	BanReason     string     `json:"ban_reason,omitempty"`
	HasAvatar     bool       `json:"has_avatar" gorm:"default:false"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Bio      *string `json:"bio,omitempty"`
	Language string  `json:"language" gorm:"default:'en'"`

	Experience     int64  `json:"experience" gorm:"default:0"`
	Level          int    `json:"level" gorm:"default:0"`
	Rank           string `json:"rank" gorm:"default:'Beginner'"`
	EloRating      int    `json:"elo_rating" gorm:"default:1000"`
	EloGamesPlayed int    `json:"elo_games_played" gorm:"default:0"`

	Status        string     `json:"status" gorm:"default:'offline'"`
	IsOnline      bool       `json:"is_online" gorm:"default:false"`
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
	CurrentGameID *uuid.UUID `json:"current_game_id,omitempty"`

	IsProfilePublic     bool `json:"is_profile_public" gorm:"default:true"`
	ShowOnlineStatus    bool `json:"show_online_status" gorm:"default:true"`
	AllowFriendRequests bool `json:"allow_friend_requests" gorm:"default:true"`
	AllowPartyInvites   bool `json:"allow_party_invites" gorm:"default:true"`

	LastPublicIDRegeneration *time.Time `json:"last_public_id_regeneration,omitempty"`

	GameStats *UserGameStats `json:"game_stats,omitempty" gorm:"foreignKey:UserID"`
}

type UserAdminView struct {
	ID                  uuid.UUID      `json:"id"`
	Email               string         `json:"email"`
	Username            string         `json:"username"`
	PublicID            string         `json:"public_id"`
	Role                string         `json:"role"`
	AccountStatus       string         `json:"account_status"`
	BannedUntil         *time.Time     `json:"banned_until,omitempty"`
	BanReason           string         `json:"ban_reason,omitempty"`
	HasAvatar           bool           `json:"has_avatar"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	Bio                 *string        `json:"bio,omitempty"`
	Language            string         `json:"language"`
	Experience          int64          `json:"experience"`
	Level               int            `json:"level"`
	Rank                string         `json:"rank"`
	EloRating           int            `json:"elo_rating"`
	EloGamesPlayed      int            `json:"elo_games_played"`
	Status              string         `json:"status"`
	IsOnline            bool           `json:"is_online"`
	LastSeenAt          *time.Time     `json:"last_seen_at,omitempty"`
	CurrentGameID       *uuid.UUID     `json:"current_game_id,omitempty"`
	IsProfilePublic     bool           `json:"is_profile_public"`
	ShowOnlineStatus    bool           `json:"show_online_status"`
	AllowFriendRequests bool           `json:"allow_friend_requests"`
	AllowPartyInvites   bool           `json:"allow_party_invites"`
	GameStats           *UserGameStats `json:"game_stats,omitempty"`
}

type UserView struct {
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	PublicID      string     `json:"public_id"`
	Role          string     `json:"role"`
	AccountStatus string     `json:"account_status"`
	BannedUntil   *time.Time `json:"banned_until,omitempty"`
	BanReason     string     `json:"ban_reason,omitempty"`
	HasAvatar     bool       `json:"has_avatar"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Bio           *string    `json:"bio,omitempty"`
	Language      string     `json:"language"`

	Experience     int64  `json:"experience"`
	Level          int    `json:"level"`
	Rank           string `json:"rank"`
	EloRating      int    `json:"elo_rating"`
	EloGamesPlayed int    `json:"elo_games_played"`

	Status     string     `json:"status"`
	IsOnline   bool       `json:"is_online"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`

	IsProfilePublic     bool `json:"is_profile_public"`
	ShowOnlineStatus    bool `json:"show_online_status"`
	AllowFriendRequests bool `json:"allow_friend_requests"`
	AllowPartyInvites   bool `json:"allow_party_invites"`

	GameStats *UserGameStats `json:"game_stats,omitempty"`
}

type UserBasicInfo struct {
	PublicID  string `json:"public_id"`
	Username  string `json:"username"`
	HasAvatar bool   `json:"has_avatar"`
	Language  string `json:"language"`
	Role      string `json:"role"`
	Level     int    `json:"level"`
	Rank      string `json:"rank"`
	EloRating int    `json:"elo_rating"`
	CurrentGamePublicID *string `json:"current_game_public_id,omitempty"`
}

type PublicProfile struct {
	PublicID  string    `json:"public_id"`
	Username  string    `json:"username"`
	HasAvatar bool      `json:"has_avatar"`
	CreatedAt time.Time `json:"created_at"`

	Bio      *string `json:"bio,omitempty"`
	Language string  `json:"language"`

	Level    int    `json:"level"`
	Rank     string `json:"rank"`
	Status   string `json:"status"`
	PlayTime int64  `json:"play_time"`

	TotalGames     int        `json:"total_games"`
	GamesWon       int        `json:"games_won"`
	GamesLost      int        `json:"games_lost"`
	GamesDrawn     int        `json:"games_drawn"`
	DayStreak      int        `json:"day_streak"`
	BestDayStreak  int        `json:"best_day_streak"`
	TotalScore     int64      `json:"total_score"`
	AverageScore   float64    `json:"average_score"`
	LastGameAt     *time.Time `json:"last_game_at,omitempty"`
	EloRating      int        `json:"elo_rating"`
	EloGamesPlayed int        `json:"elo_games_played"`
	Experience     int64      `json:"experience"`

	IsOnline            bool       `json:"is_online"`
	ShowOnlineStatus    bool       `json:"show_online_status"`
	LastSeenAt          *time.Time `json:"last_seen_at,omitempty"`
	AllowFriendRequests bool       `json:"allow_friend_requests"`
}

type UserProfileWithRelationship struct {
	*PublicProfile
	IsFriend            bool   `json:"is_friend"`
	FriendRequestStatus string `json:"friend_request_status,omitempty"`
	IsBlocked           bool   `json:"is_blocked"`
	IsOwnProfile        bool   `json:"is_own_profile"`
}

type UserSearchCard struct {
	PublicID  string `json:"public_id"`
	Username  string `json:"username"`
	HasAvatar bool   `json:"has_avatar"`
}

type UserGameStats struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4()" json:"-"`
	UserID        uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	TotalGames    int        `json:"total_games" gorm:"default:0"`
	GamesWon      int        `json:"games_won" gorm:"default:0"`
	GamesLost     int        `json:"games_lost" gorm:"default:0"`
	GamesDrawn    int        `json:"games_drawn" gorm:"default:0"`
	DayStreak     int        `json:"day_streak" gorm:"default:0"`
	BestDayStreak int        `json:"best_day_streak" gorm:"default:0"`
	TotalScore    int64      `json:"total_score" gorm:"default:0"`
	AverageScore  float64    `json:"average_score" gorm:"default:0"`
	PlayTime      int64      `json:"play_time" gorm:"default:0"`
	LastGameAt    *time.Time `json:"last_game_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type UserGameStatsByMode struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"-"`
	UserID       uuid.UUID `gorm:"type:uuid;index:idx_user_game_stats_by_mode,unique;not null" json:"-"`
	Mode         string    `gorm:"index:idx_user_game_stats_by_mode,unique;not null" json:"mode"`
	TotalGames   int       `json:"total_games" gorm:"default:0"`
	GamesWon     int       `json:"games_won" gorm:"default:0"`
	GamesLost    int       `json:"games_lost" gorm:"default:0"`
	GamesDrawn   int       `json:"games_drawn" gorm:"default:0"`
	TotalScore   int64     `json:"total_score" gorm:"default:0"`
	AverageScore float64   `json:"average_score" gorm:"default:0"`
	PlayTime     int64     `json:"play_time" gorm:"default:0"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserStats struct {
	TotalGames     int              `json:"total_games"`
	GamesWon       int              `json:"games_won"`
	GamesLost      int              `json:"games_lost"`
	GamesDrawn     int              `json:"games_drawn"`
	WinRate        float64          `json:"win_rate"`
	DayStreak      int              `json:"day_streak"`
	BestDayStreak  int              `json:"best_day_streak"`
	TotalScore     int64            `json:"total_score"`
	AverageScore   float64          `json:"average_score"`
	PlayTime       int64            `json:"play_time"`
	EloRating      int              `json:"elo_rating"`
	EloGamesPlayed int              `json:"elo_games_played"`
	Experience     int64            `json:"experience"`
	Level          int              `json:"level"`
	Rank           string           `json:"rank"`
	GamesByMode    []GameModeStats  `json:"games_by_mode"`
	RecentGames    []RecentGameInfo `json:"recent_games"`
}

type GameModeStats struct {
	Mode         string  `json:"mode"`
	TotalGames   int     `json:"total_games"`
	GamesWon     int     `json:"games_won"`
	GamesLost    int     `json:"games_lost"`
	GamesDrawn   int     `json:"games_drawn"`
	TotalScore   int64   `json:"total_score"`
	AverageScore float64 `json:"average_score"`
	PlayTime     int64   `json:"play_time"`
}

type RecentGameInfo struct {
	PublicID    string     `json:"public_id"`
	Mode        string     `json:"mode"`
	Status      string     `json:"status"`
	Score       int        `json:"score"`
	IsWinner    bool       `json:"is_winner"`
	CompletedAt *time.Time `json:"completed_at"`
}

type UserStatsByPeriod struct {
	TotalGames   int              `json:"total_games"`
	GamesWon     int              `json:"games_won"`
	GamesLost    int              `json:"games_lost"`
	GamesDrawn   int              `json:"games_drawn"`
	TotalScore   int64            `json:"total_score"`
	AverageScore float64          `json:"average_score"`
	PlayTime     int64            `json:"play_time"`
	RecentGames  []RecentGameInfo `json:"recent_games"`
}

type LeaderboardEntry struct {
	Rank      int    `json:"rank"`
	PublicID  string `json:"public_id"`
	Username  string `json:"username"`
	HasAvatar bool   `json:"has_avatar"`
	Score     int64  `json:"score"`
	EloRating int    `json:"elo_rating"`
}

type LeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	UserRank   *LeaderboardEntry  `json:"user_rank,omitempty"`
	Type       string             `json:"type"`
	Mode       string             `json:"mode"`
	TotalUsers int                `json:"total_users"`
}

type UserUpdate struct {
	Email         string  `json:"email"`
	Username      string  `json:"username"`
	Role          string  `json:"role"`
	AccountStatus string  `json:"account_status"`
	Language      string  `json:"language"`
	Bio           *string `json:"bio,omitempty"`

	IsProfilePublic     *bool `json:"is_profile_public,omitempty"`
	ShowOnlineStatus    *bool `json:"show_online_status,omitempty"`
	AllowFriendRequests *bool `json:"allow_friend_requests,omitempty"`
	AllowPartyInvites   *bool `json:"allow_party_invites,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

type AdminUpdatePassword struct {
	Password string `json:"password" binding:"required,min=8"`
}

type AdminBanUser struct {
	Duration string `json:"duration" binding:"required"`
	Reason   string `json:"reason"`
}

type AdminCreateUser struct {
	Email         string `json:"email" binding:"required,email"`
	Username      string `json:"username" binding:"required,min=3,max=20"`
	Role          string `json:"role" binding:"required"`
	AccountStatus string `json:"account_status" binding:"required,oneof=active suspended banned inactive"`
	Password      string `json:"password" binding:"required,min=8"`
}

func (u *User) ToAdminView() *UserAdminView {
	return &UserAdminView{
		ID:                  u.ID,
		Email:               u.Email,
		Username:            u.Username,
		PublicID:            u.PublicID,
		Role:                u.Role,
		AccountStatus:       u.AccountStatus,
		BannedUntil:         u.BannedUntil,
		BanReason:           u.BanReason,
		HasAvatar:           u.HasAvatar,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
		Bio:                 u.Bio,
		Language:            u.Language,
		Experience:          u.Experience,
		Level:               u.Level,
		Rank:                u.Rank,
		EloRating:           u.EloRating,
		EloGamesPlayed:      u.EloGamesPlayed,
		Status:              u.Status,
		IsOnline:            u.IsOnline,
		LastSeenAt:          u.LastSeenAt,
		CurrentGameID:       u.CurrentGameID,
		IsProfilePublic:     u.IsProfilePublic,
		ShowOnlineStatus:    u.ShowOnlineStatus,
		AllowFriendRequests: u.AllowFriendRequests,
		AllowPartyInvites:   u.AllowPartyInvites,
		GameStats:           u.GameStats,
	}
}

func (u *User) ToUserView() *UserView {
	return &UserView{
		Username:            u.Username,
		Email:               u.Email,
		PublicID:            u.PublicID,
		Role:                u.Role,
		AccountStatus:       u.AccountStatus,
		BannedUntil:         u.BannedUntil,
		BanReason:           u.BanReason,
		HasAvatar:           u.HasAvatar,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
		Bio:                 u.Bio,
		Language:            u.Language,
		Experience:          u.Experience,
		Level:               u.Level,
		Rank:                u.Rank,
		EloRating:           u.EloRating,
		EloGamesPlayed:      u.EloGamesPlayed,
		Status:              u.Status,
		IsOnline:            u.IsOnline,
		LastSeenAt:          u.LastSeenAt,
		IsProfilePublic:     u.IsProfilePublic,
		ShowOnlineStatus:    u.ShowOnlineStatus,
		AllowFriendRequests: u.AllowFriendRequests,
		AllowPartyInvites:   u.AllowPartyInvites,
		GameStats:           u.GameStats,
	}
}
