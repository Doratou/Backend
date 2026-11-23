package broker

import "errors"

var (
	// ErrReject indicates that the message should be permanently rejected (e.g. sent to DLQ or discarded).
	ErrReject = errors.New("broker: reject message without requeue")

	// ErrRetry indicates that the message encountered a transient failure and should be routed to retry queue.
	ErrRetry = errors.New("broker: retry message")

	// ErrConnectionClosed indicates that the broker connection is currently closed.
	ErrConnectionClosed = errors.New("broker: connection is closed")

	// ErrPublishTimeout indicates that message publishing timed out.
	ErrPublishTimeout = errors.New("broker: publish timeout")

	// ErrMaxRetriesExceeded indicates that a message has reached the maximum allowed retry attempts.
	ErrMaxRetriesExceeded = errors.New("broker: max retry attempts exceeded")
)
