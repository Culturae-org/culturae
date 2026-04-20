package admin

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminLogsRepositoryInterface interface {
	GetAdminActionLogs(limit, offset int, adminID *uuid.UUID, action, resource *string, isSuccess *bool, resourceID *uuid.UUID, startDate, endDate *time.Time) ([]model.AdminActionLog, int64, error)
	GetUserActionLogs(limit, offset int, userID *uuid.UUID, action, category *string, startDate, endDate *time.Time) ([]model.UserActionLog, int64, error)
	GetSecurityEventLogs(limit, offset int, userID *uuid.UUID, eventType *string, startDate, endDate *time.Time) ([]model.SecurityEventLog, int64, error)
	GetConnectionLogs(limit, offset int, userID *uuid.UUID, isSuccess *bool, startDate, endDate *time.Time) ([]model.UserConnectionLog, int64, error)
}

type AdminLogsRepository struct {
	DB *gorm.DB
}

func NewAdminLogsRepository(db *gorm.DB) *AdminLogsRepository {
	return &AdminLogsRepository{DB: db}
}

func (r *AdminLogsRepository) GetAdminActionLogs(limit, offset int, adminID *uuid.UUID, action, resource *string, isSuccess *bool, resourceID *uuid.UUID, startDate, endDate *time.Time) ([]model.AdminActionLog, int64, error) {
	var logs []model.AdminActionLog
	var total int64

	query := r.DB.Model(&model.AdminActionLog{})

	if adminID != nil {
		query = query.Where("admin_id = ?", adminID)
	}
	if action != nil && *action != "" {
		query = query.Where("action ILIKE ?", "%"+*action+"%")
	}
	if resource != nil && *resource != "" {
		query = query.Where("resource ILIKE ?", "%"+*resource+"%")
	}
	if isSuccess != nil {
		query = query.Where("is_success = ?", *isSuccess)
	}
	if resourceID != nil {
		query = query.Where("resource_id = ?", resourceID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

var userActionCategories = map[string][]string{
	"profile": {"profile_update", "public_id_regeneration", "change_password", "delete_account"},
	"avatar":  {"avatar_upload", "avatar_delete"},
	"auth":    {"login"},
	"game":    {"create_game", "invite_to_game", "accept_game_invite", "reject_game_invite", "join_game", "leave_game", "start_game", "cancel_game"},
	"friends": {"send_friend_request", "accept_friend_request", "reject_friend_request", "cancel_friend_request", "block_friend_request", "remove_friend", "block_user", "unblock_user"},
}

func (r *AdminLogsRepository) GetUserActionLogs(limit, offset int, userID *uuid.UUID, action *string, category *string, startDate, endDate *time.Time) ([]model.UserActionLog, int64, error) {
	var logs []model.UserActionLog
	var total int64

	query := r.DB.Model(&model.UserActionLog{})

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if action != nil && *action != "" {
		query = query.Where("action ILIKE ?", "%"+*action+"%")
	}
	if category != nil && *category != "" {
		if actions, ok := userActionCategories[*category]; ok {
			query = query.Where("action IN ?", actions)
		}
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *AdminLogsRepository) GetSecurityEventLogs(limit, offset int, userID *uuid.UUID, eventType *string, startDate, endDate *time.Time) ([]model.SecurityEventLog, int64, error) {
	var logs []model.SecurityEventLog
	var total int64

	query := r.DB.Model(&model.SecurityEventLog{})

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if eventType != nil && *eventType != "" {
		query = query.Where("event_type ILIKE ?", "%"+*eventType+"%")
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *AdminLogsRepository) GetConnectionLogs(limit, offset int, userID *uuid.UUID, isSuccess *bool, startDate, endDate *time.Time) ([]model.UserConnectionLog, int64, error) {
	var logs []model.UserConnectionLog
	var total int64

	query := r.DB.Model(&model.UserConnectionLog{})

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if isSuccess != nil {
		query = query.Where("is_success = ?", *isSuccess)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *AdminLogsRepository) GetSystemMetrics() (*model.SystemMetrics, error) {
	var metrics model.SystemMetrics

	r.DB.Model(&model.User{}).Count(&metrics.TotalUsers)

	r.DB.Model(&model.Session{}).Where("is_active = ? AND expires_at > ?", true, time.Now()).Count(&metrics.ActiveSessions)

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

func (r *AdminLogsRepository) GetAPIRequestStats(startDate, endDate *time.Time) (*model.APIRequestStats, error) {
	stats := &model.APIRequestStats{
		RequestsByStatus: make(map[string]int64),
		RequestsByMethod: make(map[string]int64),
		RequestsByPath:   make(map[string]int64),
	}

	var totalRequests int64
	query := r.DB.Model(&model.APIRequestLog{}).Select("COUNT(*)")

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	err := query.Scan(&totalRequests).Error
	if err != nil {
		return nil, err
	}
	stats.TotalRequests = totalRequests

	var errorCount int64
	queryError := r.DB.Model(&model.APIRequestLog{}).Where("is_error = ?", true)
	if startDate != nil {
		queryError = queryError.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		queryError = queryError.Where("created_at <= ?", *endDate)
	}
	err = queryError.Count(&errorCount).Error
	if err != nil {
		return nil, err
	}
	if totalRequests > 0 {
		stats.ErrorRate = float64(errorCount) / float64(totalRequests) * 100
	}

	var avgResponseTime float64
	queryAvg := r.DB.Model(&model.APIRequestLog{}).Select("AVG(duration)")
	if startDate != nil {
		queryAvg = queryAvg.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		queryAvg = queryAvg.Where("created_at <= ?", *endDate)
	}
	err = queryAvg.Scan(&avgResponseTime).Error
	if err != nil {
		return nil, err
	}
	stats.AvgResponseTime = avgResponseTime

	var statusResults []struct {
		StatusCode string
		Count      int64
	}
	query = r.DB.Model(&model.APIRequestLog{}).Select("CAST(status_code AS TEXT) as status_code, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("status_code").Scan(&statusResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range statusResults {
		stats.RequestsByStatus[result.StatusCode] = result.Count
	}

	var methodResults []struct {
		Method string
		Count  int64
	}
	query = r.DB.Model(&model.APIRequestLog{}).Select("method, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("method").Scan(&methodResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range methodResults {
		stats.RequestsByMethod[result.Method] = result.Count
	}

	var pathResults []struct {
		Path  string
		Count int64
	}
	query = r.DB.Model(&model.APIRequestLog{}).Select("path, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("path").Order("count DESC").Limit(20).Scan(&pathResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range pathResults {
		stats.RequestsByPath[result.Path] = result.Count
	}

	var days int
	if startDate != nil && endDate != nil {
		days = int(endDate.Sub(*startDate).Hours()/24) + 1
	} else {
		days = 30
	}
	if days > 0 {
		stats.DailyAverage = float64(totalRequests) / float64(days)
	}

	return stats, nil
}

func (r *AdminLogsRepository) GetAdminActionStats(startDate, endDate *time.Time) (*model.AdminActionStats, error) {
	stats := &model.AdminActionStats{
		ActionsByType:     make(map[string]int64),
		ActionsByResource: make(map[string]int64),
		ActionsByAdmin:    make(map[string]int64),
	}

	var totalActions int64
	query := r.DB.Model(&model.AdminActionLog{}).Select("COUNT(*)")

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	err := query.Scan(&totalActions).Error
	if err != nil {
		return nil, err
	}
	stats.TotalActions = totalActions

	var successCount int64
	querySuccess := r.DB.Model(&model.AdminActionLog{}).Where("is_success = ?", true)
	if startDate != nil {
		querySuccess = querySuccess.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		querySuccess = querySuccess.Where("created_at <= ?", *endDate)
	}
	err = querySuccess.Count(&successCount).Error
	if err != nil {
		return nil, err
	}
	if totalActions > 0 {
		stats.SuccessRate = float64(successCount) / float64(totalActions) * 100
	}

	var typeResults []struct {
		Action string
		Count  int64
	}
	query = r.DB.Model(&model.AdminActionLog{}).Select("action, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("action").Scan(&typeResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range typeResults {
		stats.ActionsByType[result.Action] = result.Count
	}

	var resourceResults []struct {
		Resource string
		Count    int64
	}
	query = r.DB.Model(&model.AdminActionLog{}).Select("resource, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("resource").Scan(&resourceResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range resourceResults {
		stats.ActionsByResource[result.Resource] = result.Count
	}

	var adminResults []struct {
		AdminName string
		Count     int64
	}
	query = r.DB.Model(&model.AdminActionLog{}).Select("admin_name, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("admin_name").Scan(&adminResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range adminResults {
		stats.ActionsByAdmin[result.AdminName] = result.Count
	}

	var days int
	if startDate != nil && endDate != nil {
		days = int(endDate.Sub(*startDate).Hours()/24) + 1
	} else {
		days = 30
	}
	if days > 0 {
		stats.DailyAverage = float64(totalActions) / float64(days)
	}

	return stats, nil
}

func (r *AdminLogsRepository) GetUserActionStats(startDate, endDate *time.Time) (*model.UserActionStats, error) {
	stats := &model.UserActionStats{
		ActionsByType: make(map[string]int64),
		ActionsByUser: make(map[string]int64),
	}

	var totalActions int64
	query := r.DB.Model(&model.UserActionLog{}).Select("COUNT(*)")

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	err := query.Scan(&totalActions).Error
	if err != nil {
		return nil, err
	}
	stats.TotalActions = totalActions

	var successCount int64
	querySuccess := r.DB.Model(&model.UserActionLog{}).Where("is_success = ?", true)
	if startDate != nil {
		querySuccess = querySuccess.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		querySuccess = querySuccess.Where("created_at <= ?", *endDate)
	}
	err = querySuccess.Count(&successCount).Error
	if err != nil {
		return nil, err
	}
	if totalActions > 0 {
		stats.SuccessRate = float64(successCount) / float64(totalActions) * 100
	}

	var typeResults []struct {
		Action string
		Count  int64
	}
	query = r.DB.Model(&model.UserActionLog{}).Select("action, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("action").Scan(&typeResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range typeResults {
		stats.ActionsByType[result.Action] = result.Count
	}

	var userResults []struct {
		Username string
		Count    int64
	}
	query = r.DB.Model(&model.UserActionLog{}).Select("username, COUNT(*) as count")
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	err = query.Group("username").Scan(&userResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range userResults {
		stats.ActionsByUser[result.Username] = result.Count
	}

	var days int
	if startDate != nil && endDate != nil {
		days = int(endDate.Sub(*startDate).Hours()/24) + 1
	} else {
		days = 30
	}
	if days > 0 {
		stats.DailyAverage = float64(totalActions) / float64(days)
	}

	return stats, nil
}

func (r *AdminLogsRepository) GetAPIRequestTimestamps(method *string, statusCode *int, startDate, endDate *time.Time) ([]time.Time, error) {
	var timestamps []time.Time
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

	err := query.Select("created_at").
		Order("created_at ASC").
		Pluck("created_at", &timestamps).Error
	return timestamps, err
}

func (r *AdminLogsRepository) CheckDatabaseHealth() error {
	return r.DB.Exec("SELECT 1").Error
}

func (r *AdminLogsRepository) GetDatabaseInfo() (map[string]interface{}, error) {
	details := make(map[string]interface{})

	var version string
	err := r.DB.Raw("SELECT version()").Scan(&version).Error
	if err != nil {
		return nil, err
	}
	details["version"] = version

	var activeConnections int64
	err = r.DB.Raw("SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConnections).Error
	if err != nil {
		return nil, err
	}
	details["active_connections"] = activeConnections

	var totalConnections int64
	err = r.DB.Raw("SELECT count(*) FROM pg_stat_activity").Scan(&totalConnections).Error
	if err != nil {
		return nil, err
	}
	details["total_connections"] = totalConnections

	var dbSize string
	err = r.DB.Raw("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize).Error
	if err != nil {
		return nil, err
	}
	details["database_size"] = dbSize

	var cacheHitRatio float64
	err = r.DB.Raw(`
		SELECT 
			round(100 * sum(blks_hit) / (sum(blks_hit) + sum(blks_read)), 2) as cache_hit_ratio
		FROM pg_stat_database
		WHERE datname = current_database()
	`).Scan(&cacheHitRatio).Error
	if err == nil {
		details["cache_hit_ratio_percent"] = cacheHitRatio
	}

	var lockCount int64
	err = r.DB.Raw("SELECT count(*) FROM pg_locks").Scan(&lockCount).Error
	if err == nil {
		details["active_locks"] = lockCount
	}

	var longQueries int64
	err = r.DB.Raw(`
		SELECT count(*) 
		FROM pg_stat_activity 
		WHERE state = 'active' 
		AND now() - query_start > interval '1 minute'
	`).Scan(&longQueries).Error
	if err == nil {
		details["long_running_queries"] = longQueries
	}

	return details, nil
}
