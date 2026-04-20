package admin

import (
	"github.com/Culturae-org/culturae/internal/model"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/google/uuid"
)

type AdminFriendsUsecase struct {
	Repo adminRepo.AdminFriendsRepositoryInterface
}

func NewAdminFriendsUsecase(repo adminRepo.AdminFriendsRepositoryInterface) *AdminFriendsUsecase {
	return &AdminFriendsUsecase{Repo: repo}
}

// -----------------------------------------------
// Admin Friends Usecase Methods
//
// - GetFriendRequestsForUser
// - GetFriendsForUser
//
// -----------------------------------------------

func (uc *AdminFriendsUsecase) GetFriendRequestsForUser(userID uuid.UUID, limit, offset int, statusFilter *string, direction *string) ([]model.AdminFriendRequest, int64, error) {
	return uc.Repo.GetFriendRequestsForUser(userID, limit, offset, statusFilter, direction)
}

func (uc *AdminFriendsUsecase) GetFriendsForUser(userID uuid.UUID, limit, offset int) ([]model.AdminFriendship, int64, error) {
	return uc.Repo.GetFriendsForUser(userID, limit, offset)
}
