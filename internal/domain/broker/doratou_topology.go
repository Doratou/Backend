package broker

const (
	// NotificationExchange is the direct exchange responsible for routing notification events.
	NotificationExchange = "ex.notifications.direct"

	// NotificationRetryExchange is the direct exchange responsible for routing delayed retries.
	NotificationRetryExchange = "ex.notifications.retry.direct"

	// NotificationDLX is the dead-letter exchange for unrecoverable or exhausted poison messages.
	NotificationDLX = "ex.notifications.dlx"

	// Main queue names for notification channels.
	EmailQueue = "q.email.classic"
	SMSQueue   = "q.sms.classic"

	// Retry queue names for notification channels (buffer with delay before returning to main queue).
	EmailRetryQueue = "q.email.retry"
	SMSRetryQueue   = "q.sms.retry"

	// Dead letter queue (DLQ) names for permanent failure inspection.
	EmailDLQ = "q.email.dlq"
	SMSDLQ   = "q.sms.dlq"

	// Routing keys bound to notification queues.
	EmailRoutingKey = "rk.email"
	SMSRoutingKey   = "rk.sms"
)
