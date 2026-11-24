package payload

import "time"

// TargetType defines the delivery scope for real-time WebSocket events.
type TargetType string

const (
	// TargetUser targets a single connected user by ID.
	TargetUser TargetType = "user"

	// TargetGroup targets all connected members of a group.
	TargetGroup TargetType = "group"

	// TargetBroadcast targets all active connected sessions.
	TargetBroadcast TargetType = "broadcast"
)

// WSPayload represents the message contract published to q.ws.classic.
type WSPayload struct {
	Event      string     `json:"event"`
	TargetType TargetType `json:"target_type"`
	TargetID   string     `json:"target_id,omitempty"`
	Data       any        `json:"data"`
	SenderID   string     `json:"sender_id,omitempty"`
	Timestamp  time.Time  `json:"timestamp"`
}
