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
)

// SMSSender defines the interface for delivering SMS messages.
type SMSSender interface {
	Send(ctx context.Context, p payload.SMSPayload) error
}

// mockSMSSender implements SMSSender by logging delivery details.
type mockSMSSender struct{}

func (m *mockSMSSender) Send(ctx context.Context, p payload.SMSPayload) error {
	log.Printf("[MOCK SMS SENDER] Successfully dispatched SMS to %s | TemplateID: %s | Tokens: %v | Message: %q",
		p.Recipient, p.TemplateID, p.Tokens, p.Message)
	return nil
}

func main() {
	runner.Run("SMS", func(ctx context.Context, consumer *brokerInfra.Consumer, cfg *bootstrap.Config) error {
		smsSender := &mockSMSSender{}

		smsHandler := func(ctx context.Context, d domain_broker.Delivery) error {
			var p payload.SMSPayload
			if err := json.Unmarshal(d.Message().Body, &p); err != nil {
				log.Printf("[SMS WORKER ERROR] invalid payload format: %v", err)
				return domain_broker.ErrReject
			}

			log.Printf("[SMS WORKER] processing SMS for recipient %s", p.Recipient)
			return smsSender.Send(ctx, p)
		}

		wCfg := cfg.Environment.Workers
		return consumer.Consume(
			ctx,
			domain_broker.SMSQueue,
			smsHandler,
			domain_broker.WithConcurrency(wCfg.SMSConcurrency),
			domain_broker.WithPrefetchCount(wCfg.SMSPrefetch),
			domain_broker.WithRetryPolicy(domain_broker.RetryPolicy{
				Enabled:            true,
				MaxRetries:         wCfg.SMSMaxRetries,
				InitialDelay:       wCfg.SMSRetryDelay,
				RetryExchange:      domain_broker.NotificationRetryExchange,
				DeadLetterExchange: domain_broker.NotificationDLX,
			}),
		)
	})
}
