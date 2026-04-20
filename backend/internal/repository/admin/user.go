// backend/internal/repository/admin/user.go

package admin

import (
	"context"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/identifier"
	"github.com/Culturae-org/culturae/internal/service"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AdminUserRepositoryInterface interface {
	GetAllUsers(roleFilter string, rankFilter string, accountStatusFilter string, isOnlineFilter *bool, limit int, offset int) ([]model.UserAdminView, error)
	GetUserCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error)
	GetUserOnlineCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error)
	GetWeeklyActiveUserCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error)
	GetUserByID(id string) (*model.User, error)
	UpdateUserByID(id string, userUpdate model.UserUpdate) (*model.User, error)
	DeactivateUserByID(id string) error
	GetUserConnectionLogs(id string, successFilter *bool) ([]model.UserConnectionLog, error)
	GetUserActiveSessions(id string) ([]model.Session, error)
	UpdateUserPassword(id string, hashedPassword string) error
	GetUserLevelStats() (map[string]int, error)
	GetUserRoleStats() (map[string]int, error)
	CreateUser(user model.User) (*model.User, error)
	GetUserCreationDates(startDate *time.Time, endDate *time.Time) ([]string, error)
	SearchUsers(query string, limit int, offset int) ([]model.UserAdminView, error)
	SearchUserCount(query string) (int, error)
	DeleteUserByID(id string) error
	UpdateAvatar(id string, hasAvatar bool) error
	RegeneratePublicID(id string) error
	UpdateBanFields(id string, bannedUntil *time.Time, banReason string) error
}

type AdminUserRepository struct {
	DB               *gorm.DB
	UserCacheService *service.UserCacheService
	logger           *zap.Logger
}

func NewAdminUserRepository(
	db *gorm.DB,
	userCacheService *service.UserCacheService,
	logger *zap.Logger,
) *AdminUserRepository {
	return &AdminUserRepository{
		DB:               db,
		UserCacheService: userCacheService,
		logger:           logger,
	}
}

func (r *AdminUserRepository) GetAllUsers(roleFilter string, rankFilter string, accountStatusFilter string, isOnlineFilter *bool, limit int, offset int) ([]model.UserAdminView, error) {
	var users []model.User
	query := r.DB.Preload("GameStats")
	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}
	if rankFilter != "" {
		query = query.Where("rank = ?", rankFilter)
	}
	if accountStatusFilter != "" {
		query = query.Where("account_status = ?", accountStatusFilter)
	}
	if isOnlineFilter != nil {
		query = query.Where("is_online = ?", *isOnlineFilter)
	}
	err := query.Limit(limit).Offset(offset).Find(&users).Error

	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, user := range users {
			if cacheErr := r.UserCacheService.SetUser(ctx, &user); cacheErr != nil {
				r.logger.Warn("Failed to cache user during GetAllUsers", zap.String("userID", user.ID.String()), zap.Error(cacheErr))
			}
		}
	}

	adminUsers := make([]model.UserAdminView, len(users))
	for i, user := range users {
		adminUsers[i] = *user.ToAdminView()
	}

	return adminUsers, err
}

func (r *AdminUserRepository) GetUserByID(id string) (*model.User, error) {
	var user model.User
	err := r.DB.Preload("GameStats").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminUserRepository) UpdateUserByID(id string, userUpdate model.UserUpdate) (*model.User, error) {
	var user model.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if userUpdate.Username != "" {
		user.Username = userUpdate.Username
	}
	if userUpdate.Email != "" {
		user.Email = userUpdate.Email
	}
	if userUpdate.Role != "" {
		user.Role = userUpdate.Role
	}
	if userUpdate.AccountStatus != "" {
		user.AccountStatus = userUpdate.AccountStatus
	}
	if userUpdate.Language != "" {
		user.Language = userUpdate.Language
	}
	if userUpdate.Bio != nil {
		user.Bio = userUpdate.Bio
	}

	if userUpdate.IsProfilePublic != nil {
		user.IsProfilePublic = *userUpdate.IsProfilePublic
	}
	if userUpdate.ShowOnlineStatus != nil {
		user.ShowOnlineStatus = *userUpdate.ShowOnlineStatus
	}
	if userUpdate.AllowFriendRequests != nil {
		user.AllowFriendRequests = *userUpdate.AllowFriendRequests
	}
	if userUpdate.AllowPartyInvites != nil {
		user.AllowPartyInvites = *userUpdate.AllowPartyInvites
	}

	if err := r.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cacheErr := r.UserCacheService.InvalidateUser(ctx, user.ID.String()); cacheErr != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", user.ID.String()), zap.Error(cacheErr))
	}

	return &user, nil
}

func (r *AdminUserRepository) DeactivateUserByID(id string) error {
	var user model.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		return err
	}

	user.AccountStatus = model.AccountStatusInactive
	if err := r.DB.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func (r *AdminUserRepository) GetUserConnectionLogs(id string, successFilter *bool) ([]model.UserConnectionLog, error) {
	var logs []model.UserConnectionLog
	query := r.DB.Where("user_id = ?", id)
	if successFilter != nil {
		query = query.Where("is_success = ?", *successFilter)
	}
	err := query.Find(&logs).Error
	return logs, err
}

func (r *AdminUserRepository) GetUserActiveSessions(id string) ([]model.Session, error) {
	var sessions []model.Session
	err := r.DB.Where("user_id = ?", id).Find(&sessions).Error
	return sessions, err
}

func (r *AdminUserRepository) GetUserCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error) {
	var count int64
	query := r.DB.Model(&model.User{})
	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}
	if rankFilter != "" {
		query = query.Where("rank = ?", rankFilter)
	}
	if accountStatusFilter != "" {
		query = query.Where("account_status = ?", accountStatusFilter)
	}
	err := query.Count(&count).Error
	return int(count), err
}

func (r *AdminUserRepository) GetUserOnlineCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error) {
	var count int64
	query := r.DB.Model(&model.User{}).Where("is_online = ?", true)
	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}
	if rankFilter != "" {
		query = query.Where("rank = ?", rankFilter)
	}
	if accountStatusFilter != "" {
		query = query.Where("account_status = ?", accountStatusFilter)
	}
	err := query.Count(&count).Error
	return int(count), err
}

func (r *AdminUserRepository) GetWeeklyActiveUserCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error) {
	var count int64
	query := r.DB.Model(&model.Session{}).
		Select("COUNT(DISTINCT user_id)").
		Where("is_active = ? AND last_used >= ?", true, time.Now().AddDate(0, 0, -7))
	if roleFilter != "" || rankFilter != "" || accountStatusFilter != "" {
		query = query.Joins("JOIN users ON users.id = sessions.user_id")
		if roleFilter != "" {
			query = query.Where("users.role = ?", roleFilter)
		}
		if rankFilter != "" {
			query = query.Where("users.rank = ?", rankFilter)
		}
		if accountStatusFilter != "" {
			query = query.Where("users.account_status = ?", accountStatusFilter)
		}
	}
	err := query.Count(&count).Error
	return int(count), err
}

func (r *AdminUserRepository) UpdateUserPassword(id string, hashedPassword string) error {
	if err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("password", hashedPassword).Error; err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cacheErr := r.UserCacheService.InvalidateUser(ctx, id); cacheErr != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", id), zap.Error(cacheErr))
	}

	return nil
}

func (r *AdminUserRepository) GetUserLevelStats() (map[string]int, error) {
	var results []struct {
		Rank  string
		Count int
	}
	err := r.DB.Model(&model.User{}).Select("rank, count(*) as count").Group("rank").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Rank] = result.Count
	}

	return stats, nil
}

func (r *AdminUserRepository) GetUserRoleStats() (map[string]int, error) {
	var results []struct {
		Role  string
		Count int
	}
	err := r.DB.Model(&model.User{}).Select("role, count(*) as count").Group("role").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Role] = result.Count
	}

	return stats, nil
}

func (r *AdminUserRepository) CreateUser(user model.User) (*model.User, error) {
	if err := r.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cacheErr := r.UserCacheService.SetUser(ctx, &user); cacheErr != nil {
		r.logger.Warn("Failed to cache user during CreateUser", zap.String("userID", user.ID.String()), zap.Error(cacheErr))
	}

	return &user, nil
}

func (r *AdminUserRepository) GetUserCreationDates(startDate *time.Time, endDate *time.Time) ([]string, error) {
	var dates []string
	query := r.DB.Model(&model.User{})

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		endOfDay := endDate.Add(24 * time.Hour)
		query = query.Where("created_at < ?", endOfDay)
	}

	err := query.Select("DATE(created_at) as date").Order("DATE(created_at)").Pluck("date", &dates).Error
	return dates, err
}

func (r *AdminUserRepository) SearchUsers(query string, limit int, offset int) ([]model.UserAdminView, error) {
	var users []model.User
	dbQuery := r.DB.Model(&model.User{}).Where("username ILIKE ? OR email ILIKE ? OR id::text ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")

	err := dbQuery.Limit(limit).Offset(offset).Find(&users).Error

	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, user := range users {
			if cacheErr := r.UserCacheService.SetUser(ctx, &user); cacheErr != nil {
				r.logger.Warn("Failed to cache user during SearchUsers", zap.String("userID", user.ID.String()), zap.Error(cacheErr))
			}
		}
	}

	adminUsers := make([]model.UserAdminView, len(users))
	for i, user := range users {
		adminUsers[i] = *user.ToAdminView()
	}

	return adminUsers, err
}

func (r *AdminUserRepository) SearchUserCount(query string) (int, error) {
	var count int64
	dbQuery := r.DB.Model(&model.User{}).Where("username ILIKE ? OR email ILIKE ? OR id::text ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")

	err := dbQuery.Count(&count).Error
	return int(count), err
}

func (r *AdminUserRepository) DeleteUserByID(id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&model.Session{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&model.GamePlayer{}).Error; err != nil {
			return err
		}
		if err := tx.Where("from_user_id = ? OR to_user_id = ?", id, id).Delete(&model.GameInvite{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id1 = ? OR user_id2 = ?", id, id).Delete(&model.Friend{}).Error; err != nil {
			return err
		}
		if err := tx.Where("from_user_id = ? OR to_user_id = ?", id, id).Delete(&model.FriendRequest{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.UserConnectionLog{}).Where("user_id = ?", id).Update("user_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.UserActionLog{}).Where("user_id = ?", id).Update("user_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.SecurityEventLog{}).Where("user_id = ?", id).Update("user_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.APIRequestLog{}).Where("user_id = ?", id).Update("user_id", nil).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.User{}, "id = ?", id).Error; err != nil {
			return err
		}

		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()
		if cacheErr := r.UserCacheService.InvalidateUser(cacheCtx, id); cacheErr != nil {
			r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", id), zap.Error(cacheErr))
		}

		return nil
	})
}

func (r *AdminUserRepository) UpdateAvatar(id string, hasAvatar bool) error {
	if err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("has_avatar", hasAvatar).Error; err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cacheErr := r.UserCacheService.InvalidateUser(ctx, id); cacheErr != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", id), zap.Error(cacheErr))
	}

	return nil
}

func (r *AdminUserRepository) RegeneratePublicID(id string) error {
	newPublicID := identifier.GeneratePublicID()
	return r.DB.Model(&model.User{}).Where("id = ?", id).Update("public_id", newPublicID).Error
}

func (r *AdminUserRepository) UpdateBanFields(id string, bannedUntil *time.Time, banReason string) error {
	updates := map[string]interface{}{
		"banned_until": bannedUntil,
		"ban_reason":   banReason,
	}
	if err := r.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cacheErr := r.UserCacheService.InvalidateUser(ctx, id); cacheErr != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", id), zap.Error(cacheErr))
	}

	return nil
}
