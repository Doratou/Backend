package broker

const (
	// NotificationExchange is the direct exchange responsible for routing notification events.
	NotificationExchange = "ex.notifications.direct"

	// Queue names for notification channels.
	EmailQueue = "q.email.classic"
	SMSQueue   = "q.sms.classic"
	WSQueue    = "q.ws.classic"

	// Routing keys bound to notification queues.
	EmailRoutingKey = "rk.email"
	SMSRoutingKey   = "rk.sms"
	WSRoutingKey    = "rk.ws"
)
