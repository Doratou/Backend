package mail

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/wneessen/go-mail"
)

type EmailConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort string
}

type SendOptions struct {
	To                []string
	Subject           string
	PlainBody         string
	HTMLBody          string
	InlineAttachments []InlineAttachment
}

type InlineAttachment struct {
	Name      string
	ContentID string
	Data      []byte
}

type Sender interface {
	Send(ctx context.Context, opts SendOptions) error
}

type mailer struct {
	cfg EmailConfig
}

func NewMailer(cfg EmailConfig) Sender {
	return &mailer{cfg: cfg}
}

func (m *mailer) Send(ctx context.Context, opts SendOptions) error {
	message := mail.NewMsg()
	if err := message.From(m.cfg.From); err != nil {
		return fmt.Errorf("failed to set sender email: %w", err)
	}
	if err := message.To(opts.To...); err != nil {
		return fmt.Errorf("failed to set recipient emails: %w", err)
	}
	message.Subject(opts.Subject)
	message.SetBodyString(mail.TypeTextPlain, opts.PlainBody)
	if opts.HTMLBody != "" {
		message.AddAlternativeString(mail.TypeTextHTML, opts.HTMLBody)
	}
	for _, attachment := range opts.InlineAttachments {
		if err := message.EmbedReader(attachment.Name, bytes.NewReader(attachment.Data), mail.WithFileContentID("<"+attachment.ContentID+">")); err != nil {
			return fmt.Errorf("failed to embed inline email attachment: %w", err)
		}
	}

	port, err := strconv.Atoi(m.cfg.SMTPPort)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}
	client, err := mail.NewClient(
		m.cfg.SMTPHost,
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.cfg.From),
		mail.WithPassword(m.cfg.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}
	if err := client.DialAndSendWithContext(ctx, message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
