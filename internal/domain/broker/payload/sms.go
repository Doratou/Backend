package payload

// SMSPayload represents the message contract published to q.sms.classic.
type SMSPayload struct {
	Recipient  string            `json:"recipient"`
	Message    string            `json:"message,omitempty"`
	TemplateID string            `json:"template_id,omitempty"`
	Tokens     map[string]string `json:"tokens,omitempty"`
}
