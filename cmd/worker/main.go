package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	domain_broker "github.com/Doratou/Backend/internal/domain/broker"
	brokerInfra "github.com/Doratou/Backend/internal/infrastructure/broker/rabbitmq"
)

func main() {
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("worker bootstrap error: %v", err)
	}

	log.Print("Starting Doratou Standalone Consumer Worker...")

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
	log.Print("RabbitMQ connection established for workers")

	// 2. Initialize Broker Infrastructure
	rmqPublisher := brokerInfra.NewPublisher(rmqConnMgr)
	rmqConsumer := brokerInfra.NewConsumer(rmqConnMgr, rmqPublisher)
	topology := brokerInfra.NewTopology(rmqConnMgr)

	_ = topology

	// 3. Graceful Shutdown & Context Management
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// Sample Consumer Registration Skeleton:
	// In the next phase, register specific business queues and handlers here:
	//
	// retryPolicy := domain_broker.RetryPolicy{
	// 	Enabled:           true,
	// 	MaxRetries:        cfg.Constant.RabbitMQ.DefaultMaxRetries,
	// 	InitialDelay:      cfg.Constant.RabbitMQ.DefaultRetryDelay,
	// 	BackoffMultiplier: 2.0,
	// 	RetryExchange:     "doratou.retry.exchange",
	// }
	//
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	_ = rmqConsumer.Consume(ctx, "sample.queue", sampleHandler,
	// 		domain_broker.WithConcurrency(2),
	// 		domain_broker.WithPrefetchCount(cfg.Constant.RabbitMQ.DefaultPrefetchCount),
	// 		domain_broker.WithRetryPolicy(retryPolicy),
	// 	)
	// }()

	_ = rmqConsumer
	_ = domain_broker.DefaultRetryPolicy

	log.Print("Worker process initialized and ready for registered consumers.")

	<-ctx.Done()
	log.Print("Shutdown signal received in worker, waiting for in-flight tasks to complete...")

	shutdownWaitCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Print("All worker consumers shut down cleanly.")
	case <-shutdownWaitCtx.Done():
		log.Print("Shutdown timed out waiting for consumers.")
	}

	log.Print("Worker process stopped.")
}
