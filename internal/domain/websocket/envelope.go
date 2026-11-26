package websocket

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Built-in infrastructure control event names (transport-level, non-business).
const (
	EventConnected    = "ws.connected"
	EventPing         = "ws.ping"
	EventPong         = "ws.pong"
	EventSubscribe    = "ws.subscribe"
	EventUnsubscribe  = "ws.unsubscribe"
	EventSubscribed   = "ws.subscribed"
	EventUnsubscribed = "ws.unsubscribed"
	EventError        = "ws.error"
)

// Envelope represents the generic JSON wire contract for WebSocket frames.
type Envelope struct {
	ID        string          `json:"id,omitempty"`
	Event     string          `json:"event"`
	Channel   string          `json:"channel,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// ConnectedData is emitted to a client upon successful WebSocket registration.
type ConnectedData struct {
	ConnID string `json:"conn_id"`
	UserID int64  `json:"user_id"`
}

// ChannelAckData is emitted when a client subscribes to or unsubscribes from a channel.
type ChannelAckData struct {
	Channel string `json:"channel"`
}

// ErrorData represents a structured transport-level error frame sent to a client.
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewEnvelope constructs a generic Envelope with UTC timestamp and optional JSON-encoded data.
func NewEnvelope(event string, data any) (Envelope, error) {
	event = strings.TrimSpace(event)
	if event == "" {
		return Envelope{}, fmt.Errorf("%w: event must not be empty", ErrInvalidEnvelope)
	}

	env := Envelope{
		Event:     event,
		Timestamp: time.Now().UTC(),
	}

	if data != nil {
		switch v := data.(type) {
		case json.RawMessage:
			env.Data = v
		case []byte:
			if len(v) > 0 {
				if !json.Valid(v) {
					return Envelope{}, fmt.Errorf("%w: data bytes are not valid JSON", ErrInvalidEnvelope)
				}
				env.Data = json.RawMessage(v)
			}
		default:
			raw, err := json.Marshal(data)
			if err != nil {
				return Envelope{}, fmt.Errorf("%w: marshal data: %v", ErrInvalidEnvelope, err)
			}
			env.Data = raw
		}
	}

	return env, nil
}

// WithID returns a copy of the envelope with the given message ID.
func (e Envelope) WithID(id string) Envelope {
	e.ID = strings.TrimSpace(id)
	return e
}

// WithChannel returns a copy of the envelope scoped to the given channel/room.
func (e Envelope) WithChannel(channel string) Envelope {
	e.Channel = strings.TrimSpace(channel)
	return e
}

// Validate verifies that the envelope contains a non-empty event name.
func (e Envelope) Validate() error {
	if strings.TrimSpace(e.Event) == "" {
		return fmt.Errorf("%w: event must not be empty", ErrInvalidEnvelope)
	}
	return nil
}

// DecodeData unmarshals the envelope's Data payload into target.
func (e Envelope) DecodeData(target any) error {
	if len(e.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(e.Data, target); err != nil {
		return fmt.Errorf("%w: decode data: %v", ErrInvalidEnvelope, err)
	}
	return nil
}
