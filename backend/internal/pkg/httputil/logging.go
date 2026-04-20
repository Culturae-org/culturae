// backend/internal/pkg/httputil/logging.go

package httputil

import (
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var logger *zap.Logger

const (
	boolFalse = "false"
	boolTrue  = "true"
)

func SetLogger(l *zap.Logger) {
	logger = l
}

func init() {
	logger = zap.NewNop()
}

func LogAdminAction(loggingService service.LoggingServiceInterface, adminID uuid.UUID, adminName, action, resource string, resourceID *uuid.UUID, ip, userAgent string, details interface{}, success bool, errorMsg *string) {
	if err := loggingService.LogAdminAction(adminID, adminName, action, resource, resourceID, ip, userAgent, details, success, errorMsg); err != nil {
		logger.Error("Failed to log admin action", zap.Error(err))
	}
}

func LogFailedAuthAttempt(loggingService service.LoggingServiceInterface, c *gin.Context, userID *uuid.UUID, reason string) {
	if err := loggingService.LogConnectionAttempt(
		userID,
		nil,
		GetRealIP(c),
		c.GetHeader("User-Agent"),
		false,
		&reason,
	); err != nil {
		logger.Error("Failed to log failed auth attempt", zap.Error(err))
	}
}

func LogSuccessfulAuthAttempt(loggingService service.LoggingServiceInterface, c *gin.Context, userID uuid.UUID, sessionID *uuid.UUID) {
	if err := loggingService.LogConnectionAttempt(
		&userID,
		sessionID,
		GetRealIP(c),
		c.GetHeader("User-Agent"),
		true,
		nil,
	); err != nil {
		logger.Error("Failed to log successful auth attempt", zap.Error(err))
	}
}

func GenerateChangeSummary(changes map[string]interface{}) string {
	if len(changes) == 0 {
		return "No changes made"
	}

	summaries := []string{}
	for field, change := range changes {
		switch v := change.(type) {
		case map[string]string:
			summaries = append(summaries, field+": "+v["from"]+" → "+v["to"])
		case map[string]bool:
			from := boolFalse
			to := boolFalse
			if v["from"] {
				from = boolTrue
			}
			if v["to"] {
				to = boolTrue
			}
			summaries = append(summaries, field+": "+from+" → "+to)
		}
	}

	result := ""
	for i, s := range summaries {
		if i > 0 {
			result += "; "
		}
		result += s
	}
	return result
}
