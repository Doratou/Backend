package broker

import (
	"time"
)

// Headers represents key-value metadata attached to a message.
type Headers map[string]any

// Message is a domain-level, transport-agnostic message representation.
type Message struct {
	ID            string
	ContentType   string
	Body          []byte
	Headers       Headers
	CorrelationID string
	ReplyTo       string
	Timestamp     time.Time
}

// NewMessage creates a standard Message with default timestamp and headers.
func NewMessage(id string, body []byte) Message {
	return Message{
		ID:        id,
		Body:      body,
		Headers:   make(Headers),
		Timestamp: time.Now().UTC(),
	}
}

// NewJSONMessage creates a Message with application/json content type.
func NewJSONMessage(id string, body []byte) Message {
	msg := NewMessage(id, body)
	msg.ContentType = "application/json"
	return msg
}
