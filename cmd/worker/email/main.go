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
	mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"
)

func main() {
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("email worker bootstrap error: %v", err)
	}

	log.Print("Starting Doratou Email Consumer Worker...")

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
	log.Print("RabbitMQ connection established for email worker")

	// 2. Initialize Broker Infrastructure and Setup Topology
	rmqPublisher := brokerInfra.NewPublisher(rmqConnMgr)
	rmqConsumer := brokerInfra.NewConsumer(rmqConnMgr, rmqPublisher)
	topology := brokerInfra.NewTopology(rmqConnMgr)

	setupCtx, cancelSetup := context.WithTimeout(context.Background(), 10*time.Second)
	if err := topology.SetupDoratouTopology(setupCtx); err != nil {
		log.Printf("[EMAIL WORKER WARNING] failed to declare doratou topology: %v", err)
	} else {
		log.Print("Doratou notification topology verified and ready")
	}
	cancelSetup()

	// 3. Initialize Outgoing Email Sender
	emailSender := mail.NewMailer(mail.EmailConfig{
		From:     cfg.Environment.Email.DoratouEmail,
		Password: cfg.Environment.Email.DoratouEmailAppPassword,
		SMTPHost: cfg.Environment.Email.SMTPHost,
		SMTPPort: cfg.Environment.Email.SMTPPort,
	})

	// 4. Graceful Shutdown & Context Management
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// 5. Register Email Queue Consumer Handler
	emailHandler := func(ctx context.Context, d domain_broker.Delivery) error {
		var p payload.EmailPayload
		if err := json.Unmarshal(d.Message().Body, &p); err != nil {
			log.Printf("[EMAIL WORKER ERROR] invalid payload format: %v", err)
			return domain_broker.ErrReject
		}

		log.Printf("[EMAIL WORKER] delivering email to %v with subject %q", p.To, p.Subject)

		// Map domain EmailPayload into infrastructure mail.SendOptions
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rmqConsumer.Consume(
			ctx,
			domain_broker.EmailQueue,
			emailHandler,
			domain_broker.WithConcurrency(2),
			domain_broker.WithPrefetchCount(cfg.Constant.RabbitMQ.DefaultPrefetchCount),
		); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[EMAIL WORKER ERROR] consumer stopped: %v", err)
		}
	}()

	log.Print("Email worker running and listening on q.email.classic.")

	<-ctx.Done()
	log.Print("Shutdown signal received in email worker, waiting for in-flight tasks to complete...")

	shutdownWaitCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Print("All email worker consumers shut down cleanly.")
	case <-shutdownWaitCtx.Done():
		log.Print("Shutdown timed out waiting for email consumers.")
	}

	log.Print("Email worker process stopped.")
}
