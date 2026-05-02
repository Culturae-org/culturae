// backend/internal/usecase/friends.go

package usecase

import (
	"errors"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/repository"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FriendsUsecase struct {
	friendsRepo repository.FriendsRepositoryInterface
	userRepo    repository.UserRepositoryInterface
	loggingSvc  service.LoggingServiceInterface
	wsService   service.WebSocketServiceInterface
	notifRepo   repository.NotificationRepositoryInterface
}

func NewFriendsUsecase(
	friendRepo repository.FriendsRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	loggingSvc service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
	notifRepo repository.NotificationRepositoryInterface,
) *FriendsUsecase {
	return &FriendsUsecase{
		friendsRepo: friendRepo,
		userRepo:    userRepo,
		loggingSvc:  loggingSvc,
		wsService:   wsService,
		notifRepo:   notifRepo,
	}
}

// -----------------------------------------------
// Friends Usecase Methods
//
// - SendFriendRequest
// - AcceptFriendRequest
// - RejectFriendRequest
// - CancelFriendRequest
// - BlockFriendRequest
// - ListFriendRequests
// - ListFriends
// - RemoveFriend
// - IsFriend
// - GetUserUUIDByPublicID
// - GetBlockedUsers
// - BlockUser
// - UnblockUser
// - GetUserProfileWithRelationship
//
// -----------------------------------------------

func (u *FriendsUsecase) SendFriendRequest(c *gin.Context, fromUserID uuid.UUID, toUserPublicID string) (*model.FriendRequestResponse, error) {
	toUserID, err := u.GetUserUUIDByPublicID(toUserPublicID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(fromUserID, "send_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyToUserPublicID: toUserPublicID}, false, &errorMsg)
		return nil, errors.New("user not found")
	}

	toUser, err := u.userRepo.GetByID(toUserID.String())
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(fromUserID, "send_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyToUserPublicID: toUserPublicID}, false, &errorMsg)
		return nil, errors.New("user not found")
	}
	if !toUser.AllowFriendRequests {
		errMsg := "user does not allow friend requests"
		_ = u.loggingSvc.LogUserAction(fromUserID, "send_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyToUserPublicID: toUserPublicID}, false, &errMsg)
		return nil, errors.New(errMsg)
	}

	blocked, err := u.friendsRepo.IsBlocked(fromUserID, toUserID)
	if err == nil && blocked {
		errMsg := "cannot send friend request to blocked user"
		_ = u.loggingSvc.LogUserAction(fromUserID, "send_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyToUserPublicID: toUserPublicID}, false, &errMsg)
		return nil, errors.New(errMsg)
	}

	request, err := u.friendsRepo.SendFriendRequest(fromUserID, toUserID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(fromUserID, "send_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyToUserPublicID: toUserPublicID}, false, &errorMsg)
		return nil, err
	}

	_ = u.loggingSvc.LogUserAction(fromUserID, "send_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyToUserPublicID: toUserPublicID, keyRequestID: request.ID}, true, nil)

	fromUser, _ := u.userRepo.GetByID(fromUserID.String())

	fromPublicID := ""
	if fromUser != nil {
		fromPublicID = fromUser.PublicID
	}

	if u.wsService != nil {
		go func() {
			_ = u.wsService.SendToUser(toUserID, map[string]interface{}{
				keyType:                "friend_request_received",
				keyRequestID:         request.ID.String(),
				"from_user_public_id": fromPublicID,
			})
		}()
	}

	if u.notifRepo != nil {
		username := toUserPublicID
		if fromUser != nil {
			username = fromUser.Username
		}
		_ = u.notifRepo.Create(&model.Notification{
			UserID: toUserID,
			Type:   "friend_request",
			Title:  "New friend request",
			Body:   username + " sent you a friend request",
		})
	}

	return &model.FriendRequestResponse{
		ID:               request.ID,
		FromUserPublicID: fromPublicID,
		ToUserPublicID:   toUserPublicID,
		Status:           request.Status,
		CreatedAt:        request.CreatedAt,
		UpdatedAt:        request.UpdatedAt,
	}, nil
}

func (u *FriendsUsecase) AcceptFriendRequest(c *gin.Context, requestID, userID uuid.UUID) error {
	request, reqErr := u.friendsRepo.GetFriendRequestByID(requestID)

	err := u.friendsRepo.AcceptFriendRequest(requestID, userID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(userID, "accept_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(userID, "accept_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, true, nil)

	if u.notifRepo != nil && reqErr == nil {
		_ = u.notifRepo.Create(&model.Notification{
			UserID: request.FromUserID,
			Type:   "friend_request_accepted",
			Title:  "Friend request accepted",
			Body:   "Your friend request was accepted",
		})
	}

	return nil
}

func (u *FriendsUsecase) RejectFriendRequest(c *gin.Context, requestID, userID uuid.UUID) error {
	request, reqErr := u.friendsRepo.GetFriendRequestByID(requestID)

	err := u.friendsRepo.RejectFriendRequest(requestID, userID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(userID, "reject_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(userID, "reject_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, true, nil)

	if u.wsService != nil && reqErr == nil {
		go func() {
			_ = u.wsService.SendToUser(request.FromUserID, map[string]interface{}{
				keyType:       "friend_request_rejected",
				keyRequestID: requestID.String(),
			})
		}()
	}

	return nil
}

func (u *FriendsUsecase) CancelFriendRequest(c *gin.Context, requestID, userID uuid.UUID) error {
	request, reqErr := u.friendsRepo.GetFriendRequestByID(requestID)

	err := u.friendsRepo.CancelFriendRequest(requestID, userID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(userID, "cancel_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(userID, "cancel_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, true, nil)

	if u.wsService != nil && reqErr == nil {
		go func() {
			_ = u.wsService.SendToUser(request.ToUserID, map[string]interface{}{
				keyType:       "friend_request_cancelled",
				keyRequestID: requestID.String(),
			})
		}()
	}

	return nil
}

func (u *FriendsUsecase) BlockFriendRequest(c *gin.Context, requestID, userID uuid.UUID) error {
	err := u.friendsRepo.BlockFriendRequest(requestID, userID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(userID, "block_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(userID, "block_friend_request", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyRequestID: requestID}, true, nil)
	return nil
}

func (u *FriendsUsecase) ListFriendRequests(c *gin.Context, userID uuid.UUID, status string, requestType string) ([]model.FriendRequestWithUser, error) {
	return u.friendsRepo.ListFriendRequests(userID, status, requestType)
}

func (u *FriendsUsecase) ListFriends(c *gin.Context, userID uuid.UUID) ([]model.FriendUserResponse, error) {
	friends, err := u.friendsRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}

	var result []model.FriendUserResponse
	for _, friend := range friends {
		var user model.User
		if friend.UserID1 == userID {
			user = friend.User2
		} else {
			user = friend.User1
		}
		result = append(result, model.FriendUserResponse{
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
		})
	}

	return result, nil
}

func (u *FriendsUsecase) RemoveFriend(c *gin.Context, userID uuid.UUID, friendUserPublicID string) error {
	friendUserID, err := u.GetUserUUIDByPublicID(friendUserPublicID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(userID, "remove_friend", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyFriendUserPublicID: friendUserPublicID}, false, &errorMsg)
		return errors.New("user not found")
	}

	err = u.friendsRepo.RemoveFriend(userID, friendUserID)
	if err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(userID, "remove_friend", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyFriendUserPublicID: friendUserPublicID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(userID, "remove_friend", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyFriendUserPublicID: friendUserPublicID}, true, nil)

	if u.wsService != nil {
		remover, _ := u.userRepo.GetByID(userID.String())
		removerPublicID := ""
		if remover != nil {
			removerPublicID = remover.PublicID
		}
		go func() {
			_ = u.wsService.SendToUser(friendUserID, map[string]interface{}{
				keyType:              "friend_removed",
				"removed_by_public_id": removerPublicID,
			})
		}()
	}

	return nil
}

func (u *FriendsUsecase) IsFriend(userID1, userID2 uuid.UUID) (bool, error) {
	return u.friendsRepo.IsFriend(userID1, userID2)
}

func (u *FriendsUsecase) GetUserUUIDByPublicID(publicID string) (uuid.UUID, error) {
	user, err := u.userRepo.GetByPublicID(publicID)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func (u *FriendsUsecase) GetBlockedUsers(userID uuid.UUID) ([]model.UserBasicInfo, error) {
	blocked, err := u.friendsRepo.GetBlockedUsers(userID)
	if err != nil {
		return nil, err
	}

	var users []model.UserBasicInfo
	for _, b := range blocked {
		user, err := u.userRepo.GetByID(b.ToUserID.String())
		if err != nil {
			continue
		}
		users = append(users, model.UserBasicInfo{
			PublicID:  user.PublicID,
			Username:  user.Username,
			HasAvatar: user.HasAvatar,
			Role:      user.Role,
		})
	}
	return users, nil
}

func (u *FriendsUsecase) BlockUser(c *gin.Context, blockerID uuid.UUID, blockedPublicID string) error {
	blockedID, err := u.GetUserUUIDByPublicID(blockedPublicID)
	if err != nil {
		return errors.New("user not found")
	}

	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	if err := u.friendsRepo.BlockUserDirect(blockerID, blockedID); err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(blockerID, "block_user", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyBlockedUserPublicID: blockedPublicID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(blockerID, "block_user", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyBlockedUserPublicID: blockedPublicID}, true, nil)
	return nil
}

func (u *FriendsUsecase) UnblockUser(c *gin.Context, blockerID uuid.UUID, blockedPublicID string) error {
	blockedID, err := u.GetUserUUIDByPublicID(blockedPublicID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := u.friendsRepo.UnblockUser(blockerID, blockedID); err != nil {
		errorMsg := err.Error()
		_ = u.loggingSvc.LogUserAction(blockerID, "unblock_user", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyBlockedUserPublicID: blockedPublicID}, false, &errorMsg)
		return err
	}

	_ = u.loggingSvc.LogUserAction(blockerID, "unblock_user", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{keyBlockedUserPublicID: blockedPublicID}, true, nil)
	return nil
}

func (u *FriendsUsecase) GetUserProfileWithRelationship(viewerID uuid.UUID, targetPublicID string) (*model.UserProfileWithRelationship, error) {
	targetUser, err := u.userRepo.GetByPublicID(targetPublicID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	isOwnProfile := viewerID == targetUser.ID
	isFriend := false
	friendRequestStatus := ""
	isBlocked := false

	if !isOwnProfile && viewerID != uuid.Nil {
		isFriend, _ = u.friendsRepo.IsFriend(viewerID, targetUser.ID)

		sentRequest, _ := u.friendsRepo.HasPendingRequest(viewerID, targetUser.ID)
		if sentRequest {
			friendRequestStatus = "sent"
		}

		receivedRequest, _ := u.friendsRepo.HasPendingRequest(targetUser.ID, viewerID)
		if receivedRequest {
			friendRequestStatus = "received"
		}

		isBlocked, _ = u.friendsRepo.IsBlocked(viewerID, targetUser.ID)
		if isBlocked {
			return nil, errors.New("user blocked")
		}

		blockedByTarget, _ := u.friendsRepo.IsBlocked(targetUser.ID, viewerID)
		if blockedByTarget {
			return nil, errors.New("user not found")
		}
	}

	canSeeFullProfile := isOwnProfile || isFriend || targetUser.IsProfilePublic

	if !canSeeFullProfile {
		if !targetUser.AllowFriendRequests {
			return nil, errors.New("profile is private")
		}
		return &model.UserProfileWithRelationship{
			PublicProfile: &model.PublicProfile{
				PublicID:            targetUser.PublicID,
				Username:            targetUser.Username,
				HasAvatar:           targetUser.HasAvatar,
				AllowFriendRequests: targetUser.AllowFriendRequests,
			},
			IsFriend:            false,
			FriendRequestStatus: friendRequestStatus,
			IsBlocked:           false,
			IsOwnProfile:        false,
		}, nil
	}

	profile := &model.PublicProfile{
		PublicID:            targetUser.PublicID,
		Username:            targetUser.Username,
		HasAvatar:           targetUser.HasAvatar,
		CreatedAt:           targetUser.CreatedAt,
		Bio:                 targetUser.Bio,
		Language:            targetUser.Language,
		Level:               targetUser.Level,
		Rank:                targetUser.Rank,
		Status:              targetUser.Status,
		EloRating:           targetUser.EloRating,
		EloGamesPlayed:      targetUser.EloGamesPlayed,
		Experience:          targetUser.Experience,
		IsOnline:            targetUser.IsOnline,
		ShowOnlineStatus:    targetUser.ShowOnlineStatus,
		LastSeenAt:          targetUser.LastSeenAt,
		AllowFriendRequests: targetUser.AllowFriendRequests,
	}

	if stats, err := u.userRepo.GetUserGameStats(targetUser.ID); err == nil {
		profile.TotalGames = stats.TotalGames
		profile.GamesWon = stats.GamesWon
		profile.GamesLost = stats.GamesLost
		profile.DayStreak = stats.DayStreak
		profile.BestDayStreak = stats.BestDayStreak
		profile.TotalScore = stats.TotalScore
		profile.AverageScore = stats.AverageScore
		profile.PlayTime = stats.PlayTime
		profile.LastGameAt = stats.LastGameAt
	}

	return &model.UserProfileWithRelationship{
		PublicProfile:       profile,
		IsFriend:            isFriend,
		FriendRequestStatus: friendRequestStatus,
		IsBlocked:           isBlocked,
		IsOwnProfile:        isOwnProfile,
	}, nil
}
