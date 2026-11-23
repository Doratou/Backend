package broker

// QueueType specifies the underlying storage engine and clustering model for a RabbitMQ queue.
type QueueType string

const (
	// QueueTypeClassic is the traditional, lightweight standard RabbitMQ queue.
	QueueTypeClassic QueueType = "classic"

	// QueueTypeQuorum is a durable, replicated, Raft-based high-availability queue (production recommended).
	QueueTypeQuorum QueueType = "quorum"

	// QueueTypeStream is an append-only, high-throughput log-based queue (similar to Apache Kafka).
	QueueTypeStream QueueType = "stream"
)

// ExchangeType specifies the routing algorithm of an AMQP exchange.
type ExchangeType string

const (
	// ExchangeTypeDirect routes messages based on exact routing key match.
	ExchangeTypeDirect ExchangeType = "direct"

	// ExchangeTypeFanout broadcasts messages to all bound queues unconditionally.
	ExchangeTypeFanout ExchangeType = "fanout"

	// ExchangeTypeTopic routes messages using wildcard pattern matching (* and #) on routing keys.
	ExchangeTypeTopic ExchangeType = "topic"

	// ExchangeTypeHeaders routes messages using message header attributes instead of routing keys.
	ExchangeTypeHeaders ExchangeType = "headers"
)
