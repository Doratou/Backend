package websocket

import (
	"context"
	"net/http"
)

// Identity carries the authenticated subject resolved during the WebSocket handshake.
type Identity struct {
	UserID  int64
	Role    string
	TokenID string
	Meta    map[string]any
}

// Client represents a single active WebSocket connection managed by a Hub.
type Client interface {
	ID() string
	UserID() int64
	Identity() *Identity
	Channels() []string
	Send(env Envelope) error
	Close(code int, reason string) error
}

// Authenticator validates an incoming HTTP upgrade request and resolves the client's Identity.
type Authenticator interface {
	Authenticate(ctx context.Context, r *http.Request) (*Identity, error)
}

// InboundHandler processes application-level (non-control) envelopes received from a client.
type InboundHandler interface {
	HandleMessage(ctx context.Context, client Client, env Envelope) error
}

// InboundHandlerFunc adapts a function to the InboundHandler interface.
type InboundHandlerFunc func(ctx context.Context, client Client, env Envelope) error

// HandleMessage calls f(ctx, client, env).
func (f InboundHandlerFunc) HandleMessage(ctx context.Context, client Client, env Envelope) error {
	return f(ctx, client, env)
}

// HubStats exposes a point-in-time snapshot of active WebSocket resources.
type HubStats struct {
	ActiveConnections int `json:"active_connections"`
	UniqueUsers       int `json:"unique_users"`
	ActiveChannels    int `json:"active_channels"`
}

// Hub defines the transport-agnostic registry and message router for WebSocket connections.
type Hub interface {
	Register(client Client) error
	Unregister(connID string)
	Subscribe(connID string, channel string) error
	Unsubscribe(connID string, channel string) error
	SendToConn(ctx context.Context, connID string, env Envelope) error
	SendToUser(ctx context.Context, userID int64, env Envelope) error
	SendToUsers(ctx context.Context, userIDs []int64, env Envelope) error
	SendToChannel(ctx context.Context, channel string, env Envelope) error
	Broadcast(ctx context.Context, env Envelope) error
	Stats() HubStats
	Close(ctx context.Context) error
}
