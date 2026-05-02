// backend/internal/repository/user.go

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepositoryInterface interface {
	Exists(email, username string) bool
	Create(user *model.User) error
	GetByIdentifier(identifier string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	GetByPublicID(publicID string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
	SearchByUsername(query string) ([]model.User, error)
	GetPublicProfile(userID string) (*model.PublicProfile, error)
	GetPublicProfiles(page, limit int) ([]model.UserSearchCard, error)
	SearchPublicProfiles(query string, page, limit int) ([]model.UserSearchCard, error)
	UpdateAvatar(userID string, hasAvatar bool) error
	GetTotalUserCount() (int64, error)
	GetConnectionLogs(userID uuid.UUID) ([]model.UserConnectionLog, error)
	GetAllUsers() ([]model.User, error)
	UpdateUserByID(id string, userUpdate model.UserUpdate) error
	UpdateUserRole(id string, role string) error
	GetUserConnectionLogs(userID string) ([]model.UserConnectionLog, error)
	UpdateUserStatus(userID uuid.UUID, status string) error
	UpdateUserOnlineStatus(userID uuid.UUID, isOnline bool) error
	UpdateUserGameStatus(userID uuid.UUID, gameID *uuid.UUID) error
	CreateWithoutHash(user *model.User) error
	UpdateDates(userID uuid.UUID, createdAt, updatedAt time.Time) error
	AddExperience(userID uuid.UUID, xp int64, cfg model.XPConfig) error
	UpdateGameStats(userID uuid.UUID, isWinner bool, isDrawn bool, score int, duration int64, mode string) error
	UpdateEloRating(userID uuid.UUID, newRating int) error
	GetUserGameStats(userID uuid.UUID) (*model.UserGameStats, error)
	GetUserGameStatsByMode(userID uuid.UUID) ([]model.UserGameStatsByMode, error)
	SetUserGameStats(userID uuid.UUID, stats model.UserGameStats) error
	GetLeaderboardGlobal(limit, offset int) ([]model.LeaderboardEntry, error)
	GetLeaderboardByElo(limit, offset int) ([]model.LeaderboardEntry, error)
	GetUserRankByScore(userID uuid.UUID) (int, error)
	GetUserRankByElo(userID uuid.UUID) (int, error)
}

type UserReader interface {
	GetByIdentifier(identifier string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	GetByPublicID(publicID string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetPublicProfile(userID string) (*model.PublicProfile, error)
	GetPublicProfiles(page, limit int) ([]model.UserSearchCard, error)
	SearchPublicProfiles(query string, page, limit int) ([]model.UserSearchCard, error)
	SearchByUsername(query string) ([]model.User, error)
	GetAllUsers() ([]model.User, error)
	GetTotalUserCount() (int64, error)
	Exists(email, username string) bool
}

type UserWriter interface {
	Create(user *model.User) error
	CreateWithoutHash(user *model.User) error
	Update(user *model.User) error
	Delete(id string) error
	UpdateUserByID(id string, userUpdate model.UserUpdate) error
	UpdateUserRole(id string, role string) error
	UpdateUserStatus(userID uuid.UUID, status string) error
	UpdateUserOnlineStatus(userID uuid.UUID, isOnline bool) error
	UpdateUserGameStatus(userID uuid.UUID, gameID *uuid.UUID) error
	UpdateAvatar(userID string, hasAvatar bool) error
	UpdateDates(userID uuid.UUID, createdAt, updatedAt time.Time) error
	AddExperience(userID uuid.UUID, xp int64, cfg model.XPConfig) error
	UpdateGameStats(userID uuid.UUID, isWinner bool, isDrawn bool, score int, duration int64, mode string) error
}


type UserRepository struct {
	DB               *gorm.DB
	UserCacheService *service.UserCacheService
	logger           *zap.Logger
}

func NewUserRepository(
	db *gorm.DB,
	userCacheService *service.UserCacheService,
	logger *zap.Logger,
) *UserRepository {
	return &UserRepository{
		DB:               db,
		UserCacheService: userCacheService,
		logger:           logger,
	}
}

func (r *UserRepository) Exists(email, username string) bool {
	var count int64
	err := r.DB.Model(&model.User{}).Where("email = ? OR username = ?", email, username).Count(&count).Error
	if err != nil {
		return false
	}
	return count > 0
}

func (r *UserRepository) Create(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Create(user).Error; err != nil {
		r.logger.Error("Failed to create user in database", zap.Error(err))
		return err
	}

	if err := r.DB.Create(&model.UserGameStats{UserID: user.ID}).Error; err != nil {
		r.logger.Warn("Failed to initialize user game stats", zap.String("userID", user.ID.String()), zap.Error(err))
	}

	if err := r.UserCacheService.SetUser(ctx, user); err != nil {
		r.logger.Warn("Failed to cache new user", zap.String("userID", user.ID.String()), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) GetByIdentifier(identifier string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if cacheErr := r.UserCacheService.SetUser(ctx, &user); cacheErr != nil {
			r.logger.Warn("Failed to cache user after username lookup", zap.String("userID", user.ID.String()), zap.Error(cacheErr))
		}
	}
	return &user, err
}

func (r *UserRepository) GetByID(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := r.UserCacheService.GetOrSetUser(ctx, id, func() (*model.User, error) {
		var user model.User
		if err := r.DB.Where("id = ?", id).First(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	})
	if err != nil {
		return nil, err
	}

	if user.Password == "" {
		var dbUser model.User
		if err := r.DB.Select("password").Where("id = ?", id).First(&dbUser).Error; err == nil {
			user.Password = dbUser.Password
		}
	}
	return user, nil
}

func (r *UserRepository) GetByPublicID(publicID string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("public_id = ?", publicID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if user.Password == "" {
		var existing model.User
		if err := r.DB.Select("password").Where("id = ?", user.ID).First(&existing).Error; err == nil {
			user.Password = existing.Password
		}
	}

	if err := r.DB.Save(user).Error; err != nil {
		r.logger.Error("Failed to update user in database", zap.String("userID", user.ID.String()), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.UpdateUserCache(ctx, user); err != nil {
		r.logger.Warn("Failed to update cache after DB update", zap.String("userID", user.ID.String()), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
		r.logger.Error("Failed to delete user from database", zap.String("userID", id), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, id); err != nil {
		r.logger.Warn("Failed to invalidate cache for deleted user", zap.String("userID", id), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) SearchByUsername(query string) ([]model.User, error) {
	var users []model.User
	if err := r.DB.Where("LOWER(username) LIKE LOWER(?)", "%"+query+"%").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetPublicProfile(userID string) (*model.PublicProfile, error) {
	var user model.User
	err := r.DB.Preload("GameStats").
		Where("id = ? AND is_profile_public = ?", userID, true).First(&user).Error
	if err != nil {
		return nil, err
	}

	publicProfile := &model.PublicProfile{
		PublicID:  user.PublicID,
		Username:  user.Username,
		HasAvatar: user.HasAvatar,
		CreatedAt: user.CreatedAt,
		Bio:       user.Bio,
		Language:  user.Language,
		Level:     user.Level,
		Rank:      user.Rank,
		Status:    user.Status,
		EloRating: user.EloRating,

		IsOnline:            user.IsOnline,
		ShowOnlineStatus:    user.ShowOnlineStatus,
		LastSeenAt:          user.LastSeenAt,
		AllowFriendRequests: user.AllowFriendRequests,
	}

	if user.GameStats != nil {
		publicProfile.TotalGames = user.GameStats.TotalGames
		publicProfile.GamesWon = user.GameStats.GamesWon
		publicProfile.GamesLost = user.GameStats.GamesLost
		publicProfile.DayStreak = user.GameStats.DayStreak
		publicProfile.BestDayStreak = user.GameStats.BestDayStreak
		publicProfile.TotalScore = user.GameStats.TotalScore
		publicProfile.AverageScore = user.GameStats.AverageScore
		publicProfile.PlayTime = user.GameStats.PlayTime
		publicProfile.LastGameAt = user.GameStats.LastGameAt
	}

	return publicProfile, nil
}

func (r *UserRepository) GetPublicProfiles(page, limit int) ([]model.UserSearchCard, error) {
	var cards []model.UserSearchCard
	offset := (page - 1) * limit

	err := r.DB.Model(&model.User{}).
		Select("public_id, username, has_avatar").
		Where("is_profile_public = ? AND account_status = ?", true, model.AccountStatusActive).
		Offset(offset).Limit(limit).
		Scan(&cards).Error
	return cards, err
}

func (r *UserRepository) SearchPublicProfiles(query string, page, limit int) ([]model.UserSearchCard, error) {
	var cards []model.UserSearchCard
	offset := (page - 1) * limit

	err := r.DB.Model(&model.User{}).
		Select("public_id, username, has_avatar").
		Where("LOWER(username) LIKE LOWER(?) AND is_profile_public = ? AND account_status = ?", "%"+query+"%", true, model.AccountStatusActive).
		Offset(offset).Limit(limit).
		Scan(&cards).Error
	return cards, err
}

func (r *UserRepository) UpdateAvatar(userID string, hasAvatar bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Model(&model.User{}).Where("id = ?", userID).Update("has_avatar", hasAvatar).Error; err != nil {
		r.logger.Error("Failed to update avatar status in database", zap.String("userID", userID), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, userID); err != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", userID), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) GetTotalUserCount() (int64, error) {
	var count int64
	err := r.DB.Model(&model.User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) GetConnectionLogs(userID uuid.UUID) ([]model.UserConnectionLog, error) {
	var logs []model.UserConnectionLog
	err := r.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *UserRepository) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := r.DB.Preload("GameStats").Find(&users).Error
	return users, err
}

func (r *UserRepository) UpdateUserByID(id string, userUpdate model.UserUpdate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Model(&model.User{}).Where("id = ?", id).Updates(userUpdate).Error; err != nil {
		r.logger.Error("Failed to update user in database", zap.String("userID", id), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, id); err != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", id), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) UpdateUserRole(id string, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error; err != nil {
		r.logger.Error("Failed to update role for user in database", zap.String("userID", id), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, id); err != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", id), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) GetUserConnectionLogs(userID string) ([]model.UserConnectionLog, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return r.GetConnectionLogs(id)
}

func (r *UserRepository) UpdateUserStatus(userID uuid.UUID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Model(&model.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		r.logger.Error("Failed to update user status in database", zap.String("userID", userID.String()), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, userID.String()); err != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", userID.String()), zap.Error(err))
	}

	r.logger.Info("User status updated", zap.String("userID", userID.String()), zap.String("status", status))
	return nil
}

func (r *UserRepository) UpdateUserOnlineStatus(userID uuid.UUID, isOnline bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := "offline"
	if isOnline {
		status = "online"
	}

	if err := r.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		keyStatus:    status,
		"is_online": isOnline,
	}).Error; err != nil {
		r.logger.Error("Failed to update user online status in database", zap.String("userID", userID.String()), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, userID.String()); err != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", userID.String()), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) UpdateUserGameStatus(userID uuid.UUID, gameID *uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Model(&model.User{}).Where("id = ?", userID).Update("current_game_id", gameID).Error; err != nil {
		r.logger.Error("Failed to update user game status in database", zap.String("userID", userID.String()), zap.Error(err))
		return err
	}

	if err := r.UserCacheService.InvalidateUser(ctx, userID.String()); err != nil {
		r.logger.Warn("Failed to invalidate cache for user", zap.String("userID", userID.String()), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateWithoutHash(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.DB.Create(user).Error; err != nil {
		r.logger.Error("Failed to create user from import", zap.Error(err))
		return err
	}

	if err := r.DB.Create(&model.UserGameStats{UserID: user.ID}).Error; err != nil {
		r.logger.Warn("Failed to initialize game stats for imported user", zap.String("userID", user.ID.String()), zap.Error(err))
	}

	if err := r.UserCacheService.SetUser(ctx, user); err != nil {
		r.logger.Warn("Failed to cache imported user", zap.String("userID", user.ID.String()), zap.Error(err))
	}

	return nil
}

func (r *UserRepository) UpdateDates(userID uuid.UUID, createdAt, updatedAt time.Time) error {
	return r.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"created_at": createdAt,
		keyUpdatedAt: updatedAt,
	}).Error
}

func (r *UserRepository) AddExperience(userID uuid.UUID, xp int64, cfg model.XPConfig) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Clauses(clause.Locking{Strength: lockStrengthUpdate}).First(&user, "id = ?", userID).Error; err != nil {
			return err
		}

		newXP := user.Experience + xp
		if newXP < 0 {
			newXP = 0
		}

		level := cfg.CalculateLevel(newXP)
		rank := cfg.RankFromLevel(level)

		return tx.Model(&user).Updates(map[string]interface{}{
			"experience": newXP,
			"level":      level,
			"rank":       rank,
		}).Error
	})
}

func (r *UserRepository) UpdateGameStats(userID uuid.UUID, isWinner bool, isDrawn bool, score int, duration int64, mode string) error {
	isSolo := mode == string(model.GameModeSolo)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return r.DB.Transaction(func(tx *gorm.DB) error {
		var stats model.UserGameStats
		if err := tx.Clauses(clause.Locking{Strength: lockStrengthUpdate}).
			Where("user_id = ?", userID).
			FirstOrCreate(&stats, model.UserGameStats{UserID: userID}).Error; err != nil {
			return err
		}

		stats.TotalGames++
		stats.TotalScore += int64(score)
		stats.PlayTime += duration

		if !isSolo {
			switch {
			case isDrawn:
				stats.GamesDrawn++
			case isWinner:
				stats.GamesWon++
			default:
				stats.GamesLost++
			}
		}

		if stats.LastGameAt != nil {
			lastDay := time.Date(stats.LastGameAt.Year(), stats.LastGameAt.Month(), stats.LastGameAt.Day(), 0, 0, 0, 0, stats.LastGameAt.Location())
			days := int(today.Sub(lastDay).Hours() / 24)
			switch days {
			case 0:
			case 1:
				stats.DayStreak++
				if stats.DayStreak > stats.BestDayStreak {
					stats.BestDayStreak = stats.DayStreak
				}
			default:
				stats.DayStreak = 1
			}
		} else {
			stats.DayStreak = 1
			if stats.BestDayStreak < 1 {
				stats.BestDayStreak = 1
			}
		}

		if stats.TotalGames > 0 {
			stats.AverageScore = float64(stats.TotalScore) / float64(stats.TotalGames)
		}
		stats.LastGameAt = &now
		stats.UpdatedAt = now

		if err := tx.Save(&stats).Error; err != nil {
			return err
		}

		if mode == "" {
			return nil
		}

		var modeStats model.UserGameStatsByMode
		if err := tx.Clauses(clause.Locking{Strength: lockStrengthUpdate}).
			Where("user_id = ? AND mode = ?", userID, mode).
			FirstOrCreate(&modeStats, model.UserGameStatsByMode{UserID: userID, Mode: mode}).Error; err != nil {
			return err
		}

		modeStats.TotalGames++
		modeStats.TotalScore += int64(score)
		modeStats.PlayTime += duration

		if !isSolo {
			switch {
			case isDrawn:
				modeStats.GamesDrawn++
			case isWinner:
				modeStats.GamesWon++
			default:
				modeStats.GamesLost++
			}
		}

		if modeStats.TotalGames > 0 {
			modeStats.AverageScore = float64(modeStats.TotalScore) / float64(modeStats.TotalGames)
		}
		modeStats.UpdatedAt = now

		return tx.Save(&modeStats).Error
	})
}

func (r *UserRepository) GetUserGameStats(userID uuid.UUID) (*model.UserGameStats, error) {
	var stats model.UserGameStats
	err := r.DB.Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *UserRepository) GetUserGameStatsByMode(userID uuid.UUID) ([]model.UserGameStatsByMode, error) {
	var stats []model.UserGameStatsByMode
	err := r.DB.Where("user_id = ?", userID).Find(&stats).Error
	return stats, err
}

func (r *UserRepository) SetUserGameStats(userID uuid.UUID, stats model.UserGameStats) error {
	now := time.Now()
	return r.DB.Exec(`
		INSERT INTO user_game_stats (id, user_id, total_games, games_won, games_lost, win_streak, best_win_streak, total_score, average_score, play_time, last_game_at, updated_at)
		VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			total_games = EXCLUDED.total_games,
			games_won = EXCLUDED.games_won,
			games_lost = EXCLUDED.games_lost,
			win_streak = EXCLUDED.win_streak,
			best_win_streak = EXCLUDED.best_win_streak,
			total_score = EXCLUDED.total_score,
			average_score = EXCLUDED.average_score,
			play_time = EXCLUDED.play_time,
			last_game_at = EXCLUDED.last_game_at,
			updated_at = EXCLUDED.updated_at
	`, userID,
		stats.TotalGames, stats.GamesWon, stats.GamesLost,
		stats.DayStreak, stats.BestDayStreak,
		stats.TotalScore, stats.AverageScore, stats.PlayTime,
		stats.LastGameAt, now).Error
}

func (r *UserRepository) UpdateEloRating(userID uuid.UUID, newRating int) error {
	return r.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"elo_rating":       newRating,
		"elo_games_played": gorm.Expr("elo_games_played + 1"),
	}).Error
}

func (r *UserRepository) GetLeaderboardGlobal(limit, offset int) ([]model.LeaderboardEntry, error) {
	var entries []model.LeaderboardEntry
	err := r.DB.Model(&model.User{}).
		Select("users.username, users.public_id, users.has_avatar, COALESCE(s.total_score, 0) as score, users.elo_rating").
		Joins("LEFT JOIN user_game_stats s ON s.user_id = users.id").
		Where("users.account_status = ? AND users.is_profile_public = ?", model.AccountStatusActive, true).
		Order("score DESC").
		Limit(limit).Offset(offset).
		Scan(&entries).Error
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Rank = offset + i + 1
	}
	return entries, nil
}

func (r *UserRepository) GetLeaderboardByElo(limit, offset int) ([]model.LeaderboardEntry, error) {
	var entries []model.LeaderboardEntry
	err := r.DB.Model(&model.User{}).
		Select("users.username, users.public_id, users.has_avatar, COALESCE(s.total_score, 0) as score, users.elo_rating").
		Joins("LEFT JOIN user_game_stats s ON s.user_id = users.id").
		Where("users.account_status = ? AND users.is_profile_public = ? AND users.elo_games_played > 0", model.AccountStatusActive, true).
		Order("users.elo_rating DESC").
		Limit(limit).Offset(offset).
		Scan(&entries).Error
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Rank = offset + i + 1
	}
	return entries, nil
}

func (r *UserRepository) GetUserRankByScore(userID uuid.UUID) (int, error) {
	var totalScore int64
	if err := r.DB.Model(&model.UserGameStats{}).
		Select("total_score").
		Where("user_id = ?", userID).
		Scan(&totalScore).Error; err != nil {
		return 0, err
	}
	var count int64
	err := r.DB.Model(&model.User{}).
		Joins("LEFT JOIN user_game_stats s ON s.user_id = users.id").
		Where("users.account_status = ? AND users.is_profile_public = ? AND COALESCE(s.total_score, 0) > ?", model.AccountStatusActive, true, totalScore).
		Count(&count).Error
	return int(count) + 1, err
}

func (r *UserRepository) GetUserRankByElo(userID uuid.UUID) (int, error) {
	var user model.User
	if err := r.DB.Select("elo_rating").First(&user, "id = ?", userID).Error; err != nil {
		return 0, err
	}
	var count int64
	err := r.DB.Model(&model.User{}).
		Where("account_status = ? AND is_profile_public = ? AND elo_games_played > 0 AND elo_rating > ?", model.AccountStatusActive, true, user.EloRating).
		Count(&count).Error
	return int(count) + 1, err
}
