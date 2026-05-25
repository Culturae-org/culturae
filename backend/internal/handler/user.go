// backend/internal/handler/user.go

package handler

import (
	"errors"
	"net/http"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	Usecase        *usecase.UserUsecase
	FriendsUsecase *usecase.FriendsUsecase
}

func NewUserHandler(
	usecase *usecase.UserUsecase,
	friendsUsecase *usecase.FriendsUsecase,
) *UserHandler {
	return &UserHandler{
		Usecase:        usecase,
		FriendsUsecase: friendsUsecase,
	}
}

// -----------------------------------------------------
// User Handlers
//
// - SearchPublicProfiles
// - GetUserProfileWithRelationship
// -----------------------------------------------------

func (pc *UserHandler) SearchPublicProfiles(ctx *gin.Context) {
	pag := pagination.Parse(ctx, pagination.Config{DefaultLimit: 20, MaxLimit: 100})

	query := ctx.Query("q")

	var cards []model.UserSearchCard
	var err error

	if query != "" {
		cards, err = pc.Usecase.SearchPublicProfiles(query, pag.Page, pag.Limit)
	} else {
		cards, err = pc.Usecase.GetPublicProfiles(pag.Page, pag.Limit)
	}

	if err != nil {
		httputil.Error(ctx, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to search profiles")
		return
	}

	if cards == nil {
		cards = []model.UserSearchCard{}
	}

	if len(cards) == pag.Limit {
		pag.WithTotal(int64(pag.Offset + len(cards) + 1))
	} else {
		pag.WithTotal(int64(pag.Offset + len(cards)))
	}
	httputil.SuccessList(ctx, cards, &pag)
}

func (pc *UserHandler) GetUserProfileWithRelationship(ctx *gin.Context) {
	publicID := ctx.Param("publicID")
	if publicID == "" {
		httputil.Error(ctx, http.StatusBadRequest, httputil.ErrCodeMissingField, "Public ID is required")
		return
	}

	var viewerID uuid.UUID
	userID, exists := ctx.Get("user_id")
	if exists {
		if vid, ok := userID.(uuid.UUID); ok {
			viewerID = vid
		}
	}

	profile, err := pc.FriendsUsecase.GetUserProfileWithRelationship(viewerID, publicID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) || errors.Is(err, model.ErrProfilePrivate) || errors.Is(err, model.ErrUserBlocked) {
			httputil.Error(ctx, http.StatusNotFound, httputil.ErrCodeNotFound, "Profile not found")
			return
		}
		httputil.Error(ctx, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(ctx, http.StatusOK, profile)
}
