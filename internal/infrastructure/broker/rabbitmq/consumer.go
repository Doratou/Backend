package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer implements domain_broker.Consumer with support for concurrent workers,
// resilient auto-resubscription, retry queues, dead-letter exchanges, and graceful shutdown.
type Consumer struct {
	cm        ConnectionManager
	publisher *Publisher
}

// NewConsumer creates a new RabbitMQ consumer instance.
func NewConsumer(cm ConnectionManager, pub *Publisher) *Consumer {
	return &Consumer{
		cm:        cm,
		publisher: pub,
	}
}

// Consume starts listening on queue and dispatches received deliveries to handler.
// It automatically recovers when connection is restored, and shuts down cleanly when ctx is canceled.
func (c *Consumer) Consume(ctx context.Context, queue string, handler domain_broker.HandlerFunc, opts ...domain_broker.ConsumeOption) error {
	options := domain_broker.ConsumeOptions{
		PrefetchCount: 10,
		Concurrency:   1,
		AutoAck:       false,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if options.ConsumerTag == "" {
		options.ConsumerTag = fmt.Sprintf("consumer-%s-%s", queue, uuid.NewString()[:8])
	}

	reconnectCh := c.cm.SubscribeReconnect()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := c.runConsumerSession(ctx, queue, handler, options)
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}

		log.Printf("[RABBITMQ WARNING] consumer session for %q failed: %v. Waiting for reconnect...", queue, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-reconnectCh:
			log.Printf("[RABBITMQ INFO] reconnect detected, restarting consumer on queue %q...", queue)
		case <-time.After(2 * time.Second):
			// Check if connection is already healthy
			if c.cm.IsConnected() {
				log.Printf("[RABBITMQ INFO] connection active, re-establishing consumer on %q...", queue)
			}
		}
	}
}

func (c *Consumer) runConsumerSession(ctx context.Context, queue string, handler domain_broker.HandlerFunc, opts domain_broker.ConsumeOptions) error {
	ch, err := c.cm.Channel()
	if err != nil {
		return fmt.Errorf("open consumer channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(opts.PrefetchCount, 0, false); err != nil {
		return fmt.Errorf("set qos (prefetch %d): %w", opts.PrefetchCount, err)
	}

	deliveries, err := ch.Consume(
		queue,
		opts.ConsumerTag,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register consumer on %q: %w", queue, err)
	}

	log.Printf("[RABBITMQ INFO] consumer %q started on queue %q (concurrency: %d, prefetch: %d)",
		opts.ConsumerTag, queue, opts.Concurrency, opts.PrefetchCount)

	channelCloseChan := make(chan *amqp.Error, 1)
	ch.NotifyClose(channelCloseChan)

	var wg sync.WaitGroup
	workersCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	// Dispatch deliveries across workers
	deliveryChan := make(chan amqp.Delivery, opts.PrefetchCount)

	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.workerLoop(workersCtx, queue, handler, opts, deliveryChan)
		}(i + 1)
	}

	// Session loop feeding deliveries to workers
	var sessionErr error
	keepConsuming := true

	for keepConsuming {
		select {
		case <-ctx.Done():
			log.Printf("[RABBITMQ INFO] shutdown requested for consumer %q. Canceling subscription...", opts.ConsumerTag)
			_ = ch.Cancel(opts.ConsumerTag, false)
			keepConsuming = false

		case closeErr, ok := <-channelCloseChan:
			if !ok || closeErr == nil {
				sessionErr = amqp.ErrClosed
			} else {
				sessionErr = fmt.Errorf("channel closed: %w", closeErr)
			}
			keepConsuming = false

		case d, ok := <-deliveries:
			if !ok {
				sessionErr = errors.New("delivery channel closed")
				keepConsuming = false
				break
			}

			select {
			case deliveryChan <- d:
			case <-ctx.Done():
				// If shutting down, nack unhandled message back to queue so it is not lost
				if !opts.AutoAck {
					_ = d.Nack(false, true)
				}
				keepConsuming = false
			}
		}
	}

	// Close worker dispatch and wait for in-flight jobs
	close(deliveryChan)
	wg.Wait()

	return sessionErr
}

func (c *Consumer) workerLoop(ctx context.Context, queue string, handler domain_broker.HandlerFunc, opts domain_broker.ConsumeOptions, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		delivery := NewDelivery(d)

		// Process delivery
		err := handler(ctx, delivery)

		if opts.AutoAck {
			continue
		}

		if err == nil {
			if ackErr := delivery.Ack(false); ackErr != nil {
				log.Printf("[RABBITMQ WARNING] failed to ack message %s: %v", d.MessageId, ackErr)
			}
			continue
		}

		// Handle failed execution
		c.handleFailure(ctx, queue, delivery, err, opts.RetryPolicy)
	}
}

func (c *Consumer) handleFailure(ctx context.Context, queue string, d domain_broker.Delivery, err error, policy *domain_broker.RetryPolicy) {
	msg := d.Message()

	// 1. Explicit permanent rejection requested
	if errors.Is(err, domain_broker.ErrReject) {
		log.Printf("[RABBITMQ NOTICE] message %s rejected permanently: %v", msg.ID, err)
		if rejectErr := d.Reject(false); rejectErr != nil {
			log.Printf("[RABBITMQ WARNING] failed to reject message %s: %v", msg.ID, rejectErr)
		}
		return
	}

	// 2. Retry mechanism
	if policy != nil && policy.Enabled && c.publisher != nil {
		retryCount := d.RetryCount()
		if retryCount < policy.MaxRetries {
			delay := c.calculateBackoffDelay(policy, retryCount)
			log.Printf("[RABBITMQ NOTICE] retrying message %s (attempt %d/%d) in %v...",
				msg.ID, retryCount+1, policy.MaxRetries, delay)

			retryExchange := policy.RetryExchange
			retryErr := c.publisher.PublishToRetry(ctx, retryExchange, queue, msg, retryCount, delay)
			if retryErr == nil {
				// Safely acknowledge original delivery since it is now buffered in retry queue
				_ = d.Ack(false)
				return
			}
			log.Printf("[RABBITMQ ERROR] failed to publish message %s to retry exchange: %v. Falling back to nack.", msg.ID, retryErr)
		} else {
			log.Printf("[RABBITMQ WARNING] message %s exceeded max retries (%d). Rejecting to DLQ/discard.",
				msg.ID, policy.MaxRetries)
			_ = d.Reject(false)
			return
		}
	}

	// 3. Default fallback: Nack with requeue
	if nackErr := d.Nack(false, true); nackErr != nil {
		log.Printf("[RABBITMQ WARNING] failed to nack message %s: %v", msg.ID, nackErr)
	}
}

func (c *Consumer) calculateBackoffDelay(policy *domain_broker.RetryPolicy, retryCount int) time.Duration {
	delay := policy.InitialDelay
	if delay <= 0 {
		delay = 5 * time.Second
	}

	multiplier := policy.BackoffMultiplier
	if multiplier < 1.0 {
		multiplier = 1.0
	}

	backoff := float64(delay) * math.Pow(multiplier, float64(retryCount))
	calculated := time.Duration(backoff)

	if policy.MaxDelay > 0 && calculated > policy.MaxDelay {
		calculated = policy.MaxDelay
	}

	return calculated
}
