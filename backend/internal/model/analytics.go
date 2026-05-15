package model

type DailyActiveUserStat struct {
	Date           string `json:"date"`
	ActiveUsers    int    `json:"active_users"`
	ActiveSessions int    `json:"active_sessions"`
}

type DailyUserGrowthStat struct {
	Date       string `json:"date"`
	NewUsers   int    `json:"new_users"`
	TotalUsers int    `json:"total_users"`
}
