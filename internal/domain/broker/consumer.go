package broker

import (
	"context"
	"time"
)

// HandlerFunc is the signature of a message processing function executed by consumers.
// If the handler returns nil, the message is automatically acknowledged (Ack).
// If it returns ErrReject, the message is rejected without requeuing (routed to DLQ if configured).
// If it returns ErrRetry or an error and RetryPolicy is enabled, it is routed to the retry queue.
type HandlerFunc func(ctx context.Context, d Delivery) error

// RetryPolicy specifies retry behavior for messages that fail execution in a consumer.
type RetryPolicy struct {
	Enabled            bool
	MaxRetries         int
	InitialDelay       time.Duration
	BackoffMultiplier  float64
	MaxDelay           time.Duration
	RetryExchange      string
	DeadLetterExchange string
}

// DefaultRetryPolicy returns sensible defaults for retrying failed messages.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		Enabled:           true,
		MaxRetries:        3,
		InitialDelay:      5 * time.Second,
		BackoffMultiplier: 2.0,
		MaxDelay:          5 * time.Minute,
	}
}

// ConsumeOptions contains parameters used when registering a consumer on a queue.
type ConsumeOptions struct {
	ConsumerTag   string
	AutoAck       bool
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	PrefetchCount int
	Concurrency   int
	RetryPolicy   *RetryPolicy
}

// ConsumeOption configures consumer registration.
type ConsumeOption func(*ConsumeOptions)

// WithConsumerTag sets an explicit tag for this consumer instance.
func WithConsumerTag(tag string) ConsumeOption {
	return func(o *ConsumeOptions) {
		o.ConsumerTag = tag
	}
}

// WithAutoAck configures whether messages are automatically acknowledged by broker upon delivery.
func WithAutoAck(autoAck bool) ConsumeOption {
	return func(o *ConsumeOptions) {
		o.AutoAck = autoAck
	}
}

// WithPrefetchCount sets the QoS prefetch window for this consumer.
func WithPrefetchCount(count int) ConsumeOption {
	return func(o *ConsumeOptions) {
		if count > 0 {
			o.PrefetchCount = count
		}
	}
}

// WithConcurrency sets the number of concurrent worker goroutines handling deliveries from this queue.
func WithConcurrency(concurrency int) ConsumeOption {
	return func(o *ConsumeOptions) {
		if concurrency > 0 {
			o.Concurrency = concurrency
		}
	}
}

// WithRetryPolicy attaches a retry and DLQ policy to the consumer.
func WithRetryPolicy(policy RetryPolicy) ConsumeOption {
	return func(o *ConsumeOptions) {
		o.RetryPolicy = &policy
	}
}

// Consumer is the interface for registering queue message consumers.
type Consumer interface {
	// Consume starts consuming messages from the specified queue and dispatches them to handler.
	// It blocks until ctx is canceled or a fatal unrecoverable error occurs.
	Consume(ctx context.Context, queue string, handler HandlerFunc, opts ...ConsumeOption) error
}
