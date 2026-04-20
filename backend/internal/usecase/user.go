// backend/internal/usecase/user.go

package usecase

import (
	"errors"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/crypto"
	"github.com/Culturae-org/culturae/internal/pkg/identifier"
	"github.com/Culturae-org/culturae/internal/repository"
)

type UserUsecase struct {
	Reader      repository.UserReader
	Writer      repository.UserWriter
	SessionRepo repository.SessionRepositoryInterface
}

func NewUserUsecase(
	reader repository.UserReader,
	writer repository.UserWriter,
	sessionRepo repository.SessionRepositoryInterface,
) *UserUsecase {
	return &UserUsecase{
		Reader:      reader,
		Writer:      writer,
		SessionRepo: sessionRepo,
	}
}

// -----------------------------------------------
// User Usecase Methods
//
// - GetByIdentifier
// - GetByEmail
// - CheckUserExists
// - GetByID
// - UpdateUser
// - CreateUser
// - GetByPublicID
// - GetPublicProfiles
// - SearchPublicProfiles
// - GetPublicProfile
// - GetUserCount
// - ChangePassword
//
// -----------------------------------------------

func (uc *UserUsecase) GetByIdentifier(identifier string) (*model.User, error) {
	return uc.Reader.GetByIdentifier(identifier)
}

func (uc *UserUsecase) GetByEmail(email string) (*model.User, error) {
	return uc.Reader.GetByEmail(email)
}

func (uc *UserUsecase) CheckUserExists(email, username string) bool {
	return uc.Reader.Exists(email, username)
}

func (uc *UserUsecase) GetByID(id string) (*model.User, error) {
	return uc.Reader.GetByID(id)
}

func (uc *UserUsecase) UpdateUser(user *model.User) error {
	return uc.Writer.Update(user)
}

func (uc *UserUsecase) CreateUser(user *model.User) error {
	publicID, err := identifier.GenerateUniquePublicID(uc.Reader)
	if err != nil {
		return err
	}
	user.PublicID = publicID
	hashed, err := crypto.HashPassword(user.Password, model.GetArgonParams())
	if err != nil {
		return err
	}
	user.Password = hashed
	return uc.Writer.Create(user)
}


func (uc *UserUsecase) GetByPublicID(publicID string) (*model.User, error) {
	return uc.Reader.GetByPublicID(publicID)
}

func (uc *UserUsecase) GetPublicProfiles(page, limit int) ([]model.UserSearchCard, error) {
	return uc.Reader.GetPublicProfiles(page, limit)
}

func (uc *UserUsecase) SearchPublicProfiles(query string, page, limit int) ([]model.UserSearchCard, error) {
	return uc.Reader.SearchPublicProfiles(query, page, limit)
}

func (uc *UserUsecase) GetPublicProfile(userID string) (*model.PublicProfile, error) {
	return uc.Reader.GetPublicProfile(userID)
}

func (uc *UserUsecase) GetUserCount() (int64, error) {
	return uc.Reader.GetTotalUserCount()
}

func (uc *UserUsecase) ChangePassword(userID, currentPassword, newPassword string) error {
	user, err := uc.Reader.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	ok, err := crypto.CheckPassword(currentPassword, user.Password)
	if err != nil {
		return errors.New("failed to verify password")
	}
	if !ok {
		return errors.New("current password is incorrect")
	}

	if !crypto.IsValidPassword(newPassword) {
		return errors.New("new password does not meet requirements")
	}

	hashed, err := crypto.HashPassword(newPassword, model.GetArgonParams())
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashed
	return uc.Writer.Update(user)
}

func (uc *UserUsecase) SoftDeleteAccount(userID, password string) error {
	user, err := uc.Reader.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	ok, err := crypto.CheckPassword(password, user.Password)
	if err != nil {
		return errors.New("failed to verify password")
	}
	if !ok {
		return errors.New("password is incorrect")
	}

	shortID := user.ID.String()[:8]
	user.AccountStatus = model.AccountStatusDeleted
	user.Username = "deleted_" + shortID
	user.Email = "deleted_" + shortID + "@deleted.local"
	user.Bio = nil
	user.IsProfilePublic = false
	user.ShowOnlineStatus = false
	user.AllowFriendRequests = false
	user.AllowPartyInvites = false

	return uc.Writer.Update(user)
}

func (uc *UserUsecase) Authenticate(identifier, password string) (*model.User, error) {
	user, err := uc.Reader.GetByIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	ok, checkErr := crypto.CheckPassword(password, user.Password)
	if checkErr != nil {
		return nil, checkErr
	}
	if !ok {
		return user, model.ErrInvalidCredentials
	}

	if user.AccountStatus != model.AccountStatusActive {
		return nil, &model.ErrAccountStatus{Status: user.AccountStatus, UserID: user.ID}
	}

	return user, nil
}
