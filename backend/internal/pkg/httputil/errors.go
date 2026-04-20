// backend/internal/pkg/httputil/errors.go

package httputil

const (
	ErrCodeMissingToken       = "AUTH_MISSING_TOKEN"
	ErrCodeEmptyToken         = "AUTH_EMPTY_TOKEN"
	ErrCodeInvalidToken       = "AUTH_INVALID_TOKEN"
	ErrCodeExpiredToken       = "AUTH_EXPIRED_TOKEN"
	ErrCodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	ErrCodeAccountBanned      = "AUTH_ACCOUNT_BANNED"
	ErrCodeAccountSuspended   = "AUTH_ACCOUNT_SUSPENDED"
	ErrCodeAccountInactive    = "AUTH_ACCOUNT_INACTIVE"
	ErrCodeSessionRevoked     = "AUTH_SESSION_REVOKED"
	ErrCodeRefreshExpired     = "AUTH_REFRESH_EXPIRED"

	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeInvalidFormat = "INVALID_FORMAT"
	ErrCodeMissingField  = "MISSING_FIELD"

	ErrCodeNotFound  = "NOT_FOUND"
	ErrCodeConflict  = "CONFLICT"
	ErrCodeForbidden = "FORBIDDEN"

	ErrCodeRateLimited = "RATE_LIMITED"

	ErrCodeGameNotFound       = "GAME_NOT_FOUND"
	ErrCodeGameFull           = "GAME_FULL"
	ErrCodeGameAlreadyStarted = "GAME_ALREADY_STARTED"
	ErrCodeGameNotReady       = "GAME_NOT_READY"

	ErrCodePasswordMismatch = "PASSWORD_MISMATCH"

	ErrCodeAccountDeleted = "ACCOUNT_DELETED"

	ErrCodeSelfAction = "SELF_ACTION"

	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"

	ErrCodeMaintenance = "MAINTENANCE"
)
