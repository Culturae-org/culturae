package admin

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceLoggingRepository implements service.LoggingRepositoryInterface.
// The interface is defined in the service package to avoid import cycles.
type ServiceLoggingRepository struct {
	DB *gorm.DB
}

func NewServiceLoggingRepository(db *gorm.DB) *ServiceLoggingRepository {
	return &ServiceLoggingRepository{DB: db}
}

func (r *ServiceLoggingRepository) CreateAdminActionLog(log *model.AdminActionLog) error {
	return r.DB.Create(log).Error
}

func (r *ServiceLoggingRepository) GetUsernameByID(userID uuid.UUID) (string, error) {
	var user model.User
	if err := r.DB.Select("username").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	return user.Username, nil
}

func (r *ServiceLoggingRepository) CreateUserActionLog(log *model.UserActionLog) error {
	return r.DB.Create(log).Error
}

func (r *ServiceLoggingRepository) CreateSecurityEventLog(log *model.SecurityEventLog) error {
	return r.DB.Create(log).Error
}

func (r *ServiceLoggingRepository) CreateAPIRequestLog(log *model.APIRequestLog) error {
	return r.DB.Create(log).Error
}

func (r *ServiceLoggingRepository) CreateConnectionLog(log *model.UserConnectionLog) error {
	return r.DB.Create(log).Error
}

func (r *ServiceLoggingRepository) CheckDatabaseHealth() error {
	return r.DB.Exec("SELECT 1").Error
}

// GetSystemMetrics is a bonus method used by the LoggingService for legacy compatibility.
func (r *ServiceLoggingRepository) GetSystemMetrics() (*model.SystemMetrics, error) {
	var metrics model.SystemMetrics

	r.DB.Model(&model.User{}).Count(&metrics.TotalUsers)
	r.DB.Model(&model.Session{}).Where("is_active = ?", true).Count(&metrics.ActiveSessions)
	r.DB.Model(&model.Session{}).Count(&metrics.TotalSessions)

	yesterday := time.Now().Add(-24 * time.Hour)
	var apiStats struct {
		TotalRequests   int64
		ErrorCount      int64
		AvgResponseTime float64
	}
	r.DB.Model(&model.APIRequestLog{}).
		Where("created_at > ?", yesterday).
		Select("COUNT(*) as total_requests, SUM(CASE WHEN is_error THEN 1 ELSE 0 END) as error_count, AVG(duration) as avg_response_time").
		Scan(&apiStats)

	metrics.TotalAPIRequests = apiStats.TotalRequests
	if apiStats.TotalRequests > 0 {
		metrics.ErrorRate = float64(apiStats.ErrorCount) / float64(apiStats.TotalRequests) * 100
	}
	metrics.AvgResponseTime = apiStats.AvgResponseTime

	return &metrics, nil
}
