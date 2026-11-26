package rabbitmq

import (
	"context"
	"fmt"
	"time"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupDoratouTopology declares the notification direct exchange, retry exchange, and DLX,
// along with main, delayed retry, and channel-specific dead-letter queues idempotently.
func (t *Topology) SetupDoratouTopology(ctx context.Context, emailTTL, smsTTL, wsTTL, dlqTTL time.Duration) error {
	// 1. Declare Exchanges: Main Notification, Retry, and Dead-Letter Exchanges
	exchanges := []string{
		domain_broker.NotificationExchange,
		domain_broker.NotificationRetryExchange,
		domain_broker.NotificationDLX,
	}
	for _, ex := range exchanges {
		if err := t.DeclareExchange(ctx, ex, domain_broker.ExchangeTypeDirect, true, false, nil); err != nil {
			return fmt.Errorf("setup exchange %s: %w", ex, err)
		}
	}

	// 2. Setup Notification Channels Topology
	channels := []struct {
		mainQueue  string
		retryQueue string
		dlqQueue   string
		routingKey string
		ttl        time.Duration
	}{
		{
			mainQueue:  domain_broker.EmailQueue,
			retryQueue: domain_broker.EmailRetryQueue,
			dlqQueue:   domain_broker.EmailDLQ,
			routingKey: domain_broker.EmailRoutingKey,
			ttl:        emailTTL,
		},
		{
			mainQueue:  domain_broker.SMSQueue,
			retryQueue: domain_broker.SMSRetryQueue,
			dlqQueue:   domain_broker.SMSDLQ,
			routingKey: domain_broker.SMSRoutingKey,
			ttl:        smsTTL,
		},
		{
			mainQueue:  domain_broker.WSQueue,
			retryQueue: domain_broker.WSRetryQueue,
			dlqQueue:   domain_broker.WSDLQ,
			routingKey: domain_broker.WSRoutingKey,
			ttl:        wsTTL,
		},
	}

	for _, ch := range channels {
		// A. Declare Main Queue & Bind to NotificationExchange
		mainArgs := amqp.Table{
			"x-max-priority": int64(domain_broker.MaxQueuePriority),
		}
		if ch.ttl > 0 {
			mainArgs["x-message-ttl"] = int64(ch.ttl.Milliseconds())
		}
		if _, err := t.DeclareQueue(ctx, ch.mainQueue, domain_broker.QueueTypeClassic, true, false, mainArgs); err != nil {
			return fmt.Errorf("setup main queue %s: %w", ch.mainQueue, err)
		}
		if err := t.BindQueue(ctx, ch.mainQueue, ch.routingKey, domain_broker.NotificationExchange, nil); err != nil {
			return fmt.Errorf("bind main queue %s to %s: %w", ch.mainQueue, domain_broker.NotificationExchange, err)
		}

		// B. Declare Retry Queue (dead-letters back to NotificationExchange upon delay expiration)
		retryArgs := amqp.Table{
			"x-dead-letter-exchange":    domain_broker.NotificationExchange,
			"x-dead-letter-routing-key": ch.routingKey,
			"x-max-priority":            int64(domain_broker.MaxQueuePriority),
		}
		if _, err := t.DeclareQueue(ctx, ch.retryQueue, domain_broker.QueueTypeClassic, true, false, retryArgs); err != nil {
			return fmt.Errorf("setup retry queue %s: %w", ch.retryQueue, err)
		}
		// Bind retry queue by both routing key and queue name for flexible retry routing
		if err := t.BindQueue(ctx, ch.retryQueue, ch.routingKey, domain_broker.NotificationRetryExchange, nil); err != nil {
			return fmt.Errorf("bind retry queue %s to %s with key %s: %w", ch.retryQueue, domain_broker.NotificationRetryExchange, ch.routingKey, err)
		}
		if err := t.BindQueue(ctx, ch.retryQueue, ch.mainQueue, domain_broker.NotificationRetryExchange, nil); err != nil {
			return fmt.Errorf("bind retry queue %s to %s with queue name %s: %w", ch.retryQueue, domain_broker.NotificationRetryExchange, ch.mainQueue, err)
		}

		// C. Declare Dead Letter Queue (DLQ) & Bind to NotificationDLX
		dlqArgs := make(amqp.Table)
		if dlqTTL > 0 {
			dlqArgs["x-message-ttl"] = int64(dlqTTL.Milliseconds())
		}
		if _, err := t.DeclareQueue(ctx, ch.dlqQueue, domain_broker.QueueTypeClassic, true, false, dlqArgs); err != nil {
			return fmt.Errorf("setup dlq queue %s: %w", ch.dlqQueue, err)
		}
		if err := t.BindQueue(ctx, ch.dlqQueue, ch.routingKey, domain_broker.NotificationDLX, nil); err != nil {
			return fmt.Errorf("bind dlq queue %s to %s with key %s: %w", ch.dlqQueue, domain_broker.NotificationDLX, ch.routingKey, err)
		}
		if err := t.BindQueue(ctx, ch.dlqQueue, ch.mainQueue, domain_broker.NotificationDLX, nil); err != nil {
			return fmt.Errorf("bind dlq queue %s to %s with queue name %s: %w", ch.dlqQueue, domain_broker.NotificationDLX, ch.mainQueue, err)
		}
	}

	return nil
}
