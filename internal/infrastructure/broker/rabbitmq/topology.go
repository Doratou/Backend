package rabbitmq

import (
	"context"
	"fmt"
	"time"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Topology provides helper methods to declare and bind AMQP exchanges and queues idempotently.
type Topology struct {
	cm ConnectionManager
}

// NewTopology creates a new Topology helper instance.
func NewTopology(cm ConnectionManager) *Topology {
	return &Topology{cm: cm}
}

// DeclareExchange idempotently declares an AMQP exchange using typed ExchangeType.
func (t *Topology) DeclareExchange(ctx context.Context, name string, kind domain_broker.ExchangeType, durable, autoDelete bool, args amqp.Table) error {
	ch, err := t.cm.Channel()
	if err != nil {
		return fmt.Errorf("topology: open channel for exchange %q: %w", name, err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		name,
		string(kind),
		durable,
		autoDelete,
		false, // internal
		false, // no-wait
		args,
	); err != nil {
		return fmt.Errorf("topology: declare exchange %q: %w", name, err)
	}
	return nil
}

// DeclareQueue idempotently declares an AMQP queue with an explicit QueueType and optional custom arguments.
func (t *Topology) DeclareQueue(ctx context.Context, name string, qType domain_broker.QueueType, durable, autoDelete bool, args amqp.Table) (*amqp.Queue, error) {
	ch, err := t.cm.Channel()
	if err != nil {
		return nil, fmt.Errorf("topology: open channel for queue %q: %w", name, err)
	}
	defer ch.Close()

	mergedArgs := make(amqp.Table, len(args)+1)
	for k, v := range args {
		mergedArgs[k] = v
	}
	if qType != "" {
		mergedArgs["x-queue-type"] = string(qType)
	}

	q, err := ch.QueueDeclare(
		name,
		durable,
		autoDelete,
		false, // exclusive
		false, // no-wait
		mergedArgs,
	)
	if err != nil {
		return nil, fmt.Errorf("topology: declare queue %q: %w", name, err)
	}
	return &q, nil
}

// BindQueue binds an AMQP queue to an exchange with a routing key.
func (t *Topology) BindQueue(ctx context.Context, queue, routingKey, exchange string, args amqp.Table) error {
	ch, err := t.cm.Channel()
	if err != nil {
		return fmt.Errorf("topology: open channel for binding queue %q: %w", queue, err)
	}
	defer ch.Close()

	if err := ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // no-wait
		args,
	); err != nil {
		return fmt.Errorf("topology: bind queue %q to exchange %q (rk: %q): %w", queue, exchange, routingKey, err)
	}
	return nil
}

// SetupDeadLetterTopology declares a Dead Letter Exchange and Queue, binding them together.
func (t *Topology) SetupDeadLetterTopology(ctx context.Context, dlxExchange, dlqQueue, routingKey string) error {
	if dlxExchange != "" {
		if err := t.DeclareExchange(ctx, dlxExchange, domain_broker.ExchangeTypeDirect, true, false, nil); err != nil {
			return err
		}
	}

	if _, err := t.DeclareQueue(ctx, dlqQueue, domain_broker.QueueTypeQuorum, true, false, nil); err != nil {
		return err
	}

	if dlxExchange != "" {
		if err := t.BindQueue(ctx, dlqQueue, routingKey, dlxExchange, nil); err != nil {
			return err
		}
	}

	return nil
}

// SetupRetryTopology configures a delayed retry queue using RabbitMQ's native TTL + Dead Letter Exchange.
// Messages routed to retryExchange with routingKey `mainQueue` will sit in `retryQueue` for `delayTTL`,
// after which RabbitMQ dead-letters them back to `mainExchange` with `mainQueue` routing key.
func (t *Topology) SetupRetryTopology(ctx context.Context, mainExchange, mainQueue, retryExchange, retryQueue string, delayTTL time.Duration) error {
	if retryExchange != "" {
		if err := t.DeclareExchange(ctx, retryExchange, domain_broker.ExchangeTypeDirect, true, false, nil); err != nil {
			return fmt.Errorf("declare retry exchange %q: %w", retryExchange, err)
		}
	}

	// Queue arguments that power delayed auto-forwarding back to main exchange/queue:
	args := amqp.Table{
		"x-message-ttl":             int64(delayTTL.Milliseconds()),
		"x-dead-letter-exchange":    mainExchange,
		"x-dead-letter-routing-key": mainQueue,
	}

	if _, err := t.DeclareQueue(ctx, retryQueue, domain_broker.QueueTypeClassic, true, false, args); err != nil {
		return fmt.Errorf("declare retry queue %q: %w", retryQueue, err)
	}

	if retryExchange != "" {
		if err := t.BindQueue(ctx, retryQueue, mainQueue, retryExchange, nil); err != nil {
			return fmt.Errorf("bind retry queue %q to retry exchange %q: %w", retryQueue, retryExchange, err)
		}
	}

	return nil
}
