package websocket

import (
	"context"
	"net/http"
	"time"
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

// HubStats exposes a point-in-time snapshot of active WebSocket resources and cumulative counters.
type HubStats struct {
	InstanceID           string    `json:"instance_id"`
	Status               string    `json:"status"`
	StartedAt            time.Time `json:"started_at"`
	UptimeSeconds        int64     `json:"uptime_seconds"`
	ActiveConnections    int       `json:"active_connections"`
	AnonymousConnections int       `json:"anonymous_connections"`
	UniqueUsers          int       `json:"unique_users"`
	ActiveChannels       int       `json:"active_channels"`
	TotalAccepted        uint64    `json:"total_accepted"`
	TotalClosed          uint64    `json:"total_closed"`
	TotalRejectedLimit   uint64    `json:"total_rejected_limit"`
	TotalRejectedAuth    uint64    `json:"total_rejected_auth"`
	MessagesEnqueued     uint64    `json:"messages_enqueued"`
	MessagesSent         uint64    `json:"messages_sent"`
	MessagesDropped      uint64    `json:"messages_dropped"`
	MessagesReceived     uint64    `json:"messages_received"`
	InboundErrors        uint64    `json:"inbound_errors"`
	HeartbeatFailures    uint64    `json:"heartbeat_failures"`
}

// SnapshotFilter constrains the connections included in a HubSnapshot.
type SnapshotFilter struct {
	UserID  int64
	Channel string
	Limit   int
}

// ConnectionSnapshot describes the live state of a single WebSocket connection.
type ConnectionSnapshot struct {
	ConnID         string    `json:"conn_id"`
	UserID         int64     `json:"user_id"`
	Role           string    `json:"role,omitempty"`
	RemoteAddr     string    `json:"remote_addr,omitempty"`
	Channels       []string  `json:"channels"`
	SendQueueLen   int       `json:"send_queue_len"`
	SendQueueCap   int       `json:"send_queue_cap"`
	ConnectedAt    time.Time `json:"connected_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// ChannelSnapshot describes an active channel and its subscriber count.
type ChannelSnapshot struct {
	Channel     string `json:"channel"`
	Subscribers int    `json:"subscribers"`
}

// HubSnapshot captures a point-in-time view of hub metrics, channels, and connections for monitoring.
type HubSnapshot struct {
	Stats       HubStats             `json:"stats"`
	Channels    []ChannelSnapshot    `json:"channels"`
	Connections []ConnectionSnapshot `json:"connections"`
	Truncated   bool                 `json:"truncated"`
}

// Monitor exposes read-only WebSocket operational metrics and snapshots.
type Monitor interface {
	Stats() HubStats
	Snapshot(filter SnapshotFilter) HubSnapshot
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
	Snapshot(filter SnapshotFilter) HubSnapshot
	Close(ctx context.Context) error
}
