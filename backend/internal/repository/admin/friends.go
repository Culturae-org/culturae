package admin

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminFriendsRepositoryInterface interface {
	GetFriendRequestsForUser(userID uuid.UUID, limit, offset int, statusFilter *string, direction *string) ([]model.AdminFriendRequest, int64, error)
	GetFriendsForUser(userID uuid.UUID, limit, offset int) ([]model.AdminFriendship, int64, error)
}

type AdminFriendsRepository struct {
	DB *gorm.DB
}

func NewAdminFriendsRepository(db *gorm.DB) *AdminFriendsRepository {
	return &AdminFriendsRepository{DB: db}
}

func (r *AdminFriendsRepository) GetFriendRequestsForUser(userID uuid.UUID, limit, offset int, statusFilter *string, direction *string) ([]model.AdminFriendRequest, int64, error) {
	type row struct {
		ID           string    `gorm:"column:id"`
		FromUserID   string    `gorm:"column:from_user_id"`
		FromUsername string    `gorm:"column:from_username"`
		ToUserID     string    `gorm:"column:to_user_id"`
		ToUsername   string    `gorm:"column:to_username"`
		Status       string    `gorm:"column:status"`
		CreatedAt    time.Time `gorm:"column:created_at"`
		UpdatedAt    time.Time `gorm:"column:updated_at"`
	}

	base := r.DB.Table("friend_requests fr").
		Select("fr.id, fr.from_user_id, u1.username as from_username, fr.to_user_id, u2.username as to_username, fr.status, fr.created_at, fr.updated_at").
		Joins("LEFT JOIN users u1 ON u1.id = fr.from_user_id").
		Joins("LEFT JOIN users u2 ON u2.id = fr.to_user_id")

	if direction != nil {
		switch *direction {
		case "sent":
			base = base.Where("fr.from_user_id = ?", userID)
		case "received":
			base = base.Where("fr.to_user_id = ?", userID)
		default:
			base = base.Where("fr.from_user_id = ? OR fr.to_user_id = ?", userID, userID)
		}
	} else {
		base = base.Where("fr.from_user_id = ? OR fr.to_user_id = ?", userID, userID)
	}

	if statusFilter != nil {
		base = base.Where("fr.status = ?", *statusFilter)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []row
	if err := base.Order("fr.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	result := make([]model.AdminFriendRequest, len(rows))
	for i, r := range rows {
		fromID, _ := uuid.Parse(r.FromUserID)
		toID, _ := uuid.Parse(r.ToUserID)
		id, _ := uuid.Parse(r.ID)
		result[i] = model.AdminFriendRequest{
			ID:           id,
			FromUserID:   fromID,
			FromUsername: r.FromUsername,
			ToUserID:     toID,
			ToUsername:   r.ToUsername,
			Status:       r.Status,
			CreatedAt:    r.CreatedAt,
			UpdatedAt:    r.UpdatedAt,
		}
	}
	return result, total, nil
}

func (r *AdminFriendsRepository) GetFriendsForUser(userID uuid.UUID, limit, offset int) ([]model.AdminFriendship, int64, error) {
	type row struct {
		User1ID       string    `gorm:"column:user_id1"`
		User1Username string    `gorm:"column:user1_username"`
		User2ID       string    `gorm:"column:user_id2"`
		User2Username string    `gorm:"column:user2_username"`
		CreatedAt     time.Time `gorm:"column:created_at"`
	}

	base := r.DB.Table("friends f").
		Select("f.user_id1, u1.username as user1_username, f.user_id2, u2.username as user2_username, f.created_at").
		Joins("LEFT JOIN users u1 ON u1.id = f.user_id1").
		Joins("LEFT JOIN users u2 ON u2.id = f.user_id2").
		Where("f.user_id1 = ? OR f.user_id2 = ?", userID, userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []row
	if err := base.Order("f.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	result := make([]model.AdminFriendship, len(rows))
	for i, r := range rows {
		id1, _ := uuid.Parse(r.User1ID)
		id2, _ := uuid.Parse(r.User2ID)
		result[i] = model.AdminFriendship{
			ID:            r.User1ID + ":" + r.User2ID,
			User1ID:       id1,
			User1Username: r.User1Username,
			User2ID:       id2,
			User2Username: r.User2Username,
			CreatedAt:     r.CreatedAt,
		}
	}
	return result, total, nil
}
