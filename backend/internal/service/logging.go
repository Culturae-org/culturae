// backend/internal/service/logging.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/infrastructure/storage"
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LoggingRepositoryInterface interface {
	CreateAdminActionLog(log *model.AdminActionLog) error
	GetUsernameByID(userID uuid.UUID) (string, error)
	CreateUserActionLog(log *model.UserActionLog) error
	CreateSecurityEventLog(log *model.SecurityEventLog) error
	CreateAPIRequestLog(log *model.APIRequestLog) error
	CreateConnectionLog(log *model.UserConnectionLog) error
	CheckDatabaseHealth() error
}

type LoggingServiceInterface interface {
	LogAdminAction(adminID uuid.UUID, adminName, action, resource string, resourceID *uuid.UUID, ip, userAgent string, details interface{}, success bool, errorMsg *string) error
	LogUserAction(userID uuid.UUID, action, ip, userAgent string, details interface{}, success bool, errorMsg *string) error
	LogSecurityEvent(userID *uuid.UUID, eventType string, details interface{}, ip, userAgent string, success bool, errorMsg *string) error
	LogAPIRequest(method, path string, statusCode int, userID *uuid.UUID, ip, userAgent string, requestSize, responseSize, duration int64, isError bool, errorMsg *string) error
	LogConnectionAttempt(userID *uuid.UUID, sessionID *uuid.UUID, ip, userAgent string, isSuccess bool, failureReason *string) error
	CheckServiceStatus() ([]model.ServiceStatus, error)
	APILoggingMiddleware() gin.HandlerFunc
}

type LoggingService struct {
	repo        LoggingRepositoryInterface
	RedisClient cache.RedisClientInterface
	MinIOClient storage.MinIOClientInterface
	logger      *zap.Logger
}

func NewLoggingService(repo LoggingRepositoryInterface, redisClient cache.RedisClientInterface, minioClient storage.MinIOClientInterface, logger *zap.Logger) *LoggingService {
	return &LoggingService{
		repo:        repo,
		RedisClient: redisClient,
		MinIOClient: minioClient,
		logger:      logger,
	}
}

func (ls *LoggingService) LogAdminAction(adminID uuid.UUID, adminName, action, resource string, resourceID *uuid.UUID, ip, userAgent string, details interface{}, success bool, errorMsg *string) error {
	detailsJSON, _ := json.Marshal(details)

	logEntry := model.AdminActionLog{
		AdminID:    adminID,
		AdminName:  adminName,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ip,
		UserAgent:  userAgent,
		Details:    detailsJSON,
		IsSuccess:  success,
		ErrorMsg:   errorMsg,
	}

	return ls.repo.CreateAdminActionLog(&logEntry)
}

func (ls *LoggingService) LogUserAction(userID uuid.UUID, action, ip, userAgent string, details interface{}, success bool, errorMsg *string) error {
	detailsJSON, _ := json.Marshal(details)

	username, err := ls.repo.GetUsernameByID(userID)
	if err != nil {
		return err
	}

	logEntry := model.UserActionLog{
		UserID:     &userID,
		Username:   username,
		Action:     action,
		Resource:   "user",
		ResourceID: &userID,
		IPAddress:  ip,
		UserAgent:  userAgent,
		Details:    detailsJSON,
		IsSuccess:  success,
		ErrorMsg:   errorMsg,
	}

	return ls.repo.CreateUserActionLog(&logEntry)
}

func (ls *LoggingService) LogSecurityEvent(userID *uuid.UUID, eventType string, details interface{}, ip, userAgent string, success bool, errorMsg *string) error {
	detailsJSON, _ := json.Marshal(details)

	logEntry := model.SecurityEventLog{
		UserID:    userID,
		EventType: eventType,
		IPAddress: ip,
		UserAgent: userAgent,
		Details:   detailsJSON,
		IsSuccess: success,
		ErrorMsg:  errorMsg,
	}

	return ls.repo.CreateSecurityEventLog(&logEntry)
}

func (ls *LoggingService) LogAPIRequest(method, path string, statusCode int, userID *uuid.UUID, ip, userAgent string, requestSize, responseSize, duration int64, isError bool, errorMsg *string) error {
	logEntry := model.APIRequestLog{
		Method:       method,
		Path:         path,
		StatusCode:   statusCode,
		UserID:       userID,
		IPAddress:    ip,
		UserAgent:    userAgent,
		RequestSize:  requestSize,
		ResponseSize: responseSize,
		Duration:     duration,
		IsError:      isError,
		ErrorMsg:     errorMsg,
	}

	return ls.repo.CreateAPIRequestLog(&logEntry)
}

func (ls *LoggingService) LogConnectionAttempt(userID *uuid.UUID, sessionID *uuid.UUID, ip, userAgent string, isSuccess bool, failureReason *string) error {
	logEntry := model.UserConnectionLog{
		UserID:        userID,
		SessionID:     sessionID,
		IPAddress:     ip,
		UserAgent:     userAgent,
		DeviceInfo:    nil,
		Location:      "",
		IsSuccess:     isSuccess,
		FailureReason: failureReason,
	}
	return ls.repo.CreateConnectionLog(&logEntry)
}

func (ls *LoggingService) CheckServiceStatus() ([]model.ServiceStatus, error) {
	services := []string{"postgres", "redis", "minio"}
	var statuses []model.ServiceStatus

	for _, service := range services {
		status := model.ServiceStatus{
			ServiceName: service,
			LastCheck:   time.Now(),
		}

		start := time.Now()
		var err error

		switch service {
		case "postgres":
			err = ls.repo.CheckDatabaseHealth()
		case "redis":
			err = ls.checkRedisHealth()
		case "minio":
			err = ls.checkMinIOHealth()
		}

		status.ResponseTime = time.Since(start).Milliseconds()

		if err != nil {
			status.Status = "unhealthy"
			errorMsg := err.Error()
			status.ErrorMsg = &errorMsg
		} else {
			status.Status = "healthy"
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (ls *LoggingService) checkRedisHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return ls.RedisClient.Ping(ctx)
}

func (ls *LoggingService) checkMinIOHealth() error {
	return ls.MinIOClient.CheckBucketExists()
}

func (ls *LoggingService) APILoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		start := time.Now()

		var requestSize int64
		if c.Request.ContentLength > 0 {
			requestSize = c.Request.ContentLength
		}

		c.Next()

		duration := time.Since(start).Milliseconds()

		responseSize := int64(c.Writer.Size())

		isError := c.Writer.Status() >= 400
		var errorMsg *string
		if isError {
			msg := fmt.Sprintf("HTTP %d", c.Writer.Status())
			errorMsg = &msg
		}

		var userID *uuid.UUID
		if val, exists := c.Get("user_id"); exists {
			switch v := val.(type) {
			case uuid.UUID:
				userID = &v
			case *uuid.UUID:
				userID = v
			case string:
				if uid, err := uuid.Parse(v); err == nil {
					userID = &uid
				}
			default:
				if s := fmt.Sprintf("%v", v); s != "" {
					if uid, err := uuid.Parse(s); err == nil {
						userID = &uid
					}
				}
			}
		}

		clientIP := c.ClientIP()
		if cfip := c.GetHeader("CF-Connecting-IP"); cfip != "" {
			clientIP = strings.TrimSpace(cfip)
		} else if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				clientIP = strings.TrimSpace(xff[:idx])
			} else {
				clientIP = strings.TrimSpace(xff)
			}
		} else if xri := c.GetHeader("X-Real-IP"); xri != "" {
			clientIP = strings.TrimSpace(xri)
		}

		go func() {
			if err := ls.LogAPIRequest(
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				userID,
				clientIP,
				c.Request.UserAgent(),
				requestSize,
				responseSize,
				duration,
				isError,
				errorMsg,
			); err != nil {
				ls.logger.Error("Failed to log API request", zap.Error(err))
			}
		}()
	}
}
