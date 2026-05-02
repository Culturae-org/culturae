// backend/internal/repository/friends.go

package repository

import (
	"errors"
	"time"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FriendsRepositoryInterface interface {
	SendFriendRequest(fromUserID, toUserID uuid.UUID) (*model.FriendRequest, error)
	GetFriendRequestByID(requestID uuid.UUID) (*model.FriendRequest, error)
	AcceptFriendRequest(requestID, userID uuid.UUID) error
	RejectFriendRequest(requestID, userID uuid.UUID) error
	CancelFriendRequest(requestID, userID uuid.UUID) error
	BlockFriendRequest(requestID, userID uuid.UUID) error
	ListFriendRequests(userID uuid.UUID, status string, requestType string) ([]model.FriendRequestWithUser, error)
	ListFriends(userID uuid.UUID) ([]model.FriendWithUser, error)
	RemoveFriend(userID1, userID2 uuid.UUID) error
	IsFriend(userID1, userID2 uuid.UUID) (bool, error)
	HasPendingRequest(fromUserID, toUserID uuid.UUID) (bool, error)
	GetBlockedUsers(userID uuid.UUID) ([]model.FriendRequest, error)
	BlockUserDirect(blockerID, blockedID uuid.UUID) error
	UnblockUser(blockerID, blockedID uuid.UUID) error
	IsBlocked(userID1, userID2 uuid.UUID) (bool, error)
}

type FriendsRepository struct {
	DB     *gorm.DB
	logger *zap.Logger
}

func NewFriendsRepository(
	db *gorm.DB,
	logger *zap.Logger,
) *FriendsRepository {
	return &FriendsRepository{
		DB:     db,
		logger: logger,
	}
}

func (r *FriendsRepository) SendFriendRequest(fromUserID, toUserID uuid.UUID) (*model.FriendRequest, error) {
	if fromUserID == toUserID {
		return nil, errors.New("cannot send friend request to yourself")
	}

	isFriend, err := r.IsFriend(fromUserID, toUserID)
	if err != nil {
		return nil, err
	}
	if isFriend {
		return nil, errors.New("already friends")
	}

	hasPending, err := r.HasPendingRequest(fromUserID, toUserID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("friend request already pending")
	}

	hasPendingReverse, err := r.HasPendingRequest(toUserID, fromUserID)
	if err != nil {
		return nil, err
	}
	if hasPendingReverse {
		return nil, errors.New("friend request already pending from the other user")
	}

	request := &model.FriendRequest{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Status:     model.FriendRequestStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := r.DB.Create(request).Error; err != nil {
		r.logger.Error("Failed to create friend request", zap.Error(err))
		return nil, err
	}

	return request, nil
}

func (r *FriendsRepository) GetFriendRequestByID(requestID uuid.UUID) (*model.FriendRequest, error) {
	var request model.FriendRequest
	err := r.DB.Where("id = ?", requestID).First(&request).Error
	return &request, err
}

func (r *FriendsRepository) AcceptFriendRequest(requestID, userID uuid.UUID) error {
	var request model.FriendRequest
	if err := r.DB.Where("id = ? AND to_user_id = ?", requestID, userID).First(&request).Error; err != nil {
		return err
	}

	if request.Status != model.FriendRequestStatusPending {
		return errors.New("request is not pending")
	}

	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&request).Updates(map[string]interface{}{
		keyStatus:     model.FriendRequestStatusAccepted,
		keyUpdatedAt: time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	userID1 := request.FromUserID
	userID2 := request.ToUserID
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}
	friend := &model.Friend{
		UserID1:   userID1,
		UserID2:   userID2,
		CreatedAt: time.Now(),
	}
	if err := tx.Create(friend).Error; err != nil {
		tx.Rollback()
		r.logger.Error("Failed to create friend relationship", zap.Error(err))
		return err
	}

	return tx.Commit().Error
}

func (r *FriendsRepository) RejectFriendRequest(requestID, userID uuid.UUID) error {
	return r.DB.Model(&model.FriendRequest{}).Where("id = ? AND to_user_id = ?", requestID, userID).Updates(map[string]interface{}{
		keyStatus:     model.FriendRequestStatusRejected,
		keyUpdatedAt: time.Now(),
	}).Error
}

func (r *FriendsRepository) CancelFriendRequest(requestID, userID uuid.UUID) error {
	return r.DB.Model(&model.FriendRequest{}).Where("id = ? AND from_user_id = ?", requestID, userID).Updates(map[string]interface{}{
		keyStatus:     model.FriendRequestStatusCancelled,
		keyUpdatedAt: time.Now(),
	}).Error
}

func (r *FriendsRepository) BlockFriendRequest(requestID, userID uuid.UUID) error {
	return r.DB.Model(&model.FriendRequest{}).Where("id = ? AND (from_user_id = ? OR to_user_id = ?)", requestID, userID, userID).Updates(map[string]interface{}{
		keyStatus:     model.FriendRequestStatusBlocked,
		keyUpdatedAt: time.Now(),
	}).Error
}

func (r *FriendsRepository) ListFriendRequests(userID uuid.UUID, status string, requestType string) ([]model.FriendRequestWithUser, error) {
	var requests []model.FriendRequest
	query := r.DB.Model(&model.FriendRequest{})
	
	switch requestType {
	case "incoming":
		query = query.Where("to_user_id = ?", userID)
	case "outgoing":
		query = query.Where("from_user_id = ?", userID)
	default:
		query = query.Where("from_user_id = ? OR to_user_id = ?", userID, userID)
	}
	
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Find(&requests).Error; err != nil {
		return nil, err
	}
	
	if len(requests) == 0 {
		return []model.FriendRequestWithUser{}, nil
	}
	
	var userIDs []uuid.UUID
	for _, req := range requests {
		userIDs = append(userIDs, req.FromUserID, req.ToUserID)
	}
	
	var users []model.User
	if err := r.DB.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	
	userMap := make(map[uuid.UUID]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}
	
	var result []model.FriendRequestWithUser
	for _, req := range requests {
		result = append(result, model.FriendRequestWithUser{
			ID:        req.ID,
			FromUser:  model.UserToFriendUserResponse(userMap[req.FromUserID]),
			ToUser:    model.UserToFriendUserResponse(userMap[req.ToUserID]),
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
		})
	}
	
	return result, nil
}

func (r *FriendsRepository) ListFriends(userID uuid.UUID) ([]model.FriendWithUser, error) {
	var friends []model.Friend
	if err := r.DB.Where("user_id1 = ? OR user_id2 = ?", userID, userID).Find(&friends).Error; err != nil {
		return nil, err
	}
	
	if len(friends) == 0 {
		return []model.FriendWithUser{}, nil
	}
	
	var userIDs []uuid.UUID
	for _, f := range friends {
		userIDs = append(userIDs, f.UserID1, f.UserID2)
	}
	
	var users []model.User
	if err := r.DB.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	
	userMap := make(map[uuid.UUID]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}
	
	var result []model.FriendWithUser
	for _, f := range friends {
		result = append(result, model.FriendWithUser{
			UserID1:   f.UserID1,
			UserID2:   f.UserID2,
			User1:     userMap[f.UserID1],
			User2:     userMap[f.UserID2],
			CreatedAt: f.CreatedAt,
		})
	}
	
	return result, nil
}

func (r *FriendsRepository) RemoveFriend(userID1, userID2 uuid.UUID) error {
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}
	return r.DB.Where("user_id1 = ? AND user_id2 = ?", userID1, userID2).Delete(&model.Friend{}).Error
}

func (r *FriendsRepository) IsFriend(userID1, userID2 uuid.UUID) (bool, error) {
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}
	var count int64
	err := r.DB.Model(&model.Friend{}).Where("user_id1 = ? AND user_id2 = ?", userID1, userID2).Count(&count).Error
	return count > 0, err
}

func (r *FriendsRepository) HasPendingRequest(fromUserID, toUserID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Model(&model.FriendRequest{}).Where("from_user_id = ? AND to_user_id = ? AND status = ?", fromUserID, toUserID, model.FriendRequestStatusPending).Count(&count).Error
	return count > 0, err
}

func (r *FriendsRepository) GetBlockedUsers(userID uuid.UUID) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := r.DB.Where("from_user_id = ? AND status = ?", userID, model.FriendRequestStatusBlocked).Find(&requests).Error
	return requests, err
}

func (r *FriendsRepository) BlockUserDirect(blockerID, blockedID uuid.UUID) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	blocked, err := r.IsBlocked(blockerID, blockedID)
	if err != nil {
		return err
	}
	if blocked {
		return errors.New("user already blocked")
	}

	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	uid1, uid2 := blockerID, blockedID
	if uid1.String() > uid2.String() {
		uid1, uid2 = uid2, uid1
	}
	tx.Where("user_id1 = ? AND user_id2 = ?", uid1, uid2).Delete(&model.Friend{})

	var existing model.FriendRequest
	err = tx.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		blockerID, blockedID, blockedID, blockerID).First(&existing).Error

	if err == nil {
		existing.FromUserID = blockerID
		existing.ToUserID = blockedID
		existing.Status = model.FriendRequestStatusBlocked
		existing.UpdatedAt = time.Now()
		if err := tx.Save(&existing).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		req := &model.FriendRequest{
			FromUserID: blockerID,
			ToUserID:   blockedID,
			Status:     model.FriendRequestStatusBlocked,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := tx.Create(req).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *FriendsRepository) UnblockUser(blockerID, blockedID uuid.UUID) error {
	result := r.DB.Where("from_user_id = ? AND to_user_id = ? AND status = ?",
		blockerID, blockedID, model.FriendRequestStatusBlocked).Delete(&model.FriendRequest{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("block not found")
	}
	return nil
}

func (r *FriendsRepository) IsBlocked(userID1, userID2 uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Model(&model.FriendRequest{}).Where(
		"((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)) AND status = ?",
		userID1, userID2, userID2, userID1, model.FriendRequestStatusBlocked,
	).Count(&count).Error
	return count > 0, err
}
