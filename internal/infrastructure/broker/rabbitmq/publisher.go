package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher implements domain_broker.Publisher with thread-safe channel access and persistent deliveries.
type Publisher struct {
	cm        ConnectionManager
	chMu      sync.Mutex
	pubChan   *amqp.Channel
	closeChan chan *amqp.Error
}

// NewPublisher creates an initialized RabbitMQ publisher.
func NewPublisher(cm ConnectionManager) *Publisher {
	return &Publisher{
		cm: cm,
	}
}

func (p *Publisher) getChannel() (*amqp.Channel, error) {
	p.chMu.Lock()
	defer p.chMu.Unlock()

	if p.pubChan != nil && !p.pubChan.IsClosed() {
		return p.pubChan, nil
	}

	ch, err := p.cm.Channel()
	if err != nil {
		if errors.Is(err, amqp.ErrClosed) {
			return nil, domain_broker.ErrConnectionClosed
		}
		return nil, fmt.Errorf("publisher: open channel: %w", err)
	}

	p.pubChan = ch
	p.closeChan = make(chan *amqp.Error, 1)
	p.pubChan.NotifyClose(p.closeChan)

	return p.pubChan, nil
}

// Publish sends a domain Message to the target exchange with the specified routing key.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, msg domain_broker.Message, opts ...domain_broker.PublishOption) error {
	options := domain_broker.PublishOptions{
		Persistent: true,
		Timeout:    10 * time.Second,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	ch, err := p.getChannel()
	if err != nil {
		return err
	}

	// Merge message headers with publish options headers
	headers := make(amqp.Table)
	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}
	if options.Headers != nil {
		for k, v := range options.Headers {
			headers[k] = v
		}
	}

	msgID := msg.ID
	if msgID == "" {
		msgID = uuid.NewString()
	}

	timestamp := msg.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	contentType := msg.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	deliveryMode := amqp.Transient
	if options.Persistent {
		deliveryMode = amqp.Persistent
	}

	publishing := amqp.Publishing{
		Headers:         headers,
		ContentType:     contentType,
		ContentEncoding: "utf-8",
		DeliveryMode:    deliveryMode,
		MessageId:       msgID,
		Timestamp:       timestamp,
		CorrelationId:   msg.CorrelationID,
		ReplyTo:         msg.ReplyTo,
		Body:            msg.Body,
	}

	if options.TTL > 0 {
		publishing.Expiration = strconv.FormatInt(options.TTL.Milliseconds(), 10)
	}

	p.chMu.Lock()
	defer p.chMu.Unlock()

	publishDone := make(chan error, 1)
	go func() {
		publishDone <- ch.PublishWithContext(
			ctx,
			exchange,
			routingKey,
			options.Mandatory,
			false, // immediate
			publishing,
		)
	}()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return domain_broker.ErrPublishTimeout
		}
		return ctx.Err()
	case err := <-publishDone:
		if err != nil {
			if errors.Is(err, amqp.ErrClosed) {
				return domain_broker.ErrConnectionClosed
			}
			return fmt.Errorf("publisher: publish failed: %w", err)
		}
		return nil
	}
}

// PublishJSON serializes payload into JSON and publishes it to the specified exchange and routing key.
func (p *Publisher) PublishJSON(ctx context.Context, exchange, routingKey string, payload any, opts ...domain_broker.PublishOption) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("publisher: json marshal error: %w", err)
	}

	msg := domain_broker.NewJSONMessage(uuid.NewString(), body)
	return p.Publish(ctx, exchange, routingKey, msg, opts...)
}

// PublishToRetry routes a failed message to a delayed retry exchange/queue with incremented retry count and optional TTL.
func (p *Publisher) PublishToRetry(ctx context.Context, retryExchange, mainQueue string, originalMsg domain_broker.Message, retryCount int, delay time.Duration) error {
	headers := make(domain_broker.Headers)
	if originalMsg.Headers != nil {
		for k, v := range originalMsg.Headers {
			headers[k] = v
		}
	}
	headers["x-retry-count"] = retryCount + 1
	headers["x-original-queue"] = mainQueue
	headers["x-retried-at"] = time.Now().UTC().Format(time.RFC3339)

	originalMsg.Headers = headers

	var publishOpts []domain_broker.PublishOption
	publishOpts = append(publishOpts, domain_broker.WithPersistent(true))

	// If a specific per-message delay was supplied and not handled solely by the queue TTL:
	if delay > 0 {
		return p.PublishWithExpiration(ctx, retryExchange, mainQueue, originalMsg, delay, publishOpts...)
	}

	return p.Publish(ctx, retryExchange, mainQueue, originalMsg, publishOpts...)
}

// PublishToDLQ routes an unrecoverable poison message to the DLX with failure metadata headers.
func (p *Publisher) PublishToDLQ(ctx context.Context, dlx, routingKey string, originalMsg domain_broker.Message, retryCount int, reason error) error {
	headers := make(domain_broker.Headers)
	if originalMsg.Headers != nil {
		for k, v := range originalMsg.Headers {
			headers[k] = v
		}
	}
	if reason != nil {
		headers["x-death-reason"] = reason.Error()
	}
	headers["x-failed-retries"] = retryCount
	headers["x-dead-lettered-at"] = time.Now().UTC().Format(time.RFC3339)
	originalMsg.Headers = headers

	return p.Publish(ctx, dlx, routingKey, originalMsg, domain_broker.WithPersistent(true))
}

// PublishWithExpiration publishes a message with per-message TTL (expiration).
func (p *Publisher) PublishWithExpiration(ctx context.Context, exchange, routingKey string, msg domain_broker.Message, expiration time.Duration, opts ...domain_broker.PublishOption) error {
	ch, err := p.getChannel()
	if err != nil {
		return err
	}

	headers := make(amqp.Table)
	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}

	msgID := msg.ID
	if msgID == "" {
		msgID = uuid.NewString()
	}

	timestamp := msg.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	contentType := msg.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	publishing := amqp.Publishing{
		Headers:         headers,
		ContentType:     contentType,
		ContentEncoding: "utf-8",
		DeliveryMode:    amqp.Persistent,
		MessageId:       msgID,
		Timestamp:       timestamp,
		CorrelationId:   msg.CorrelationID,
		ReplyTo:         msg.ReplyTo,
		Body:            msg.Body,
		Expiration:      fmt.Sprintf("%d", expiration.Milliseconds()),
	}

	p.chMu.Lock()
	defer p.chMu.Unlock()

	return ch.PublishWithContext(ctx, exchange, routingKey, false, false, publishing)
}
