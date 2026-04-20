// backend/internal/repository/admin/metrics.go

package admin

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"

	"gorm.io/gorm"
)

type MetricsRepositoryInterface interface {
	GetSystemMetrics() (*model.SystemMetrics, error)
	GetAPIRequestStats(startDate, endDate *time.Time) (*model.APIRequestStats, error)
	GetAdminActionStats(startDate, endDate *time.Time) (*model.AdminActionStats, error)
	GetUserActionStats(startDate, endDate *time.Time) (*model.UserActionStats, error)
	GetAPIRequestTimestamps(method *string, statusCode *int, startDate, endDate *time.Time) ([]time.Time, error)
	GetDatabaseInfo() (map[string]interface{}, error)
	CheckDatabaseHealth() error
}

type MetricsRepository struct {
	DB *gorm.DB
}

func NewMetricsRepository(db *gorm.DB) *MetricsRepository {
	return &MetricsRepository{DB: db}
}

func (r *MetricsRepository) GetSystemMetrics() (*model.SystemMetrics, error) {
	var metrics model.SystemMetrics

	if err := r.DB.Model(&model.User{}).Count(&metrics.TotalUsers).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&model.User{}).
		Where("account_status = ?", model.AccountStatusActive).
		Count(&metrics.ActiveUsers).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&model.Session{}).Count(&metrics.TotalSessions).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&model.Session{}).
		Where("is_active = ?", true).
		Count(&metrics.ActiveSessions).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Model(&model.APIRequestLog{}).Count(&metrics.TotalAPIRequests).Error; err != nil {
		return nil, err
	}

	if metrics.TotalAPIRequests > 0 {
		var errorCount int64
		if err := r.DB.Model(&model.APIRequestLog{}).
			Where("is_error = ?", true).
			Count(&errorCount).Error; err != nil {
			return nil, err
		}
		metrics.ErrorRate = float64(errorCount) / float64(metrics.TotalAPIRequests) * 100

		var avgDuration *float64
		if err := r.DB.Model(&model.APIRequestLog{}).
			Select("AVG(duration)").
			Scan(&avgDuration).Error; err != nil {
			return nil, err
		}
		if avgDuration != nil {
			metrics.AvgResponseTime = *avgDuration
		}
	}

	return &metrics, nil
}

func (r *MetricsRepository) GetAPIRequestStats(startDate, endDate *time.Time) (*model.APIRequestStats, error) {
	var stats model.APIRequestStats
	query := r.DB.Model(&model.APIRequestLog{})

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	type summary struct {
		Total       int64
		ErrorCount  int64
		AvgDuration float64
	}
	var s summary
	if err := query.Select("COUNT(*) as total, SUM(CASE WHEN is_error THEN 1 ELSE 0 END) as error_count, AVG(duration) as avg_duration").Scan(&s).Error; err != nil {
		return nil, err
	}

	stats.TotalRequests = s.Total
	if s.Total > 0 {
		stats.ErrorRate = float64(s.ErrorCount) / float64(s.Total) * 100
	}
	stats.AvgResponseTime = s.AvgDuration

	type kv struct {
		Key   string
		Count int64
	}
	var byStatus []kv
	if err := query.Select("status_code::text as key, COUNT(*) as count").Group("status_code").Scan(&byStatus).Error; err != nil {
		return nil, err
	}
	stats.RequestsByStatus = make(map[string]int64, len(byStatus))
	for _, row := range byStatus {
		stats.RequestsByStatus[row.Key] = row.Count
	}

	var byMethod []kv
	if err := query.Select("method as key, COUNT(*) as count").Group("method").Scan(&byMethod).Error; err != nil {
		return nil, err
	}
	stats.RequestsByMethod = make(map[string]int64, len(byMethod))
	for _, row := range byMethod {
		stats.RequestsByMethod[row.Key] = row.Count
	}

	var byPath []kv
	if err := query.Select("path as key, COUNT(*) as count").Group("path").Order("count DESC").Limit(20).Scan(&byPath).Error; err != nil {
		return nil, err
	}
	stats.RequestsByPath = make(map[string]int64, len(byPath))
	for _, row := range byPath {
		stats.RequestsByPath[row.Key] = row.Count
	}

	return &stats, nil
}

func (r *MetricsRepository) GetAdminActionStats(startDate, endDate *time.Time) (*model.AdminActionStats, error) {
	var stats model.AdminActionStats

	base := r.DB.Model(&model.AdminActionLog{})
	if startDate != nil {
		base = base.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		base = base.Where("created_at <= ?", *endDate)
	}

	type summary struct {
		Total   int64
		Success int64
	}
	var s summary
	if err := base.Select("COUNT(*) as total, SUM(CASE WHEN is_success THEN 1 ELSE 0 END) as success").Scan(&s).Error; err != nil {
		return nil, err
	}
	stats.TotalActions = s.Total
	if s.Total > 0 {
		stats.SuccessRate = float64(s.Success) / float64(s.Total) * 100
	}

	type kv struct {
		Key   string
		Count int64
	}
	var byType []kv
	if err := base.Select("action as key, COUNT(*) as count").Group("action").Scan(&byType).Error; err != nil {
		return nil, err
	}
	stats.ActionsByType = make(map[string]int64, len(byType))
	for _, row := range byType {
		stats.ActionsByType[row.Key] = row.Count
	}

	var byResource []kv
	if err := base.Select("resource as key, COUNT(*) as count").Group("resource").Scan(&byResource).Error; err != nil {
		return nil, err
	}
	stats.ActionsByResource = make(map[string]int64, len(byResource))
	for _, row := range byResource {
		stats.ActionsByResource[row.Key] = row.Count
	}

	var byAdmin []kv
	if err := base.Select("admin_name as key, COUNT(*) as count").Group("admin_name").Scan(&byAdmin).Error; err != nil {
		return nil, err
	}
	stats.ActionsByAdmin = make(map[string]int64, len(byAdmin))
	for _, row := range byAdmin {
		stats.ActionsByAdmin[row.Key] = row.Count
	}

	return &stats, nil
}

func (r *MetricsRepository) GetUserActionStats(startDate, endDate *time.Time) (*model.UserActionStats, error) {
	var stats model.UserActionStats
	query := r.DB.Model(&model.UserActionLog{})

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var totalActions int64
	if err := query.Count(&totalActions).Error; err != nil {
		return nil, err
	}

	stats.TotalActions = totalActions

	return &stats, nil
}

func (r *MetricsRepository) GetAPIRequestTimestamps(method *string, statusCode *int, startDate, endDate *time.Time) ([]time.Time, error) {
	query := r.DB.Model(&model.APIRequestLog{})

	if method != nil && *method != "" {
		query = query.Where("UPPER(method) = UPPER(?)", *method)
	}
	if statusCode != nil {
		query = query.Where("status_code = ?", *statusCode)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var results []time.Time
	if err := query.Select("created_at").Order("created_at ASC").Pluck("created_at", &results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MetricsRepository) GetDatabaseInfo() (map[string]interface{}, error) {
	sqlDB, err := r.DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	info := map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use_connections":   stats.InUse,
		"idle_connections":     stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"total_connections":    stats.MaxLifetimeClosed + stats.MaxIdleClosed,
	}

	return info, nil
}

func (r *MetricsRepository) CheckDatabaseHealth() error {
	sqlDB, err := r.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
