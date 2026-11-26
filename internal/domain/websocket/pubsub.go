package websocket

import (
	"context"
	"encoding/json"
	"time"
)

// RelayTargetType identifies how a relayed event should be routed across connected WebSocket clients.
type RelayTargetType string

const (
	RelayTargetUsers     RelayTargetType = "users"
	RelayTargetChannel   RelayTargetType = "channel"
	RelayTargetBroadcast RelayTargetType = "broadcast"
)

// IsValid reports whether the relay target type is supported.
func (t RelayTargetType) IsValid() bool {
	switch t {
	case RelayTargetUsers, RelayTargetChannel, RelayTargetBroadcast:
		return true
	default:
		return false
	}
}

// RelayMessage represents a cross-instance WebSocket dispatch command sent over Pub/Sub.
type RelayMessage struct {
	ID         string          `json:"id,omitempty"`
	Event      string          `json:"event"`
	TargetType RelayTargetType `json:"target_type"`
	TargetIDs  []int64         `json:"target_ids,omitempty"`
	Channel    string          `json:"channel,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
	SenderID   int64           `json:"sender_id,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

// Publisher publishes real-time WebSocket relay messages to connected backend instances.
type Publisher interface {
	Publish(ctx context.Context, msg RelayMessage) error
}
