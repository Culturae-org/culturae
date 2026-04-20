// backend/internal/service/websocket_messages.go

package service

type BaseMessage struct {
	Type          string  `json:"type"`
	CorrelationID *string `json:"correlation_id,omitempty"`
}

type JoinGameMessage struct {
	BaseMessage
	GamePublicID string `json:"game_public_id"`
}

type LeaveGameMessage struct {
	BaseMessage
	GamePublicID *string `json:"game_public_id,omitempty"`
}

type AnswerPayload struct {
	Type       string      `json:"type"`
	Answer     interface{} `json:"answer"`
	TimeSpent *int64 `json:"time_spent,omitempty"`
	QuestionID *string     `json:"question_id,omitempty"`
}

type SubmitAnswerMessage struct {
	BaseMessage
	GamePublicID string        `json:"game_public_id"`
	Answer       AnswerPayload `json:"answer"`
}

type PlayerReadyMessage struct {
	BaseMessage
	GamePublicID string `json:"game_public_id"`
	Ready        bool   `json:"ready"`
}

type StartGameMessage struct {
	BaseMessage
	GamePublicID string `json:"game_public_id"`
}

type PingMessage struct {
	BaseMessage
}

type AckMessage struct {
	Type          string  `json:"type"`
	Action        string  `json:"action"`
	Success       bool    `json:"success"`
	Error         *string `json:"error,omitempty"`
	CorrelationID *string `json:"correlation_id,omitempty"`
}

type PongMessage struct {
	Type string `json:"type"`
}

type HelloMessage struct {
	Type              string `json:"type"`
	ProtocolVersion   int    `json:"protocol_version"`
	ServerTime        string `json:"server_time"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
}

type AdminNotification struct {
	Event      string                 `json:"event"`
	Data       map[string]interface{} `json:"data"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	ActionURL  string                 `json:"action_url,omitempty"`
}
