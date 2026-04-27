// backend/internal/repository/notification.go

package repository

import (
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepositoryInterface interface {
	Create(n *model.Notification) error
	GetByUserID(userID uuid.UUID, limit, offset int, unreadOnly bool) ([]model.Notification, error)
	CountByUserID(userID uuid.UUID, unreadOnly bool) (int64, error)
	MarkAsRead(id, userID uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error
	DeleteGameInviteNotification(toUserID, inviteID uuid.UUID) error
}

type NotificationRepository struct {
	DB *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (r *NotificationRepository) Create(n *model.Notification) error {
	return r.DB.Create(n).Error
}

func (r *NotificationRepository) GetByUserID(userID uuid.UUID, limit, offset int, unreadOnly bool) ([]model.Notification, error) {
	var notifs []model.Notification
	q := r.DB.Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = false")
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifs).Error
	return notifs, err
}

func (r *NotificationRepository) CountByUserID(userID uuid.UUID, unreadOnly bool) (int64, error) {
	var count int64
	q := r.DB.Model(&model.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = false")
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkAsRead(id, userID uuid.UUID) error {
	result := r.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	return r.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) DeleteGameInviteNotification(toUserID, inviteID uuid.UUID) error {
	return r.DB.Where(
		"user_id = ? AND type = 'game_invite' AND data->>'invite_id' = ?",
		toUserID, inviteID.String(),
	).Delete(&model.Notification{}).Error
}
