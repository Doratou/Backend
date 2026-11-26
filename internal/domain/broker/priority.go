package broker

import "math"

// Urgency defines the business importance and latency tolerance of a message.
type Urgency uint8

const (
	// UrgencyNormal is for routine, bulk, or marketing messages (newsletters, digests).
	UrgencyNormal Urgency = 1

	// UrgencyStandard is for transactional and state notifications (receipts, friend requests).
	UrgencyStandard Urgency = 2

	// UrgencyRealtime is for latency-critical, high-priority dispatches (OTP, 2FA, security codes).
	UrgencyRealtime Urgency = 3
)

const (
	// MaxQueuePriority is the maximum priority configured on classic notification queues.
	MaxQueuePriority = 10

	// HeaderUrgency carries the message urgency level in message headers.
	HeaderUrgency = "x-urgency"
)

// CalculatePriority dynamically computes the integer AMQP priority (1..10) based on
// business urgency and the current retry attempt count relative to maxRetries.
//
// Priority Bands:
//   - UrgencyRealtime: Fresh = 7, Retried = [8 .. 10]
//   - UrgencyStandard: Fresh = 4, Retried = [5 .. 6]
//   - UrgencyNormal:   Fresh = 1, Retried = [2 .. 3]
func CalculatePriority(urgency Urgency, retryCount, maxRetries int) uint8 {
	if retryCount < 0 {
		retryCount = 0
	}
	if maxRetries <= 0 {
		maxRetries = 1
	}
	if retryCount > maxRetries {
		retryCount = maxRetries
	}

	progress := float64(retryCount) / float64(maxRetries)

	switch urgency {
	case UrgencyRealtime:
		if retryCount == 0 {
			return 7 // Fresh OTP delivery
		}
		// In-flight retries span [8 .. 10]
		return 8 + uint8(math.Round(progress*2.0))

	case UrgencyStandard:
		if retryCount == 0 {
			return 4 // Fresh standard notification
		}
		// In-flight retries span [5 .. 6]
		return 5 + uint8(math.Round(progress*1.0))

	default: // UrgencyNormal
		if retryCount == 0 {
			return 1 // Fresh normal / marketing message
		}
		// In-flight retries span [2 .. 3]
		return 2 + uint8(math.Round(progress*1.0))
	}
}
