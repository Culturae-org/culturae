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

	ErrAlreadyFriends           = errors.New("already friends")
	ErrFriendRequestPending     = errors.New("friend request already pending")
	ErrFriendRequestNotPending  = errors.New("request is not pending")
	ErrSelfFriendRequest        = errors.New("cannot send friend request to yourself")
	ErrAlreadyBlocked           = errors.New("user already blocked")
	ErrBlockNotFound            = errors.New("block not found")
)
