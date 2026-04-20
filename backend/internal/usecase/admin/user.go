// backend/internal/usecase/admin/user.go

package admin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/crypto"
	"github.com/Culturae-org/culturae/internal/pkg/identifier"
	"github.com/Culturae-org/culturae/internal/repository"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/google/uuid"
)

type AdminUserUsecase struct {
	Repo        adminRepo.AdminUserRepositoryInterface
	SessionRepo repository.SessionRepositoryInterface
	LoggingSvc  service.LoggingServiceInterface
}

func NewAdminUserUsecase(
	repo adminRepo.AdminUserRepositoryInterface,
	sessionRepo repository.SessionRepositoryInterface,
	loggingSvc service.LoggingServiceInterface,
) *AdminUserUsecase {
	return &AdminUserUsecase{
		Repo:        repo,
		SessionRepo: sessionRepo,
		LoggingSvc:  loggingSvc,
	}
}

// -----------------------------------------------
// Admin User Usecase Methods
//
// - GetAllUsers
// - GetUserCount
// - GetUserOnlineCount
// - GetWeeklyActiveUserCount
// - SearchUsers
// - GetUserByID
// - UpdateUserByID
// - DeactivateUserByID
// - GetUserConnectionLogs
// - GetUserActiveSessions
// - UpdateUserPassword
// - UpdateUserStatusByID
// - RevokeAllUserSessions
// - GetUserLevelStats
// - GetUserRoleStats
// - CreateUser
// - CheckAdminExists
// - CreateDefaultAdmin
// - GetUserCreationDates
// - SearchUserCount
// - DeleteUserByID
// - RegeneratePublicID
// - BanUser
// - UnbanUser
// - CheckAndLiftExpiredBan
//
// -----------------------------------------------

func (uc *AdminUserUsecase) GetAllUsers(roleFilter string, rankFilter string, accountStatusFilter string, isOnlineFilter *bool, limit int, offset int) ([]model.UserAdminView, error) {
	return uc.Repo.GetAllUsers(roleFilter, rankFilter, accountStatusFilter, isOnlineFilter, limit, offset)
}

func (uc *AdminUserUsecase) GetUserCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error) {
	return uc.Repo.GetUserCount(roleFilter, rankFilter, accountStatusFilter)
}

func (uc *AdminUserUsecase) GetUserOnlineCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error) {
	return uc.Repo.GetUserOnlineCount(roleFilter, rankFilter, accountStatusFilter)
}

func (uc *AdminUserUsecase) GetWeeklyActiveUserCount(roleFilter string, rankFilter string, accountStatusFilter string) (int, error) {
	return uc.Repo.GetWeeklyActiveUserCount(roleFilter, rankFilter, accountStatusFilter)
}

func (uc *AdminUserUsecase) SearchUsers(query string, limit int, offset int) ([]model.UserAdminView, error) {
	return uc.Repo.SearchUsers(query, limit, offset)
}

func (uc *AdminUserUsecase) GetUserByID(id string) (*model.UserAdminView, error) {
	user, err := uc.Repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return user.ToAdminView(), nil
}

func (uc *AdminUserUsecase) UpdateUserByID(id string, userUpdate model.UserUpdate) (*model.UserAdminView, error) {
	// Fetch current to detect status changes
	current, err := uc.Repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	updatedUser, err := uc.Repo.UpdateUserByID(id, userUpdate)
	if err != nil {
		return nil, err
	}

	if userUpdate.AccountStatus != "" && userUpdate.AccountStatus != current.AccountStatus {
		if userUpdate.AccountStatus == model.AccountStatusInactive || userUpdate.AccountStatus == model.AccountStatusBanned {
			if err := uc.SessionRepo.RevokeAllUserSessions(updatedUser.ID); err != nil {
				if uc.LoggingSvc != nil {
					_ = uc.LoggingSvc.LogAdminAction(uuid.Nil, "system", "revoke_sessions_after_update", "user", &updatedUser.ID, "", "", map[string]interface{}{"new_status": userUpdate.AccountStatus}, false, nil)
				}
			}
		}
	}

	return updatedUser.ToAdminView(), nil
}

func (uc *AdminUserUsecase) DeactivateUserByID(id string) error {
	// Deactivate and revoke sessions when status changes to inactive
	user, err := uc.Repo.GetUserByID(id)
	if err != nil {
		return err
	}

	if user.AccountStatus == model.AccountStatusInactive {
		return nil
	}

	if err := uc.Repo.DeactivateUserByID(id); err != nil {
		return err
	}

	// Revoke all sessions for this user after deactivation
	if err := uc.SessionRepo.RevokeAllUserSessions(user.ID); err != nil {
		// best-effort log
		if uc.LoggingSvc != nil {
			_ = uc.LoggingSvc.LogAdminAction(uuid.Nil, "system", "revoke_sessions_after_deactivate", "user", &user.ID, "", "", nil, false, nil)
		}
	}

	return nil
}

func (uc *AdminUserUsecase) GetUserConnectionLogs(id string, successFilter *bool) ([]model.UserConnectionLog, error) {
	return uc.Repo.GetUserConnectionLogs(id, successFilter)
}

func (uc *AdminUserUsecase) GetUserActiveSessions(id string) ([]model.Session, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return uc.SessionRepo.GetActiveSessionsByUserID(userID)
}

func (uc *AdminUserUsecase) UpdateUserPassword(id string, password string) error {
	hashedPassword, err := crypto.HashPassword(password, model.GetArgonParams())
	if err != nil {
		return err
	}
	return uc.Repo.UpdateUserPassword(id, hashedPassword)
}

func (uc *AdminUserUsecase) UpdateUserStatusByID(id string, accountStatus string) error {
	user, err := uc.Repo.GetUserByID(id)
	if err != nil {
		return err
	}

	if user.AccountStatus == accountStatus {
		return nil
	}

	userUpdate := model.UserUpdate{AccountStatus: accountStatus}
	updatedUser, err := uc.Repo.UpdateUserByID(id, userUpdate)
	if err != nil {
		return err
	}

	if accountStatus == model.AccountStatusInactive || accountStatus == model.AccountStatusBanned {
		if err := uc.SessionRepo.RevokeAllUserSessions(updatedUser.ID); err != nil {
			if uc.LoggingSvc != nil {
				_ = uc.LoggingSvc.LogAdminAction(uuid.Nil, "system", "revoke_sessions_after_status_change", "user", &updatedUser.ID, "", "", map[string]interface{}{"new_status": accountStatus}, false, nil)
			}
		}
	}

	return nil
}

func (uc *AdminUserUsecase) RevokeAllUserSessions(userID uuid.UUID) error {
	return uc.SessionRepo.RevokeAllUserSessions(userID)
}

func (uc *AdminUserUsecase) GetUserLevelStats() (map[string]int, error) {
	return uc.Repo.GetUserLevelStats()
}

func (uc *AdminUserUsecase) GetUserRoleStats() (map[string]int, error) {
	return uc.Repo.GetUserRoleStats()
}

func (uc *AdminUserUsecase) CreateUser(createUser model.AdminCreateUser) (*model.UserAdminView, error) {
	hashedPassword, err := crypto.HashPassword(createUser.Password, model.GetArgonParams())
	if err != nil {
		return nil, err
	}

	user := model.User{
		Email:         createUser.Email,
		Username:      createUser.Username,
		Role:          createUser.Role,
		AccountStatus: createUser.AccountStatus,
		Password:      hashedPassword,
		PublicID:      identifier.GeneratePublicID(),
	}

	created, err := uc.Repo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return created.ToAdminView(), nil
}

func (uc *AdminUserUsecase) CheckAdminExists() (bool, error) {
	stats, err := uc.Repo.GetUserRoleStats()
	if err != nil {
		return false, err
	}
	count, exists := stats["administrator"]
	return exists && count > 0, nil
}

func (uc *AdminUserUsecase) CreateDefaultAdmin() error {
	hashedPassword, err := crypto.HashPassword("password123", model.GetArgonParams())
	if err != nil {
		return err
	}

	adminUser := model.User{
		Email:         "admin@culturae.me",
		Username:      "culturae-admin",
		Role:          "administrator",
		AccountStatus: model.AccountStatusActive,
		Password:      hashedPassword,
		PublicID:      identifier.GeneratePublicID(),
	}

	_, err = uc.Repo.CreateUser(adminUser)
	return err
}

func (uc *AdminUserUsecase) GetUserCreationDates(startDate *time.Time, endDate *time.Time) ([]string, error) {
	return uc.Repo.GetUserCreationDates(startDate, endDate)
}

func (uc *AdminUserUsecase) SearchUserCount(query string) (int, error) {
	return uc.Repo.SearchUserCount(query)
}

func (uc *AdminUserUsecase) DeleteUserByID(id string) error {
	user, err := uc.Repo.GetUserByID(id)
	if err != nil {
		return err
	}

	if err := uc.SessionRepo.RevokeAllUserSessions(user.ID); err != nil {
		if uc.LoggingSvc != nil {
			_ = uc.LoggingSvc.LogAdminAction(uuid.Nil, "system", "revoke_sessions_before_delete", "user", &user.ID, "", "", nil, false, nil)
		}
	}

	return uc.Repo.DeleteUserByID(id)
}


func (uc *AdminUserUsecase) RegeneratePublicID(id string) error {
	return uc.Repo.RegeneratePublicID(id)
}

// ParseBanDuration parses a ban duration string into a time.Duration.
// Accepts: "1h", "5h", "24h", "72h", "168h", "720h", "permanent", or custom like "30m", "2h", "7d".
func ParseBanDuration(duration string) (time.Duration, bool, error) {
	if duration == "permanent" {
		return 0, true, nil
	}

	// Support "Xd" notation for days
	if strings.HasSuffix(duration, "d") {
		daysStr := strings.TrimSuffix(duration, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return 0, false, fmt.Errorf("invalid duration: %s", duration)
		}
		return time.Duration(days) * 24 * time.Hour, false, nil
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0, false, fmt.Errorf("invalid duration: %s", duration)
	}
	if d <= 0 {
		return 0, false, fmt.Errorf("duration must be positive")
	}
	return d, false, nil
}

func (uc *AdminUserUsecase) BanUser(id string, duration string, reason string) (*model.UserAdminView, error) {
	user, err := uc.Repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	d, permanent, err := ParseBanDuration(duration)
	if err != nil {
		return nil, err
	}

	var bannedUntil *time.Time
	if !permanent {
		t := time.Now().Add(d)
		bannedUntil = &t
	}
	// permanent: bannedUntil stays nil, account_status=banned with no expiry

	userUpdate := model.UserUpdate{AccountStatus: model.AccountStatusBanned}
	updated, err := uc.Repo.UpdateUserByID(id, userUpdate)
	if err != nil {
		return nil, err
	}

	if err := uc.Repo.UpdateBanFields(id, bannedUntil, reason); err != nil {
		return nil, err
	}
	updated.BannedUntil = bannedUntil
	updated.BanReason = reason

	// Revoke sessions
	if user.AccountStatus != model.AccountStatusBanned {
		if err := uc.SessionRepo.RevokeAllUserSessions(updated.ID); err != nil {
			if uc.LoggingSvc != nil {
				_ = uc.LoggingSvc.LogAdminAction(uuid.Nil, "system", "revoke_sessions_after_ban", "user", &updated.ID, "", "", nil, false, nil)
			}
		}
	}

	return updated.ToAdminView(), nil
}

func (uc *AdminUserUsecase) UnbanUser(id string) (*model.UserAdminView, error) {
	userUpdate := model.UserUpdate{AccountStatus: model.AccountStatusActive}
	updated, err := uc.Repo.UpdateUserByID(id, userUpdate)
	if err != nil {
		return nil, err
	}

	if err := uc.Repo.UpdateBanFields(id, nil, ""); err != nil {
		return nil, err
	}
	updated.BannedUntil = nil
	updated.BanReason = ""

	return updated.ToAdminView(), nil
}

// CheckAndLiftExpiredBan checks if a user's ban has expired and lifts it if so.
// Returns true if the ban was lifted.
func (uc *AdminUserUsecase) CheckAndLiftExpiredBan(user *model.User) bool {
	if user.AccountStatus != model.AccountStatusBanned {
		return false
	}
	if user.BannedUntil == nil {
		return false // permanent ban
	}
	if time.Now().Before(*user.BannedUntil) {
		return false // ban not yet expired
	}

	// Ban expired, lift it
	userUpdate := model.UserUpdate{AccountStatus: model.AccountStatusActive}
	if _, err := uc.Repo.UpdateUserByID(user.ID.String(), userUpdate); err != nil {
		return false
	}
	if err := uc.Repo.UpdateBanFields(user.ID.String(), nil, ""); err != nil {
		return false
	}
	return true
}
