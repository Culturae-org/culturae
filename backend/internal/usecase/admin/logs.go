// backend/internal/usecase/admin/logs.go

package admin

import (
	"context"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"
	"github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
)

type AdminLogsUsecase struct {
	logsRepo      admin.AdminLogsRepositoryInterface
	metricsRepo   admin.MetricsRepositoryInterface
	cacheHealth   repository.CacheHealthRepositoryInterface
	storageHealth repository.StorageHealthRepositoryInterface
}

func NewAdminLogsUsecase(
	logsRepo admin.AdminLogsRepositoryInterface,
	metricsRepo admin.MetricsRepositoryInterface,
	cacheHealth repository.CacheHealthRepositoryInterface,
	storageHealth repository.StorageHealthRepositoryInterface,
) *AdminLogsUsecase {
	return &AdminLogsUsecase{
		logsRepo:      logsRepo,
		metricsRepo:   metricsRepo,
		cacheHealth:   cacheHealth,
		storageHealth: storageHealth,
	}
}

// -----------------------------------------------
// Admin Logs Usecase Methods
//
// - GetAdminActionLogs
// - GetUserActionLogs
// - GetSecurityEventLogs
// - GetConnectionLogs
// - GetSystemMetrics
// - GetAPIRequestStats
// - GetAdminActionStats
// - GetUserActionStats
// - GetAPIRequestTimestamps
// - CheckServiceStatus
//
// -----------------------------------------------

func (uc *AdminLogsUsecase) GetAdminActionLogs(limit, offset int, adminID *uuid.UUID, action, resource *string, isSuccess *bool, resourceID *uuid.UUID, startDate, endDate *time.Time) ([]model.AdminActionLog, int64, error) {
	return uc.logsRepo.GetAdminActionLogs(limit, offset, adminID, action, resource, isSuccess, resourceID, startDate, endDate)
}

func (uc *AdminLogsUsecase) GetUserActionLogs(limit, offset int, userID *uuid.UUID, action, category *string, startDate, endDate *time.Time) ([]model.UserActionLog, int64, error) {
	return uc.logsRepo.GetUserActionLogs(limit, offset, userID, action, category, startDate, endDate)
}

func (uc *AdminLogsUsecase) GetSecurityEventLogs(limit, offset int, userID *uuid.UUID, eventType *string, startDate, endDate *time.Time) ([]model.SecurityEventLog, int64, error) {
	return uc.logsRepo.GetSecurityEventLogs(limit, offset, userID, eventType, startDate, endDate)
}

func (uc *AdminLogsUsecase) GetConnectionLogs(limit, offset int, userID *uuid.UUID, isSuccess *bool, startDate, endDate *time.Time) ([]model.UserConnectionLog, int64, error) {
	return uc.logsRepo.GetConnectionLogs(limit, offset, userID, isSuccess, startDate, endDate)
}

func (uc *AdminLogsUsecase) GetSystemMetrics() (*model.SystemMetrics, error) {
	return uc.metricsRepo.GetSystemMetrics()
}

func (uc *AdminLogsUsecase) GetAPIRequestStats(startDate, endDate *time.Time) (*model.APIRequestStats, error) {
	return uc.metricsRepo.GetAPIRequestStats(startDate, endDate)
}

func (uc *AdminLogsUsecase) GetAdminActionStats(startDate, endDate *time.Time) (*model.AdminActionStats, error) {
	return uc.metricsRepo.GetAdminActionStats(startDate, endDate)
}

func (uc *AdminLogsUsecase) GetUserActionStats(startDate, endDate *time.Time) (*model.UserActionStats, error) {
	return uc.metricsRepo.GetUserActionStats(startDate, endDate)
}

func (uc *AdminLogsUsecase) GetAPIRequestTimestamps(method *string, statusCode *int, startDate, endDate *time.Time) ([]time.Time, error) {
	return uc.metricsRepo.GetAPIRequestTimestamps(method, statusCode, startDate, endDate)
}

func (uc *AdminLogsUsecase) CheckServiceStatus() ([]model.ServiceStatus, error) {
	services := []string{"postgres", "redis", "minio"}
	var statuses []model.ServiceStatus

	for _, service := range services {
		status := model.ServiceStatus{
			ServiceName: service,
			LastCheck:   time.Now(),
			Details:     make(map[string]interface{}),
		}

		start := time.Now()
		var err error
		var details map[string]interface{}

		switch service {
		case "postgres":
			err = uc.checkDatabaseHealth()
			if err == nil {
				details, _ = uc.metricsRepo.GetDatabaseInfo()
			}
		case "redis":
			err = uc.checkRedisHealth()
			if err == nil {
				redisInfoCtx, redisInfoCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer redisInfoCancel()
				details, _ = uc.cacheHealth.GetInfo(redisInfoCtx)
			}
		case "minio":
			err = uc.checkMinIOHealth()
			if err == nil {
				details, _ = uc.storageHealth.GetBucketInfo()
			}
		}

		status.ResponseTime = time.Since(start).Milliseconds()

		if err != nil {
			status.Status = "unhealthy"
			errorMsg := err.Error()
			status.ErrorMsg = &errorMsg
		} else {
			status.Status = "healthy"
			status.Details = details
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (uc *AdminLogsUsecase) checkDatabaseHealth() error {
	return uc.metricsRepo.CheckDatabaseHealth()
}

func (uc *AdminLogsUsecase) checkRedisHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return uc.cacheHealth.Ping(ctx)
}

func (uc *AdminLogsUsecase) checkMinIOHealth() error {
	return uc.storageHealth.CheckBucketExists()
}
