// backend/internal/pkg/httputil/response.go

package httputil

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/pkg/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type successEnvelope struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

type paginatedEnvelope struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type errorBody struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, successEnvelope{Data: data})
}

func SuccessWithMessage(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, successEnvelope{Message: message, Data: data})
}

func SuccessList(c *gin.Context, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, paginatedEnvelope{Data: data, Pagination: pagination})
}

func SuccessPaginated(c *gin.Context, data gin.H, pag pagination.Params) {
	c.JSON(http.StatusOK, pag.MergeIntoGinH(data))
}

func SuccessRaw(c *gin.Context, status int, data gin.H) {
	c.JSON(status, data)
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func ErrorWithDetails(c *gin.Context, status int, code string, message string, details map[string]interface{}) {
	c.JSON(status, errorEnvelope{Error: errorBody{Code: code, Message: message, Details: details}})
}

func AbortWithError(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(status, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func AbortWithErrorDetails(c *gin.Context, status int, code string, message string, details map[string]interface{}) {
	c.AbortWithStatusJSON(status, errorEnvelope{Error: errorBody{Code: code, Message: message, Details: details}})
}

func HandleAccountStatus(c *gin.Context, accountStatus string, userID uuid.UUID, logFailed func(reason string)) bool {
	switch accountStatus {
	case "suspended":
		logFailed("account_suspended")
		Error(c, http.StatusForbidden, ErrCodeAccountSuspended, "Your account is temporarily suspended")
		return true
	case "banned":
		logFailed("account_banned")
		Error(c, http.StatusForbidden, ErrCodeAccountBanned, "Your account has been banned")
		return true
	case "inactive":
		logFailed("account_inactive")
		Error(c, http.StatusForbidden, ErrCodeAccountInactive, "Your account is inactive")
		return true
	case "deleted":
		logFailed("account_deleted")
		Error(c, http.StatusForbidden, ErrCodeAccountDeleted, "This account has been deleted")
		return true
	case "active":
		return false
	default:
		logFailed("invalid_account_status")
		Error(c, http.StatusForbidden, ErrCodeForbidden, "Invalid account status")
		return true
	}
}
