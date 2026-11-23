package broker

import (
	"context"
	"time"
)

// PublishOptions holds configuration for a publish call.
type PublishOptions struct {
	Persistent bool
	Mandatory  bool
	Headers    Headers
	Timeout    time.Duration
}

// PublishOption configures a publish operation.
type PublishOption func(*PublishOptions)

// WithPersistent sets the message delivery mode to persistent.
func WithPersistent(persistent bool) PublishOption {
	return func(o *PublishOptions) {
		o.Persistent = persistent
	}
}

// WithMandatory requires that the message must be routed to a queue or returned.
func WithMandatory(mandatory bool) PublishOption {
	return func(o *PublishOptions) {
		o.Mandatory = mandatory
	}
}

// WithPublishHeaders appends or overrides headers on the published message.
func WithPublishHeaders(headers Headers) PublishOption {
	return func(o *PublishOptions) {
		if o.Headers == nil {
			o.Headers = make(Headers)
		}
		for k, v := range headers {
			o.Headers[k] = v
		}
	}
}

// WithPublishTimeout sets a custom timeout for the publish operation.
func WithPublishTimeout(timeout time.Duration) PublishOption {
	return func(o *PublishOptions) {
		o.Timeout = timeout
	}
}

// Publisher is the domain interface for publishing events/messages across the system.
type Publisher interface {
	// Publish sends a domain Message to the target exchange with the specified routing key.
	Publish(ctx context.Context, exchange, routingKey string, msg Message, opts ...PublishOption) error

	// PublishJSON serializes payload as JSON and publishes it to the target exchange with the routing key.
	PublishJSON(ctx context.Context, exchange, routingKey string, payload any, opts ...PublishOption) error
}
