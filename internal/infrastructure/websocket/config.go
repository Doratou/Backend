package websocket

import (
	"time"

	"github.com/coder/websocket"
)

// SlowConsumerPolicy defines how the hub reacts when a client's outbound send buffer is full.
type SlowConsumerPolicy string

const (
	// SlowConsumerDisconnect closes the slow client connection so it does not stall or silently miss state.
	SlowConsumerDisconnect SlowConsumerPolicy = "disconnect"

	// SlowConsumerDrop drops the newest message for that client while keeping the connection alive.
	SlowConsumerDrop SlowConsumerPolicy = "drop"
)

// Config holds tuning parameters for the WebSocket server, hub, and per-connection pumps.
type Config struct {
	ReadLimit           int64
	SendBufferSize      int
	WriteTimeout        time.Duration
	PingInterval        time.Duration
	PongTimeout         time.Duration
	ShutdownTimeout     time.Duration
	MaxConnsPerUser     int
	OriginPatterns      []string
	InsecureSkipVerify  bool
	Subprotocols        []string
	CompressionMode     websocket.CompressionMode
	SlowConsumerPolicy  SlowConsumerPolicy
	RequireAuth         bool
	EnableControlEvents bool
	SendConnectedEvent  bool
	PubSubChannel       string
}

// DefaultConfig returns production-safe defaults for WebSocket infrastructure.
func DefaultConfig() Config {
	return Config{
		ReadLimit:           64 * 1024, // 64 KiB
		SendBufferSize:      64,
		WriteTimeout:        5 * time.Second,
		PingInterval:        30 * time.Second,
		PongTimeout:         10 * time.Second,
		ShutdownTimeout:     5 * time.Second,
		MaxConnsPerUser:     0,
		CompressionMode:     websocket.CompressionDisabled,
		SlowConsumerPolicy:  SlowConsumerDisconnect,
		RequireAuth:         true,
		EnableControlEvents: true,
		SendConnectedEvent:  true,
		PubSubChannel:       "doratou:ws:events",
	}
}

// withDefaults normalizes zero or invalid config fields to safe defaults.
func (c Config) withDefaults() Config {
	def := DefaultConfig()
	if c.ReadLimit == 0 {
		c.ReadLimit = def.ReadLimit
	}
	if c.SendBufferSize <= 0 {
		c.SendBufferSize = def.SendBufferSize
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = def.WriteTimeout
	}
	if c.PingInterval < 0 {
		c.PingInterval = def.PingInterval
	}
	if c.PongTimeout <= 0 {
		c.PongTimeout = def.PongTimeout
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = def.ShutdownTimeout
	}
	if c.SlowConsumerPolicy != SlowConsumerDrop && c.SlowConsumerPolicy != SlowConsumerDisconnect {
		c.SlowConsumerPolicy = def.SlowConsumerPolicy
	}
	if c.PubSubChannel == "" {
		c.PubSubChannel = def.PubSubChannel
	}
	return c
}
