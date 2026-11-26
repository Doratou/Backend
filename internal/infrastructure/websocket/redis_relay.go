package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/redis/go-redis/v9"
)

const defaultRedisRelayChannel = "doratou:ws:events"

// RedisRelay implements domain_ws.Publisher and subscribes to Redis Pub/Sub to fan out
// WebSocket events across all backend instances to their local hubs.
type RedisRelay struct {
	client  *redis.Client
	channel string
}

// NewRedisRelay constructs a RedisRelay bound to the given Redis client and Pub/Sub channel.
func NewRedisRelay(client *redis.Client, channel string) *RedisRelay {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		channel = defaultRedisRelayChannel
	}
	return &RedisRelay{
		client:  client,
		channel: channel,
	}
}

// Channel returns the Redis Pub/Sub channel name used by this relay.
func (r *RedisRelay) Channel() string {
	return r.channel
}

// Publish serializes msg and broadcasts it over Redis Pub/Sub to all listening app instances.
func (r *RedisRelay) Publish(ctx context.Context, msg domain_ws.RelayMessage) error {
	if r == nil || r.client == nil {
		return errors.New("websocket redis relay: redis client is nil")
	}
	if strings.TrimSpace(msg.Event) == "" {
		return fmt.Errorf("%w: event must not be empty", domain_ws.ErrInvalidEnvelope)
	}
	if !msg.TargetType.IsValid() {
		return fmt.Errorf("%w: unsupported relay target type %q", domain_ws.ErrInvalidEnvelope, msg.TargetType)
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("websocket redis relay: marshal message: %w", err)
	}

	if err := r.client.Publish(ctx, r.channel, raw).Err(); err != nil {
		return fmt.Errorf("websocket redis relay: publish to %q: %w", r.channel, err)
	}
	return nil
}

// Run continuously listens to the Redis Pub/Sub channel and dispatches incoming events to hub
// until ctx is canceled. It automatically resubscribes after transient Redis failures.
func (r *RedisRelay) Run(ctx context.Context, hub domain_ws.Hub) error {
	if r == nil || r.client == nil {
		return errors.New("websocket redis relay: redis client is nil")
	}
	if hub == nil {
		return errors.New("websocket redis relay: hub is nil")
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := r.consumeSession(ctx, hub)
		if err == nil || errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return ctx.Err()
		}

		log.Printf("[WS REDIS RELAY] session ended (%v); retrying in 2s...", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (r *RedisRelay) consumeSession(ctx context.Context, hub domain_ws.Hub) error {
	pubsub := r.client.Subscribe(ctx, r.channel)
	defer func() {
		_ = pubsub.Close()
	}()

	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe %q: %w", r.channel, err)
	}
	log.Printf("[WS REDIS RELAY] listening on channel %q", r.channel)

	msgCh := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgCh:
			if !ok {
				return errors.New("pubsub channel closed")
			}
			if err := DispatchRelayPayload(ctx, hub, []byte(msg.Payload)); err != nil {
				log.Printf("[WS REDIS RELAY] dispatch error: %v", err)
			}
		}
	}
}

// DispatchRelayPayload parses a JSON-encoded RelayMessage and routes it through hub.
func DispatchRelayPayload(ctx context.Context, hub domain_ws.Hub, raw []byte) error {
	if hub == nil {
		return errors.New("websocket redis relay: hub is nil")
	}

	var msg domain_ws.RelayMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal relay message: %v", domain_ws.ErrInvalidEnvelope, err)
	}

	env, err := domain_ws.NewEnvelope(msg.Event, msg.Data)
	if err != nil {
		return err
	}
	if msg.ID != "" {
		env = env.WithID(msg.ID)
	}
	if msg.Channel != "" {
		env = env.WithChannel(msg.Channel)
	}
	if !msg.Timestamp.IsZero() {
		env.Timestamp = msg.Timestamp.UTC()
	}

	switch msg.TargetType {
	case domain_ws.RelayTargetUsers:
		if len(msg.TargetIDs) == 0 {
			return nil
		}
		return hub.SendToUsers(ctx, msg.TargetIDs, env)

	case domain_ws.RelayTargetChannel:
		if strings.TrimSpace(msg.Channel) == "" {
			return domain_ws.ErrInvalidChannel
		}
		return hub.SendToChannel(ctx, msg.Channel, env)

	case domain_ws.RelayTargetBroadcast:
		return hub.Broadcast(ctx, env)

	default:
		return fmt.Errorf("%w: unsupported relay target type %q", domain_ws.ErrInvalidEnvelope, msg.TargetType)
	}
}
