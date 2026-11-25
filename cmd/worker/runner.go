package runner

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	brokerInfra "github.com/Doratou/Backend/internal/infrastructure/broker/rabbitmq"
)

// WorkerFunc represents the consumer execution logic for an individual worker.
type WorkerFunc func(ctx context.Context, consumer *brokerInfra.Consumer, cfg *bootstrap.Config) error

// Run provides a standardized, DRY lifecycle manager for standalone workers.
// It initializes bootstrap configuration, connects to RabbitMQ, verifies topology,
// dispatches the worker function, and handles graceful shutdown upon OS signals.
func Run(workerName string, fn WorkerFunc) {
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("[%s WORKER] bootstrap error: %v", workerName, err)
	}

	log.Printf("[%s WORKER] Starting Doratou %s consumer worker...", workerName, workerName)

	// 1. Establish RabbitMQ Connection
	rmqEnv := cfg.Environment.RabbitMQ
	rmqCfg := brokerInfra.DefaultConfig(rmqEnv.Host, rmqEnv.Port, rmqEnv.User, rmqEnv.Password, rmqEnv.VHost)
	rmqCfg.ConnectionTimeout = cfg.Constant.RabbitMQ.DefaultPublishTimeout
	rmqCfg.InitialReconnectDelay = cfg.Constant.RabbitMQ.DefaultReconnectDelay

	connCtx, cancelConn := context.WithTimeout(context.Background(), 10*time.Second)
	rmqConnMgr, err := brokerInfra.NewConnectionManager(connCtx, rmqCfg)
	cancelConn()
	if err != nil {
		log.Fatalf("[%s WORKER] failed to connect to RabbitMQ broker: %v", workerName, err)
	}
	defer rmqConnMgr.Close()
	log.Printf("[%s WORKER] RabbitMQ connection established", workerName)

	// 2. Initialize Broker Infrastructure & Ensure Topology
	rmqPublisher := brokerInfra.NewPublisher(rmqConnMgr)
	rmqConsumer := brokerInfra.NewConsumer(rmqConnMgr, rmqPublisher)
	topology := brokerInfra.NewTopology(rmqConnMgr)

	setupCtx, cancelSetup := context.WithTimeout(context.Background(), 10*time.Second)
	if err := topology.SetupDoratouTopology(setupCtx); err != nil {
		log.Printf("[%s WORKER WARNING] failed to setup doratou topology: %v", workerName, err)
	} else {
		log.Printf("[%s WORKER] Doratou notification topology verified and ready", workerName)
	}
	cancelSetup()

	// 3. Graceful Shutdown & Context Management
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := fn(ctx, rmqConsumer, cfg); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[%s WORKER ERROR] consumer stopped with error: %v", workerName, err)
		}
	}()

	log.Printf("[%s WORKER] Running and awaiting messages.", workerName)

	<-ctx.Done()
	log.Printf("[%s WORKER] Shutdown signal received, waiting for in-flight tasks...", workerName)

	shutdownWaitCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Printf("[%s WORKER] Clean shutdown completed.", workerName)
	case <-shutdownWaitCtx.Done():
		log.Printf("[%s WORKER] Shutdown timed out waiting for in-flight tasks.", workerName)
	}

	log.Printf("[%s WORKER] Stopped.", workerName)
}
