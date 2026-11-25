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
	brokerInfra "github.com/Doratou/Backend/internal/infrastructure/broker/rabbitmq"
)

func main() {
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("ws worker bootstrap error: %v", err)
	}

	log.Print("Starting Doratou WebSocket Gateway Worker...")

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
	log.Print("RabbitMQ connection established for ws worker")

	// 2. Initialize Broker Infrastructure and Setup Topology
	rmqPublisher := brokerInfra.NewPublisher(rmqConnMgr)
	rmqConsumer := brokerInfra.NewConsumer(rmqConnMgr, rmqPublisher)
	topology := brokerInfra.NewTopology(rmqConnMgr)

	setupCtx, cancelSetup := context.WithTimeout(context.Background(), 10*time.Second)
	if err := topology.SetupDoratouTopology(setupCtx); err != nil {
		log.Printf("[WS WORKER WARNING] failed to declare doratou topology: %v", err)
	} else {
		log.Print("Doratou notification topology verified and ready")
	}
	cancelSetup()

	_ = rmqConsumer

	// 3. Graceful Shutdown & Context Management
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	log.Print("WS worker initialized and waiting for future WebSocket hub integration.")

	<-ctx.Done()
	log.Print("Shutdown signal received in ws worker...")

	shutdownWaitCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Print("All ws worker components shut down cleanly.")
	case <-shutdownWaitCtx.Done():
		log.Print("Shutdown timed out waiting for ws components.")
	}

	log.Print("WS worker process stopped.")
}
