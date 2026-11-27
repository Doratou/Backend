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

// WSSender defines the interface for delivering real-time WebSocket events.
type WSSender interface {
	Send(ctx context.Context, p payload.WSPayload) error
}

// mockWSSender implements WSSender by logging delivery details.
type mockWSSender struct{}

func (m *mockWSSender) Send(ctx context.Context, p payload.WSPayload) error {
	log.Printf("[MOCK WS SENDER] Successfully dispatched WS event %q | Scope: %s (IDs: %v) | Sender: %v | Data: %v",
		p.Event, p.TargetType, p.TargetIDs, p.SenderID, p.Data)
	return nil
}

func main() {
	runner.Run("WS", func(ctx context.Context, consumer *brokerInfra.Consumer, cfg *bootstrap.Config) error {
		wsSender := &mockWSSender{}

		wsHandler := func(ctx context.Context, d domain_broker.Delivery) error {
			var p payload.WSPayload
			if err := json.Unmarshal(d.Message().Body, &p); err != nil {
				log.Printf("[WS WORKER ERROR] invalid payload format: %v", err)
				return domain_broker.ErrReject
			}

			log.Printf("[WS WORKER] processing WS event %q for targets %s:%v", p.Event, p.TargetType, p.TargetIDs)
			return wsSender.Send(ctx, p)
		}

		wCfg := cfg.Environment.Workers
		return consumer.Consume(
			ctx,
			domain_broker.WSQueue,
			wsHandler,
			domain_broker.WithConcurrency(wCfg.WSConcurrency),
			domain_broker.WithPrefetchCount(wCfg.WSPrefetch),
			domain_broker.WithRetryPolicy(domain_broker.RetryPolicy{
				Enabled:            true,
				MaxRetries:         wCfg.WSMaxRetries,
				InitialDelay:       wCfg.WSRetryDelay,
				RetryExchange:      domain_broker.NotificationRetryExchange,
				DeadLetterExchange: domain_broker.NotificationDLX,
			}),
		)
	})
}
