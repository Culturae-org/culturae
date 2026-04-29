// backend/internal/handler/user.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
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
	const defaultLimit = 20
	const maxLimit = 100

	limit := defaultLimit
	if l, err := strconv.Atoi(ctx.Query("limit")); err == nil {
		if l < 1 {
			l = 1
		}
		if l > maxLimit {
			l = maxLimit
		}
		limit = l
	}

	page := 1
	if p, err := strconv.Atoi(ctx.Query("page")); err == nil && p > 0 {
		page = p
	}

	query := ctx.Query("q")

	var cards []model.UserSearchCard
	var err error

	if query != "" {
		cards, err = pc.Usecase.SearchPublicProfiles(query, page, limit)
	} else {
		cards, err = pc.Usecase.GetPublicProfiles(page, limit)
	}

	if err != nil {
		httputil.Error(ctx, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to search profiles")
		return
	}

	if cards == nil {
		cards = []model.UserSearchCard{}
	}

	httputil.Success(ctx, http.StatusOK, gin.H{
		"profiles": cards,
		"page":     page,
		"limit":    limit,
		"has_more": len(cards) == limit,
	})
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
		if err.Error() == "user not found" || err.Error() == "profile is private" || err.Error() == "user blocked" {
			httputil.Error(ctx, http.StatusNotFound, httputil.ErrCodeNotFound, "Profile not found")
			return
		}
		httputil.Error(ctx, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(ctx, http.StatusOK, profile)
}
