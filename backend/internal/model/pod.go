// backend/internal/model/pod.go

package model

import "time"

type PodType string

const (
	PodTypeMain     PodType = "main"
	PodTypeHeadless PodType = "headless"
)

type PodStatus string

const (
	PodStatusHealthy  PodStatus = "healthy"
	PodStatusDegraded PodStatus = "degraded"
	PodStatusOffline  PodStatus = "offline"
)

type PodInfo struct {
	PodID            string    `json:"pod_id"`
	PodType         PodType   `json:"pod_type"`
	Status          PodStatus `json:"status"`
	IsCurrent       bool      `json:"is_current"`
	ConnectedClients int64     `json:"connected_clients"`
	OnlineUsers     int64     `json:"online_users"`
	ActiveGames     int64     `json:"active_games"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
	StartedAt       time.Time `json:"started_at"`
}

type PodsDiscovery struct {
	Pods []PodInfo `json:"pods"`
	Meta PodsMeta `json:"meta"`
}

type PodsMeta struct {
	TotalPods     int64 `json:"total_pods"`
	MainPods     int64 `json:"main_pods"`
	HeadlessPods int64 `json:"headless_pods"`
	TotalClients int64 `json:"total_clients"`
	TotalUsers  int64 `json:"total_users"`
}