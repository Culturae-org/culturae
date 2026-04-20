// backend/internal/model/metrics.go

package model

import "time"

type ServiceStatus struct {
	ServiceName  string                 `json:"service_name"`
	Status       string                 `json:"status"`
	LastCheck    time.Time              `json:"last_check"`
	ResponseTime int64                  `json:"response_time_ms"`
	ErrorMsg     *string                `json:"error_msg,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

type SystemMetrics struct {
	TotalUsers       int64   `json:"total_users"`
	ActiveUsers      int64   `json:"active_users"`
	TotalSessions    int64   `json:"total_sessions"`
	ActiveSessions   int64   `json:"active_sessions"`
	TotalAPIRequests int64   `json:"total_api_requests"`
	ErrorRate        float64 `json:"error_rate"`
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
}
