package payload

import mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"

// InlineAttachment represents an embedded file (such as a header logo with cid).
type InlineAttachment struct {
	Name      string `json:"name"`
	ContentID string `json:"content_id"`
	Data      []byte `json:"data"`
}

// EmailPayload represents the message contract published to q.email.classic.
type EmailPayload struct {
	To                []string           `json:"to"`
	Subject           string             `json:"subject"`
	PlainBody         string             `json:"plain_body"`
	HTMLBody          string             `json:"html_body,omitempty"`
	InlineAttachments []InlineAttachment `json:"inline_attachments,omitempty"`
}

// ToSendOptions converts the payload into mail.SendOptions used by the email sender.
func (p EmailPayload) ToSendOptions() mail.SendOptions {
	attachments := make([]mail.InlineAttachment, len(p.InlineAttachments))
	for i, att := range p.InlineAttachments {
		attachments[i] = mail.InlineAttachment{
			Name:      att.Name,
			ContentID: att.ContentID,
			Data:      att.Data,
		}
	}

	return mail.SendOptions{
		To:                p.To,
		Subject:           p.Subject,
		PlainBody:         p.PlainBody,
		HTMLBody:          p.HTMLBody,
		InlineAttachments: attachments,
	}
}

// FromSendOptions constructs an EmailPayload from mail.SendOptions.
func FromSendOptions(opts mail.SendOptions) EmailPayload {
	attachments := make([]InlineAttachment, len(opts.InlineAttachments))
	for i, att := range opts.InlineAttachments {
		attachments[i] = InlineAttachment{
			Name:      att.Name,
			ContentID: att.ContentID,
			Data:      att.Data,
		}
	}

	return EmailPayload{
		To:                opts.To,
		Subject:           opts.Subject,
		PlainBody:         opts.PlainBody,
		HTMLBody:          opts.HTMLBody,
		InlineAttachments: attachments,
	}
}
