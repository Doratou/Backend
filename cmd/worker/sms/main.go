package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Doratou/Backend/bootstrap"
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
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("sms worker bootstrap error: %v", err)
	}

	log.Print("Starting Doratou SMS Consumer Worker...")

	// 1. Establish RabbitMQ Connection
	rmqEnv := cfg.Environment.RabbitMQ
	rmqCfg := brokerInfra.DefaultConfig(rmqEnv.Host, rmqEnv.Port, rmqEnv.User, rmqEnv.Password, rmqEnv.VHost)
	rmqCfg.ConnectionTimeout = cfg.Constant.RabbitMQ.DefaultPublishTimeout
	rmqCfg.InitialReconnectDelay = cfg.Constant.RabbitMQ.DefaultReconnectDelay

	connCtx, cancelConn := context.WithTimeout(context.Background(), 10*time.Second)
	rmqConnMgr, err := brokerInfra.NewConnectionManager(connCtx, rmqCfg)
	cancelConn()
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ broker: %v", err)
	}
	defer rmqConnMgr.Close()
	log.Print("RabbitMQ connection established for sms worker")

	// 2. Initialize Broker Infrastructure and Setup Topology
	rmqPublisher := brokerInfra.NewPublisher(rmqConnMgr)
	rmqConsumer := brokerInfra.NewConsumer(rmqConnMgr, rmqPublisher)
	topology := brokerInfra.NewTopology(rmqConnMgr)

	setupCtx, cancelSetup := context.WithTimeout(context.Background(), 10*time.Second)
	if err := topology.SetupDoratouTopology(setupCtx); err != nil {
		log.Printf("[SMS WORKER WARNING] failed to declare doratou topology: %v", err)
	} else {
		log.Print("Doratou notification topology verified and ready")
	}
	cancelSetup()

	// 3. Initialize SMS Sender (Mock implementation for now)
	smsSender := &mockSMSSender{}

	// 4. Graceful Shutdown & Context Management
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// 5. Register SMS Queue Consumer Handler
	smsHandler := func(ctx context.Context, d domain_broker.Delivery) error {
		var p payload.SMSPayload
		if err := json.Unmarshal(d.Message().Body, &p); err != nil {
			log.Printf("[SMS WORKER ERROR] invalid payload format: %v", err)
			return domain_broker.ErrReject
		}

		log.Printf("[SMS WORKER] processing SMS for recipient %s", p.Recipient)
		if err := smsSender.Send(ctx, p); err != nil {
			log.Printf("[SMS WORKER ERROR] failed to deliver SMS: %v", err)
			return err
		}

		return nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rmqConsumer.Consume(
			ctx,
			domain_broker.SMSQueue,
			smsHandler,
			domain_broker.WithConcurrency(2),
			domain_broker.WithPrefetchCount(cfg.Constant.RabbitMQ.DefaultPrefetchCount),
		); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[SMS WORKER ERROR] consumer stopped: %v", err)
		}
	}()

	log.Print("SMS worker running and listening on q.sms.classic.")

	<-ctx.Done()
	log.Print("Shutdown signal received in sms worker, waiting for in-flight tasks to complete...")

	shutdownWaitCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Print("All sms worker consumers shut down cleanly.")
	case <-shutdownWaitCtx.Done():
		log.Print("Shutdown timed out waiting for sms consumers.")
	}

	log.Print("SMS worker process stopped.")
}
