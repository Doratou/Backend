package rabbitmq

import (
	"context"
	"fmt"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
)

// SetupDoratouTopology declares the notification direct exchange and the 3 classic queues,
// binding each queue to the exchange with its corresponding routing key idempotently.
func (t *Topology) SetupDoratouTopology(ctx context.Context) error {
	// 1. Declare Direct Exchange
	if err := t.DeclareExchange(ctx, domain_broker.NotificationExchange, domain_broker.ExchangeTypeDirect, true, false, nil); err != nil {
		return fmt.Errorf("setup exchange %s: %w", domain_broker.NotificationExchange, err)
	}

	// 2. Setup Queues and Bindings
	bindings := []struct {
		queue      string
		routingKey string
	}{
		{queue: domain_broker.EmailQueue, routingKey: domain_broker.EmailRoutingKey},
		{queue: domain_broker.SMSQueue, routingKey: domain_broker.SMSRoutingKey},
		{queue: domain_broker.WSQueue, routingKey: domain_broker.WSRoutingKey},
	}

	for _, b := range bindings {
		if _, err := t.DeclareQueue(ctx, b.queue, domain_broker.QueueTypeClassic, true, false, nil); err != nil {
			return fmt.Errorf("setup queue %s: %w", b.queue, err)
		}

		if err := t.BindQueue(ctx, b.queue, b.routingKey, domain_broker.NotificationExchange, nil); err != nil {
			return fmt.Errorf("bind queue %s to %s with key %s: %w", b.queue, domain_broker.NotificationExchange, b.routingKey, err)
		}
	}

	return nil
}
