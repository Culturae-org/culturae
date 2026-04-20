// backend/internal/model/friends.go

package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	FriendRequestStatusPending  = "pending"
	FriendRequestStatusAccepted = "accepted"
	FriendRequestStatusRejected = "rejected"
	FriendRequestStatusBlocked  = "blocked"
	FriendRequestStatusCancelled = "cancelled"
)

type Friend struct {
	UserID1   uuid.UUID `gorm:"type:uuid;primaryKey;check:user_id1 < user_id2" json:"user_id_1"`
	UserID2   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id_2"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

type FriendWithUser struct {
	UserID1   uuid.UUID    `json:"user_id_1"`
	UserID2   uuid.UUID    `json:"user_id_2"`
	User1     User         `json:"user_1"`
	User2     User         `json:"user_2"`
	CreatedAt time.Time    `json:"created_at"`
}

type FriendRequestWithUser struct {
	ID         uuid.UUID            `json:"id"`
	FromUser   FriendUserResponse   `json:"from_user"`
	ToUser     FriendUserResponse   `json:"to_user"`
	Status     string               `json:"status"`
	CreatedAt  time.Time            `json:"created_at"`
}

type FriendRequest struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"-"`
	FromUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	ToUserID   uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Status     string    `gorm:"type:citext;not null;default:'pending'" json:"-"`
	CreatedAt  time.Time `gorm:"not null" json:"-"`
	UpdatedAt  time.Time `gorm:"not null" json:"-"`
}

type FriendRequestResponse struct {
	ID               uuid.UUID `json:"id"`
	FromUserPublicID string    `json:"from_user_public_id"`
	ToUserPublicID   string    `json:"to_user_public_id"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type FriendUserResponse struct {
	PublicID          string    `json:"public_id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	AccountStatus     string    `json:"account_status"`
	HasAvatar         bool      `json:"has_avatar"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Bio               *string   `json:"bio"`
	Language          string    `json:"language"`
	Experience        int64     `json:"experience"`
	Level             int       `json:"level"`
	Rank              string    `json:"rank"`
	EloRating         int       `json:"elo_rating"`
	EloGamesPlayed    int       `json:"elo_games_played"`
	Status            string    `json:"status"`
	IsOnline          bool      `json:"is_online"`
	IsProfilePublic   bool      `json:"is_profile_public"`
	ShowOnlineStatus  bool      `json:"show_online_status"`
	AllowFriendRequests bool    `json:"allow_friend_requests"`
	AllowPartyInvites bool      `json:"allow_party_invites"`
}

type AdminFriendRequest struct {
	ID           uuid.UUID `json:"id"`
	FromUserID   uuid.UUID `json:"from_user_id"`
	FromUsername string    `json:"from_username"`
	ToUserID     uuid.UUID `json:"to_user_id"`
	ToUsername   string    `json:"to_username"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdminFriendship struct {
	ID            string    `json:"id"`
	User1ID       uuid.UUID `json:"user1_id"`
	User1Username string    `json:"user1_username"`
	User2ID       uuid.UUID `json:"user2_id"`
	User2Username string    `json:"user2_username"`
	CreatedAt     time.Time `json:"created_at"`
}

func UserToFriendUserResponse(user User) FriendUserResponse {
	return FriendUserResponse{
		PublicID:            user.PublicID,
		Username:            user.Username,
		Email:               user.Email,
		Role:                user.Role,
		AccountStatus:       user.AccountStatus,
		HasAvatar:           user.HasAvatar,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
		Bio:                 user.Bio,
		Language:            user.Language,
		Experience:          user.Experience,
		Level:               user.Level,
		Rank:                user.Rank,
		EloRating:           user.EloRating,
		EloGamesPlayed:      user.EloGamesPlayed,
		Status:              user.Status,
		IsOnline:            user.IsOnline,
		IsProfilePublic:     user.IsProfilePublic,
		ShowOnlineStatus:    user.ShowOnlineStatus,
		AllowFriendRequests: user.AllowFriendRequests,
		AllowPartyInvites:   user.AllowPartyInvites,
	}
}
