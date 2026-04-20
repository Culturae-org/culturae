// backend/internal/usecase/admin/history.go

package admin

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HistoryEvent struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Version   string    `json:"version,omitempty"`
	Details   string    `json:"details"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	AdminName string    `json:"admin_name,omitempty"`

	Added   int `json:"added,omitempty"`
	Updated int `json:"updated,omitempty"`
	Skipped int `json:"skipped,omitempty"`
	Errors  int `json:"errors,omitempty"`
}

// -----------------------------------------------
// Admin History Methods
//
// - ListHistory
//
// -----------------------------------------------

func (u *AdminDatasetsUsecase) ListHistory(limit int, offset int, typeFilter *string) ([]HistoryEvent, int64, error) {
	logs, total, err := u.adminLogsRepo.GetAdminActionLogs(limit, offset, nil, nil, typeFilter, nil, nil, nil, nil)
	if err != nil {
		u.logger.Error("Failed to list admin actions", zap.Error(err))
		return nil, 0, err
	}

	var events []HistoryEvent
	for _, action := range logs {
		details := ""
		if action.Details != nil {
			details = string(action.Details)
		}
		if details == "" {
			details = fmt.Sprintf("%s on %s", action.Action, action.Resource)
		}

		event := HistoryEvent{
			ID:        action.ID,
			Type:      action.Resource,
			Action:    action.Action,
			Target:    action.Resource,
			Details:   details,
			Success:   action.IsSuccess,
			Timestamp: action.CreatedAt,
			AdminName: action.AdminName,
		}
		events = append(events, event)
	}

	return events, total, nil
}
