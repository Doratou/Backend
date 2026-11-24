package payload

// InlineAttachment represents an embedded inline file (e.g. header logo with cid).
type InlineAttachment struct {
	Name      string `json:"name"`
	ContentID string `json:"content_id"`
	Data      []byte `json:"data"`
}

// EmailPayload represents the pure domain message contract published to q.email.classic.
type EmailPayload struct {
	To                []string           `json:"to"`
	Subject           string             `json:"subject"`
	PlainBody         string             `json:"plain_body"`
	HTMLBody          string             `json:"html_body,omitempty"`
	InlineAttachments []InlineAttachment `json:"inline_attachments,omitempty"`
}
