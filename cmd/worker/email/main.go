package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Doratou/Backend/bootstrap"
	runner "github.com/Doratou/Backend/cmd/worker"
	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	brokerInfra "github.com/Doratou/Backend/internal/infrastructure/broker/rabbitmq"
	mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"
)

func main() {
	runner.Run("EMAIL", func(ctx context.Context, consumer *brokerInfra.Consumer, cfg *bootstrap.Config) error {
		emailSender := mail.NewMailer(mail.EmailConfig{
			From:     cfg.Environment.Email.DoratouEmail,
			Password: cfg.Environment.Email.DoratouEmailAppPassword,
			SMTPHost: cfg.Environment.Email.SMTPHost,
			SMTPPort: cfg.Environment.Email.SMTPPort,
		})

		emailHandler := func(ctx context.Context, d domain_broker.Delivery) error {
			var p payload.EmailPayload
			if err := json.Unmarshal(d.Message().Body, &p); err != nil {
				log.Printf("[EMAIL WORKER ERROR] invalid payload format: %v", err)
				return domain_broker.ErrReject
			}

			log.Printf("[EMAIL WORKER] delivering email to %v with subject %q", p.To, p.Subject)

			attachments := make([]mail.InlineAttachment, len(p.InlineAttachments))
			for i, att := range p.InlineAttachments {
				attachments[i] = mail.InlineAttachment{
					Name:      att.Name,
					ContentID: att.ContentID,
					Data:      att.Data,
				}
			}

			sendOpts := mail.SendOptions{
				To:                p.To,
				Subject:           p.Subject,
				PlainBody:         p.PlainBody,
				HTMLBody:          p.HTMLBody,
				InlineAttachments: attachments,
			}

			if err := emailSender.Send(ctx, sendOpts); err != nil {
				log.Printf("[EMAIL WORKER ERROR] failed to deliver email: %v", err)
				return err
			}

			log.Printf("[EMAIL WORKER] email successfully sent to %v", p.To)
			return nil
		}

		wCfg := cfg.Environment.Workers
		return consumer.Consume(
			ctx,
			domain_broker.EmailQueue,
			emailHandler,
			domain_broker.WithConcurrency(wCfg.EmailConcurrency),
			domain_broker.WithPrefetchCount(wCfg.EmailPrefetch),
			domain_broker.WithRetryPolicy(domain_broker.RetryPolicy{
				Enabled:      true,
				MaxRetries:   wCfg.EmailMaxRetries,
				InitialDelay: wCfg.EmailRetryDelay,
			}),
		)
	})
}
