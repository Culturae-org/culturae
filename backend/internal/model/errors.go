package model

import "errors"

var (
	ErrGameNotFound     = errors.New("game not found")
	ErrGameFull         = errors.New("game is full")
	ErrAlreadyInGame    = errors.New("you are already in this game")
	ErrGameNotCompleted = errors.New("game is not completed")

	ErrInviteNotFound  = errors.New("invite not found")
	ErrInviteNotPending = errors.New("invite is not pending")

	ErrForbidden = errors.New("forbidden")

	ErrPasswordMismatch = errors.New("current password is incorrect")
	ErrWrongPassword    = errors.New("password is incorrect")

	ErrSelfBlock = errors.New("cannot block yourself")

	ErrAlreadyFriends              = errors.New("already friends")
	ErrFriendRequestPending        = errors.New("friend request already pending")
	ErrFriendRequestNotPending     = errors.New("request is not pending")
	ErrSelfFriendRequest           = errors.New("cannot send friend request to yourself")
	ErrAlreadyBlocked              = errors.New("user already blocked")
	ErrBlockNotFound               = errors.New("block not found")
	ErrFriendRequestsDisabled      = errors.New("user does not allow friend requests")
	ErrCannotRequestBlockedUser    = errors.New("cannot send friend request to blocked user")

	ErrUserNotFound         = errors.New("user not found")
	ErrUserBlocked          = errors.New("user blocked")
	ErrProfilePrivate       = errors.New("profile is private")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrSessionNotFound      = errors.New("session not found")

	ErrInvalidGameMode       = errors.New("invalid game mode")
	ErrGameTemplateNotActive = errors.New("game template is not active")
	ErrNotInGame             = errors.New("not in game")
	ErrGameAlreadyFinished   = errors.New("game already finished")
	ErrGameNotActive         = errors.New("game not active")
	ErrGameInvitesDisabled   = errors.New("user does not accept game invites")
	ErrCannotInviteUser      = errors.New("cannot invite this user")
	ErrCannotJoinGame        = errors.New("cannot join this game")

	ErrAlreadyReported = errors.New("question already reported by this user")
)
