package bootstrap

import (
	"strings"
	"time"
)

// ResolvedWS contains WebSocket infrastructure configuration with environment overrides
// merged cleanly with default constants.
type ResolvedWS struct {
	Enabled            bool
	AuthRequired       bool
	InsecureSkipVerify bool
	OriginPatterns     []string
	ReadLimitBytes     int64
	SendBufferSize     int
	WriteTimeout       time.Duration
	PingInterval       time.Duration
	PongTimeout        time.Duration
	ShutdownTimeout    time.Duration
	MaxConnsPerUser    int
	SlowConsumerPolicy string
	PubSubChannel      string
}

// ResolveWS merges Environment.WS settings with Constant.WS defaults.
func (c *Config) ResolveWS() ResolvedWS {
	env := c.Environment.WS
	def := c.Constant.WS

	return ResolvedWS{
		Enabled:            pickBool(env.Enabled, def.DefaultEnabled),
		AuthRequired:       pickBool(env.AuthRequired, def.DefaultAuthRequired),
		InsecureSkipVerify: env.InsecureSkipVerify,
		OriginPatterns:     splitCSV(env.OriginPatterns),
		ReadLimitBytes:     pickPositive(env.ReadLimitBytes, def.DefaultReadLimitBytes),
		SendBufferSize:     pickPositive(env.SendBufferSize, def.DefaultSendBufferSize),
		WriteTimeout:       pickPositive(env.WriteTimeout, def.DefaultWriteTimeout),
		PingInterval:       pickPositive(env.PingInterval, def.DefaultPingInterval),
		PongTimeout:        pickPositive(env.PongTimeout, def.DefaultPongTimeout),
		ShutdownTimeout:    pickPositive(env.ShutdownTimeout, def.DefaultShutdownTimeout),
		MaxConnsPerUser:    pickPositive(env.MaxConnsPerUser, def.DefaultMaxConnsPerUser),
		SlowConsumerPolicy: pickNonEmpty(env.SlowConsumerPolicy, def.DefaultSlowConsumerPolicy),
		PubSubChannel:      pickNonEmpty(env.PubSubChannel, def.DefaultPubSubChannel),
	}
}

func pickBool(val *bool, fallback bool) bool {
	if val != nil {
		return *val
	}
	return fallback
}

func pickPositive[T ~int | ~int64](val, fallback T) T {
	if val > 0 {
		return val
	}
	return fallback
}

func pickNonEmpty(val, fallback string) string {
	if trimmed := strings.TrimSpace(val); trimmed != "" {
		return trimmed
	}
	return fallback
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
