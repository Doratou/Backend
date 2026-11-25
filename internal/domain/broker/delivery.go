package broker

// Delivery is the interface exposed to consumers to inspect and acknowledge incoming messages.
type Delivery interface {
	// Message returns the unpacked domain Message.
	Message() Message

	// RoutingKey returns the routing key with which the message was published.
	RoutingKey() string

	// Exchange returns the exchange through which the message was routed.
	Exchange() string

	// Redelivered returns true if the message was delivered before.
	Redelivered() bool

	// RetryCount returns the number of times this message has been retried.
	RetryCount() int

	// Urgency returns the business urgency tier of this message.
	Urgency() Urgency

	// Ack acknowledges the delivery of this message.
	Ack(multiple bool) error

	// Nack negatively acknowledges the message, with option to requeue.
	Nack(multiple bool, requeue bool) error

	// Reject rejects the message without acknowledgment, with option to requeue.
	Reject(requeue bool) error
}
