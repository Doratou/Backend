package rabbitmq

import (
	"context"
	"fmt"
	"time"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupDoratouTopology declares the notification direct exchange and the 3 classic queues,
// configuring safety net message TTLs and binding each queue to the exchange idempotently.
func (t *Topology) SetupDoratouTopology(ctx context.Context, emailTTL, smsTTL, wsTTL time.Duration) error {
	// 1. Declare Direct Exchange
	if err := t.DeclareExchange(ctx, domain_broker.NotificationExchange, domain_broker.ExchangeTypeDirect, true, false, nil); err != nil {
		return fmt.Errorf("setup exchange %s: %w", domain_broker.NotificationExchange, err)
	}

	// 2. Setup Queues and Bindings with Queue-Level Message TTL
	bindings := []struct {
		queue      string
		routingKey string
		ttl        time.Duration
	}{
		{queue: domain_broker.EmailQueue, routingKey: domain_broker.EmailRoutingKey, ttl: emailTTL},
		{queue: domain_broker.SMSQueue, routingKey: domain_broker.SMSRoutingKey, ttl: smsTTL},
		{queue: domain_broker.WSQueue, routingKey: domain_broker.WSRoutingKey, ttl: wsTTL},
	}

	for _, b := range bindings {
		args := make(amqp.Table)
		if b.ttl > 0 {
			args["x-message-ttl"] = int64(b.ttl.Milliseconds())
		}

		if _, err := t.DeclareQueue(ctx, b.queue, domain_broker.QueueTypeClassic, true, false, args); err != nil {
			return fmt.Errorf("setup queue %s: %w", b.queue, err)
		}

		if err := t.BindQueue(ctx, b.queue, b.routingKey, domain_broker.NotificationExchange, nil); err != nil {
			return fmt.Errorf("bind queue %s to %s with key %s: %w", b.queue, domain_broker.NotificationExchange, b.routingKey, err)
		}
	}

	return nil
}
