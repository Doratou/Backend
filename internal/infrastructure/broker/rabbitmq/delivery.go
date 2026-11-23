package rabbitmq

import (
	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

// deliveryWrapper wraps amqp.Delivery into the domain_broker.Delivery interface.
type deliveryWrapper struct {
	d amqp.Delivery
}

// NewDelivery creates a domain delivery from an AMQP delivery.
func NewDelivery(d amqp.Delivery) domain_broker.Delivery {
	return &deliveryWrapper{d: d}
}

func (dw *deliveryWrapper) Message() domain_broker.Message {
	headers := make(domain_broker.Headers)
	if dw.d.Headers != nil {
		for k, v := range dw.d.Headers {
			headers[k] = v
		}
	}

	return domain_broker.Message{
		ID:            dw.d.MessageId,
		ContentType:   dw.d.ContentType,
		Body:          dw.d.Body,
		Headers:       headers,
		CorrelationID: dw.d.CorrelationId,
		ReplyTo:       dw.d.ReplyTo,
		Timestamp:     dw.d.Timestamp,
	}
}

func (dw *deliveryWrapper) RoutingKey() string {
	return dw.d.RoutingKey
}

func (dw *deliveryWrapper) Exchange() string {
	return dw.d.Exchange
}

func (dw *deliveryWrapper) Redelivered() bool {
	return dw.d.Redelivered
}

func (dw *deliveryWrapper) RetryCount() int {
	if dw.d.Headers == nil {
		return 0
	}

	// 1. Check custom retry header
	if count, ok := dw.d.Headers["x-retry-count"]; ok {
		switch v := count.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}

	// 2. Check standard AMQP x-death header if message was dead-lettered
	if xDeath, ok := dw.d.Headers["x-death"]; ok {
		if deaths, ok := xDeath.([]any); ok && len(deaths) > 0 {
			if deathMap, ok := deaths[0].(amqp.Table); ok {
				if count, ok := deathMap["count"]; ok {
					if c, ok := count.(int64); ok {
						return int(c)
					}
				}
			}
		}
	}

	return 0
}

func (dw *deliveryWrapper) Ack(multiple bool) error {
	return dw.d.Ack(multiple)
}

func (dw *deliveryWrapper) Nack(multiple bool, requeue bool) error {
	return dw.d.Nack(multiple, requeue)
}

func (dw *deliveryWrapper) Reject(requeue bool) error {
	return dw.d.Reject(requeue)
}
